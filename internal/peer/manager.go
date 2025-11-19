// CRC: crc-PeerManager.md, Spec: main.md
package peer

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/boxo/ipld/unixfs"
	uio "github.com/ipfs/boxo/ipld/unixfs/io"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	discoveryrouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/multiformats/go-multiaddr"
)

const (
	// P2PWebAppProtocol is the reserved libp2p protocol for file list queries
	P2PWebAppProtocol = "/p2p-webapp/1.0.0"
)

// FileEntry represents a file or directory entry with metadata
type FileEntry struct {
	Type     string `json:"type"`     // "file" or "directory"
	CID      string `json:"cid"`      // Content identifier
	MimeType string `json:"mimeType,omitempty"` // MIME type for files
}

// GetFileListMessage is sent to request a peer's file list
type GetFileListMessage struct {
	// Empty for now, can add fields if needed
}

// FileListMessage is the response containing a peer's file list
type FileListMessage struct {
	CID     string                `json:"cid"`     // Root directory CID
	Entries map[string]FileEntry `json:"entries"` // Full pathname tree
}

// Manager manages multiple peers
// CRC: crc-PeerManager.md
type Manager struct {
	ctx              context.Context
	mu               sync.RWMutex
	peers            map[string]*Peer
	onPeerData       func(receiverPeerID, senderPeerID, protocol string, data any)
	onTopicData      func(receiverPeerID, topic, senderPeerID string, data any)
	onPeerChange     func(receiverPeerID, topic, changedPeerID string, joined bool)
	onPeerFiles      func(receiverPeerID, targetPeerID, dirCID string, entries map[string]FileEntry)
	onGotFile        func(receiverPeerID string, cid string, success bool, content any)
	peerAliases      map[string]string // peerID -> alias
	aliasCounter     int
	verbosity        int
	ipfsPeer         *ipfslite.Peer // IPFS peer for file storage
	fileListHandlers map[string]func() // peerID -> handler for pending listFiles request (single handler per peer)
}

// Peer represents a single libp2p peer with its own host and state
// CRC: crc-PeerManager.md
type Peer struct {
	ctx              context.Context
	host             host.Host
	pubsub           *pubsub.PubSub
	dht              *dht.IpfsDHT
	mdnsService      mdns.Service
	peerID           peer.ID
	alias            string
	mu               sync.RWMutex
	connections      map[string]*Connection // key: "peerID:protocol" (legacy, kept for compatibility)
	protocols        map[protocol.ID]*ProtocolHandler
	topics           map[string]*TopicHandler
	monitoredTopics  map[string]*TopicMonitor // topics being monitored for join/leave events
	manager          *Manager
	vcm              *VirtualConnectionManager // Virtual connection manager for reliability
	directory        *uio.HAMTDirectory        // Peer's file directory (HAMTDirectory)
	directoryCID     cid.Cid                   // Current CID of the peer's directory
}

// TopicMonitor tracks peers in a topic and monitors join/leave events
type TopicMonitor struct {
	Topic       string
	ctx         context.Context
	cancel      context.CancelFunc
	knownPeers  map[string]bool // track which peers we've seen
}

// Connection represents an active peer-to-peer stream
type Connection struct {
	Stream   network.Stream
	PeerID   peer.ID
	Protocol protocol.ID
	mu       sync.Mutex
}

// ProtocolHandler handles incoming streams for a protocol
type ProtocolHandler struct {
	Protocol protocol.ID
}

// TopicHandler handles pub/sub topic subscriptions
type TopicHandler struct {
	Topic        string
	PubsubTopic  *pubsub.Topic
	Subscription *pubsub.Subscription
	ctx          context.Context
	cancel       context.CancelFunc
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	// Try to connect to the discovered peer (best effort)
	_ = n.h.Connect(context.Background(), pi)
}

// NewManager creates a new peer manager
// CRC: crc-PeerManager.md
// Sequence: seq-server-startup.md
func NewManager(ctx context.Context, bootstrapHost host.Host, ipfsPeer *ipfslite.Peer, verbosity int) (*Manager, error) {
	return &Manager{
		ctx:              ctx,
		peers:            make(map[string]*Peer),
		peerAliases:      make(map[string]string),
		verbosity:        verbosity,
		ipfsPeer:         ipfsPeer,
		peerFiles:        make(map[string]map[string]string),
		fileListHandlers: make(map[string][]func(entries map[string]string)),
	}, nil
}

// SetCallbacks sets the callback functions for events
func (m *Manager) SetCallbacks(
	onPeerData func(receiverPeerID, senderPeerID, protocol string, data any),
	onTopicData func(receiverPeerID, topic, senderPeerID string, data any),
	onPeerChange func(receiverPeerID, topic, changedPeerID string, joined bool),
) {
	m.onPeerData = onPeerData
	m.onTopicData = onTopicData
	m.onPeerChange = onPeerChange
}

