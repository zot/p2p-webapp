// CRC: crc-Server.md, crc-WebServer.md, Spec: main.md
package server

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/zot/p2p-webapp/internal/config"
	"github.com/zot/p2p-webapp/internal/peer"
	"github.com/zot/p2p-webapp/internal/pidfile"
	"github.com/zot/p2p-webapp/internal/protocol"
)

// Server manages the HTTP server and WebSocket connections
// CRC: crc-Server.md
type Server struct {
	ctx            context.Context
	cancel         context.CancelFunc
	httpServer     *http.Server
	peerManager    *peer.Manager
	handler        *protocol.Handler
	config         *config.Config
	port           int
	fileSystem     http.FileSystem
	connections    map[*WSConnection]bool
	peerConnection map[string]*WSConnection // Maps peerID to WSConnection
	mu             sync.RWMutex
	linger         bool
	exitTimer      *time.Timer
	exitTimerMu    sync.Mutex
	verbosity      int
}

// zipFileSystem implements http.FileSystem for serving files from a ZIP archive
// CRC: crc-WebServer.md
type zipFileSystem struct {
	reader *zip.Reader
}

// Open implements http.FileSystem interface
// CRC: crc-WebServer.md
func (zfs *zipFileSystem) Open(name string) (http.File, error) {
	// Clean the path and remove leading slash
	// Use path.Clean (not filepath.Clean) to ensure forward slashes on all platforms
	// ZIP files always use forward slashes, even on Windows
	name = strings.TrimPrefix(path.Clean(name), "/")
	if name == "." {
		name = ""
	}

	// Build target path in ZIP (all files are under html/ prefix)
	// Use path.Join to ensure forward slashes
	targetPath := path.Join("html", name)

	// Find and return the file from ZIP
	for _, f := range zfs.reader.File {
		if f.Name == targetPath && !f.FileInfo().IsDir() {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}

			// Read entire file content
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}

			return &zipFile{
				name:    filepath.Base(targetPath),
				content: content,
				reader:  bytes.NewReader(content),
				info:    f.FileInfo(),
			}, nil
		}
	}

	return nil, os.ErrNotExist
}

// zipFile implements http.File interface
// CRC: crc-WebServer.md
type zipFile struct {
	name    string
	content []byte
	reader  *bytes.Reader
	info    os.FileInfo
}

// Read implements http.File interface
// CRC: crc-WebServer.md
func (zf *zipFile) Read(p []byte) (int, error) {
	return zf.reader.Read(p)
}

func (zf *zipFile) Seek(offset int64, whence int) (int64, error) {
	return zf.reader.Seek(offset, whence)
}

func (zf *zipFile) Close() error {
	return nil
}

