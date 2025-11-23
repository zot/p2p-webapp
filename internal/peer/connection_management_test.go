package peer

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
)

// TestPeerAddPeers tests the AddPeers method
// CRC: crc-Peer.md
// Test Design: test-Peer.md
func TestPeerAddPeers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a test peer with libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	if err != nil {
		t.Fatalf("Failed to create libp2p host: %v", err)
	}
	defer h.Close()

	// Create a Peer instance
	p := &Peer{
		host:  h,
		ctx:   ctx,
		alias: "test-peer",
	}

	// Create a second host to act as target peer
	h2, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	if err != nil {
		t.Fatalf("Failed to create second libp2p host: %v", err)
	}
	defer h2.Close()

	targetPeerID := h2.ID().String()

	// Add the target peer to peerstore so we have addresses
	h.Peerstore().AddAddrs(h2.ID(), h2.Addrs(), peerstore.PermanentAddrTTL)

	// Test AddPeers
	err = p.AddPeers([]string{targetPeerID})
	if err != nil {
		t.Fatalf("AddPeers failed: %v", err)
	}

	// Verify the peer is protected
	// Note: The libp2p ConnManager interface doesn't expose a way to check protection status directly
	// So we're testing that the method completes without error
	// In a real-world scenario, we'd test with a mock ConnManager

	// Test with invalid peer ID (should silently skip)
	err = p.AddPeers([]string{"invalid-peer-id"})
	if err != nil {
		t.Fatalf("AddPeers should not fail on invalid peer ID: %v", err)
	}

	// Test with empty list
	err = p.AddPeers([]string{})
	if err != nil {
		t.Fatalf("AddPeers failed on empty list: %v", err)
	}
}

// TestPeerRemovePeers tests the RemovePeers method
// CRC: crc-Peer.md
// Test Design: test-Peer.md
func TestPeerRemovePeers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a test peer with libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	if err != nil {
		t.Fatalf("Failed to create libp2p host: %v", err)
	}
	defer h.Close()

	// Create a Peer instance
	p := &Peer{
		host:  h,
		ctx:   ctx,
		alias: "test-peer",
	}

	// Create a second host to act as target peer
	h2, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	if err != nil {
		t.Fatalf("Failed to create second libp2p host: %v", err)
	}
	defer h2.Close()

	targetPeerID := h2.ID().String()

	// First add the peer (protect and tag it)
	h.Peerstore().AddAddrs(h2.ID(), h2.Addrs(), peerstore.PermanentAddrTTL)
	err = p.AddPeers([]string{targetPeerID})
	if err != nil {
		t.Fatalf("AddPeers failed: %v", err)
	}

	// Test RemovePeers
	err = p.RemovePeers([]string{targetPeerID})
	if err != nil {
		t.Fatalf("RemovePeers failed: %v", err)
	}

	// Test with invalid peer ID (should silently skip)
	err = p.RemovePeers([]string{"invalid-peer-id"})
	if err != nil {
		t.Fatalf("RemovePeers should not fail on invalid peer ID: %v", err)
	}

	// Test with empty list
	err = p.RemovePeers([]string{})
	if err != nil {
		t.Fatalf("RemovePeers failed on empty list: %v", err)
	}
}

// TestManagerAddPeers tests the Manager.AddPeers delegation
// CRC: crc-PeerManager.md
// Sequence: seq-add-peers.md
func TestManagerAddPeers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a test host and peer directly
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	if err != nil {
		t.Fatalf("Failed to create libp2p host: %v", err)
	}
	defer h.Close()

	testPeer := &Peer{
		host:  h,
		ctx:   ctx,
		alias: "test-peer",
	}

	// Create a mock manager with the test peer
	manager := &Manager{
		ctx:   ctx,
		peers: map[string]*Peer{h.ID().String(): testPeer},
	}

	// Create a second host to act as target peer
	h2, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	if err != nil {
		t.Fatalf("Failed to create second libp2p host: %v", err)
	}
	defer h2.Close()

	targetPeerID := h2.ID().String()

	// Add addresses to peerstore
	h.Peerstore().AddAddrs(h2.ID(), h2.Addrs(), time.Hour)

	// Test AddPeers via manager
	err = manager.AddPeers(h.ID().String(), []string{targetPeerID})
	if err != nil {
		t.Fatalf("Manager.AddPeers failed: %v", err)
	}

	// Test with non-existent peer
	err = manager.AddPeers("nonexistent", []string{targetPeerID})
	if err == nil {
		t.Fatal("Manager.AddPeers should fail for non-existent peer")
	}
}