// LogVerbose logs a message if the level is within the verbosity threshold
func (m *Manager) LogVerbose(peerID string, level int, format string, args ...any) {
	if level > m.verbosity {
		return
	}
	alias := m.getOrCreateAlias(peerID)
	message := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] %s\n", alias, message)
}

// getOrCreateAliasLocked returns the alias for a peer, creating one if it doesn't exist
// Caller must hold m.mu lock
func (m *Manager) getOrCreateAliasLocked(peerID string) string {
	if alias, exists := m.peerAliases[peerID]; exists {
		return alias
	}

	// Generate new alias (peer-a, peer-b, etc.)
	letter := rune('a' + m.aliasCounter)
	alias := fmt.Sprintf("peer-%c", letter)
	m.peerAliases[peerID] = alias
	m.aliasCounter++

	return alias
}

// getOrCreateAlias returns the alias for a peer, creating one if it doesn't exist
// This version acquires the lock
func (m *Manager) getOrCreateAlias(peerID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getOrCreateAliasLocked(peerID)
}

// allowPrivateGater is a ConnectionGater that allows all connections,
// including those on private/local addresses
type allowPrivateGater struct{}

// Ensure allowPrivateGater implements connmgr.ConnectionGater
var _ connmgr.ConnectionGater = (*allowPrivateGater)(nil)

func (g *allowPrivateGater) InterceptPeerDial(p peer.ID) (allow bool) {
	return true
}

func (g *allowPrivateGater) InterceptAddrDial(p peer.ID, m multiaddr.Multiaddr) (allow bool) {
	return true
}

func (g *allowPrivateGater) InterceptAccept(n network.ConnMultiaddrs) (allow bool) {
	return true
}

func (g *allowPrivateGater) InterceptSecured(dir network.Direction, p peer.ID, n network.ConnMultiaddrs) (allow bool) {
	return true
}

func (g *allowPrivateGater) InterceptUpgraded(c network.Conn) (allow bool, reason control.DisconnectReason) {
	return true, 0
}

// prepareCreatePeer generates/parses the private key and checks for duplicates
// Must be called with m.mu locked
func (m *Manager) prepareCreatePeer(requestedPeerKey string) (priv crypto.PrivKey, err error) {
	// Generate or parse peer identity
	if requestedPeerKey != "" {
		// Decode and unmarshal the private key
		keyBytes, err := crypto.ConfigDecodeKey(requestedPeerKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode peer key: %w", err)
		}
		priv, err = crypto.UnmarshalPrivateKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal peer key: %w", err)
		}
	} else {
		// Generate new identity
		priv, _, err = crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}
	}

	// Derive peer ID to check for duplicates
	pid, err := peer.IDFromPublicKey(priv.GetPublic())
	if err != nil {
		return nil, fmt.Errorf("failed to derive peer ID: %w", err)
	}

	// Check if peer ID is already registered (duplicate)
	if _, exists := m.peers[pid.String()]; exists {
		return nil, fmt.Errorf("peer ID already in use (possible duplicate browser tab)")
	}

	return priv, nil
}

