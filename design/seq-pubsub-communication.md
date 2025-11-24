# Sequence: PubSub Communication

**Source Spec:** main.md
**Use Case:** Browsers subscribe to topic and publish/receive messages with DHT-based global discovery

## Participants

- Browser1: First web application instance
- Peer1: Handles pub/sub operations (has GossipSub1)
- GossipSub1: libp2p GossipSub instance for peer1
- DHT: Distributed Hash Table for peer discovery and topic advertisement
- GossipSub2: libp2p GossipSub instance for peer2
- Peer2: Handles pub/sub operations (has GossipSub2)
- Browser2: Second web application instance

## Sequence

         ┌─┐                                                                                                                                          ┌─┐
         ║"│                                                                                                                                          ║"│
         └┬┘                                                                                                                                          └┬┘
         ┌┼┐                                                                                                                                          ┌┼┐
          │                            ┌─────┐          ┌──────────┐           ┌───┐          ┌──────────┐           ┌─────┐                           │
         ┌┴┐                           │Peer1│          │GossipSub1│           │DHT│          │GossipSub2│           │Peer2│                          ┌┴┐
      Browser1                         └──┬──┘          └─────┬────┘           └─┬─┘          └─────┬────┘           └──┬──┘                       Browser2
          │       subscribe(topic)        │                   │                  │                  │                   │                              │
          │──────────────────────────────>│                   │                  │                  │                   │                              │
          │                               │                   │                  │                  │                   │                              │
          │                               │    join topic     │                  │                  │                   │                              │
          │                               │──────────────────>│                  │                  │                   │                              │
          │                               │                   │                  │                  │                   │                              │
          │                               │    subscribe      │                  │                  │                   │                              │
          │                               │──────────────────>│                  │                  │                   │                              │
          │                               │                   │                  │                  │                   │                              │
          │                               │      advertise topic (async)        ┌┴┐                 │                   │                              │
          │                               │───────────────────────────────────> │ │                 │                   │                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │      discover peers (async)         │ │                 │                   │                              │
          │                               │───────────────────────────────────> │ │                 │                   │                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │         advertisement OK            │ │                 │                   │                              │
          │                               │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │ │                 │                   │                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │  monitor peers    │                 │ │                 │                   │                              │
          │                               │──────────────────>│                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │                   │                              │
          │           success             │                   │                 │ │                 │                   │                              │
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │                   │      subscribe(topic)        │
          │                               │                   │                 │ │                 │                   │<─────────────────────────────│
          │                               │                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │    join topic     │                              │
          │                               │                   │                 │ │                 │<──────────────────│                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │    subscribe      │                              │
          │                               │                   │                 │ │                 │<──────────────────│                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │       advertise topic (async)       │                              │
          │                               │                   │                 │ │ <───────────────────────────────────│                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │       discover peers (async)        │                              │
          │                               │                   │                 │ │ <───────────────────────────────────│                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │          advertisement OK           │                              │
          │                               │                   │                 │ │  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ >│                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │  monitor peers    │                              │
          │                               │                   │                 │ │                 │<──────────────────│                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │                   │           success            │
          │                               │                   │                 │ │                 │                   │ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ >│
          │                               │                   │                 │ │                 │                   │                              │
          │                               │                   │      ╔══════════╧═╧══════════╗      │                   │                              │
          │                               │                   │      ║Peer2 discovers Peer1 ░║      │                   │                              │
          │                               │                   │      ║via DHT topic query    ║      │                   │                              │
          │                               │                   │      ╚══════════╤═╤══════════╝      │                   │                              │
          │                               │                   │                 │ │             peer1 info              │                              │
          │                               │                   │                 │ │  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ >│                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │                   │────┐                         │
          │                               │                   │                 │ │                 │                   │    │ connect to peer1        │
          │                               │                   │                 │ │                 │                   │<───┘                         │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │            ╔══════╧═════════════════╧═╧═════════════════╧══════╗            │                              │
          │                               │            ║GossipSub mesh                                    ░║            │                              │
          │                               │            ║forms via libp2p                                   ║            │                              │
          │                               │            ╚══════╤═════════════════╤═╤═════════════════╤══════╝            │                              │
          │                               │                   │          mesh connection            │                   │                              │
          │                               │                   │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─>│                   │                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │ peerJoin(peer2)   │                 │ │                 │                   │                              │
          │                               │<──────────────────│                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │                   │                              │
          │peerChange(topic, peer2, true) │                   │                 │ │                 │                   │                              │
          │<──────────────────────────────│                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │                   │                              │
          │     publish(topic, data)      │                   │                 │ │                 │                   │                              │
          │──────────────────────────────>│                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │     publish       │                 │ │                 │                   │                              │
          │                               │──────────────────>│                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │                   │         broadcast message           │                   │                              │
          │                               │                   │────────────────────────────────────>│                   │                              │
          │                               │                   │                 │ │                 │                   │                              │
          │                               │                   │                 │ │                 │ receive message   │                              │
          │                               │                   │                 │ │                 │──────────────────>│                              │
          │                               │                   │                 └┬┘                 │                   │                              │
          │                               │                   │                  │                  │                   │topicData(topic, peer1, data) │
          │                               │                   │                  │                  │                   │─────────────────────────────>│
      Browser1                         ┌──┴──┐          ┌─────┴────┐           ┌─┴─┐          ┌─────┴────┐           ┌──┴──┐                       Browser2
         ┌─┐                           │Peer1│          │GossipSub1│           │DHT│          │GossipSub2│           │Peer2│                          ┌─┐
         ║"│                           └─────┘          └──────────┘           └───┘          └──────────┘           └─────┘                          ║"│
         └┬┘                                                                                                                                          └┬┘
         ┌┼┐                                                                                                                                          ┌┼┐
          │                                                                                                                                            │
         ┌┴┐                                                                                                                                          ┌┴┐

