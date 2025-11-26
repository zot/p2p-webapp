// CRC: crc-PeerManager.md, Spec: main.md
package peer

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/boxo/ipld/unixfs"
	uio "github.com/ipfs/boxo/ipld/unixfs/io"
	blocks "github.com/ipfs/go-block-format"
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
	"github.com/libp2p/go-libp2p/core/peerstore"
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

// PeerOperations defines the interface for peer operations
type PeerOperations interface {
	// Protocol operations
	Start(protocol string) error
	Stop(protocol string) error
	SendToPeer(targetPeerIDStr, protocolStr string, data any) error

	// Topic operations
	Subscribe(topic string) error
	Publish(topic string, data any) error
	Unsubscribe(topic string) error
	ListPeers(topic string) ([]string, error)
	Monitor(topic string) error
	StopMonitor(topic string) error

	// File operations
	ListFiles(targetPeerID string) error
	GetFile(cidStr, fallbackPeerID string) error
	StoreFile(filepath string, content []byte, directory bool) (string, string, error)
	RemoveFile(filepath string) error
}

// FileEntry represents a file or directory entry with metadata
type FileEntry struct {
	Type     string `json:"type"`               // "file" or "directory"
	CID      string `json:"cid"`                // Content identifier
	MimeType string `json:"mimeType,omitempty"` // MIME type for files
}

// GetFileListMessage is sent to request a peer's file list
type GetFileListMessage struct {
	// Empty for now, can add fields if needed
}

// FileListMessage is the response containing a peer's file list
type FileListMessage struct {
	CID     string               `json:"cid"`     // Root directory CID
	Entries map[string]FileEntry `json:"entries"` // Full pathname tree
}

// Manager manages multiple peers
// CRC: crc-PeerManager.md
type Manager struct {
	ctx                   context.Context
	mu                    sync.RWMutex
	peers                 map[string]*Peer
	onPeerData            func(receiverPeerID, senderPeerID, protocol string, data any)
	onTopicData           func(receiverPeerID, topic, senderPeerID string, data any)
	onPeerChange          func(receiverPeerID, topic, changedPeerID string, joined bool)
	onPeerFiles           func(receiverPeerID, targetPeerID, dirCID string, entries map[string]any)
	onGotFile             func(receiverPeerID string, cid string, success bool, content any)
	peerAliases           map[string]string // peerID -> alias
	aliasCounter          int
	verbosity             int
	ipfsPeer              *ipfslite.Peer // IPFS peer for file storage
	fileUpdateNotifyTopic string         // Optional topic for file update notifications
	ipfsGetTimeout        time.Duration  // Timeout for IPFS Get operations
	streamTimeout         time.Duration  // Timeout for opening streams to peers
}

// Peer represents a single libp2p peer with its own host and state
// CRC: crc-PeerManager.md
type Peer struct {
	ctx             context.Context
	host            host.Host
	pubsub          *pubsub.PubSub
	dht             *dht.IpfsDHT
	dhtReady        chan struct{}             // Closed when DHT is bootstrapped and ready for operations
	dhtOperations   []func()                  // Queue of pending DHT operations (executed when DHT ready)
	dhtOpMu         sync.Mutex                // Protects dhtOperations queue
	mdnsService     mdns.Service
	peerID          peer.ID
	alias           string
	mu              sync.RWMutex
	protocols       map[protocol.ID]*ProtocolHandler
	topics          map[string]*TopicHandler
	monitoredTopics map[string]*TopicMonitor // topics being monitored for join/leave events
	manager         *Manager
	vcm             *VirtualConnectionManager // Virtual connection manager for reliability
	directory       *uio.HAMTDirectory        // Peer's file directory (HAMTDirectory)
	directoryCID    cid.Cid                   // Current CID of the peer's directory
	fileListHandler func()                    // Handler for pending listFiles request (single handler per peer)
	addedPeers      map[peer.ID]bool          // Track peers added via AddPeers (for retry attempts)
}