// CreatePeer creates a new peer with its own libp2p host
// CRC: crc-PeerManager.md
// Sequence: seq-peer-creation.md
func (m *Manager) CreatePeer(requestedPeerKey string, rootDirectory string) (peerID string, peerKey string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Prepare and validate peer creation (checks for duplicates)
	priv, err := m.prepareCreatePeer(requestedPeerKey)
	if err != nil {
		return "", "", err
	}

	// Marshal and encode the private key for return
	keyBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal private key: %w", err)
	}
	encodedKey := crypto.ConfigEncodeKey(keyBytes)

	// Variable to store DHT reference
	var kdht *dht.IpfsDHT

	// Create libp2p host
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"), // Random port
		libp2p.ConnectionGater(&allowPrivateGater{}),   // Allow private/local addresses
		libp2p.EnableRelay(),                            // Enable relay for NAT traversal
		libp2p.EnableAutoRelayWithStaticRelays([]peer.AddrInfo{}), // Use public relays
		libp2p.NATPortMap(),                             // Try NAT port mapping
		libp2p.EnableNATService(),                       // Help other peers with NAT detection
		libp2p.EnableHolePunching(),                     // Enable hole punching for direct connections
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			// Create DHT for global discovery
			var err error
			kdht, err = dht.New(m.ctx, h, dht.Mode(dht.ModeAutoServer))
			return kdht, err
		}),
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to create host: %w", err)
	}

	// Bootstrap DHT with IPFS nodes for global discovery
	if kdht != nil {
		bootstrapPeers := dht.GetDefaultBootstrapPeerAddrInfos()
		connected := 0
		for _, peerinfo := range bootstrapPeers {
			if err := h.Connect(m.ctx, peerinfo); err == nil {
				connected++
			}
			// Stop after connecting to 3 bootstrap nodes (sufficient for DHT)
			if connected >= 3 {
				break
			}
		}
		if err := kdht.Bootstrap(m.ctx); err != nil {
			// Log but don't fail - DHT will continue trying to bootstrap
			if m.verbosity >= 1 {
				fmt.Printf("DHT bootstrap warning: %v\n", err)
			}
		}
	}

	// Setup mDNS for local discovery
	mdnsService := mdns.NewMdnsService(h, "p2p-webapp", &discoveryNotifee{h: h})
	if err := mdnsService.Start(); err != nil {
		if kdht != nil {
			kdht.Close()
		}
		h.Close()
		return "", "", fmt.Errorf("failed to start mDNS: %w", err)
	}

	// Create pubsub with DHT-based discovery for global peer finding
	var ps *pubsub.PubSub
	if kdht != nil {
		// Use DHT for topic-based peer discovery (enables global connectivity)
		routingDiscovery := discoveryrouting.NewRoutingDiscovery(kdht)
		ps, err = pubsub.NewGossipSub(m.ctx, h, pubsub.WithDiscovery(routingDiscovery))
	} else {
		// Fallback without discovery
		ps, err = pubsub.NewGossipSub(m.ctx, h)
	}
	if err != nil {
		mdnsService.Close()
		if kdht != nil {
			kdht.Close()
		}
		h.Close()
		return "", "", fmt.Errorf("failed to create pubsub: %w", err)
	}

	// Create peer
	p := &Peer{
		ctx:             m.ctx,
		host:            h,
		pubsub:          ps,
		dht:             kdht,
		mdnsService:     mdnsService,
		peerID:          h.ID(),
		connections:     make(map[string]*Connection),
		protocols:       make(map[protocol.ID]*ProtocolHandler),
		topics:          make(map[string]*TopicHandler),
		monitoredTopics: make(map[string]*TopicMonitor),
		manager:         m,
	}

	// Initialize virtual connection manager
	p.vcm = NewVirtualConnectionManager(m.ctx, p)

	// Connect to other peers in the same manager for local pubsub
	for _, otherPeer := range m.peers {
		// Try to connect peers to each other
		addrs := otherPeer.host.Addrs()
		if len(addrs) > 0 {
			peerInfo := peer.AddrInfo{
				ID:    otherPeer.peerID,
				Addrs: addrs,
			}
			// Best effort connection, ignore errors
			_ = h.Connect(m.ctx, peerInfo)
		}
	}

	// Set alias for the peer (we already hold the lock)
	p.alias = m.getOrCreateAliasLocked(p.peerID.String())

	m.peers[p.peerID.String()] = p

	// Initialize file storage for this peer
	// CRC: crc-PeerManager.md
	// Sequence: seq-store-file.md, seq-list-files.md
	if rootDirectory != "" {
		// Restore from existing directory CID
		dirCID, err := cid.Decode(rootDirectory)
		if err != nil {
			// Clean up peer on error
			delete(m.peers, p.peerID.String())
			p.Close()
			return "", "", fmt.Errorf("failed to parse root directory CID: %w", err)
		}

		// Load directory node from IPFS
		dirNode, err := m.ipfsPeer.GetNode(m.ctx, dirCID)
		if err != nil {
			delete(m.peers, p.peerID.String())
			p.Close()
			return "", "", fmt.Errorf("failed to load directory from IPFS: %w", err)
		}

		// Create HAMTDirectory from existing node
		dir, err := uio.NewHAMTDirectoryFromNode(m.ipfsPeer, dirNode)
		if err != nil {
			delete(m.peers, p.peerID.String())
			p.Close()
			return "", "", fmt.Errorf("failed to create directory from node: %w", err)
		}

		p.directory = dir
		p.directoryCID = dirCID

		// Pin the directory
		if err := m.ipfsPeer.AddPin(m.ctx, dirCID); err != nil {
			delete(m.peers, p.peerID.String())
			p.Close()
			return "", "", fmt.Errorf("failed to pin directory: %w", err)
		}
	} else {
		// Create new empty HAMTDirectory
		dir, err := uio.NewHAMTDirectory(m.ipfsPeer, 0)
		if err != nil {
			delete(m.peers, p.peerID.String())
			p.Close()
			return "", "", fmt.Errorf("failed to create directory: %w", err)
		}

		// Get the node and CID
		dirNode, err := dir.GetNode()
		if err != nil {
			delete(m.peers, p.peerID.String())
			p.Close()
			return "", "", fmt.Errorf("failed to get directory node: %w", err)
		}

		dirCID := dirNode.Cid()
		p.directory = dir
		p.directoryCID = dirCID

		// Pin the directory
		if err := m.ipfsPeer.AddPin(m.ctx, dirCID); err != nil {
			delete(m.peers, p.peerID.String())
			p.Close()
			return "", "", fmt.Errorf("failed to pin directory: %w", err)
		}
	}

	// Register protocol handler for file list queries
	h.SetStreamHandler(protocol.ID(P2PWebAppProtocol), p.handleP2PWebAppStream)

	// Log peer creation with connectivity info
	if m.verbosity >= 1 {
		fmt.Printf("[%s] Created peer\n", p.alias)
		if m.verbosity >= 2 {
			fmt.Printf("[%s] Listen addresses:\n", p.alias)
			for _, addr := range h.Addrs() {
				fmt.Printf("[%s]   %s\n", p.alias, addr)
			}
		}
	}

	return p.peerID.String(), encodedKey, nil
}