func (zf *zipFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (zf *zipFile) Stat() (os.FileInfo, error) {
	return zf.info, nil
}

// NewServerFromDir creates a new HTTP server serving from a filesystem directory
// CRC: crc-Server.md
// Sequence: seq-server-startup.md
func NewServerFromDir(ctx context.Context, pm *peer.Manager, cfg *config.Config, htmlDir string) *Server {
	ctx, cancel := context.WithCancel(ctx)
	s := &Server{
		ctx:            ctx,
		cancel:         cancel,
		peerManager:    pm,
		config:         cfg,
		port:           cfg.Server.Port,
		fileSystem:     http.Dir(htmlDir),
		connections:    make(map[*WSConnection]bool),
		peerConnection: make(map[string]*WSConnection),
		linger:         cfg.Behavior.Linger,
		verbosity:      cfg.Behavior.Verbosity,
	}

	// Create protocol handler
	s.handler = protocol.NewHandler(pm)

	// Set handler ack callback
	s.handler.SetAckCallback(s.onSendAck)

	// Set peer manager callbacks to send messages to WebSocket clients
	pm.SetCallbacks(
		s.onPeerData,
		s.onTopicData,
		s.onPeerChange,
	)

	// Set file operation callbacks
	pm.SetPeerFilesCallback(s.onPeerFiles)
	pm.SetGotFileCallback(s.onGotFile)

	return s
}

// NewServerFromBundle creates a new HTTP server serving from a bundled ZIP archive
// CRC: crc-Server.md
// Sequence: seq-server-startup.md
func NewServerFromBundle(ctx context.Context, pm *peer.Manager, cfg *config.Config, bundleReader *zip.Reader) *Server {
	ctx, cancel := context.WithCancel(ctx)
	s := &Server{
		ctx:            ctx,
		cancel:         cancel,
		peerManager:    pm,
		config:         cfg,
		port:           cfg.Server.Port,
		fileSystem:     &zipFileSystem{reader: bundleReader},
		connections:    make(map[*WSConnection]bool),
		peerConnection: make(map[string]*WSConnection),
		linger:         cfg.Behavior.Linger,
		verbosity:      cfg.Behavior.Verbosity,
	}

	// Create protocol handler
	s.handler = protocol.NewHandler(pm)

	// Set handler ack callback
	s.handler.SetAckCallback(s.onSendAck)

	// Set peer manager callbacks to send messages to WebSocket clients
	pm.SetCallbacks(
		s.onPeerData,
		s.onTopicData,
		s.onPeerChange,
	)

	// Set file operation callbacks
	pm.SetPeerFilesCallback(s.onPeerFiles)
	pm.SetGotFileCallback(s.onGotFile)

	return s
}

// Start starts the HTTP server
// CRC: crc-Server.md
// Sequence: seq-server-startup.md
func (s *Server) Start() error {
	// Determine starting port from config
	startPort := s.config.Server.Port
	if startPort == 0 {
		startPort = 10000
	}

	// Create HTTP server mux
	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.handleWebSocket(w, r)
	})

	// Static file server with SPA routing fallback
	mux.Handle("/", s.spaHandler(s.fileSystem))

	// Try to find an available port
	var listener net.Listener
	var err error
	maxAttempts := s.config.Server.PortRange

	for attempt := 0; attempt < maxAttempts; attempt++ {
		port := startPort + attempt
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			s.port = port
			break
		}
	}

	if listener == nil {
		return fmt.Errorf("failed to find available port starting from %d", startPort)
	}

	s.httpServer = &http.Server{
		Handler:           mux,
		ReadTimeout:       s.config.Server.Timeouts.Read.Duration,
		WriteTimeout:      s.config.Server.Timeouts.Write.Duration,
		IdleTimeout:       s.config.Server.Timeouts.Idle.Duration,
		ReadHeaderTimeout: s.config.Server.Timeouts.ReadHeader.Duration,
		MaxHeaderBytes:    s.config.Server.MaxHeaderBytes,
	}

	// Start server in goroutine with the listener
	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Register this process in the PID tracking file
	if err := pidfile.Register(); err != nil {
		fmt.Printf("Warning: failed to register process: %v\n", err)
	}

	fmt.Printf("Server started on http://localhost:%d\n", s.port)
	return nil
}

