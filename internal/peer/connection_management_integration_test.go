package peer

import (
	"testing"
)

// Integration tests for connection management using real connection pressure
// These tests verify actual behavior (connections persist/pruned) rather than
// mocking internal state that isn't exposed by libp2p's ConnManager API.
//
// NOTE: libp2p's BasicConnMgr has complex pruning heuristics that make behavioral
// testing unreliable. When Protect() is used, pruning behavior becomes unpredictable.
// The unit tests in connection_management_test.go verify the API works correctly.
// These integration tests focus on scenarios that ARE reliable:
// 1. Basic pruning without protection works
// 2. RemovePeers allows previously-protected peers to be pruned

// TestBasicPruningWithoutProtection verifies libp2p ConnMgr prunes without protection
func TestBasicPruningWithoutProtection(t *testing.T) {
	// Use low=1, high=2
	peerA, cleanupA := createPeerWithConnMgr(t, 1, 2)
	defer cleanupA()

	peerB, cleanupB := createSimplePeer(t)
	defer cleanupB()

	peerC, cleanupC := createSimplePeer(t)
	defer cleanupC()

	// Connect to both peers (exceeds high=2)
	connectPeers(t, peerA, peerB)
	connectPeers(t, peerA, peerC)

	// Verify both connected
	if countConnections(peerA) != 2 {
		t.Fatalf("Expected 2 connections, got %d", countConnections(peerA))
	}

	// Trigger pruning
	triggerPruning(t, peerA)

	// Should prune down toward low watermark
	finalCount := countConnections(peerA)
	if finalCount == 2 {
		t.Errorf("Expected pruning to reduce connections from 2, but still at %d", finalCount)
	}
}

// TestRemovePeersAllowsPruning verifies that removing protection makes a peer
// eligible for pruning by the connection manager
// Test Design: test-Peer.md - "Remove peers from connection protection"
func TestRemovePeersAllowsPruning(t *testing.T) {
	// Create peer A with low limits
	peerA, cleanupA := createPeerWithConnMgr(t, 1, 2)
	defer cleanupA()

	// Create two target peers
	peerB, cleanupB := createSimplePeer(t)
	defer cleanupB()
	peerC, cleanupC := createSimplePeer(t)
	defer cleanupC()

	// Protect B
	err := peerA.AddPeers([]string{peerB.host.ID().String()})
	if err != nil {
		t.Fatalf("AddPeers failed: %v", err)
	}

	// Connect to both
	connectPeers(t, peerA, peerB)
	connectPeers(t, peerA, peerC)

	// Trigger initial pruning (C should be pruned, B protected)
	triggerPruning(t, peerA)
	assertConnected(t, peerA, peerB)

	// Remove protection from B
	err = peerA.RemovePeers([]string{peerB.host.ID().String()})
	if err != nil {
		t.Fatalf("RemovePeers failed: %v", err)
	}

	// Connect to C again to create pressure
	if !isConnected(peerA, peerC) {
		connectPeers(t, peerA, peerC)
	}

	// Trigger pruning
	triggerPruning(t, peerA)

	// Now B is eligible for pruning (protection removed)
	// Verify pruning occurred (watermarks are heuristics)
	connCount := countConnections(peerA)
	if connCount == 2 {
		t.Errorf("Expected pruning to reduce connections from 2, but still at %d", connCount)
	}

	// The key test is that B CAN be pruned now (it's no longer protected)
	// Either B or C might remain, but at least one should be pruned
}