// getPeer retrieves a peer by ID
func (m *Manager) getPeer(peerID string) (*Peer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, exists := m.peers[peerID]
	if !exists {
		return nil, fmt.Errorf("peer not found: %s", peerID)
	}
	return p, nil
}

// Start starts a protocol handler for a peer
// CRC: crc-PeerManager.md
// Sequence: seq-protocol-communication.md
func (m *Manager) Start(peerID, protocolStr string) error {
	p, err := m.getPeer(peerID)
	if err != nil {
		return err
	}
	return p.Start(protocolStr)
}

// Stop removes a protocol handler for a peer
func (m *Manager) Stop(peerID, protocolStr string) error {
	p, err := m.getPeer(peerID)
	if err != nil {
		return err
	}
	return p.Stop(protocolStr)
}

// Send sends data to a peer on a protocol
// CRC: crc-PeerManager.md
// Sequence: seq-protocol-communication.md
func (m *Manager) Send(peerID, targetPeerID, protocolStr string, data any) error {
	p, err := m.getPeer(peerID)
	if err != nil {
		return err
	}
	return p.SendToPeer(targetPeerID, protocolStr, data)
}

// Subscribe subscribes a peer to a pub/sub topic
// CRC: crc-PeerManager.md
// Sequence: seq-pubsub-communication.md
func (m *Manager) Subscribe(peerID, topic string) error {
	p, err := m.getPeer(peerID)
	if err != nil {
		return err
	}
	return p.Subscribe(topic)
}

// Publish publishes data to a topic from a peer
// CRC: crc-PeerManager.md
// Sequence: seq-pubsub-communication.md
func (m *Manager) Publish(peerID, topic string, data any) error {
	p, err := m.getPeer(peerID)
	if err != nil {
		return err
	}
	return p.Publish(topic, data)
}

// Unsubscribe unsubscribes a peer from a topic
func (m *Manager) Unsubscribe(peerID, topic string) error {
	p, err := m.getPeer(peerID)
	if err != nil {
		return err
	}
	return p.Unsubscribe(topic)
}

// ListPeers returns the list of peers subscribed to a topic
func (m *Manager) ListPeers(peerID, topic string) ([]string, error) {
	p, err := m.getPeer(peerID)
	if err != nil {
		return nil, err
	}
	return p.ListPeers(topic)
}

// Monitor starts monitoring a topic for peer join/leave events
func (m *Manager) Monitor(peerID, topic string) error {
	p, err := m.getPeer(peerID)
	if err != nil {
		return err
	}
	return p.Monitor(topic)
}

// StopMonitor stops monitoring a topic for peer join/leave events
func (m *Manager) StopMonitor(peerID, topic string) error {
	p, err := m.getPeer(peerID)
	if err != nil {
		return err
	}
	return p.StopMonitor(topic)
}

// RemovePeer removes a peer and cleans up its resources
func (m *Manager) RemovePeer(peerID string) error {
	m.mu.Lock()
	p, exists := m.peers[peerID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("peer not found: %s", peerID)
	}
	delete(m.peers, peerID)
	delete(m.peerFiles, peerID) // Clean up file storage
	m.mu.Unlock()

	// Clean up peer resources
	return p.Close()
}

// File operations