// TopicMonitor tracks peers in a topic and monitors join/leave events
type TopicMonitor struct {
	Topic      string
	ctx        context.Context
	cancel     context.CancelFunc
	knownPeers map[string]bool // track which peers we've seen
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
func NewManager(ctx context.Context, bootstrapHost host.Host, ipfsPeer *ipfslite.Peer, verbosity int, fileUpdateNotifyTopic string, ipfsGetTimeout time.Duration, streamTimeout time.Duration) (*Manager, error) {
	return &Manager{
		ctx:                   ctx,
		peers:                 make(map[string]*Peer),
		peerAliases:           make(map[string]string),
		verbosity:             verbosity,
		ipfsPeer:              ipfsPeer,
		fileUpdateNotifyTopic: fileUpdateNotifyTopic,
		ipfsGetTimeout:        ipfsGetTimeout,
		streamTimeout:         streamTimeout,
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

// GetPeer returns a peer by its ID
func (m *Manager) GetPeer(peerID string) (PeerOperations, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, exists := m.peers[peerID]
	if !exists {
		return nil, fmt.Errorf("peer not found: %s", peerID)
	}
	return p, nil
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
	// ============================================================
	// PHASE 1: Validate and get peer snapshot (minimal lock)
	// ============================================================
	m.mu.Lock()
	// Prepare and validate peer creation (checks for duplicates)
	priv, err := m.prepareCreatePeer(requestedPeerKey)
	if err != nil {
		m.mu.Unlock()
		return "", "", err
	}

	// Get snapshot of existing peers for later connection
	existingPeers := make([]*Peer, 0, len(m.peers))
	for _, p := range m.peers {
		existingPeers = append(existingPeers, p)
	}
	m.mu.Unlock()

	// ============================================================
	// PHASE 2: Do all network/IPFS I/O WITHOUT holding lock
	// ============================================================

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
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),            // Random port
		libp2p.ConnectionGater(&allowPrivateGater{}),              // Allow private/local addresses
		libp2p.EnableRelay(),                                      // Enable relay for NAT traversal
		libp2p.EnableAutoRelayWithStaticRelays([]peer.AddrInfo{}), // Use public relays
		libp2p.NATPortMap(),                                       // Try NAT port mapping
		libp2p.EnableNATService(),                                 // Help other peers with NAT detection
		libp2p.EnableHolePunching(),                               // Enable hole punching for direct connections
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

	// Note: DHT bootstrap is started later after peer creation (see below)

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
	// Configure GossipSub with faster heartbeat for quicker mesh formation
	var ps *pubsub.PubSub
	gossipSubParams := pubsub.DefaultGossipSubParams()
	// Use default mesh parameters (D=6, Dlo=5, Dhi=12, Dout=2) which are validated and work well
	// Only adjust heartbeat timing for faster mesh formation in local networks
	gossipSubParams.HeartbeatInitialDelay = 50 * time.Millisecond // Faster initial heartbeat for quick mesh formation
	gossipSubParams.HeartbeatInterval = 500 * time.Millisecond    // More frequent heartbeats (default: 1s)

	// Build direct peer list from existing peers in the same Manager
	// This guarantees localhost peers are always in each other's mesh
	directPeerInfos := make([]peer.AddrInfo, 0, len(existingPeers))
	for _, otherPeer := range existingPeers {
		directPeerInfos = append(directPeerInfos, peer.AddrInfo{
			ID:    otherPeer.peerID,
			Addrs: otherPeer.host.Addrs(),
		})
	}
	m.LogVerbose(h.ID().String(), 2, "Configuring GossipSub with %d direct peers", len(directPeerInfos))

	if kdht != nil {
		// Use DHT for topic-based peer discovery (enables global connectivity)
		routingDiscovery := discoveryrouting.NewRoutingDiscovery(kdht)
		ps, err = pubsub.NewGossipSub(
			m.ctx,
			h,
			pubsub.WithDiscovery(routingDiscovery),
			pubsub.WithPeerExchange(true), // Enable peer exchange
			pubsub.WithFloodPublish(true), // Flood publish for reliability in small networks
			pubsub.WithGossipSubParams(gossipSubParams),
			pubsub.WithDirectPeers(directPeerInfos), // Guarantee mesh inclusion for localhost peers
		)
	} else {
		// Fallback without discovery
		ps, err = pubsub.NewGossipSub(
			m.ctx,
			h,
			pubsub.WithPeerExchange(true), // Enable peer exchange
			pubsub.WithFloodPublish(true), // Flood publish for reliability in small networks
			pubsub.WithGossipSubParams(gossipSubParams),
			pubsub.WithDirectPeers(directPeerInfos), // Guarantee mesh inclusion for localhost peers
		)
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
		dhtReady:        make(chan struct{}), // Will be closed when DHT bootstrap completes
		dhtOperations:   make([]func(), 0),   // Queue for DHT operations
		mdnsService:     mdnsService,
		peerID:          h.ID(),
		protocols:       make(map[protocol.ID]*ProtocolHandler),
		topics:          make(map[string]*TopicHandler),
		monitoredTopics: make(map[string]*TopicMonitor),
		manager:         m,
		addedPeers:      make(map[peer.ID]bool),
	}

	// Initialize virtual connection manager
	p.vcm = NewVirtualConnectionManager(m.ctx, p)

	// Connect to other peers in the same manager for local pubsub (network I/O - no lock held!)
	for _, otherPeer := range existingPeers {
		// Try to connect peers to each other with exponential backoff retries
		addrs := otherPeer.host.Addrs()
		m.LogVerbose(h.ID().String(), 2, "Manual connect: new peer -> %s (addrs: %d)", otherPeer.peerID.String(), len(addrs))
		if len(addrs) > 0 {
			peerInfo := peer.AddrInfo{
				ID:    otherPeer.peerID,
				Addrs: addrs,
			}
			// Retry with exponential backoff for up to 15 seconds
			// Backoff sequence: 0ms, 100ms, 200ms, 400ms, 800ms, 1600ms, 3200ms, 6400ms, ...
			const maxRetryDuration = 15 * time.Second
			startTime := time.Now()
			attempt := 0
			connected := false
			backoff := 100 * time.Millisecond

			for time.Since(startTime) < maxRetryDuration {
				attempt++
				if attempt > 1 {
					m.LogVerbose(h.ID().String(), 2, "Manual connect retry attempt %d after %v (elapsed: %v)", attempt, backoff, time.Since(startTime))
					time.Sleep(backoff)
					backoff *= 2 // Double backoff for next attempt
					if backoff > 10*time.Second {
						backoff = 10 * time.Second // Cap at 10 seconds
					}
				}
				if err := h.Connect(m.ctx, peerInfo); err != nil {
					m.LogVerbose(h.ID().String(), 2, "Manual connect attempt %d FAILED: new peer -> %s: %v", attempt, otherPeer.peerID.String(), err)
				} else {
					m.LogVerbose(h.ID().String(), 2, "Manual connect SUCCESS on attempt %d (elapsed: %v): new peer -> %s", attempt, time.Since(startTime), otherPeer.peerID.String())
					connected = true
					break
				}
			}
			if !connected {
				m.LogVerbose(h.ID().String(), 1, "Manual connect failed after %v: new peer -> %s", time.Since(startTime), otherPeer.peerID.String())
			}
		}
	}

	// Make existing peers connect back to new peer for bidirectional discovery
	// This enables fast localhost peer discovery without waiting for mDNS/DHT
	go func() {
		newPeerAddrs := h.Addrs()
		newPeerID := h.ID()
		m.LogVerbose(newPeerID.String(), 2, "Bidirectional connect: starting reverse connections (addrs: %d)", len(newPeerAddrs))
		if len(newPeerAddrs) > 0 {
			newPeerInfo := peer.AddrInfo{
				ID:    newPeerID,
				Addrs: newPeerAddrs,
			}
			for _, otherPeer := range existingPeers {
				// Retry with exponential backoff for up to 15 seconds
				// Backoff sequence: 0ms, 100ms, 200ms, 400ms, 800ms, 1600ms, 3200ms, 6400ms, ...
				const maxRetryDuration = 15 * time.Second
				startTime := time.Now()
				attempt := 0
				connected := false
				backoff := 100 * time.Millisecond

				for time.Since(startTime) < maxRetryDuration {
					attempt++
					if attempt > 1 {
						m.LogVerbose(newPeerID.String(), 2, "Bidirectional connect retry attempt %d after %v (elapsed: %v)", attempt, backoff, time.Since(startTime))
						time.Sleep(backoff)
						backoff *= 2 // Double backoff for next attempt
						if backoff > 10*time.Second {
							backoff = 10 * time.Second // Cap at 10 seconds
						}
					}
					m.LogVerbose(newPeerID.String(), 2, "Bidirectional connect attempt %d: %s -> new peer", attempt, otherPeer.peerID.String())
					if err := otherPeer.host.Connect(m.ctx, newPeerInfo); err != nil {
						m.LogVerbose(newPeerID.String(), 2, "Bidirectional connect attempt %d FAILED: %s -> new peer: %v", attempt, otherPeer.peerID.String(), err)
					} else {
						m.LogVerbose(newPeerID.String(), 2, "Bidirectional connect SUCCESS on attempt %d (elapsed: %v): %s -> new peer", attempt, time.Since(startTime), otherPeer.peerID.String())
						connected = true
						break
					}
				}
				if !connected {
					m.LogVerbose(newPeerID.String(), 1, "Bidirectional connect failed after %v: %s -> new peer", time.Since(startTime), otherPeer.peerID.String())
				}
			}
		}
	}()

	// Initialize file storage for this peer (IPFS I/O - no lock held!)
	// CRC: crc-PeerManager.md
	// Sequence: seq-store-file.md, seq-list-files.md
	if rootDirectory != "" {
		// Restore from existing directory CID
		dirCID, err := cid.Decode(rootDirectory)
		if err != nil {
			// Clean up peer on error (peer not yet added to m.peers)
			p.Close()
			return "", "", fmt.Errorf("failed to parse root directory CID: %w", err)
		}

		// Load directory node from IPFS
		dirNode, err := m.ipfsPeer.Get(m.ctx, dirCID)
		if err != nil {
			p.Close()
			return "", "", fmt.Errorf("failed to load directory from IPFS: %w", err)
		}

		// Create HAMTDirectory from existing node
		dir, err := uio.NewHAMTDirectoryFromNode(m.ipfsPeer, dirNode)
		if err != nil {
			p.Close()
			return "", "", fmt.Errorf("failed to create directory from node: %w", err)
		}

		p.directory = dir
		p.directoryCID = dirCID
	} else {
		// Create new empty HAMTDirectory
		dir, err := uio.NewHAMTDirectory(m.ipfsPeer, 0)
		if err != nil {
			p.Close()
			return "", "", fmt.Errorf("failed to create directory: %w", err)
		}

		// Get the node and CID
		dirNode, err := dir.GetNode()
		if err != nil {
			p.Close()
			return "", "", fmt.Errorf("failed to get directory node: %w", err)
		}

		dirCID := dirNode.Cid()
		p.directory = dir
		p.directoryCID = dirCID
	}

	// Register protocol handler for file list queries
	h.SetStreamHandler(protocol.ID(P2PWebAppProtocol), p.handleP2PWebAppStream)

	// ============================================================
	// PHASE 3: Update Manager state (minimal lock)
	// ============================================================
	m.mu.Lock()
	p.alias = m.getOrCreateAliasLocked(p.peerID.String())
	m.peers[p.peerID.String()] = p
	m.mu.Unlock()

	// ============================================================
	// PHASE 4: Post-processing (no lock needed)
	// ============================================================

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

	// Start background retry goroutine for added peers that haven't connected yet
	go p.retryAddedPeersLoop()

	// Start DHT bootstrap goroutine (signals readiness and processes queued operations)
	if kdht != nil {
		go p.bootstrapDHT(kdht, h)
	} else {
		// No DHT - close dhtReady immediately so operations don't wait
		close(p.dhtReady)
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
	m.mu.Unlock()

	// Clean up peer resources
	return p.Close()
}

// AddPeers protects and tags peer connections to ensure they remain active
// CRC: crc-PeerManager.md
// Sequence: seq-add-peers.md
func (m *Manager) AddPeers(peerID string, targetPeerIDs []string) error {
	peer, err := m.getPeer(peerID)
	if err != nil {
		return err
	}
	return peer.AddPeers(targetPeerIDs)
}

// RemovePeers unprotects and untags peer connections
// CRC: crc-PeerManager.md
// Sequence: seq-remove-peers.md
func (m *Manager) RemovePeers(peerID string, targetPeerIDs []string) error {
	peer, err := m.getPeer(peerID)
	if err != nil {
		return err
	}
	return peer.RemovePeers(targetPeerIDs)
}

// Shutdown closes all peers
// CRC: crc-PeerManager.md
func (m *Manager) Shutdown() error {
	if m.verbosity >= 3 {
		fmt.Println("[DEBUG] Manager.Shutdown() called")
	}

	m.mu.Lock()
	peers := make([]*Peer, 0, len(m.peers))
	for _, p := range m.peers {
		peers = append(peers, p)
	}
	m.peers = make(map[string]*Peer) // Clear the map
	m.mu.Unlock()

	if m.verbosity >= 3 {
		fmt.Printf("[DEBUG] Closing %d peers...\n", len(peers))
	}

	// Close all peers (each peer closes its host, DHT, mDNS, etc.)
	for i, p := range peers {
		if m.verbosity >= 3 {
			fmt.Printf("[DEBUG] Closing peer %d/%d (%s)...\n", i+1, len(peers), p.alias)
		}
		if err := p.Close(); err != nil {
			fmt.Printf("Warning: failed to close peer: %v\n", err)
		}
		if m.verbosity >= 3 {
			fmt.Printf("[DEBUG] Peer %d/%d closed\n", i+1, len(peers))
		}
	}

	if m.verbosity >= 3 {
		fmt.Println("[DEBUG] Manager.Shutdown() complete")
	}

	// Note: IPFS peer is closed by the caller (ipfsNode.Close() in main.go)
	return nil
}

// File operations

// StoreFile stores file or directory in IPFS and adds it to the peer's HAMTDirectory
// CRC: crc-PeerManager.md
// Sequence: seq-store-file.md
func (m *Manager) StoreFile(peerID, filepath string, content []byte, directory bool) error {
	if m.ipfsPeer == nil {
		return fmt.Errorf("IPFS peer not initialized")
	}

	// Get peer
	p, err := m.getPeer(peerID)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Validate parameters
	if directory && content != nil {
		return fmt.Errorf("directory cannot have content")
	}
	if !directory && content == nil {
		return fmt.Errorf("file must have content")
	}

	var newNode ipld.Node

	// Create node based on type
	if directory {
		// Create empty HAMTDirectory
		dir, err := uio.NewHAMTDirectory(m.ipfsPeer, 0)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		newNode, err = dir.GetNode()
		if err != nil {
			return fmt.Errorf("failed to get directory node: %w", err)
		}
	} else {
		// Create file node
		newNode, err = m.ipfsPeer.AddFile(m.ctx, bytes.NewReader(content), nil)
		if err != nil {
			return fmt.Errorf("failed to add file to IPFS: %w", err)
		}
	}

	// Parse path to find parent directory and name
	parentPath, name := path.Split(filepath)
	if name == "" {
		return fmt.Errorf("invalid path: must include file/directory name")
	}

	// Clean parent path
	parentPath = strings.Trim(parentPath, "/")

	// Helper to add/update child in directory and get new directory
	updateDir := func(dir *uio.HAMTDirectory, childName string, childNode ipld.Node) (*uio.HAMTDirectory, error) {
		// Remove existing child if present (for updates)
		if err := dir.RemoveChild(m.ctx, childName); err != nil && err != os.ErrNotExist {
			// Ignore not exist errors, fail on other errors
			if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "no such file") {
				return nil, err
			}
		}

		// Add the new/updated child
		if err := dir.AddChild(m.ctx, childName, childNode); err != nil {
			return nil, err
		}

		// Return the same directory (it was modified in place)
		return dir, nil
	}

	// Navigate down the path, keeping track of directories for rebuild
	type dirLevel struct {
		dir  *uio.HAMTDirectory
		name string
	}
	dirStack := []dirLevel{{dir: p.directory, name: ""}}

	if parentPath != "" {
		pathParts := strings.Split(parentPath, "/")
		currentDir := p.directory

		for _, part := range pathParts {
			// Try to find existing subdirectory
			links, err := currentDir.Links(m.ctx)
			if err != nil {
				return fmt.Errorf("failed to read directory: %w", err)
			}

			found := false
			for _, link := range links {
				if link.Name == part {
					// Found subdirectory, navigate into it
					node, err := m.ipfsPeer.Get(m.ctx, link.Cid)
					if err != nil {
						return fmt.Errorf("failed to get subdirectory: %w", err)
					}
					currentDir, err = uio.NewHAMTDirectoryFromNode(m.ipfsPeer, node)
					if err != nil {
						return fmt.Errorf("failed to create directory from node: %w", err)
					}
					found = true
					break
				}
			}

			if !found {
				// Create new subdirectory
				currentDir, err = uio.NewHAMTDirectory(m.ipfsPeer, 0)
				if err != nil {
					return fmt.Errorf("failed to create subdirectory: %w", err)
				}
			}

			dirStack = append(dirStack, dirLevel{dir: currentDir, name: part})
		}
	}

	// Add the new file/directory to the leaf directory
	leafDir := dirStack[len(dirStack)-1].dir
	leafDir, err = updateDir(leafDir, name, newNode)
	if err != nil {
		return fmt.Errorf("failed to add child: %w", err)
	}

	// Rebuild the tree from leaf to root
	for i := len(dirStack) - 1; i > 0; i-- {
		childDir := dirStack[i].dir
		childName := dirStack[i].name
		parentDir := dirStack[i-1].dir

		// Get updated child node
		childNode, err := childDir.GetNode()
		if err != nil {
			return fmt.Errorf("failed to get child directory node: %w", err)
		}

		// Update parent to point to new child
		parentDir, err = updateDir(parentDir, childName, childNode)
		if err != nil {
			return fmt.Errorf("failed to update parent directory: %w", err)
		}

		dirStack[i-1].dir = parentDir
	}

	// Update peer's root directory
	p.directory = dirStack[0].dir
	rootNode, err := p.directory.GetNode()
	if err != nil {
		return fmt.Errorf("failed to get updated directory node: %w", err)
	}

	newRootCID := rootNode.Cid()
	p.directoryCID = newRootCID

	typeStr := "file"
	if directory {
		typeStr = "directory"
	}
	m.LogVerbose(peerID, 2, "Stored %s: %s -> %s", typeStr, filepath, newNode.Cid().String())

	return nil
}

// RemoveFile removes a file or directory from the peer's HAMTDirectory
// CRC: crc-PeerManager.md
// Sequence: seq-store-file.md
func (m *Manager) RemoveFile(peerID, filepath string) error {
	if m.ipfsPeer == nil {
		return fmt.Errorf("IPFS peer not initialized")
	}

	// Get peer
	p, err := m.getPeer(peerID)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Parse path to find parent directory and name
	parentPath, name := path.Split(filepath)
	if name == "" {
		return fmt.Errorf("invalid path: must include file/directory name")
	}

	// Clean parent path
	parentPath = strings.Trim(parentPath, "/")

	// Navigate to parent directory
	parentDir := p.directory
	if parentPath != "" {
		pathParts := strings.Split(parentPath, "/")
		for _, part := range pathParts {
			// Find subdirectory
			links, err := parentDir.Links(m.ctx)
			if err != nil {
				return fmt.Errorf("failed to read directory: %w", err)
			}

			found := false
			for _, link := range links {
				if link.Name == part {
					// Found subdirectory, navigate into it
					node, err := m.ipfsPeer.Get(m.ctx, link.Cid)
					if err != nil {
						return fmt.Errorf("failed to get subdirectory: %w", err)
					}
					parentDir, err = uio.NewHAMTDirectoryFromNode(m.ipfsPeer, node)
					if err != nil {
						return fmt.Errorf("failed to create directory from node: %w", err)
					}
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("parent directory not found: %s", part)
			}
		}
	}

	// Remove child from parent directory
	if err := parentDir.RemoveChild(m.ctx, name); err != nil {
		return fmt.Errorf("failed to remove child: %w", err)
	}

	// Get updated root directory node and CID
	rootNode, err := p.directory.GetNode()
	if err != nil {
		return fmt.Errorf("failed to get updated directory node: %w", err)
	}

	newRootCID := rootNode.Cid()

	// Update peer's directory CID
	p.directoryCID = newRootCID

	m.LogVerbose(peerID, 2, "Removed file/directory: %s", filepath)

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

	// Advertise topic to DHT for global discovery (enables geographically distant peers to find each other)
	if p.dht != nil {
		go p.advertiseTopic(topic, handler)
		go p.discoverTopicPeers(topic)
	}

	// Wait for gossip mesh to form before returning
	// This ensures peers can communicate immediately after Subscribe() returns
	p.waitForMeshFormation(t)

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

// waitForMeshFormation waits for the gossip mesh to form after subscribing to a topic
// This ensures peers can communicate immediately after Subscribe() returns
func (p *Peer) waitForMeshFormation(t *pubsub.Topic) {
	// Create timeout context - increased from 2s to 5s to handle slower mesh formation
	ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
	defer cancel()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Timeout reached - mesh may not have formed, but proceed anyway
			// This can happen if there are no other peers subscribed yet
			return
		case <-ticker.C:
			// Check if any peers are in the GossipSub mesh for this topic
			meshPeers := t.ListPeers()
			if len(meshPeers) > 0 {
				// Found peers in mesh - wait one heartbeat cycle for mesh to stabilize
				// Heartbeat interval is 500ms, so wait 600ms to be safe
				time.Sleep(600 * time.Millisecond)
				return
			}
		}
	}
}

// advertiseTopic advertises a topic subscription to the DHT for global peer discovery
// Runs continuously and re-advertises periodically as DHT advertisements expire
// Queues operation if DHT not ready yet
// Sequence: seq-pubsub-communication.md
func (p *Peer) advertiseTopic(topic string, handler *TopicHandler) {
	// Wrap the main logic so it can be queued
	p.enqueueDHTOperation(func() {
		routingDiscovery := discoveryrouting.NewRoutingDiscovery(p.dht)

		// Initial advertisement with retry
		var ttl time.Duration
		var err error
		for attempt := 1; attempt <= 3; attempt++ {
			ttl, err = routingDiscovery.Advertise(p.ctx, topic)
			if err == nil {
				break
			}
			p.logVerbose(1, "Failed to advertise topic %s to DHT (attempt %d/3): %v", topic, attempt, err)
			if attempt < 3 {
				select {
				case <-p.ctx.Done():
					return
				case <-handler.ctx.Done():
					return
				case <-time.After(time.Duration(attempt*2) * time.Second):
					// Exponential backoff: 2s, 4s
				}
			}
		}
		if err != nil {
			p.logVerbose(1, "Giving up advertising topic %s to DHT after 3 attempts", topic)
			return
		}
		p.logVerbose(2, "Advertised topic %s to DHT (TTL: %v)", topic, ttl)

		// Re-advertise periodically (before TTL expires)
		// Use half the TTL to ensure we don't miss the window
		ticker := time.NewTicker(ttl / 2)
		defer ticker.Stop()

		for {
			select {
			case <-p.ctx.Done():
				return
			case <-handler.ctx.Done():
				// Topic unsubscribed, stop advertising
				return
			case <-ticker.C:
				ttl, err = routingDiscovery.Advertise(p.ctx, topic)
				if err != nil {
					p.logVerbose(2, "Failed to re-advertise topic %s to DHT: %v", topic, err)
				} else {
					p.logVerbose(3, "Re-advertised topic %s to DHT (TTL: %v)", topic, ttl)
				}
			}
		}
	})
}

// discoverTopicPeers discovers and connects to peers subscribed to a topic via DHT
// Queues operation if DHT not ready yet
// Sequence: seq-pubsub-communication.md
func (p *Peer) discoverTopicPeers(topic string) {
	// Wrap the main logic so it can be queued
	p.enqueueDHTOperation(func() {
		routingDiscovery := discoveryrouting.NewRoutingDiscovery(p.dht)
		p.logVerbose(2, "Discovering peers for topic %s via DHT...", topic)

		// FindPeers queries the DHT for peers advertising this topic
		peerChan, err := routingDiscovery.FindPeers(p.ctx, topic)
		if err != nil {
			p.logVerbose(1, "Failed to start DHT peer discovery for topic %s: %v", topic, err)
			return
		}

		// Process discovered peers
		for peer := range peerChan {
			if peer.ID == p.peerID {
				continue // Skip self
			}

			p.logVerbose(2, "Discovered peer %s for topic %s via DHT (addrs: %d)", peer.ID.String(), topic, len(peer.Addrs))

			// Skip peers with no addresses - they can't be connected to
			if len(peer.Addrs) == 0 {
				continue
			}

			// Add addresses to peerstore with temporary TTL
			p.host.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.TempAddrTTL)

			// Attempt connection (best effort)
			if err := p.host.Connect(p.ctx, peer); err != nil {
				p.logVerbose(2, "Failed to connect to DHT-discovered peer %s: %v", peer.ID.String(), err)
			} else {
				p.logVerbose(1, "Connected to peer %s via DHT topic discovery", peer.ID.String())
			}
		}
	})
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

// AddPeers protects and tags peer connections using the libp2p BasicConnMgr
// For each peer ID: Protect, TagPeer (value 100), and attempt connection (best-effort)
// CRC: crc-Peer.md
// Sequence: seq-add-peers.md
func (p *Peer) AddPeers(targetPeerIDs []string) error {
	// libp2p host includes a BasicConnMgr (github.com/libp2p/go-libp2p/p2p/net/connmgr)
	// accessed via h.ConnManager() which provides Protect/TagPeer methods
	connMgr := p.host.ConnManager()

	for _, peerIDStr := range targetPeerIDs {
		// Parse peer ID
		targetPeerID, err := peer.Decode(peerIDStr)
		if err != nil {
			// Silently skip invalid peer IDs
			continue
		}

		// Add to tracking map for retry mechanism
		p.mu.Lock()
		p.addedPeers[targetPeerID] = true
		p.mu.Unlock()

		// Protect the connection to prevent the connection manager from closing it
		connMgr.Protect(targetPeerID, "connected")

		// Tag the peer with priority value 100
		connMgr.TagPeer(targetPeerID, "connected", 100)

		// Attempt to connect if not already connected (best-effort)
		// Get addresses from peerstore
		addrs := p.host.Peerstore().Addrs(targetPeerID)
		if len(addrs) > 0 {
			// Have addresses, attempt connection
			addrInfo := peer.AddrInfo{
				ID:    targetPeerID,
				Addrs: addrs,
			}
			// Connect with context (ignore errors per spec)
			_ = p.host.Connect(p.ctx, addrInfo)
		} else if p.dht != nil {
			// No addresses in peerstore - use DHT to actively find this peer
			// This is critical for geographically separated peers where mDNS won't work
			go func(pid peer.ID) {
				ctx, cancel := context.WithTimeout(p.ctx, 30*time.Second)
				defer cancel()

				p.logVerbose(2, "Looking up peer %s via DHT...", pid.String())
				addrInfo, err := p.dht.FindPeer(ctx, pid)
				if err != nil {
					// Peer not found via DHT - they may not be online or not connected to DHT
					p.logVerbose(1, "Could not find peer %s via DHT: %v", pid.String(), err)
					return
				}

				// Found peer - add their addresses to peerstore and connect
				p.logVerbose(2, "Found peer %s via DHT with %d addresses", pid.String(), len(addrInfo.Addrs))
				p.host.Peerstore().AddAddrs(addrInfo.ID, addrInfo.Addrs, peerstore.PermanentAddrTTL)
				_ = p.host.Connect(p.ctx, addrInfo)
			}(targetPeerID)
		}
	}

	return nil
}

// RemovePeers unprotects and untags peer connections using the libp2p BasicConnMgr
// For each peer ID: Unprotect and UntagPeer
// Does NOT disconnect the peers, only removes protection and priority
// CRC: crc-Peer.md
// Sequence: seq-remove-peers.md
func (p *Peer) RemovePeers(targetPeerIDs []string) error {
	// libp2p host includes a BasicConnMgr (github.com/libp2p/go-libp2p/p2p/net/connmgr)
	// accessed via h.ConnManager() which provides Unprotect/UntagPeer methods
	connMgr := p.host.ConnManager()

	for _, peerIDStr := range targetPeerIDs {
		// Parse peer ID
		targetPeerID, err := peer.Decode(peerIDStr)
		if err != nil {
			// Silently skip invalid peer IDs
			continue
		}

		// Remove from tracking map to stop retry attempts
		p.mu.Lock()
		delete(p.addedPeers, targetPeerID)
		p.mu.Unlock()

		// Unprotect the connection to allow normal connection management
		connMgr.Unprotect(targetPeerID, "connected")

		// Untag the peer to remove priority value
		connMgr.UntagPeer(targetPeerID, "connected")
	}

	return nil
}

// retryAddedPeersLoop periodically retries connecting to added peers that are disconnected
// Runs every 5 seconds and attempts DHT lookup + connection for disconnected added peers
func (p *Peer) retryAddedPeersLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.retryDisconnectedPeers()
		}
	}
}