## Notes

- **DHT integration enables global peer discovery**: When a peer subscribes to a topic, it automatically advertises the subscription to the DHT and discovers other peers subscribed to the same topic
- **DHT operation queuing**: advertiseTopic and discoverTopicPeers use enqueueDHTOperation() to queue if DHT not ready (see seq-dht-bootstrap.md)
  - Operations queue automatically if DHT bootstrap hasn't completed
  - Operations execute immediately if DHT already ready
  - Prevents "no peers in table" errors during early subscriptions
- **advertiseTopic runs continuously**: Launched as async goroutine, re-advertises periodically (every TTL/2) to keep DHT advertisement alive until topic unsubscribed
- **discoverTopicPeers runs once**: Launched as async goroutine, queries DHT for peers advertising the topic, connects to discovered peers
- **Enables geographically distant peers to connect**: DHT provides global reach beyond local mDNS discovery, allowing peers on different networks to find each other via topic subscriptions
- **Automatic connection establishment**: When a peer discovers another via DHT, it automatically attempts connection and adds addresses to peerstore with temporary TTL
- **Subscribe() waits for gossip mesh formation** (up to 5 seconds):
  - GossipSub mesh forms via periodic heartbeats (50ms initial, 500ms interval)
  - Subscribe() blocks until at least one peer appears in mesh or timeout occurs
  - Ensures peers can communicate immediately after subscribe returns
  - Optimized GossipSub parameters for small/local networks (D=2, Dlo=1)
- **Automatic peer join/leave monitoring**: Enabled by default for subscribed topics, no separate monitoring command needed
- **Messages broadcast to all subscribed peers**: Topic data includes sender peerID for identification
- **Unsubscribe stops DHT advertisement**: The advertiseTopic goroutine stops when topic unsubscribed (handler.ctx.Done())
- **Bootstrap integration**: See seq-dht-bootstrap.md for complete DHT bootstrap and queuing behavior
