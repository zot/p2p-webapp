# Sequence: PubSub Communication

**Source Spec:** main.md
**Use Case:** Browsers subscribe to topic and publish/receive messages

## Participants

- Browser1: First web application instance
- WebSocketHandler1: Routes requests to Peer1 via PeerManager1
- PeerManager1: Provides Peer1 instance
- Peer1: Handles pub/sub operations (has GossipSub1)
- GossipSub1: libp2p GossipSub instance for peer1
- DHT: Distributed Hash Table for peer discovery
- GossipSub2: libp2p GossipSub instance for peer2
- Peer2: Handles pub/sub operations (has GossipSub2)
- PeerManager2: Provides Peer2 instance
- WebSocketHandler2: Routes data to Browser2
- Browser2: Second web application instance

## Sequence

┌────────┐                    ┌────────────┐           ┌──────────┐            ┌───┐          ┌──────────┐           ┌────────────┐                   ┌────────┐
     │Browser1│                    │PeerManager1│           │GossipSub1│            │DHT│          │GossipSub2│           │PeerManager2│                   │Browser2│
     └────┬───┘                    └──────┬─────┘           └─────┬────┘            └─┬─┘          └─────┬────┘           └──────┬─────┘                   └────┬───┘
          │       subscribe(topic)        │                       │                   │                  │                       │                              │
          │──────────────────────────────>│                       │                   │                  │                       │                              │
          │                               │                       │                   │                  │                       │                              │
          │                               │      join topic       │                   │                  │                       │                              │
          │                               │──────────────────────>│                   │                  │                       │                              │
          │                               │                       │                   │                  │                       │                              │
          │                               │                       │ advertise topic   │                  │                       │                              │
          │                               │                       │──────────────────>│                  │                       │                              │
          │                               │                       │                   │                  │                       │                              │
          │                               │    monitor peers      │                   │                  │                       │                              │
          │                               │<──────────────────────│                   │                  │                       │                              │
          │                               │                       │                   │                  │                       │                              │
          │           success             │                       │                   │                  │                       │                              │
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │                       │                   │                  │                       │                              │
          │                               │                       │                   │                  │                       │                              │
          │                               │                       │                   │                  │                       │      subscribe(topic)        │
          │                               │                       │                   │                  │                       │<─────────────────────────────│
          │                               │                       │                   │                  │                       │                              │
          │                               │                       │                   │                  │      join topic       │                              │
          │                               │                       │                   │                  │<──────────────────────│                              │
          │                               │                       │                   │                  │                       │                              │
          │                               │                       │                   │ advertise topic  │                       │                              │
          │                               │                       │                   │<─────────────────│                       │                              │
          │                               │                       │                   │                  │                       │                              │
          │                               │                       │                   │                  │    monitor peers      │                              │
          │                               │                       │                   │                  │──────────────────────>│                              │
          │                               │                       │                   │                  │                       │                              │
          │                               │                       │                   │                  │                       │           success            │
          │                               │                       │                   │                  │                       │ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ >│
          │                               │                       │                   │                  │                       │                              │
          │                               │                       │peer2 joined topic │                  │                       │                              │
          │                               │                       │<─ ─ ─ ─ ─ ─ ─ ─ ─ │                  │                       │                              │
          │                               │                       │                   │                  │                       │                              │
          │                               │   peerJoin(peer2)     │                   │                  │                       │                              │
          │                               │<──────────────────────│                   │                  │                       │                              │
          │                               │                       │                   │                  │                       │                              │
          │peerChange(topic, peer2, true) │                       │                   │                  │                       │                              │
          │<──────────────────────────────│                       │                   │                  │                       │                              │
          │                               │                       │                   │                  │                       │                              │
          │     publish(topic, data)      │                       │                   │                  │                       │                              │
          │──────────────────────────────>│                       │                   │                  │                       │                              │
          │                               │                       │                   │                  │                       │                              │
          │                               │       publish         │                   │                  │                       │                              │
          │                               │──────────────────────>│                   │                  │                       │                              │
          │                               │                       │                   │                  │                       │                              │
          │                               │                       │         broadcast to topic           │                       │                              │
          │                               │                       │─────────────────────────────────────>│                       │                              │
          │                               │                       │                   │                  │                       │                              │
          │                               │                       │                   │                  │   receive message     │                              │
          │                               │                       │                   │                  │──────────────────────>│                              │
          │                               │                       │                   │                  │                       │                              │
          │                               │                       │                   │                  │                       │topicData(topic, peer1, data) │
          │                               │                       │                   │                  │                       │─────────────────────────────>│
     ┌────┴───┐                    ┌──────┴─────┐           ┌─────┴────┐            ┌─┴─┐          ┌─────┴────┐           ┌──────┴─────┐                   ┌────┴───┐
     │Browser1│                    │PeerManager1│           │GossipSub1│            │DHT│          │GossipSub2│           │PeerManager2│                   │Browser2│
     └────────┘                    └────────────┘           └──────────┘            └───┘          └──────────┘           └────────────┘                   └────────┘

## Notes

- GossipSub integrated with DHT for peer discovery
- Peers advertise topic subscriptions via DHT
- Automatic peer join/leave monitoring for subscribed topics
- Handler gets Peer from PeerManager, then calls peer.Subscribe(topic)
- **Subscribe() waits for gossip mesh formation** (up to 2 seconds):
  - GossipSub mesh forms via periodic heartbeats (50ms initial, 500ms interval)
  - Subscribe() blocks until at least one peer appears in mesh or timeout occurs
  - Ensures peers can communicate immediately after subscribe returns
  - Optimized GossipSub parameters for small/local networks (D=2, Dlo=1)
- Handler gets Peer from PeerManager, then calls peer.Publish(topic, data)
- Peer sends peerChange notifications automatically via PeerManager callbacks
- Messages broadcast to all subscribed peers
- Topic data includes sender peerID for identification
- No separate monitoring command needed - automatic with subscribe