// retryDisconnectedPeers checks which added peers are disconnected and retries connection
func (p *Peer) retryDisconnectedPeers() {
	p.mu.RLock()
	// Copy the added peers map to avoid holding lock during network operations
	peersToCheck := make([]peer.ID, 0, len(p.addedPeers))
	for peerID := range p.addedPeers {
		peersToCheck = append(peersToCheck, peerID)
	}
	p.mu.RUnlock()

	// Check each added peer's connection status
	for _, targetPeerID := range peersToCheck {
		// Check if we're connected to this peer
		connectedness := p.host.Network().Connectedness(targetPeerID)
		if connectedness == network.Connected {
			// Already connected, skip
			continue
		}

		// Not connected - try to reconnect
		p.logVerbose(2, "Retrying connection to added peer %s...", targetPeerID.String())

		// First try with existing addresses from peerstore
		addrs := p.host.Peerstore().Addrs(targetPeerID)
		if len(addrs) > 0 {
			addrInfo := peer.AddrInfo{
				ID:    targetPeerID,
				Addrs: addrs,
			}
			if err := p.host.Connect(p.ctx, addrInfo); err == nil {
				p.logVerbose(2, "Reconnected to added peer %s using peerstore addresses", targetPeerID.String())
				continue
			}
		}

		// If no addresses or connection failed, try DHT lookup
		if p.dht != nil {
			go func(pid peer.ID) {
				ctx, cancel := context.WithTimeout(p.ctx, 30*time.Second)
				defer cancel()

				addrInfo, err := p.dht.FindPeer(ctx, pid)
				if err != nil {
					p.logVerbose(2, "DHT retry: Could not find peer %s: %v", pid.String(), err)
					return
				}

				// Found peer - add addresses and connect
				p.logVerbose(2, "DHT retry: Found peer %s with %d addresses", pid.String(), len(addrInfo.Addrs))
				p.host.Peerstore().AddAddrs(addrInfo.ID, addrInfo.Addrs, peerstore.PermanentAddrTTL)
				if err := p.host.Connect(p.ctx, addrInfo); err == nil {
					p.logVerbose(1, "Successfully reconnected to added peer %s via DHT", pid.String())
				}
			}(targetPeerID)
		}
	}
}