// Stop stops the HTTP server and cleans up all resources
// CRC: crc-Server.md
func (s *Server) Stop() error {
	if s.verbosity >= 3 {
		fmt.Println("[DEBUG] Server.Stop() called")
	}

	// Cancel exit timer if running
	if s.verbosity >= 3 {
		fmt.Println("[DEBUG] Cancelling exit timer...")
	}
	s.cancelExitTimer()

	// Close all WebSocket connections
	// Copy connections to slice and release lock before closing to avoid deadlock
	// (conn.Close() calls UnregisterPeer which tries to acquire s.mu)
	if s.verbosity >= 3 {
		s.mu.RLock()
		connCount := len(s.connections)
		s.mu.RUnlock()
		fmt.Printf("[DEBUG] Closing %d WebSocket connections...\n", connCount)
	}
	s.mu.Lock()
	connsToClose := make([]*WSConnection, 0, len(s.connections))
	for conn := range s.connections {
		connsToClose = append(connsToClose, conn)
	}
	s.connections = make(map[*WSConnection]bool) // Clear the map
	s.peerConnection = make(map[string]*WSConnection) // Clear peer mappings
	s.mu.Unlock()

	// Now close connections without holding the lock
	for _, conn := range connsToClose {
		conn.Close()
	}
	if s.verbosity >= 3 {
		fmt.Println("[DEBUG] WebSocket connections closed")
	}

	// Stop HTTP server with a fresh context (don't use s.ctx which may be cancelled)
	var shutdownErr error
	if s.httpServer != nil {
		if s.verbosity >= 3 {
			fmt.Println("[DEBUG] Shutting down HTTP server...")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownErr = s.httpServer.Shutdown(ctx)
		if s.verbosity >= 3 {
			fmt.Println("[DEBUG] HTTP server shutdown complete")
		}
	}

	// Shutdown peer manager (close all peers and DHT)
	if s.peerManager != nil {
		if s.verbosity >= 3 {
			fmt.Println("[DEBUG] Shutting down peer manager...")
		}
		if err := s.peerManager.Shutdown(); err != nil {
			fmt.Printf("Warning: failed to shutdown peer manager: %v\n", err)
		}
		if s.verbosity >= 3 {
			fmt.Println("[DEBUG] Peer manager shutdown complete")
		}
	}

	// Unregister from PID tracking file AFTER everything else is shut down
	// This prevents race condition where ps shows no instances but process is still running
	if s.verbosity >= 3 {
		fmt.Println("[DEBUG] Unregistering PID...")
	}
	if err := pidfile.Unregister(); err != nil {
		fmt.Printf("Warning: failed to unregister process: %v\n", err)
	}
	if s.verbosity >= 3 {
		fmt.Println("[DEBUG] Server.Stop() complete")
	}

	return shutdownErr
}

// startExitTimer starts a countdown to exit when no connections remain
func (s *Server) startExitTimer() {
	s.exitTimerMu.Lock()
	defer s.exitTimerMu.Unlock()

	// Cancel any existing timer
	if s.exitTimer != nil {
		s.exitTimer.Stop()
	}

	timeout := s.config.Behavior.AutoExitTimeout.Duration
	fmt.Printf("Server closing in %v due to no active connections\n", timeout)

	s.exitTimer = time.AfterFunc(timeout, func() {
		fmt.Printf("Auto-exit: No connections for %v, shutting down...\n", timeout)
		s.cancel() // Cancel context to trigger server shutdown
	})
}

// cancelExitTimer cancels the exit countdown if a new connection arrives
func (s *Server) cancelExitTimer() {
	s.exitTimerMu.Lock()
	defer s.exitTimerMu.Unlock()

	if s.exitTimer != nil {
		s.exitTimer.Stop()
		s.exitTimer = nil
	}
}

// Port returns the port the server is listening on
func (s *Server) Port() int {
	return s.port
}

// Done returns a channel that is closed when the server context is cancelled
func (s *Server) Done() <-chan struct{} {
	return s.ctx.Done()
}

// RegisterPeer registers a peer with its WebSocket connection
func (s *Server) RegisterPeer(peerID string, conn *WSConnection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peerConnection[peerID] = conn
}

// UnregisterPeer removes a peer's connection mapping
func (s *Server) UnregisterPeer(peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.peerConnection, peerID)
}

// IsPeerRegistered checks if a peer ID is already registered
func (s *Server) IsPeerRegistered(peerID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.peerConnection[peerID]
	return exists
}

