# Sequence: Peer Creation

**Source Spec:** main.md
**Use Case:** Browser connects to server and creates a peer

## Participants

- Browser: Web application in user's browser
- WebSocketHandler: Manages WebSocket connections and JSON-RPC protocol
- PeerManager: Creates and manages libp2p peers
- Peer: Individual libp2p peer instance

## Sequence

                    ┌───────┐          ┌────────────────┐           ┌───────────┐                    ┌────┐
                    │Browser│          │WebSocketHandler│           │PeerManager│                    │Peer│
                    └───┬───┘          └────────┬───────┘           └─────┬─────┘                    └──┬─┘
                        │  WebSocket connect    │                         │                             │
                        │──────────────────────>│                         │                             │
                        │                       │                         │                             │
                        │connection established │                         │                             │
                        │<──────────────────────│                         │                             │
                        │                       │                         │                             │
                        │    Peer(peerKey?)     │                         │                             │
                        │──────────────────────>│                         │                             │
                        │                       │                         │                             │
                        │                       │  createPeer(peerKey)    │                             │
                        │                       │────────────────────────>│                             │
                        │                       │                         │                             │
                        │                       │                         │        new(peerKey)         │
                        │                       │                         │────────────────────────────>│
                        │                       │                         │                             │
                        │                       │                         │                             │ ╔═════════════════════╗
                        │                       │                         │                             │ ║Generate or use key ░║
                        │                       │                         │                             │ ╚═════════════════════╝
                        │                       │                         │                             │────┐
                        │                       │                         │                             │    │ initialize libp2p
                        │                       │                         │                             │<───┘
                        │                       │                         │                             │
                        │                       │                         │                             │────┐
                        │                       │                         │                             │    │ enable discovery
                        │                       │                         │                             │<───┘
                        │                       │                         │                             │
                        │                       │                         │       peer instance         │
                        │                       │                         │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │
                        │                       │                         │                             │
                        │                       │                         │────┐                        │
                        │                       │                         │    │ check duplicate peerID │
                        │                       │                         │<───┘                        │
                        │                       │                         │                             │
                        │                       │                         │                             │
          ╔══════╤══════╪═══════════════════════╪═════════════════════════╪═════════════════════════╗   │
          ║ ALT  │  peerID already registered   │                         │                         ║   │
          ╟──────┘      │                       │                         │                         ║   │
          ║             │                       │ error: duplicate peer   │                         ║   │
          ║             │                       │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │                         ║   │
          ║             │                       │                         │                         ║   │
          ║             │    error response     │                         │                         ║   │
          ║             │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │                         │                         ║   │
          ╠═════════════╪═══════════════════════╪═════════════════════════╪═════════════════════════╣   │
          ║ [new peer]  │                       │                         │                         ║   │
          ║             │                       │                         │────┐                    ║   │
          ║             │                       │                         │    │ store peer         ║   │
          ║             │                       │                         │<───┘                    ║   │
          ║             │                       │                         │                         ║   │
          ║             │                       │                         │────┐                    ║   │
          ║             │                       │                         │    │ generate alias     ║   │
          ║             │                       │                         │<───┘                    ║   │
          ║             │                       │                         │                         ║   │
          ║             │                       │   [peerID, peerKey]     │                         ║   │
          ║             │                       │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │                         ║   │
          ║             │                       │                         │                         ║   │
          ║             │ {peerid,peerkey,ver}  │                         │                         ║   │
          ║             │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │                         │                         ║   │
          ╚═════════════╪═══════════════════════╪═════════════════════════╪═════════════════════════╝   │
                    ┌───┴───┐          ┌────────┴───────┐           ┌─────┴─────┐                    ┌──┴─┐
                    │Browser│          │WebSocketHandler│           │PeerManager│                    │Peer│
                    └───────┘          └────────────────┘           └───────────┘                    └────┘

## Notes

- Peer() must be the first command from browser after WebSocket connection
- Cannot be sent more than once per connection
- If peerKey not provided, generates fresh key
- Duplicate peerID check prevents multiple tabs using same peer identity
- PeerManager generates human-readable aliases (peer-a, peer-b, etc.) for logging
- Peer discovery (mDNS + DHT) is enabled automatically during initialization
- Response includes server version for client to store and expose via version getter