// ListFiles requests a file list for a peer (async, uses onPeerFiles callback)
// CRC: crc-PeerManager.md
// Sequence: seq-list-files.md
func (m *Manager) ListFiles(targetPeerID string) error {
	m.mu.RLock()
	p, exists := m.peers[targetPeerID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("peer not found: %s", targetPeerID)
	}

	// Check if this is requesting local peer's files
	if p.peerID.String() == targetPeerID {
		// Build entries for local peer
		entries, err := p.buildFileEntries()
		if err != nil {
			return err
		}

		// Call callback asynchronously
		if m.onPeerFiles != nil {
			go m.onPeerFiles(targetPeerID, targetPeerID, p.directoryCID.String(), entries)
		}
		return nil
	}

	// For remote peer, check if handler already exists
	m.mu.Lock()
	_, handlerExists := m.fileListHandlers[targetPeerID]
	if handlerExists {
		// Already pending, just return
		m.mu.Unlock()
		return nil
	}

	// Register handler
	m.fileListHandlers[targetPeerID] = func() {
		// Handler will be called when response arrives
	}
	m.mu.Unlock()

	// Open stream to remote peer
	targetPeer, err := peer.Decode(targetPeerID)
	if err != nil {
		m.mu.Lock()
		delete(m.fileListHandlers, targetPeerID)
		m.mu.Unlock()
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	stream, err := p.host.NewStream(p.ctx, targetPeer, protocol.ID(P2PWebAppProtocol))
	if err != nil {
		// Remove handler on error
		m.mu.Lock()
		delete(m.fileListHandlers, targetPeerID)
		m.mu.Unlock()
		return fmt.Errorf("failed to open stream: %w", err)
	}

	// Send GetFileList message (type 0)
	if _, err := stream.Write([]byte{0}); err != nil {
		stream.Close()
		m.mu.Lock()
		delete(m.fileListHandlers, targetPeerID)
		m.mu.Unlock()
		return err
	}

	// Spawn goroutine to handle response
	go func() {
		defer stream.Close()
		p.handleFileList(stream)
	}()

	return nil
}

// GetFile retrieves file or directory content from IPFS (async, uses onGotFile callback)
// CRC: crc-PeerManager.md
func (m *Manager) GetFile(receiverPeerID string, cidStr string) error {
	if m.ipfsPeer == nil {
		return fmt.Errorf("IPFS peer not initialized")
	}

	c, err := cid.Decode(cidStr)
	if err != nil {
		return fmt.Errorf("invalid CID: %w", err)
	}

	// Spawn goroutine to retrieve content
	go func() {
		// Get node from IPFS
		node, err := m.ipfsPeer.GetNode(m.ctx, c)
		if err != nil {
			if m.onGotFile != nil {
				m.onGotFile(receiverPeerID, cidStr, false, map[string]any{"error": err.Error()})
			}
			return
		}

		// Check node type
		fsNode, err := unixfs.ExtractFSNode(node)
		if err != nil {
			if m.onGotFile != nil {
				m.onGotFile(receiverPeerID, cidStr, false, map[string]any{"error": err.Error()})
			}
			return
		}

		switch fsNode.Type() {
		case unixfs.TFile:
			// Read file content
			reader, err := uio.NewDagReader(m.ctx, node, m.ipfsPeer)
			if err != nil {
				if m.onGotFile != nil {
					m.onGotFile(receiverPeerID, cidStr, false, map[string]any{"error": err.Error()})
				}
				return
			}
			defer reader.Close()

			content, err := io.ReadAll(reader)
			if err != nil {
				if m.onGotFile != nil {
					m.onGotFile(receiverPeerID, cidStr, false, map[string]any{"error": err.Error()})
				}
				return
			}

			// Detect MIME type
			mimeType := http.DetectContentType(content)

			// Return file content
			if m.onGotFile != nil {
				m.onGotFile(receiverPeerID, cidStr, true, map[string]any{
					"type":     "file",
					"mimeType": mimeType,
					"content":  string(content), // Convert to string for JSON
				})
			}

		case unixfs.TDirectory, unixfs.THAMTShard:
			// Build directory entries
			dir, err := uio.NewHAMTDirectoryFromNode(m.ipfsPeer, node)
			if err != nil {
				if m.onGotFile != nil {
					m.onGotFile(receiverPeerID, cidStr, false, map[string]any{"error": err.Error()})
				}
				return
			}

			entries := make(map[string]string)
			links, err := dir.Links(m.ctx)
			if err != nil {
				if m.onGotFile != nil {
					m.onGotFile(receiverPeerID, cidStr, false, map[string]any{"error": err.Error()})
				}
				return
			}

			for _, link := range links {
				entries[link.Name] = link.Cid.String()
			}

			// Return directory content
			if m.onGotFile != nil {
				m.onGotFile(receiverPeerID, cidStr, true, map[string]any{
					"type":    "directory",
					"entries": entries,
				})
			}

		default:
			if m.onGotFile != nil {
				m.onGotFile(receiverPeerID, cidStr, false, map[string]any{"error": "unsupported file type"})
			}
		}
	}()

	return nil
}

// StoreFile stores file content in IPFS and adds it to the peer's file list
// CRC: crc-PeerManager.md
// Sequence: seq-store-file.md
func (m *Manager) StoreFile(peerID, path string, content []byte) (string, error) {
	if m.ipfsPeer == nil {
		return "", fmt.Errorf("IPFS peer not initialized")
	}

	// Add file to IPFS
	node, err := m.ipfsPeer.AddFile(m.ctx, bytes.NewReader(content), nil)
	if err != nil {
		return "", fmt.Errorf("failed to add file to IPFS: %w", err)
	}

	cidStr := node.Cid().String()

	// Update peer's file list
	m.mu.Lock()
	defer m.mu.Unlock()

	files, exists := m.peerFiles[peerID]
	if !exists {
		return "", fmt.Errorf("peer not found: %s", peerID)
	}

	files[path] = cidStr

	m.LogVerbose(peerID, 2, "Stored file: %s -> %s", path, cidStr)

	return cidStr, nil
}

// RemoveFile removes a file from the peer's file list
// CRC: crc-PeerManager.md
// Sequence: seq-store-file.md
func (m *Manager) RemoveFile(peerID, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	files, exists := m.peerFiles[peerID]
	if !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	if _, exists := files[path]; !exists {
		return fmt.Errorf("file not found: %s", path)
	}

	delete(files, path)

	m.LogVerbose(peerID, 2, "Removed file: %s", path)

	return nil
}

// Peer methods

// logVerbose logs a message from this peer if the level is within the verbosity threshold
func (p *Peer) logVerbose(level int, format string, args ...any) {
	p.manager.LogVerbose(p.peerID.String(), level, format, args...)
}

func (p *Peer) Start(protocolStr string) error {
	pid := protocol.ID(protocolStr)

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.protocols[pid]; exists {
		return fmt.Errorf("already started protocol: %s", protocolStr)
	}

	handler := &ProtocolHandler{
		Protocol: pid,
	}
	p.protocols[pid] = handler

	// Set stream handler
	p.host.SetStreamHandler(pid, func(s network.Stream) {
		p.handleIncomingStream(s)
	})

	return nil
}

func (p *Peer) Stop(protocolStr string) error {
	pid := protocol.ID(protocolStr)

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.protocols[pid]; !exists {
		return fmt.Errorf("protocol not started: %s", protocolStr)
	}

	delete(p.protocols, pid)
	p.host.RemoveStreamHandler(pid)

	return nil
}

// SendToPeer sends data to a peer on a protocol using the virtual connection manager
func (p *Peer) SendToPeer(targetPeerIDStr, protocolStr string, data any) error {
	// Use virtual connection manager for reliable delivery with queuing, retry, and ACK
	return p.vcm.SendToQueue(targetPeerIDStr, protocolStr, data)
}

func (p *Peer) Subscribe(topic string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If already subscribed, return success (idempotent)
	if _, exists := p.topics[topic]; exists {
		return nil
	}

	// Join topic
	t, err := p.pubsub.Join(topic)
	if err != nil {
		return fmt.Errorf("failed to join topic: %w", err)
	}

	// Subscribe
	sub, err := t.Subscribe()
	if err != nil {
		t.Close()
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	// Create handler
	ctx, cancel := context.WithCancel(p.ctx)
	handler := &TopicHandler{
		Topic:        topic,
		PubsubTopic:  t,
		Subscription: sub,
		ctx:          ctx,
		cancel:       cancel,
	}
	p.topics[topic] = handler

	// Start reading messages
	go p.readFromTopic(handler)

	return nil
}

func (p *Peer) Publish(topic string, data any) error {
	p.mu.RLock()
	handler, exists := p.topics[topic]
	p.mu.RUnlock()

	var t *pubsub.Topic
	var err error

	if exists {
		// Use existing topic
		t = handler.PubsubTopic
	} else {
		// Join new topic
		t, err = p.pubsub.Join(topic)
		if err != nil {
			return fmt.Errorf("failed to join topic: %w", err)
		}
		defer t.Close()
	}

	// Encode data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Publish
	if err := t.Publish(p.ctx, jsonData); err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	return nil
}

func (p *Peer) Unsubscribe(topic string) error {
	p.mu.Lock()
	handler, exists := p.topics[topic]
	if !exists {
		p.mu.Unlock()
		// If not subscribed, return success (idempotent)
		return nil
	}
	delete(p.topics, topic)
	p.mu.Unlock()

	handler.cancel()
	handler.Subscription.Cancel()
	handler.PubsubTopic.Close()

	return nil
}

func (p *Peer) ListPeers(topic string) ([]string, error) {
	p.mu.RLock()
	monitor, isMonitored := p.monitoredTopics[topic]
	p.mu.RUnlock()

	// If monitoring this topic, return the monitored peer list
	if isMonitored {
		peerStrs := make([]string, 0, len(monitor.knownPeers))
		for peerID := range monitor.knownPeers {
			peerStrs = append(peerStrs, peerID)
		}
		return peerStrs, nil
	}

	// Otherwise, query gossipsub directly
	peers := p.pubsub.ListPeers(topic)

	// Convert peer.ID slice to string slice
	peerStrs := make([]string, len(peers))
	for i, pid := range peers {
		peerStrs[i] = pid.String()
	}

	return peerStrs, nil
}

func (p *Peer) Monitor(topic string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If already monitoring, return success (idempotent)
	if _, exists := p.monitoredTopics[topic]; exists {
		return nil
	}

	// Create monitor
	ctx, cancel := context.WithCancel(p.ctx)
	monitor := &TopicMonitor{
		Topic:      topic,
		ctx:        ctx,
		cancel:     cancel,
		knownPeers: make(map[string]bool),
	}
	p.monitoredTopics[topic] = monitor

	// Start monitoring
	go p.monitorTopic(monitor)

	return nil
}

func (p *Peer) StopMonitor(topic string) error {
	p.mu.Lock()
	monitor, exists := p.monitoredTopics[topic]
	if !exists {
		p.mu.Unlock()
		// If not monitoring, return success (idempotent)
		return nil
	}
	delete(p.monitoredTopics, topic)
	p.mu.Unlock()

	monitor.cancel()

	return nil
}

func (p *Peer) Close() error {
	// Close virtual connection manager
	if p.vcm != nil {
		p.vcm.Close()
	}

	// Close all topics
	p.mu.Lock()
	for _, handler := range p.topics {
		handler.cancel()
		handler.Subscription.Cancel()
		handler.PubsubTopic.Close()
	}
	p.topics = make(map[string]*TopicHandler)

	// Close all monitors
	for _, monitor := range p.monitoredTopics {
		monitor.cancel()
	}
	p.monitoredTopics = make(map[string]*TopicMonitor)

	// Close all connections (legacy)
	for _, conn := range p.connections {
		conn.Stream.Close()
	}
	p.connections = make(map[string]*Connection)
	p.mu.Unlock()

	// Close mDNS discovery
	if p.mdnsService != nil {
		_ = p.mdnsService.Close()
	}

	// Close DHT
	if p.dht != nil {
		_ = p.dht.Close()
	}

	// Close host
	return p.host.Close()
}

// Internal methods

func (p *Peer) handleIncomingStream(s network.Stream) {
	// Route incoming stream to virtual connection manager for reliable handling
	p.vcm.HandleIncomingStream(s)
}

func (p *Peer) readFromStream(conn *Connection) {
	remotePeerID := conn.PeerID.String()
	protocolStr := string(conn.Protocol)
	connKey := fmt.Sprintf("%s:%s", remotePeerID, protocolStr)

	defer func() {
		p.mu.Lock()
		delete(p.connections, connKey)
		p.mu.Unlock()

		conn.Stream.Close()
	}()

	for {
		data, err := readMessage(conn.Stream)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading from stream: %v\n", err)
			}
			return
		}

		// Decode JSON
		var decoded any
		if err := json.Unmarshal(data, &decoded); err != nil {
			fmt.Printf("Error unmarshaling data: %v\n", err)
			continue
		}

		// Log received message
		remoteAlias := p.manager.getOrCreateAlias(remotePeerID)
		p.logVerbose(2, "Received message from %s on protocol %s", remoteAlias, protocolStr)

		if p.manager.onPeerData != nil {
			p.manager.onPeerData(p.peerID.String(), remotePeerID, protocolStr, decoded)
		}
	}
}