// OpenBrowser opens the default browser to the server URL
func (s *Server) OpenBrowser() error {
	url := fmt.Sprintf("http://localhost:%d", s.port)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// spaHandler wraps http.FileServer to provide SPA routing fallback
// For SPA routes (no file extension, file doesn't exist), serve index.html
// while preserving the URL path for client-side routing
// CRC: crc-WebServer.md
func (s *Server) spaHandler(fs http.FileSystem) http.Handler {
	fileServer := http.FileServer(fs)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Apply configured HTTP headers
		if s.config.HTTP.CacheControl != "" {
			w.Header().Set("Cache-Control", s.config.HTTP.CacheControl)
			// Add Pragma and Expires headers for compatibility if no-cache is set
			if strings.Contains(s.config.HTTP.CacheControl, "no-cache") {
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}
		}

		// Apply security headers
		if s.config.HTTP.Security.XContentTypeOptions != "" {
			w.Header().Set("X-Content-Type-Options", s.config.HTTP.Security.XContentTypeOptions)
		}
		if s.config.HTTP.Security.XFrameOptions != "" {
			w.Header().Set("X-Frame-Options", s.config.HTTP.Security.XFrameOptions)
		}
		if s.config.HTTP.Security.ContentSecurityPolicy != "" {
			w.Header().Set("Content-Security-Policy", s.config.HTTP.Security.ContentSecurityPolicy)
		}

		// Apply CORS headers if enabled
		if s.config.HTTP.CORS.Enabled {
			if s.config.HTTP.CORS.AllowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", s.config.HTTP.CORS.AllowOrigin)
			}
			if len(s.config.HTTP.CORS.AllowMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(s.config.HTTP.CORS.AllowMethods, ", "))
			}
			if len(s.config.HTTP.CORS.AllowHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(s.config.HTTP.CORS.AllowHeaders, ", "))
			}
		}

		path := r.URL.Path

		// Try to open the file
		f, err := fs.Open(path)
		if err == nil {
			// File exists, serve it normally
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// File doesn't exist - check if it looks like a SPA route
		ext := filepath.Ext(path)

		// If path has an extension, it's probably a real file request = real 404
		if ext != "" && ext != ".html" {
			http.NotFound(w, r)
			return
		}

		// No extension or .html extension - likely a SPA route
		// Serve index.html content directly (preserves URL for client-side routing)
		indexPath := "/index.html"
		indexFile, err := fs.Open(indexPath)
		if err != nil {
			// index.html doesn't exist, return 404
			http.NotFound(w, r)
			return
		}
		defer indexFile.Close()

		// Get file info for modification time
		indexInfo, err := indexFile.Stat()
		if err != nil {
			http.Error(w, "Failed to stat index.html", http.StatusInternalServerError)
			return
		}

		// Serve the content directly without changing the URL
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", indexInfo.ModTime(), indexFile.(io.ReadSeeker))
	})
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to upgrade connection: %v\n", err)
		return
	}

	wsConn := NewWSConnection(conn, s.handler, s.peerManager, s.peerManager, s)

	// Register connection
	s.mu.Lock()
	s.connections[wsConn] = true
	connectionCount := len(s.connections)
	s.mu.Unlock()

	// Cancel exit timer since we have a new connection
	if connectionCount > 0 {
		s.cancelExitTimer()
	}

	// Start connection
	wsConn.Start()

	// Cleanup on disconnect
	go func() {
		<-wsConn.closeCh
		s.mu.Lock()
		delete(s.connections, wsConn)
		connectionCount := len(s.connections)
		s.mu.Unlock()
		fmt.Printf("WebSocket connection closed\n")

		// Start exit timer if no connections remain and linger is disabled
		if connectionCount == 0 && !s.linger {
			s.startExitTimer()
		}
	}()

	fmt.Printf("New WebSocket connection established\n")
}

// Callback methods to send server messages to clients

