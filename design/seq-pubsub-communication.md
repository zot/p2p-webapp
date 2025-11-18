# Sequence: PubSub Communication

**Source Spec:** main.md
**Use Case:** Browsers subscribe to topic and publish/receive messages

## Participants

- Browser1: First web application instance
- PeerManager1 (PM1): Manages first peer
- GossipSub1: libp2p GossipSub instance for peer1
- DHT: Distributed Hash Table for peer discovery
- GossipSub2: libp2p GossipSub instance for peer2
- PeerManager2 (PM2): Manages second peer
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
- PeerManager sends peerChange notifications automatically
- Messages broadcast to all subscribed peers
- Topic data includes sender peerID for identification
- No separate monitoring command needed - automatic with subscribe