// bootstrapDHT bootstraps the DHT and signals readiness when complete
// Queued DHT operations are executed once the DHT has peers in its routing table
// CRC: crc-Peer.md
// Spec: main.md
// Sequence: seq-dht-bootstrap.md
func (p *Peer) bootstrapDHT(kdht *dht.IpfsDHT, h host.Host) {
	// Connect to bootstrap peers
	bootstrapPeers := dht.GetDefaultBootstrapPeerAddrInfos()
	connected := 0
	for _, peerinfo := range bootstrapPeers {
		if err := h.Connect(p.ctx, peerinfo); err == nil {
			connected++
			p.logVerbose(3, "Connected to bootstrap peer %s", peerinfo.ID.String())
		}
		// Stop after connecting to 3 bootstrap nodes (sufficient for DHT)
		if connected >= 3 {
			break
		}
	}

	if connected == 0 {
		p.logVerbose(1, "Warning: Failed to connect to any bootstrap peers")
	} else {
		p.logVerbose(2, "Connected to %d bootstrap peers", connected)
	}

	// Bootstrap the DHT
	if err := kdht.Bootstrap(p.ctx); err != nil {
		p.logVerbose(1, "DHT bootstrap warning: %v", err)
	}

	// Wait for DHT to have peers in routing table (up to 30 seconds)
	// This ensures DHT operations (Advertise, FindPeers) will succeed
	p.logVerbose(2, "Waiting for DHT routing table to populate...")
	deadline := time.Now().Add(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			// Context cancelled, close dhtReady anyway
			close(p.dhtReady)
			return
		case <-ticker.C:
			// Check if DHT has peers
			if kdht.RoutingTable().Size() > 0 {
				p.logVerbose(1, "DHT ready with %d peers in routing table", kdht.RoutingTable().Size())
				// Signal readiness
				close(p.dhtReady)
				// Process queued operations
				p.processQueuedDHTOperations()
				return
			}
			// Check timeout
			if time.Now().After(deadline) {
				p.logVerbose(1, "DHT bootstrap timeout (no peers after 30s), proceeding anyway")
				// Close anyway so operations don't wait forever
				close(p.dhtReady)
				// Try to process queued operations anyway (they might fail but will be logged)
				p.processQueuedDHTOperations()
				return
			}
		}
	}
}

