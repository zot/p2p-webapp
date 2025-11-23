# Sequence: Add Peers (Connection Protection)

**Source Spec:** main.md
**Use Case:** Browser protects and tags peer connections to ensure they remain active

## Participants

- Browser: Web application instance
- P2PWebAppClient: Client library
- WebSocketHandler: Routes requests via PeerManager
- PeerManager: Provides Peer instance
- Peer: Executes connection management operations
- ConnManager: libp2p BasicConnMgr for connection protection/tagging
- Host: libp2p host for peer connections

## Sequence

┌───────┐      ┌────────────────┐      ┌────────────────┐      ┌───────────┐      ┌────┐      ┌───────────┐      ┌────┐
│Browser│      │P2PWebAppClient │      │WebSocketHandler│      │PeerManager│      │Peer│      │ConnManager│      │Host│
└───┬───┘      └───────┬────────┘      └───────┬────────┘      └─────┬─────┘      └──┬─┘      └─────┬─────┘      └──┬─┘
    │                  │                       │                     │               │              │               │
    │ addPeers(peerIds)│                       │                     │               │              │               │
    │─────────────────>│                       │                     │               │              │               │
    │                  │                       │                     │               │              │               │
    │                  │  addPeers request     │                     │               │              │               │
    │                  │──────────────────────>│                     │               │              │               │
    │                  │                       │                     │               │              │               │
    │                  │                       │ GetPeer(peerID)     │               │              │               │
    │                  │                       │────────────────────>│               │              │               │
    │                  │                       │                     │               │              │               │
    │                  │                       │      Peer instance  │               │              │               │
    │                  │                       │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │               │              │               │
    │                  │                       │                     │               │              │               │
    │                  │                       │   AddPeers(targetPeerIDs)           │              │               │
    │                  │                       │────────────────────────────────────>│              │               │
    │                  │                       │                     │               │              │               │
    │                  │                       │                     │               │ For each peerID:             │
    │                  │                       │                     │               │              │               │
    │                  │                       │                     │               │ Protect(peerID, "connected") │
    │                  │                       │                     │               │─────────────>│               │
    │                  │                       │                     │               │              │               │
    │                  │                       │                     │               │   success    │               │
    │                  │                       │                     │               │<─ ─ ─ ─ ─ ─ ─│               │
    │                  │                       │                     │               │              │               │
    │                  │                       │                     │               │ TagPeer(peerID, "connected", 100)
    │                  │                       │                     │               │─────────────>│               │
    │                  │                       │                     │               │              │               │
    │                  │                       │                     │               │   success    │               │
    │                  │                       │                     │               │<─ ─ ─ ─ ─ ─ ─│               │
    │                  │                       │                     │               │              │               │
    │                  │                       │                     │               │ Connect(ctx, peerAddrInfo)   │
    │                  │                       │                     │               │─────────────────────────────>│
    │                  │                       │                     │               │              │               │
    │                  │                       │                     │               │success/ignore errors         │
    │                  │                       │                     │               │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│
    │                  │                       │                     │               │              │               │
    │                  │                       │                     │               │   (repeat for next peerID)   │
    │                  │                       │                     │               │              │               │
    │                  │                       │                     │         success              │               │
    │                  │                       │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │              │               │
    │                  │                       │                     │               │              │               │
    │                  │        success        │                     │               │              │               │
    │                  │<──────────────────────│                     │               │              │               │
    │                  │                       │                     │               │              │               │
    │     resolved     │                       │                     │               │              │               │
    │<─ ─ ─ ─ ─ ─ ─ ─ ─│                       │                     │               │              │               │
┌───┴───┐      ┌───────┴────────┐      ┌───────┴────────┐      ┌─────┴─────┐      ┌──┴─┐      ┌─────┴─────┐      ┌──┴─┐
│Browser│      │P2PWebAppClient │      │WebSocketHandler│      │PeerManager│      │Peer│      │ConnManager│      │Host│
└───────┘      └────────────────┘      └────────────────┘      └───────────┘      └────┘      └───────────┘      └────┘

## Notes

- **BasicConnMgr**: libp2p connection manager (github.com/libp2p/go-libp2p/p2p/net/connmgr)
  - Accessed via `host.ConnManager()` method
  - Provides Protect/Unprotect and TagPeer/UntagPeer methods
- **Protect(peerID, tag)**: Prevents connection manager from closing this peer connection
  - Tag "connected" used to identify protected connections
  - Multiple protections can be applied with different tags
- **TagPeer(peerID, tag, value)**: Assigns priority value (100) to peer connection
  - Higher values indicate higher priority when manager prunes connections
  - Tag "connected" matches the protection tag for consistency
- **Connection attempt**: Best-effort - silently ignores failures
  - Uses peer store addresses if available
  - May attempt DHT lookup for addresses
  - Failures don't abort the operation
- **Error handling**: Silently skips peers that fail any operation
  - No errors returned to client for individual peer failures
  - Overall operation succeeds unless complete failure
- Handler routes addPeers request to Manager.AddPeers(peerID, targetPeerIDs)
- Manager gets Peer instance and delegates to Peer.AddPeers(targetPeerIDs)
- Operations apply to current peer's host, protecting connections to target peers