func (p *Peer) readFromTopic(handler *TopicHandler) {
	for {
		msg, err := handler.Subscription.Next(handler.ctx)
		if err != nil {
			if err != context.Canceled {
				fmt.Printf("Error reading from topic: %v\n", err)
			}
			return
		}

		// Decode JSON
		var decoded any
		if err := json.Unmarshal(msg.Data, &decoded); err != nil {
			fmt.Printf("Error unmarshaling topic data: %v\n", err)
			continue
		}

		if p.manager.onTopicData != nil {
			p.manager.onTopicData(p.peerID.String(), handler.Topic, msg.GetFrom().String(), decoded)
		}
	}
}

func (p *Peer) monitorTopic(monitor *TopicMonitor) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-monitor.ctx.Done():
			return
		case <-ticker.C:
			// Get current peers in topic
			currentPeers := p.pubsub.ListPeers(monitor.Topic)

			// Build current peer set
			currentPeerSet := make(map[string]bool)
			for _, pid := range currentPeers {
				peerIDStr := pid.String()
				currentPeerSet[peerIDStr] = true

				// Check if this is a new peer (joined)
				if !monitor.knownPeers[peerIDStr] {
					monitor.knownPeers[peerIDStr] = true
					if p.manager.onPeerChange != nil {
						p.manager.onPeerChange(p.peerID.String(), monitor.Topic, peerIDStr, true)
					}
				}
			}

			// Check for peers that left
			for peerIDStr := range monitor.knownPeers {
				if !currentPeerSet[peerIDStr] {
					delete(monitor.knownPeers, peerIDStr)
					if p.manager.onPeerChange != nil {
						p.manager.onPeerChange(p.peerID.String(), monitor.Topic, peerIDStr, false)
					}
				}
			}
		}
	}
}

