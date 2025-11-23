package peer

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
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
	h.Peerstore().AddAddrs(h2.ID(), h2.Addrs(), peer.PermanentAddrTTL)

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
	h.Peerstore().AddAddrs(h2.ID(), h2.Addrs(), peer.PermanentAddrTTL)
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

// TestManagerAddPeers tests the Manager.AddPeers method
// CRC: crc-PeerManager.md
// Sequence: seq-add-peers.md
func TestManagerAddPeers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create manager
	manager := NewManager(ctx)

	// Create a peer
	peerID, _, err := manager.CreatePeer("", "")
	if err != nil {
		t.Fatalf("Failed to create peer: %v", err)
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

	// Test AddPeers via manager
	err = manager.AddPeers(peerID, []string{targetPeerID})
	if err != nil {
		t.Fatalf("Manager.AddPeers failed: %v", err)
	}

	// Test with non-existent peer
	err = manager.AddPeers("nonexistent", []string{targetPeerID})
	if err == nil {
		t.Fatal("Manager.AddPeers should fail for non-existent peer")
	}
}

// TestManagerRemovePeers tests the Manager.RemovePeers method
// CRC: crc-PeerManager.md
// Sequence: seq-remove-peers.md
func TestManagerRemovePeers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create manager
	manager := NewManager(ctx)

	// Create a peer
	peerID, _, err := manager.CreatePeer("", "")
	if err != nil {
		t.Fatalf("Failed to create peer: %v", err)
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

	// First add the peer
	err = manager.AddPeers(peerID, []string{targetPeerID})
	if err != nil {
		t.Fatalf("Manager.AddPeers failed: %v", err)
	}

	// Test RemovePeers via manager
	err = manager.RemovePeers(peerID, []string{targetPeerID})
	if err != nil {
		t.Fatalf("Manager.RemovePeers failed: %v", err)
	}

	// Test with non-existent peer
	err = manager.RemovePeers("nonexistent", []string{targetPeerID})
	if err == nil {
		t.Fatal("Manager.RemovePeers should fail for non-existent peer")
	}
}
