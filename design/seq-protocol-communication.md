# Sequence: Protocol-Based Communication

**Source Spec:** main.md
**Use Case:** Browser sends data to peer using protocol-based addressing

## Participants

- Browser1: Sending web application
- PeerManager1 (PM1): Manages local peer
- Peer1: Local libp2p peer
- Peer2: Remote libp2p peer
- PeerManager2 (PM2): Manages remote peer
- Browser2: Receiving web application

## Sequence

                    ┌────────┐                       ┌────────────┐                        ┌─────┐           ┌─────┐          ┌────────────┐                      ┌────────┐
                    │Browser1│                       │PeerManager1│                        │Peer1│           │Peer2│          │PeerManager2│                      │Browser2│
                    └────┬───┘                       └──────┬─────┘                        └──┬──┘           └──┬──┘          └──────┬─────┘                      └────┬───┘
                         │         start(protocol)          │                                 │                 │                    │                                 │
                         │─────────────────────────────────>│                                 │                 │                    │                                 │
                         │                                  │                                 │                 │                    │                                 │
                         │                                  │────┐                            │                 │                    │                                 │
                         │                                  │    │ register protocol listener │                 │                    │                                 │
                         │                                  │<───┘                            │                 │                    │                                 │
                         │                                  │                                 │                 │                    │                                 │
                         │             success              │                                 │                 │                    │                                 │
                         │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                                 │                 │                    │                                 │
                         │                                  │                                 │                 │                    │                                 │
                         │send(peer2, protocol, data, ack?) │                                 │                 │                    │                                 │
                         │─────────────────────────────────>│                                 │                 │                    │                                 │
                         │                                  │                                 │                 │                    │                                 │
                         │                                  │────┐                            │                 │                    │                                 │
                         │                                  │    │ get or create stream       │                 │                    │                                 │
                         │                                  │<───┘                            │                 │                    │                                 │
                         │                                  │                                 │                 │                    │                                 │
                         │                                  │ ╔══════════════════════╗        │                 │                    │                                 │
                         │                                  │ ║key: peerID:protocol ░║        │                 │                    │                                 │
                         │                                  │ ╚══════════════════════╝        │                 │                    │                                 │
                         │                                  │          open stream            │                 │                    │                                 │
                         │                                  │────────────────────────────────>│                 │                    │                                 │
                         │                                  │                                 │                 │                    │                                 │
                         │                                  │                                 │ libp2p stream   │                    │                                 │
                         │                                  │                                 │────────────────>│                    │                                 │
                         │                                  │                                 │                 │                    │                                 │
                         │                                  │                                 │  stream ready   │                    │                                 │
                         │                                  │                                 │<─ ─ ─ ─ ─ ─ ─ ─ │                    │                                 │
                         │                                  │                                 │                 │                    │                                 │
                         │                                  │           write data            │                 │                    │                                 │
                         │                                  │────────────────────────────────>│                 │                    │                                 │
                         │                                  │                                 │                 │                    │                                 │
                         │                                  │                                 │data over stream │                    │                                 │
                         │                                  │                                 │────────────────>│                    │                                 │
                         │                                  │                                 │                 │                    │                                 │
                         │                                  │                                 │                 │                    │                                 │
          ╔══════╤═══════╪══════════════════════════════════╪═════════════════════════════════╪═════════════════╪════════════╗       │                                 │
          ║ ALT  │  ack requested                           │                                 │                 │            ║       │                                 │
          ╟──────┘       │                                  │                                 │                 │            ║       │                                 │
          ║              │                                  │────┐                            │                 │            ║       │                                 │
          ║              │                                  │    │ wait for confirmation      │                 │            ║       │                                 │
          ║              │                                  │<───┘                            │                 │            ║       │                                 │
          ║              │                                  │                                 │                 │            ║       │                                 │
          ║              │                                  │                                 │      ack        │            ║       │                                 │
          ║              │                                  │                                 │<─ ─ ─ ─ ─ ─ ─ ─ │            ║       │                                 │
          ║              │                                  │                                 │                 │            ║       │                                 │
          ║              │         ack(ackNumber)           │                                 │                 │            ║       │                                 │
          ║              │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                                 │                 │            ║       │                                 │
          ╚══════════════╪══════════════════════════════════╪═════════════════════════════════╪═════════════════╪════════════╝       │                                 │
                         │                                  │                                 │                 │                    │                                 │
                         │                                  │                                 │                 │   receive data     │                                 │
                         │                                  │                                 │                 │───────────────────>│                                 │
                         │                                  │                                 │                 │                    │                                 │
                         │                                  │                                 │                 │                    │────┐                            │
                         │                                  │                                 │                 │                    │    │ route to protocol listener │
                         │                                  │                                 │                 │                    │<───┘                            │
                         │                                  │                                 │                 │                    │                                 │
                         │                                  │                                 │                 │                    │peerData(peer1, protocol, data)  │
                         │                                  │                                 │                 │                    │────────────────────────────────>│
                         │                                  │                                 │                 │                    │                                 │
                         │                                  │                                 │                 │                    │            handled              │
                         │                                  │                                 │                 │                    │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │
                    ┌────┴───┐                       ┌──────┴─────┐                        ┌──┴──┐           ┌──┴──┐          ┌──────┴─────┐                      ┌────┴───┐
                    │Browser1│                       │PeerManager1│                        │Peer1│           │Peer2│          │PeerManager2│                      │Browser2│
                    └────────┘                       └────────────┘                        └─────┘           └─────┘          └────────────┘                      └────────┘

## Notes

- Protocol must be started before sending (validates protocol is registered)
- Uses virtual connection model: client addresses by (peer, protocol) tuple
- Server manages stream lifecycle transparently
- Streams created on-demand, reused for subsequent messages
- Key format: "peerID:protocol" for stream lookup
- Optional ack callback provides delivery confirmation
- Receiving peer routes data to registered protocol listener
- All server-initiated messages (peerData) processed sequentially via queue