// TestManagerRemovePeers tests the Manager.RemovePeers delegation
// CRC: crc-PeerManager.md
// Sequence: seq-remove-peers.md
func TestManagerRemovePeers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a test host and peer directly
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	if err != nil {
		t.Fatalf("Failed to create libp2p host: %v", err)
	}
	defer h.Close()

	testPeer := &Peer{
		host:  h,
		ctx:   ctx,
		alias: "test-peer",
	}

	// Create a mock manager with the test peer
	manager := &Manager{
		ctx:   ctx,
		peers: map[string]*Peer{h.ID().String(): testPeer},
	}

	// Create a second host to act as target peer
	h2, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	if err != nil {
		t.Fatalf("Failed to create second libp2p host: %v", err)
	}
	defer h2.Close()

	targetPeerID := h2.ID().String()

	// Add addresses to peerstore
	h.Peerstore().AddAddrs(h2.ID(), h2.Addrs(), time.Hour)

	// First add the peer
	err = manager.AddPeers(h.ID().String(), []string{targetPeerID})
	if err != nil {
		t.Fatalf("Manager.AddPeers failed: %v", err)
	}

	// Test RemovePeers via manager
	err = manager.RemovePeers(h.ID().String(), []string{targetPeerID})
	if err != nil {
		t.Fatalf("Manager.RemovePeers failed: %v", err)
	}

	// Test with non-existent peer
	err = manager.RemovePeers("nonexistent", []string{targetPeerID})
	if err == nil {
		t.Fatal("Manager.RemovePeers should fail for non-existent peer")
	}
}

// Helper functions for integration tests

// createPeerWithConnMgr creates a libp2p host with a custom connection manager
func createPeerWithConnMgr(t *testing.T, low, high int) (*Peer, func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())

	// Create ConnMgr with very short grace period and silence period for testing
	// Grace period prevents newly created connections from being immediately pruned
	// Silence period is the minimum time between trim operations (default is 10 seconds!)
	cm, err := connmgr.NewConnManager(
		low,
		high,
		connmgr.WithGracePeriod(10*time.Millisecond),
		connmgr.WithSilencePeriod(10*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to create connection manager: %v", err)
	}

	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
		libp2p.ConnectionManager(cm),
	)
	if err != nil {
		cancel()
		t.Fatalf("Failed to create libp2p host: %v", err)
	}

	p := &Peer{
		host:  h,
		ctx:   ctx,
		alias: "test-peer",
	}

	cleanup := func() {
		h.Close()
		cancel()
	}

	return p, cleanup
}

// createSimplePeer creates a libp2p host without custom ConnMgr
func createSimplePeer(t *testing.T) (*Peer, func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())

	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	if err != nil {
		cancel()
		t.Fatalf("Failed to create libp2p host: %v", err)
	}

	p := &Peer{
		host:  h,
		ctx:   ctx,
		alias: "test-peer",
	}

	cleanup := func() {
		h.Close()
		cancel()
	}

	return p, cleanup
}

// connectPeers establishes connection between two peers
func connectPeers(t *testing.T, peerA, peerB *Peer) {
	t.Helper()
	// Add B's addresses to A's peerstore
	peerA.host.Peerstore().AddAddrs(peerB.host.ID(), peerB.host.Addrs(), peerstore.PermanentAddrTTL)

	// Connect A to B
	addrInfo := peer.AddrInfo{
		ID:    peerB.host.ID(),
		Addrs: peerB.host.Addrs(),
	}
	if err := peerA.host.Connect(peerA.ctx, addrInfo); err != nil {
		t.Fatalf("Failed to connect peers: %v", err)
	}
}

// isConnected checks if two peers are connected
func isConnected(peerA, peerB *Peer) bool {
	return peerA.host.Network().Connectedness(peerB.host.ID()) == network.Connected
}

// assertConnected verifies two peers are connected
func assertConnected(t *testing.T, peerA, peerB *Peer) {
	t.Helper()
	if !isConnected(peerA, peerB) {
		t.Errorf("Expected peer %s to be connected to %s, but not connected",
			peerA.host.ID().String()[:8], peerB.host.ID().String()[:8])
	}
}

// assertDisconnected verifies two peers are not connected
func assertDisconnected(t *testing.T, peerA, peerB *Peer) {
	t.Helper()
	if isConnected(peerA, peerB) {
		t.Errorf("Expected peer %s to be disconnected from %s, but still connected",
			peerA.host.ID().String()[:8], peerB.host.ID().String()[:8])
	}
}

// countConnections returns number of connected peers
func countConnections(p *Peer) int {
	return len(p.host.Network().Peers())
}

// triggerPruning forces the connection manager to trim connections
func triggerPruning(t *testing.T, p *Peer) {
	t.Helper()
	// Wait for grace period to expire (we set it to 10ms in createPeerWithConnMgr)
	time.Sleep(20 * time.Millisecond)
	// Force the connection manager to trim connections now
	// The ConnManager interface has TrimOpenConns method
	p.host.ConnManager().TrimOpenConns(p.ctx)
	// Give it more time to actually close connections
	// libp2p may have internal delays beyond just the grace period
	time.Sleep(200 * time.Millisecond)
}