// Message framing helpers

func writeMessage(w io.Writer, data []byte) error {
	// Write length as 4-byte big-endian
	length := uint32(len(data))
	lengthBytes := []byte{
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}

	if _, err := w.Write(lengthBytes); err != nil {
		return err
	}

	_, err := w.Write(data)
	return err
}

func readMessage(r io.Reader) ([]byte, error) {
	// Read length
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(r, lengthBytes); err != nil {
		return nil, err
	}

	length := uint32(lengthBytes[0])<<24 |
		uint32(lengthBytes[1])<<16 |
		uint32(lengthBytes[2])<<8 |
		uint32(lengthBytes[3])

	// Read message
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	return data, nil
}

// SetPeerFilesCallback sets the callback for peer file list notifications
func (m *Manager) SetPeerFilesCallback(cb func(receiverPeerID, targetPeerID, dirCID string, entries map[string]FileEntry)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onPeerFiles = cb
}

// SetGotFileCallback sets the callback for file retrieval notifications
func (m *Manager) SetGotFileCallback(cb func(receiverPeerID string, cid string, success bool, content any)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onGotFile = cb
}

// Bootstrap connects to a bootstrap peer (helper method)
func (m *Manager) Bootstrap(peerID, bootstrapAddr string) error {
	p, err := m.getPeer(peerID)
	if err != nil {
		return err
	}

	addr, err := multiaddr.NewMultiaddr(bootstrapAddr)
	if err != nil {
		return err
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return err
	}

	return p.host.Connect(p.ctx, *peerInfo)
}