// processQueuedDHTOperations executes all queued DHT operations
// Must be called after dhtReady is closed
// CRC: crc-Peer.md
// Spec: main.md
// Sequence: seq-dht-bootstrap.md
func (p *Peer) processQueuedDHTOperations() {
	p.dhtOpMu.Lock()
	operations := p.dhtOperations
	p.dhtOperations = nil // Clear queue
	p.dhtOpMu.Unlock()

	if len(operations) > 0 {
		p.logVerbose(2, "Processing %d queued DHT operations", len(operations))
		for _, op := range operations {
			go op() // Spawn each queued operation in its own goroutine
		}
	}
}

// enqueueDHTOperation queues a DHT operation or spawns it immediately if DHT is ready
// Operations always run in their own goroutine to avoid blocking the queue
// This ensures DHT operations don't fail with "no peers in table" errors
// CRC: crc-Peer.md
// Spec: main.md
// Sequence: seq-dht-bootstrap.md
func (p *Peer) enqueueDHTOperation(op func()) {
	// Check if DHT is ready using non-blocking select
	select {
	case <-p.dhtReady:
		// DHT is ready, spawn goroutine immediately
		go op()
	default:
		// DHT not ready yet, queue the operation
		p.dhtOpMu.Lock()
		p.dhtOperations = append(p.dhtOperations, op)
		p.dhtOpMu.Unlock()
		p.logVerbose(2, "Queued DHT operation (DHT not ready yet)")
	}
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

// ListFiles requests a file list for a target peer (async, uses onPeerFiles callback)
// CRC: crc-Peer.md
// Sequence: seq-list-files.md
func (p *Peer) ListFiles(targetPeerID string) error {
	p.logVerbose(2, "ListFiles called for target=%s", targetPeerID)

	// Check if this is requesting own files
	if p.peerID.String() == targetPeerID {
		p.logVerbose(2, "Requesting own files")
		// Build entries for local peer
		entries, err := p.buildFileEntries()
		if err != nil {
			p.logVerbose(1, "Failed to build file entries: %v", err)
			return err
		}

		p.logVerbose(2, "Built %d file entries", len(entries))

		// Convert entries to map[string]any
		anyEntries := make(map[string]any)
		for path, entry := range entries {
			anyEntries[path] = map[string]any{
				"type":     entry.Type,
				"cid":      entry.CID,
				"mimeType": entry.MimeType,
			}
		}

		// Call callback asynchronously
		if p.manager.onPeerFiles != nil {
			p.logVerbose(2, "Calling onPeerFiles callback for own files")
			go p.manager.onPeerFiles(p.peerID.String(), targetPeerID, p.directoryCID.String(), anyEntries)
		}
		return nil
	}

	p.logVerbose(2, "Requesting remote peer files from %s", targetPeerID)

	// For remote peer, check if handler already exists
	p.mu.Lock()
	if p.fileListHandler != nil {
		// Already pending, just return
		p.mu.Unlock()
		p.logVerbose(2, "Request already pending for %s", targetPeerID)
		return nil
	}

	// Register handler
	p.fileListHandler = func() {
		// Handler will be called when response arrives
	}
	p.mu.Unlock()

	// Open stream to remote peer
	targetPeer, err := peer.Decode(targetPeerID)
	if err != nil {
		p.mu.Lock()
		p.fileListHandler = nil
		p.mu.Unlock()
		p.logVerbose(1, "Invalid peer ID %s: %v", targetPeerID, err)
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	p.logVerbose(2, "Opening stream to %s", targetPeerID)
	stream, err := p.host.NewStream(p.ctx, targetPeer, protocol.ID(P2PWebAppProtocol))
	if err != nil {
		// Remove handler on error
		p.mu.Lock()
		p.fileListHandler = nil
		p.mu.Unlock()
		p.logVerbose(1, "Failed to open stream to %s: %v", targetPeerID, err)
		return fmt.Errorf("failed to open stream: %w", err)
	}

	p.logVerbose(2, "Sending GetFileList message to %s", targetPeerID)
	// Send GetFileList message (type 0)
	if _, err := stream.Write([]byte{0}); err != nil {
		stream.Close()
		p.mu.Lock()
		p.fileListHandler = nil
		p.mu.Unlock()
		p.logVerbose(1, "Failed to send GetFileList: %v", err)
		return err
	}

	// Spawn goroutine to handle response
	go func() {
		defer stream.Close()
		p.logVerbose(2, "Waiting for file list response from %s", targetPeerID)
		p.handleFileList(stream)
	}()

	return nil
}

// GetFile retrieves file or directory content from IPFS (async, uses onGotFile callback)
// CRC: crc-Peer.md
// Spec: main.md
// CRC: crc-Peer.md
// Sequence: seq-get-file.md
func (p *Peer) GetFile(cidStr, fallbackPeerID string) error {
	if p.manager.ipfsPeer == nil {
		return fmt.Errorf("IPFS peer not initialized")
	}

	c, err := cid.Decode(cidStr)
	if err != nil {
		return fmt.Errorf("invalid CID: %w", err)
	}

	// Spawn goroutine to retrieve content
	go func() {
		// Get node from IPFS with configured timeout
		getCtx, cancel := context.WithTimeout(p.ctx, p.manager.ipfsGetTimeout)
		node, err := p.manager.ipfsPeer.Get(getCtx, c)
		cancel()
		if err != nil {
			// File not found locally - try fallback peer if provided
			if fallbackPeerID != "" {
				p.logVerbose(2, "File %s not found locally, trying fallback peer %s", cidStr, fallbackPeerID)
				if err := p.requestFileFromPeer(cidStr, fallbackPeerID); err != nil {
					p.logVerbose(1, "Failed to get file from fallback peer: %v", err)
					if p.manager.onGotFile != nil {
						p.manager.onGotFile(p.peerID.String(), cidStr, false, map[string]any{"error": err.Error()})
					}
				}
				// requestFileFromPeer will handle the callback when it receives the response
				return
			}
			// No fallback peer, return error
			if p.manager.onGotFile != nil {
				p.manager.onGotFile(p.peerID.String(), cidStr, false, map[string]any{"error": err.Error()})
			}
			return
		}

		// Check node type
		fsNode, err := unixfs.ExtractFSNode(node)
		if err != nil {
			if p.manager.onGotFile != nil {
				p.manager.onGotFile(p.peerID.String(), cidStr, false, map[string]any{"error": err.Error()})
			}
			return
		}

		switch fsNode.Type() {
		case unixfs.TFile:
			// Read file content
			reader, err := uio.NewDagReader(p.ctx, node, p.manager.ipfsPeer)
			if err != nil {
				if p.manager.onGotFile != nil {
					p.manager.onGotFile(p.peerID.String(), cidStr, false, map[string]any{"error": err.Error()})
				}
				return
			}
			defer reader.Close()

			content, err := io.ReadAll(reader)
			if err != nil {
				if p.manager.onGotFile != nil {
					p.manager.onGotFile(p.peerID.String(), cidStr, false, map[string]any{"error": err.Error()})
				}
				return
			}

			// Detect MIME type
			mimeType := http.DetectContentType(content)

			// Return file content (base64-encoded for safe JSON transmission)
			if p.manager.onGotFile != nil {
				p.manager.onGotFile(p.peerID.String(), cidStr, true, map[string]any{
					"type":     "file",
					"mimeType": mimeType,
					"content":  base64.StdEncoding.EncodeToString(content),
				})
			}

		case unixfs.TDirectory, unixfs.THAMTShard:
			// Build directory entries
			dir, err := uio.NewHAMTDirectoryFromNode(p.manager.ipfsPeer, node)
			if err != nil {
				if p.manager.onGotFile != nil {
					p.manager.onGotFile(p.peerID.String(), cidStr, false, map[string]any{"error": err.Error()})
				}
				return
			}

			entries := make(map[string]string)
			links, err := dir.Links(p.ctx)
			if err != nil {
				if p.manager.onGotFile != nil {
					p.manager.onGotFile(p.peerID.String(), cidStr, false, map[string]any{"error": err.Error()})
				}
				return
			}

			for _, link := range links {
				entries[link.Name] = link.Cid.String()
			}

			// Return directory content
			if p.manager.onGotFile != nil {
				p.manager.onGotFile(p.peerID.String(), cidStr, true, map[string]any{
					"type":    "directory",
					"entries": entries,
				})
			}

		default:
			if p.manager.onGotFile != nil {
				p.manager.onGotFile(p.peerID.String(), cidStr, false, map[string]any{"error": "unsupported file type"})
			}
		}
	}()

	return nil
}

// StoreFile stores file or directory in IPFS and adds it to the peer's HAMTDirectory
// CRC: crc-Peer.md
// Sequence: seq-store-file.md
func (p *Peer) StoreFile(filepath string, content []byte, directory bool) (string, string, error) {
	if p.manager.ipfsPeer == nil {
		return "", "", fmt.Errorf("IPFS peer not initialized")
	}

	// Validate parameters
	if directory && content != nil {
		return "", "", fmt.Errorf("directory cannot have content")
	}
	if !directory && content == nil {
		return "", "", fmt.Errorf("file must have content")
	}

	// Parse path to find parent directory and name
	parentPath, name := path.Split(filepath)
	if name == "" {
		return "", "", fmt.Errorf("invalid path: must include file/directory name")
	}
	parentPath = strings.Trim(parentPath, "/")

	// ============================================================
	// PHASE 1: Get current directory reference (minimal lock)
	// ============================================================
	p.mu.RLock()
	currentRootDir := p.directory
	p.mu.RUnlock()

	// ============================================================
	// PHASE 2: Do all IPFS work WITHOUT holding lock
	// ============================================================

	// Create the new file/directory node
	var newNode ipld.Node
	var err error

	if directory {
		// Create empty HAMTDirectory
		dir, err := uio.NewHAMTDirectory(p.manager.ipfsPeer, 0)
		if err != nil {
			return "", "", fmt.Errorf("failed to create directory: %w", err)
		}
		newNode, err = dir.GetNode()
		if err != nil {
			return "", "", fmt.Errorf("failed to get directory node: %w", err)
		}
	} else {
		// Create file node (IPFS network I/O - no lock held!)
		newNode, err = p.manager.ipfsPeer.AddFile(p.ctx, bytes.NewReader(content), nil)
		if err != nil {
			return "", "", fmt.Errorf("failed to add file to IPFS: %w", err)
		}
	}

	// Helper to add/update child in directory
	updateDir := func(dir *uio.HAMTDirectory, childName string, childNode ipld.Node) (*uio.HAMTDirectory, error) {
		// Remove existing child if present (for updates)
		if err := dir.RemoveChild(p.ctx, childName); err != nil && err != os.ErrNotExist {
			// Ignore not exist errors, fail on other errors
			if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "no such file") {
				return nil, err
			}
		}

		// Add the new/updated child
		if err := dir.AddChild(p.ctx, childName, childNode); err != nil {
			return nil, err
		}

		// Return the same directory (it was modified in place)
		return dir, nil
	}

	// Navigate down the path, keeping track of directories for rebuild
	type dirLevel struct {
		dir  *uio.HAMTDirectory
		name string
	}
	dirStack := []dirLevel{{dir: currentRootDir, name: ""}}

	if parentPath != "" {
		pathParts := strings.Split(parentPath, "/")
		currentDir := currentRootDir

		for _, part := range pathParts {
			// Try to find existing subdirectory (IPFS network I/O - no lock held!)
			links, err := currentDir.Links(p.ctx)
			if err != nil {
				return "", "", fmt.Errorf("failed to read directory: %w", err)
			}

			found := false
			for _, link := range links {
				if link.Name == part {
					// Found subdirectory, navigate into it (IPFS network I/O - no lock held!)
					node, err := p.manager.ipfsPeer.Get(p.ctx, link.Cid)
					if err != nil {
						return "", "", fmt.Errorf("failed to get subdirectory: %w", err)
					}
					currentDir, err = uio.NewHAMTDirectoryFromNode(p.manager.ipfsPeer, node)
					if err != nil {
						return "", "", fmt.Errorf("failed to create directory from node: %w", err)
					}
					found = true
					break
				}
			}

			if !found {
				// Create new subdirectory
				currentDir, err = uio.NewHAMTDirectory(p.manager.ipfsPeer, 0)
				if err != nil {
					return "", "", fmt.Errorf("failed to create subdirectory: %w", err)
				}
			}

			dirStack = append(dirStack, dirLevel{dir: currentDir, name: part})
		}
	}

	// Add the new file/directory to the leaf directory
	leafDir := dirStack[len(dirStack)-1].dir
	leafDir, err = updateDir(leafDir, name, newNode)
	if err != nil {
		return "", "", fmt.Errorf("failed to add child: %w", err)
	}

	// Rebuild the tree from leaf to root
	for i := len(dirStack) - 1; i > 0; i-- {
		childDir := dirStack[i].dir
		childName := dirStack[i].name
		parentDir := dirStack[i-1].dir

		// Get updated child node
		childNode, err := childDir.GetNode()
		if err != nil {
			return "", "", fmt.Errorf("failed to get child directory node: %w", err)
		}

		// Update parent to point to new child
		parentDir, err = updateDir(parentDir, childName, childNode)
		if err != nil {
			return "", "", fmt.Errorf("failed to update parent directory: %w", err)
		}

		dirStack[i-1].dir = parentDir
	}

	// Get the final root node
	newRootDir := dirStack[0].dir
	rootNode, err := newRootDir.GetNode()
	if err != nil {
		return "", "", fmt.Errorf("failed to get updated directory node: %w", err)
	}

	newRootCID := rootNode.Cid()
	resultCID := newNode.Cid().String()

	// ============================================================
	// PHASE 3: Update peer state (minimal lock)
	// ============================================================
	p.mu.Lock()
	p.directory = newRootDir
	p.directoryCID = newRootCID
	p.mu.Unlock()

	// Log the operation
	typeStr := "file"
	if directory {
		typeStr = "directory"
	}
	p.logVerbose(2, "Stored %s: %s -> %s", typeStr, filepath, resultCID)

	// ============================================================
	// PHASE 4: Publish notification (no lock needed)
	// ============================================================
	p.publishFileUpdateNotification()

	return resultCID, newRootCID.String(), nil
}

// requestFileFromPeer requests a file from a fallback peer using the reserved p2p-webapp protocol
// Spec: main.md
// CRC: crc-Peer.md
// Sequence: seq-get-file.md
func (p *Peer) requestFileFromPeer(cidStr, fallbackPeerID string) error {
	p.logVerbose(2, "Requesting file %s from peer %s", cidStr, fallbackPeerID)

	// Decode peer ID
	targetPeer, err := peer.Decode(fallbackPeerID)
	if err != nil {
		p.logVerbose(1, "Invalid fallback peer ID %s: %v", fallbackPeerID, err)
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	// Open stream to fallback peer with timeout
	p.logVerbose(2, "Opening stream to fallback peer %s", fallbackPeerID)
	streamCtx, streamCancel := context.WithTimeout(p.ctx, p.manager.streamTimeout)
	stream, err := p.host.NewStream(streamCtx, targetPeer, protocol.ID(P2PWebAppProtocol))
	streamCancel()
	if err != nil {
		p.logVerbose(1, "Failed to open stream to fallback peer %s: %v", fallbackPeerID, err)
		return fmt.Errorf("failed to open stream: %w", err)
	}

	p.logVerbose(2, "Sending GetFile message (type 2) for CID %s to %s", cidStr, fallbackPeerID)

	// Send message type (2 = GetFile)
	if _, err := stream.Write([]byte{2}); err != nil {
		stream.Close()
		p.logVerbose(1, "Failed to write GetFile message type: %v", err)
		return fmt.Errorf("failed to write message type: %w", err)
	}

	// Send CID as JSON
	msg := map[string]string{"cid": cidStr}
	data, err := json.Marshal(msg)
	if err != nil {
		stream.Close()
		return fmt.Errorf("failed to marshal GetFile message: %w", err)
	}

	if err := writeMessage(stream, data); err != nil {
		stream.Close()
		p.logVerbose(1, "Failed to write GetFile message data: %v", err)
		return fmt.Errorf("failed to write message: %w", err)
	}

	// Spawn goroutine to handle response
	go func() {
		defer stream.Close()
		p.logVerbose(2, "Waiting for file content response from %s", fallbackPeerID)
		p.handleFileContent(stream, cidStr)
	}()

	return nil
}

// RemoveFile removes a file or directory from the peer's HAMTDirectory
// CRC: crc-Peer.md
// Sequence: seq-store-file.md
func (p *Peer) RemoveFile(filepath string) error {
	if p.manager.ipfsPeer == nil {
		return fmt.Errorf("IPFS peer not initialized")
	}

	// Parse path to find parent directory and name
	parentPath, name := path.Split(filepath)
	if name == "" {
		return fmt.Errorf("invalid path: must include file/directory name")
	}
	parentPath = strings.Trim(parentPath, "/")

	// ============================================================
	// PHASE 1: Get current directory reference (minimal lock)
	// ============================================================
	p.mu.RLock()
	currentRootDir := p.directory
	p.mu.RUnlock()

	// ============================================================
	// PHASE 2: Navigate and remove WITHOUT holding lock
	// ============================================================

	// Parse path components
	var pathParts []string
	if parentPath != "" {
		pathParts = strings.Split(parentPath, "/")
	}

	// Navigate to parent directory and build stack of directories
	type dirStackEntry struct {
		dir  *uio.HAMTDirectory
		name string
	}
	dirStack := []dirStackEntry{{dir: currentRootDir, name: ""}}

	// Navigate down to parent directory (IPFS network I/O - no lock held!)
	currentDir := currentRootDir
	for _, part := range pathParts {
		// Get directory links
		links, err := currentDir.Links(p.ctx)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		// Find the subdirectory
		found := false
		for _, link := range links {
			if link.Name == part {
				// Get the subdirectory node (IPFS network I/O - no lock held!)
				node, err := p.manager.ipfsPeer.Get(p.ctx, link.Cid)
				if err != nil {
					return fmt.Errorf("failed to get subdirectory: %w", err)
				}

				// Create directory from node
				subDir, err := uio.NewHAMTDirectoryFromNode(p.manager.ipfsPeer, node)
				if err != nil {
					return fmt.Errorf("failed to create directory from node: %w", err)
				}

				dirStack = append(dirStack, dirStackEntry{dir: subDir, name: part})
				currentDir = subDir
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("parent directory not found: %s", part)
		}
	}

	// Current directory is now the parent directory where we need to remove the child
	parentDir := currentDir

	// Remove the child from parent directory
	if err := parentDir.RemoveChild(p.ctx, name); err != nil {
		return fmt.Errorf("failed to remove child: %w", err)
	}

	// Now rebuild the directory tree from bottom to top
	// Start from the parent of the removed item and work up to root
	for i := len(dirStack) - 1; i > 0; i-- {
		childDir := dirStack[i].dir
		childName := dirStack[i].name
		parentDirEntry := dirStack[i-1].dir

		// Get the updated child directory node
		childNode, err := childDir.GetNode()
		if err != nil {
			return fmt.Errorf("failed to get updated child directory node: %w", err)
		}

		// Remove old child and add updated child to parent
		if err := parentDirEntry.RemoveChild(p.ctx, childName); err != nil {
			// Ignore "not found" errors - child might not exist in parent yet
			if !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("failed to remove old child from parent: %w", err)
			}
		}

		if err := parentDirEntry.AddChild(p.ctx, childName, childNode); err != nil {
			return fmt.Errorf("failed to add updated child to parent: %w", err)
		}
	}

	// Get the final root node
	newRootDir := dirStack[0].dir
	rootNode, err := newRootDir.GetNode()
	if err != nil {
		return fmt.Errorf("failed to get updated directory node: %w", err)
	}

	newRootCID := rootNode.Cid()

	// ============================================================
	// PHASE 3: Update peer state (minimal lock)
	// ============================================================
	p.mu.Lock()
	p.directory = newRootDir
	p.directoryCID = newRootCID
	p.mu.Unlock()

	// Log the operation
	p.logVerbose(2, "Removed file/directory: %s", filepath)

	// ============================================================
	// PHASE 4: Publish notification (no lock needed)
	// ============================================================
	p.publishFileUpdateNotification()

	return nil
}

// Internal methods

// publishFileUpdateNotification publishes a file update notification if configured and subscribed
func (p *Peer) publishFileUpdateNotification() {
	// Check if notification topic is configured
	p.logVerbose(2, "publishFileUpdateNotification: fileUpdateNotifyTopic='%s'", p.manager.fileUpdateNotifyTopic)
	if p.manager.fileUpdateNotifyTopic == "" {
		p.logVerbose(2, "publishFileUpdateNotification: topic not configured, skipping")
		return
	}

	// Check if peer is subscribed to the notification topic
	_, subscribed := p.topics[p.manager.fileUpdateNotifyTopic]
	p.logVerbose(2, "publishFileUpdateNotification: subscribed to '%s'=%v", p.manager.fileUpdateNotifyTopic, subscribed)
	if !subscribed {
		p.logVerbose(2, "publishFileUpdateNotification: not subscribed to topic, skipping")
		return
	}

	// Publish notification message
	msg := map[string]string{
		"type": "p2p-webapp-file-update",
		"peer": p.peerID.String(),
	}

	p.logVerbose(2, "publishFileUpdateNotification: publishing notification to '%s'", p.manager.fileUpdateNotifyTopic)
	// Ignore publish errors (best effort notification)
	_ = p.Publish(p.manager.fileUpdateNotifyTopic, msg)
}

func (p *Peer) handleIncomingStream(s network.Stream) {
	// Route incoming stream to virtual connection manager for reliable handling
	p.vcm.HandleIncomingStream(s)
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
func (m *Manager) SetPeerFilesCallback(cb func(receiverPeerID, targetPeerID, dirCID string, entries map[string]any)) {
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

	// Read message type (first byte: 0 = GetFileList, 1 = FileList, 2 = GetFile, 3 = FileContent)
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		return
	}

	switch msgType[0] {
	case 0: // GetFileList
		p.handleGetFileList(stream)
	case 1: // FileList
		p.handleFileList(stream)
	case 2: // GetFile
		p.handleGetFile(stream)
	case 3: // FileContent
		// Type 3 is handled by handleFileContent which is called from requestFileFromPeer
		// This case should not be reached in normal flow
		p.logVerbose(1, "handleP2PWebAppStream: unexpected FileContent message (type 3)")
	}
}

// handleGetFileList processes a file list request and sends back the peer's file list
// Sequence: seq-list-files.md
func (p *Peer) handleGetFileList(stream network.Stream) {
	requesterPeerID := stream.Conn().RemotePeer().String()
	p.logVerbose(2, "handleGetFileList: received request from %s", requesterPeerID)

	// Build file list from HAMTDirectory
	entries, err := p.buildFileEntries()
	if err != nil {
		p.logVerbose(1, "handleGetFileList: failed to build file entries: %v", err)
		return
	}

	p.logVerbose(2, "handleGetFileList: built %d entries", len(entries))

	// Create response message
	response := FileListMessage{
		CID:     p.directoryCID.String(),
		Entries: entries,
	}

	// Marshal to JSON
	data, err := json.Marshal(response)
	if err != nil {
		p.logVerbose(1, "handleGetFileList: failed to marshal response: %v", err)
		return
	}

	p.logVerbose(2, "handleGetFileList: sending %d bytes to %s", len(data), requesterPeerID)

	// Send message type (1 = FileList)
	if _, err := stream.Write([]byte{1}); err != nil {
		p.logVerbose(1, "handleGetFileList: failed to write message type: %v", err)
		return
	}

	// Send JSON data
	if err := writeMessage(stream, data); err != nil {
		p.logVerbose(1, "handleGetFileList: failed to write message data: %v", err)
		return
	}

	p.logVerbose(2, "handleGetFileList: successfully sent file list to %s", requesterPeerID)
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
		node, err := p.manager.ipfsPeer.Get(p.ctx, link.Cid)
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
	p.logVerbose(2, "handleFileList called")

	// Read message type (should be 1 = FileList)
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		p.logVerbose(1, "Failed to read message type: %v", err)
		return
	}

	if msgType[0] != 1 {
		p.logVerbose(1, "Expected message type 1 (FileList), got %d", msgType[0])
		return
	}

	// Read JSON data
	data, err := readMessage(stream)
	if err != nil {
		p.logVerbose(1, "Failed to read message from stream: %v", err)
		return
	}

	p.logVerbose(2, "Read %d bytes from stream", len(data))

	// Parse FileListMessage
	var msg FileListMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		p.logVerbose(1, "Failed to unmarshal FileListMessage: %v", err)
		return
	}

	// Get sender peer ID from stream
	senderPeerID := stream.Conn().RemotePeer().String()
	p.logVerbose(2, "Received file list from %s with %d entries, CID=%s", senderPeerID, len(msg.Entries), msg.CID)

	// Look up handler for this peer
	p.mu.Lock()
	handler := p.fileListHandler
	p.fileListHandler = nil // Clear after use
	p.mu.Unlock()

	if handler == nil {
		p.logVerbose(1, "No handler registered for file list response from %s", senderPeerID)
	}

	onPeerFiles := p.manager.onPeerFiles
	if onPeerFiles == nil {
		p.logVerbose(1, "No onPeerFiles callback set")
	}

	// Call callback if we have both
	if handler != nil && onPeerFiles != nil {
		p.logVerbose(2, "Calling onPeerFiles callback: receiver=%s, target=%s", p.peerID.String(), senderPeerID)

		// Call the stored handler
		go handler()

		// Convert entries to map[string]any
		anyEntries := make(map[string]any)
		for path, entry := range msg.Entries {
			anyEntries[path] = map[string]any{
				"type":     entry.Type,
				"cid":      entry.CID,
				"mimeType": entry.MimeType,
			}
		}

		// Call onPeerFiles callback
		go onPeerFiles(p.peerID.String(), senderPeerID, msg.CID, anyEntries)
	}
}