func (s *Server) onPeerData(receiverPeerID, senderPeerID, protocol string, data any) {
	msg := s.handler.CreatePeerDataMessage(senderPeerID, protocol, data)

	// Send only to the connection that owns the receiving peer
	s.mu.RLock()
	conn, exists := s.peerConnection[receiverPeerID]
	s.mu.RUnlock()

	if exists {
		if err := conn.SendMessage(msg); err != nil {
			fmt.Printf("Failed to send peer message to peer %s: %v\n", receiverPeerID, err)
		}
	}
}

func (s *Server) onTopicData(receiverPeerID, topic, senderPeerID string, data any) {
	msg := s.handler.CreateTopicDataMessage(topic, senderPeerID, data)

	// Send only to the connection that owns the receiving peer
	s.mu.RLock()
	conn, exists := s.peerConnection[receiverPeerID]
	s.mu.RUnlock()

	if exists {
		if err := conn.SendMessage(msg); err != nil {
			fmt.Printf("Failed to send topic message to peer %s: %v\n", receiverPeerID, err)
		}
	}
}

func (s *Server) onPeerChange(receiverPeerID, topic, changedPeerID string, joined bool) {
	msg := s.handler.CreatePeerChangeMessage(topic, changedPeerID, joined)

	// Send only to the connection that owns the receiving peer
	s.mu.RLock()
	conn, exists := s.peerConnection[receiverPeerID]
	s.mu.RUnlock()

	if exists {
		if err := conn.SendMessage(msg); err != nil {
			action := "joined"
			if !joined {
				action = "left"
			}
			fmt.Printf("Failed to send %s message to peer %s: %v\n", action, receiverPeerID, err)
		}
	}
}

func (s *Server) onSendAck(peerID string, ack int) {
	msg := s.handler.CreateAckMessage(ack)

	// Send only to the connection that owns the sending peer
	s.mu.RLock()
	conn, exists := s.peerConnection[peerID]
	s.mu.RUnlock()

	if exists {
		if err := conn.SendMessage(msg); err != nil {
			fmt.Printf("Failed to send ack message to peer %s: %v\n", peerID, err)
		}
	}
}

func (s *Server) onPeerFiles(receiverPeerID, targetPeerID, dirCID string, entries map[string]any) {
	// Convert entries to FileEntryInfo format
	fileEntries := make(map[string]protocol.FileEntryInfo)
	for path, entry := range entries {
		if entryMap, ok := entry.(map[string]any); ok {
			fileEntry := protocol.FileEntryInfo{}
			if typ, ok := entryMap["type"].(string); ok {
				fileEntry.Type = typ
			}
			if cid, ok := entryMap["cid"].(string); ok {
				fileEntry.CID = cid
			}
			if mimeType, ok := entryMap["mimeType"].(string); ok {
				fileEntry.MimeType = mimeType
			}
			fileEntries[path] = fileEntry
		}
	}

	msg := s.handler.CreatePeerFilesMessage(targetPeerID, dirCID, fileEntries)

	// Send only to the connection that owns the receiving peer
	s.mu.RLock()
	conn, exists := s.peerConnection[receiverPeerID]
	s.mu.RUnlock()

	if exists {
		if err := conn.SendMessage(msg); err != nil {
			fmt.Printf("Failed to send peerFiles message to peer %s: %v\n", receiverPeerID, err)
		}
	}
}

func (s *Server) onGotFile(receiverPeerID string, cid string, success bool, content any) {
	msg := s.handler.CreateGotFileMessage(cid, success, content)

	// Send only to the connection that owns the receiving peer
	s.mu.RLock()
	conn, exists := s.peerConnection[receiverPeerID]
	s.mu.RUnlock()

	if exists {
		if err := conn.SendMessage(msg); err != nil {
			fmt.Printf("Failed to send gotFile message to peer %s: %v\n", receiverPeerID, err)
		}
	}
}

func (s *Server) broadcastMessage(msg *protocol.Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for conn := range s.connections {
		if err := conn.SendMessage(msg); err != nil {
			fmt.Printf("Failed to send message to client: %v\n", err)
		}
	}
}