// handleP2PWebAppStream handles incoming streams on the p2p-webapp protocol
// Sequence: seq-list-files.md
func (p *Peer) handleP2PWebAppStream(stream network.Stream) {
	defer stream.Close()

	// Read message type (first byte: 0 = GetFileList, 1 = FileList)
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		return
	}

	switch msgType[0] {
	case 0: // GetFileList
		p.handleGetFileList(stream)
	case 1: // FileList
		p.handleFileList(stream)
	}
}

// handleGetFileList processes a file list request and sends back the peer's file list
// Sequence: seq-list-files.md
func (p *Peer) handleGetFileList(stream network.Stream) {
	// Build file list from HAMTDirectory
	entries, err := p.buildFileEntries()
	if err != nil {
		return // Silent failure
	}

	// Create response message
	response := FileListMessage{
		CID:     p.directoryCID.String(),
		Entries: entries,
	}

	// Marshal to JSON
	data, err := json.Marshal(response)
	if err != nil {
		return
	}

	// Send message type (1 = FileList)
	if _, err := stream.Write([]byte{1}); err != nil {
		return
	}

	// Send JSON data
	_ = writeMessage(stream, data)
}

// buildFileEntries walks the HAMTDirectory tree and builds the entries map
func (p *Peer) buildFileEntries() (map[string]FileEntry, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.directory == nil {
		return make(map[string]FileEntry), nil
	}

	entries := make(map[string]FileEntry)

	// Walk directory tree
	err := p.walkDirectory(p.directory, "", entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// walkDirectory recursively walks a directory and populates entries
func (p *Peer) walkDirectory(dir *uio.HAMTDirectory, basePath string, entries map[string]FileEntry) error {
	// Get all links in this directory
	links, err := dir.Links(p.ctx)
	if err != nil {
		return err
	}

	for _, link := range links {
		fullPath := path.Join(basePath, link.Name)

		// Get the node to determine type
		node, err := p.manager.ipfsPeer.GetNode(p.ctx, link.Cid)
		if err != nil {
			continue // Skip if we can't get the node
		}

		// Check if it's a UnixFS node
		fsNode, err := unixfs.ExtractFSNode(node)
		if err != nil {
			continue
		}

		switch fsNode.Type() {
		case unixfs.TDirectory, unixfs.THAMTShard:
			// Directory
			entries[fullPath] = FileEntry{
				Type: "directory",
				CID:  link.Cid.String(),
			}

			// Recursively walk subdirectory
			subDir, err := uio.NewHAMTDirectoryFromNode(p.manager.ipfsPeer, node)
			if err != nil {
				continue
			}
			_ = p.walkDirectory(subDir, fullPath, entries)

		case unixfs.TFile:
			// File - detect MIME type
			mimeType := "application/octet-stream" // Default
			// Read first 512 bytes to detect MIME type
			fileReader, err := uio.NewDagReader(p.ctx, node, p.manager.ipfsPeer)
			if err == nil {
				buf := make([]byte, 512)
				n, _ := fileReader.Read(buf)
				if n > 0 {
					mimeType = http.DetectContentType(buf[:n])
				}
			}

			entries[fullPath] = FileEntry{
				Type:     "file",
				CID:      link.Cid.String(),
				MimeType: mimeType,
			}
		}
	}

	return nil
}

// handleFileList processes an incoming file list response
// Sequence: seq-list-files.md
func (p *Peer) handleFileList(stream network.Stream) {
	// Read JSON data
	data, err := readMessage(stream)
	if err != nil {
		return
	}

	// Parse FileListMessage
	var msg FileListMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	// Get sender peer ID from stream
	senderPeerID := stream.Conn().RemotePeer().String()

	// Look up handler
	p.manager.mu.Lock()
	handler, exists := p.manager.fileListHandlers[senderPeerID]
	if exists {
		delete(p.manager.fileListHandlers, senderPeerID)
	}
	onPeerFiles := p.manager.onPeerFiles
	p.manager.mu.Unlock()

	// Call callback if we have both
	if exists && handler != nil && onPeerFiles != nil {
		// Call the stored handler
		go handler()

		// Call onPeerFiles callback
		go onPeerFiles(p.peerID.String(), senderPeerID, msg.CID, msg.Entries)
	}
}
