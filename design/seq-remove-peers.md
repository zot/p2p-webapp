# Sequence: Remove Peers (Connection Unprotection)

**Source Spec:** main.md
**Use Case:** Browser unprotects and untags peer connections to allow normal connection management

## Participants

- Browser: Web application instance
- P2PWebAppClient: Client library
- WebSocketHandler: Routes requests via PeerManager
- PeerManager: Provides Peer instance
- Peer: Executes connection management operations
- ConnManager: libp2p BasicConnMgr for connection protection/tagging
- Host: libp2p host for peer connections

## Sequence

┌───────┐      ┌────────────────┐      ┌────────────────┐      ┌───────────┐      ┌────┐      ┌───────────┐
│Browser│      │P2PWebAppClient │      │WebSocketHandler│      │PeerManager│      │Peer│      │ConnManager│
└───┬───┘      └───────┬────────┘      └───────┬────────┘      └─────┬─────┘      └──┬─┘      └─────┬─────┘
    │                  │                       │                     │               │              │
    │removePeers(peerIds)                     │                     │               │              │
    │─────────────────>│                       │                     │               │              │
    │                  │                       │                     │               │              │
    │                  │ removePeers request   │                     │               │              │
    │                  │──────────────────────>│                     │               │              │
    │                  │                       │                     │               │              │
    │                  │                       │ GetPeer(peerID)     │               │              │
    │                  │                       │────────────────────>│               │              │
    │                  │                       │                     │               │              │
    │                  │                       │      Peer instance  │               │              │
    │                  │                       │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │               │              │
    │                  │                       │                     │               │              │
    │                  │                       │   RemovePeers(targetPeerIDs)        │              │
    │                  │                       │────────────────────────────────────>│              │
    │                  │                       │                     │               │              │
    │                  │                       │                     │               │ For each peerID:
    │                  │                       │                     │               │              │
    │                  │                       │                     │               │ Unprotect(peerID, "connected")
    │                  │                       │                     │               │─────────────>│
    │                  │                       │                     │               │              │
    │                  │                       │                     │               │   success    │
    │                  │                       │                     │               │<─ ─ ─ ─ ─ ─ ─│
    │                  │                       │                     │               │              │
    │                  │                       │                     │               │ UntagPeer(peerID, "connected")
    │                  │                       │                     │               │─────────────>│
    │                  │                       │                     │               │              │
    │                  │                       │                     │               │   success    │
    │                  │                       │                     │               │<─ ─ ─ ─ ─ ─ ─│
    │                  │                       │                     │               │              │
    │                  │                       │                     │               │   (repeat for next peerID)
    │                  │                       │                     │               │              │
    │                  │                       │                     │         success              │
    │                  │                       │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│              │
    │                  │                       │                     │               │              │
    │                  │        success        │                     │               │              │
    │                  │<──────────────────────│                     │               │              │
    │                  │                       │                     │               │              │
    │     resolved     │                       │                     │               │              │
    │<─ ─ ─ ─ ─ ─ ─ ─ ─│                       │                     │               │              │
┌───┴───┐      ┌───────┴────────┐      ┌───────┴────────┐      ┌─────┴─────┐      ┌──┴─┐      ┌─────┴─────┐
│Browser│      │P2PWebAppClient │      │WebSocketHandler│      │PeerManager│      │Peer│      │ConnManager│
└───────┘      └────────────────┘      └────────────────┘      └───────────┘      └────┘      └───────────┘

## Notes

- **BasicConnMgr**: libp2p connection manager (github.com/libp2p/go-libp2p/p2p/net/connmgr)
  - Accessed via `host.ConnManager()` method
  - Provides Protect/Unprotect and TagPeer/UntagPeer methods
- **Unprotect(peerID, tag)**: Removes protection from peer connection
  - Tag "connected" identifies the protection to remove
  - Connection manager can now close this connection if needed
  - Does NOT actively close the connection, just allows it to be closed
- **UntagPeer(peerID, tag)**: Removes priority tag from peer connection
  - Tag "connected" identifies the tag to remove
  - Connection loses its priority value (100)
  - Reverts to default priority for connection management decisions
- **Error handling**: Silently skips peers that fail any operation
  - No errors returned to client for individual peer failures
  - Overall operation succeeds unless complete failure
  - Common case: peer already unprotected/untagged (no error)
- **No disconnection**: This operation does NOT close peer connections
  - Only removes protection and priority
  - Connections remain active until manager decides to prune them
  - Or until normal connection lifecycle closes them
- Handler routes removePeers request to Manager.RemovePeers(peerID, targetPeerIDs)
- Manager gets Peer instance and delegates to Peer.RemovePeers(targetPeerIDs)
- Operations apply to current peer's host, affecting connections to target peers