// handleGetFile processes a file request from another peer and sends back the file content
// Spec: main.md
// CRC: crc-Peer.md
// Sequence: seq-get-file.md
func (p *Peer) handleGetFile(stream network.Stream) {
	requesterPeerID := stream.Conn().RemotePeer().String()
	p.logVerbose(2, "handleGetFile: received request from %s", requesterPeerID)

	// Read JSON data containing CID
	data, err := readMessage(stream)
	if err != nil {
		p.logVerbose(1, "handleGetFile: failed to read message: %v", err)
		return
	}

	// Parse GetFile message
	var msg struct {
		CID string `json:"cid"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		p.logVerbose(1, "handleGetFile: failed to unmarshal message: %v", err)
		return
	}

	p.logVerbose(2, "handleGetFile: requested CID=%s", msg.CID)

	// Decode CID
	c, err := cid.Decode(msg.CID)
	if err != nil {
		p.logVerbose(1, "handleGetFile: invalid CID %s: %v", msg.CID, err)
		// Send error response
		p.sendFileError(stream, msg.CID, fmt.Sprintf("invalid CID: %v", err))
		return
	}

	// Get node from IPFS with timeout
	getCtx, cancel := context.WithTimeout(p.ctx, p.manager.ipfsGetTimeout)
	node, err := p.manager.ipfsPeer.Get(getCtx, c)
	cancel()
	if err != nil {
		p.logVerbose(1, "handleGetFile: failed to get CID %s: %v", msg.CID, err)
		// Send error response
		p.sendFileError(stream, msg.CID, fmt.Sprintf("file not found: %v", err))
		return
	}

	// Check node type
	fsNode, err := unixfs.ExtractFSNode(node)
	if err != nil {
		p.logVerbose(1, "handleGetFile: failed to extract FS node: %v", err)
		p.sendFileError(stream, msg.CID, fmt.Sprintf("failed to extract node: %v", err))
		return
	}

	// Get raw node data for caching on the requesting peer
	rawData := node.RawData()

	switch fsNode.Type() {
	case unixfs.TFile:
		// Read file content for MIME type detection
		reader, err := uio.NewDagReader(p.ctx, node, p.manager.ipfsPeer)
		if err != nil {
			p.logVerbose(1, "handleGetFile: failed to create reader: %v", err)
			p.sendFileError(stream, msg.CID, fmt.Sprintf("failed to read file: %v", err))
			return
		}
		defer reader.Close()

		content, err := io.ReadAll(reader)
		if err != nil {
			p.logVerbose(1, "handleGetFile: failed to read content: %v", err)
			p.sendFileError(stream, msg.CID, fmt.Sprintf("failed to read content: %v", err))
			return
		}

		// Detect MIME type
		mimeType := http.DetectContentType(content)

		// Send file content response with raw node data for caching
		p.sendFileContent(stream, msg.CID, content, mimeType, rawData, false, nil)

	case unixfs.TDirectory, unixfs.THAMTShard:
		// Build directory entries for client display
		dir, err := uio.NewHAMTDirectoryFromNode(p.manager.ipfsPeer, node)
		if err != nil {
			p.logVerbose(1, "handleGetFile: failed to create directory: %v", err)
			p.sendFileError(stream, msg.CID, fmt.Sprintf("failed to read directory: %v", err))
			return
		}

		entries := make(map[string]string)
		links, err := dir.Links(p.ctx)
		if err != nil {
			p.logVerbose(1, "handleGetFile: failed to get links: %v", err)
			p.sendFileError(stream, msg.CID, fmt.Sprintf("failed to read directory: %v", err))
			return
		}

		for _, link := range links {
			entries[link.Name] = link.Cid.String()
		}

		// Send directory content response with raw node data for caching
		p.sendFileContent(stream, msg.CID, nil, "", rawData, true, entries)

	default:
		p.logVerbose(1, "handleGetFile: unsupported file type")
		p.sendFileError(stream, msg.CID, "unsupported file type")
	}
}

// sendFileContent sends a file content response (type 3) to the requesting peer
func (p *Peer) sendFileContent(stream network.Stream, cidStr string, content []byte, mimeType string, rawNodeData []byte, isDirectory bool, entries map[string]string) {
	p.logVerbose(2, "sendFileContent: sending response for CID=%s, isDirectory=%v, rawDataSize=%d", cidStr, isDirectory, len(rawNodeData))

	// Create response message
	response := map[string]any{
		"cid":         cidStr,
		"isDirectory": isDirectory,
		"rawNode":     base64.StdEncoding.EncodeToString(rawNodeData), // Raw IPFS node data for caching
	}

	if isDirectory {
		response["entries"] = entries
	} else {
		response["content"] = base64.StdEncoding.EncodeToString(content)
		response["mimeType"] = mimeType
	}

	// Marshal to JSON
	data, err := json.Marshal(response)
	if err != nil {
		p.logVerbose(1, "sendFileContent: failed to marshal response: %v", err)
		return
	}

	// Send message type (3 = FileContent)
	if _, err := stream.Write([]byte{3}); err != nil {
		p.logVerbose(1, "sendFileContent: failed to write message type: %v", err)
		return
	}

	// Send JSON data
	if err := writeMessage(stream, data); err != nil {
		p.logVerbose(1, "sendFileContent: failed to write message data: %v", err)
		return
	}

	p.logVerbose(2, "sendFileContent: successfully sent file content for %s", cidStr)
}

// sendFileError sends an error response (type 3 with error field) to the requesting peer
func (p *Peer) sendFileError(stream network.Stream, cidStr, errorMsg string) {
	p.logVerbose(2, "sendFileError: sending error for CID=%s: %s", cidStr, errorMsg)

	// Create error response
	response := map[string]string{
		"cid":   cidStr,
		"error": errorMsg,
	}

	// Marshal to JSON
	data, err := json.Marshal(response)
	if err != nil {
		p.logVerbose(1, "sendFileError: failed to marshal error response: %v", err)
		return
	}

	// Send message type (3 = FileContent/Error)
	if _, err := stream.Write([]byte{3}); err != nil {
		p.logVerbose(1, "sendFileError: failed to write message type: %v", err)
		return
	}

	// Send JSON data
	if err := writeMessage(stream, data); err != nil {
		p.logVerbose(1, "sendFileError: failed to write message data: %v", err)
		return
	}

	p.logVerbose(2, "sendFileError: successfully sent error response")
}

// handleFileContent processes a file content response from a fallback peer
// Spec: main.md
// CRC: crc-Peer.md
// Sequence: seq-get-file.md
func (p *Peer) handleFileContent(stream network.Stream, originalCID string) {
	p.logVerbose(2, "handleFileContent: waiting for response for CID=%s", originalCID)

	// Read message type (should be 3 = FileContent)
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		p.logVerbose(1, "handleFileContent: failed to read message type: %v", err)
		if p.manager.onGotFile != nil {
			p.manager.onGotFile(p.peerID.String(), originalCID, false, map[string]any{"error": fmt.Sprintf("failed to read response: %v", err)})
		}
		return
	}

	if msgType[0] != 3 {
		p.logVerbose(1, "handleFileContent: expected message type 3 (FileContent), got %d", msgType[0])
		if p.manager.onGotFile != nil {
			p.manager.onGotFile(p.peerID.String(), originalCID, false, map[string]any{"error": "invalid response type"})
		}
		return
	}

	// Read JSON data
	data, err := readMessage(stream)
	if err != nil {
		p.logVerbose(1, "handleFileContent: failed to read message: %v", err)
		if p.manager.onGotFile != nil {
			p.manager.onGotFile(p.peerID.String(), originalCID, false, map[string]any{"error": fmt.Sprintf("failed to read response: %v", err)})
		}
		return
	}

	// Parse response
	var response map[string]any
	if err := json.Unmarshal(data, &response); err != nil {
		p.logVerbose(1, "handleFileContent: failed to unmarshal response: %v", err)
		if p.manager.onGotFile != nil {
			p.manager.onGotFile(p.peerID.String(), originalCID, false, map[string]any{"error": fmt.Sprintf("invalid response: %v", err)})
		}
		return
	}

	// Check for error in response
	if errMsg, ok := response["error"].(string); ok {
		p.logVerbose(1, "handleFileContent: received error from fallback peer: %s", errMsg)
		if p.manager.onGotFile != nil {
			p.manager.onGotFile(p.peerID.String(), originalCID, false, map[string]any{"error": errMsg})
		}
		return
	}

	// Get and decode raw IPFS node data for caching
	rawNodeStr, ok := response["rawNode"].(string)
	if !ok {
		p.logVerbose(1, "handleFileContent: missing rawNode field in response")
		if p.manager.onGotFile != nil {
			p.manager.onGotFile(p.peerID.String(), originalCID, false, map[string]any{"error": "missing node data in response"})
		}
		return
	}

	rawNodeData, err := base64.StdEncoding.DecodeString(rawNodeStr)
	if err != nil {
		p.logVerbose(1, "handleFileContent: failed to decode rawNode: %v", err)
		if p.manager.onGotFile != nil {
			p.manager.onGotFile(p.peerID.String(), originalCID, false, map[string]any{"error": fmt.Sprintf("invalid node data: %v", err)})
		}
		return
	}

	// Parse the original CID
	c, err := cid.Decode(originalCID)
	if err != nil {
		p.logVerbose(1, "handleFileContent: invalid CID %s: %v", originalCID, err)
		if p.manager.onGotFile != nil {
			p.manager.onGotFile(p.peerID.String(), originalCID, false, map[string]any{"error": fmt.Sprintf("invalid CID: %v", err)})
		}
		return
	}

	// Create a block from raw data and CID, then add to blockstore
	// This caches the IPFS node so it can be served to other peers
	block, err := blocks.NewBlockWithCid(rawNodeData, c)
	if err != nil {
		p.logVerbose(1, "handleFileContent: failed to create block: %v", err)
		if p.manager.onGotFile != nil {
			p.manager.onGotFile(p.peerID.String(), originalCID, false, map[string]any{"error": fmt.Sprintf("failed to create block: %v", err)})
		}
		return
	}

	// Add block to local IPFS blockstore
	blockstore := p.manager.ipfsPeer.BlockStore()
	err = blockstore.Put(p.ctx, block)
	if err != nil {
		p.logVerbose(1, "handleFileContent: failed to add block to IPFS: %v", err)
		if p.manager.onGotFile != nil {
			p.manager.onGotFile(p.peerID.String(), originalCID, false, map[string]any{"error": fmt.Sprintf("failed to cache node: %v", err)})
		}
		return
	}

	p.logVerbose(2, "handleFileContent: cached node %s in local IPFS (%d bytes)", originalCID, len(rawNodeData))

	// Check if directory
	isDirectory, _ := response["isDirectory"].(bool)

	if isDirectory {
		// Directory content - forward to callback
		p.logVerbose(2, "handleFileContent: directory successfully cached")
		if p.manager.onGotFile != nil {
			p.manager.onGotFile(p.peerID.String(), originalCID, true, map[string]any{
				"type":    "directory",
				"entries": response["entries"],
			})
		}
	} else {
		// File content - forward to callback
		contentStr, ok := response["content"].(string)
		if !ok {
			p.logVerbose(1, "handleFileContent: missing content field for file")
			if p.manager.onGotFile != nil {
				p.manager.onGotFile(p.peerID.String(), originalCID, false, map[string]any{"error": "missing content in response"})
			}
			return
		}

		// Get MIME type from response
		mimeType, _ := response["mimeType"].(string)

		p.logVerbose(2, "handleFileContent: file successfully cached")
		// Forward to callback
		if p.manager.onGotFile != nil {
			p.manager.onGotFile(p.peerID.String(), originalCID, true, map[string]any{
				"type":     "file",
				"mimeType": mimeType,
				"content":  contentStr, // Send base64-encoded content to client
			})
		}
	}

	p.logVerbose(2, "handleFileContent: successfully retrieved and cached %s", originalCID)
}
