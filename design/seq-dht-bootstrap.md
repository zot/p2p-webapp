# Sequence: DHT Bootstrap and Operation Queuing

**Source Spec:** main.md
**Use Case:** Peer creation with DHT bootstrap and operation queuing to prevent "no peers in table" errors

## Participants

- PeerManager: Creates and manages peer instances
- Peer: libp2p peer with DHT support
- DHT: Distributed Hash Table for peer discovery
- BootstrapPeers: Default IPFS bootstrap nodes
- Browser: Web application initiating operations

## Sequence

     ┌───────────┐                       ┌────┐                                      ┌───┐          ┌──────────────┐           ┌───────┐
     │PeerManager│                       │Peer│                                      │DHT│          │BootstrapPeers│           │Browser│
     └─────┬─────┘                       └──┬─┘                                      └─┬─┘          └───────┬──────┘           └───┬───┘
           │                                │                                          │                    │                      │
           │                                │                 ╔═══════════════╗        │                    │                      │
═══════════╪════════════════════════════════╪═════════════════╣ Peer Creation ╠════════╪════════════════════╪══════════════════════╪═════════
           │                                │                 ╚═══════════════╝        │                    │                      │
           │                                │                                          │                    │                      │
           │         CreatePeer()           │                                          │                    │                      │
           │───────────────────────────────>│                                          │                    │                      │
           │                                │                                          │                    │                      │
           │                                │────┐                                     │                    │                      │
           │                                │    │ Initialize dhtReady channel         │                    │                      │
           │                                │<───┘                                     │                    │                      │
           │                                │                                          │                    │                      │
           │                                │────┐                                     │                    │                      │
           │                                │    │ Initialize dhtOperations queue      │                    │                      │
           │                                │<───┘                                     │                    │                      │
           │                                │                                          │                    │                      │
           │                                │────┐                                     │                    │                      │
           │                                │    │ Initialize dhtOpMu mutex            │                    │                      │
           │                                │<───┘                                     │                    │                      │
           │                                │                                          │                    │                      │
           │Launch bootstrapDHT() goroutine │                                          │                    │                      │
           │───────────────────────────────>│                                          │                    │                      │
           │                                │                                          │                    │                      │
           │                                │                                          │                    │                      │
           │                                │             ╔═══════════════════════╗    │                    │                      │
═══════════╪════════════════════════════════╪═════════════╣ DHT Bootstrap (Async) ╠════╪════════════════════╪══════════════════════╪═════════
           │                                │             ╚═══════════════════════╝    │                    │                      │
           │                                │                                          │                    │                      │
           │                                │                Connect to 3+ bootstrap peers                  │                      │
           │                                │──────────────────────────────────────────────────────────────>│                      │
           │                                │                                          │                    │                      │
           │                                │                   connections established│                    │                      │
           │                                │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │                      │
           │                                │                                          │                    │                      │
           │                                │               Bootstrap()                │                    │                      │
           │                                │─────────────────────────────────────────>│                    │                      │
           │                                │                                          │                    │                      │
           │                                │            bootstrap started             │                    │                      │
           │                                │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                    │                      │
           │                                │                                          │                    │                      │
           │                                │   Wait for RoutingTable().Size() > 0     │ ╔══════════════════╧════════╗             │
           │                                │   (max 30 seconds, poll every 500ms)     │ ║Ensures DHT has peers     ░║             │
           │                                │─────────────────────────────────────────>│ ║before operations execute  ║             │
           │                                │                                          │ ╚══════════════════╤════════╝             │
           │                                │         routing table populated          │                    │                      │
           │                                │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                    │                      │
           │                                │                                          │                    │                      │
           │                                │────┐                                     │                    │                      │
           │                                │    │ Close dhtReady channel              │                    │                      │
           │                                │<───┘                                     │                    │                      │
           │                                │                                          │                    │                      │
           │                                │────┐                                ╔════╧════════════════════╧═════╗                │
           │                                │    │ processQueuedDHTOperations()   ║Execute all queued operations ░║                │
           │                                │<───┘                                ╚════╤════════════════════╤═════╝                │
           │                                │                                          │                    │                      │
           │                                │                                          │                    │                      │
           │                                │       ╔══════════════════════════════════╗                    │                      │
═══════════╪════════════════════════════════╪═══════╣ DHT Operation (Before Bootstrap) ╠════════════════════╪══════════════════════╪═════════
           │                                │       ╚══════════════════════════════════╝                    │                      │
           │                                │                                          │                    │                      │
           │                                │                                  subscribe(topic)             │                      │
           │                                │<─────────────────────────────────────────────────────────────────────────────────────│
           │                                │                                          │                    │                      │
           │                                │────┐                                     │                    │                      │
           │                                │    │ advertiseTopic(topic)               │                    │                      │
           │                                │<───┘                                     │                    │                      │
           │                                │                                          │                    │                      │
           │                                │────┐                                     │                    │                      │
           │                                │    │ enqueueDHTOperation(advertise_func) │                    │                      │
           │                                │<───┘                                     │                    │                      │
           │                                │                                          │                    │                      │
           │                                │────┐                                 ╔═══╧════════════════════╧╗                     │
           │                                │    │ Check dhtReady (non-blocking)   ║dhtReady not closed yet ░║                     │
           │                                │<───┘                                 ╚═══╤════════════════════╤╝                     │
           │                                │                                          │                    │                      │
           │                                │────┐                   ╔═════════════════╧═════════════╗      │                      │
           │                                │    │ Queue operation   ║Operation waits for bootstrap ░║      │                      │
           │                                │<───┘                   ╚═════════════════╤═════════════╝      │                      │
           │                                │                                          │                    │                      │
           │                                │                                subscription success           │                      │
           │                                │ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ >│
           │                                │                                          │                    │                      │
           │                                │                                          │                    │                      │
           │                                │        ╔═════════════════════════════════╗                    │                      │
═══════════╪════════════════════════════════╪════════╣ DHT Operation (After Bootstrap) ╠════════════════════╪══════════════════════╪═════════
           │                                │        ╚═════════════════════════════════╝                    │                      │
           │                                │                                          │                    │                      │
           │                                │                              subscribe(another_topic)         │                      │
           │                                │<─────────────────────────────────────────────────────────────────────────────────────│
           │                                │                                          │                    │                      │
           │                                │────┐                                     │                    │                      │
           │                                │    │ advertiseTopic(another_topic)       │                    │                      │
           │                                │<───┘                                     │                    │                      │
           │                                │                                          │                    │                      │
           │                                │────┐                                     │                    │                      │
           │                                │    │ enqueueDHTOperation(advertise_func) │                    │                      │
           │                                │<───┘                                     │                    │                      │
           │                                │                                          │                    │                      │
           │                                │────┐                                 ╔═══╧════════════════════╧╗                     │
           │                                │    │ Check dhtReady (non-blocking)   ║dhtReady already closed ░║                     │
           │                                │<───┘                                 ╚═══╤════════════════════╤╝                     │
           │                                │                                          │                    │                      │
           │                                │      Advertise(topic) [immediate]        │                    │                      │
           │                                │─────────────────────────────────────────>│                    │                      │
           │                                │                                          │                    │                      │
           │                                │            advertisement OK              │                    │                      │
           │                                │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                    │                      │
           │                                │                                          │                    │                      │
           │                                │                                subscription success           │                      │
           │                                │ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ >│
     ┌─────┴─────┐                       ┌──┴─┐                                      ┌─┴─┐          ┌───────┴──────┐           ┌───┴───┐
     │PeerManager│                       │Peer│                                      │DHT│          │BootstrapPeers│           │Browser│
     └───────────┘                       └────┘                                      └───┘          └──────────────┘           └───────┘

## Notes

- **Problem solved**: DHT operations (Advertise, FindPeers) fail with "no peers in table" if called before DHT bootstrap completes
- **Solution**: Queue-based approach using channel signaling and function queue
- **dhtReady channel**: Closed when DHT routing table has peers, signals readiness to all waiters
- **dhtOperations queue**: Stores operations (as functions) that execute when DHT ready
- **dhtOpMu mutex**: Protects the operations queue following synchronization hygiene (lock → extract → unlock → process)
- **enqueueDHTOperation**: Non-blocking check of dhtReady channel, queues if not ready, executes immediately if ready
- **Bootstrap process**: Connects to 3+ bootstrap peers, calls DHT.Bootstrap(), waits max 30s for routing table to populate
- **Routing table check**: Polls RoutingTable().Size() every 500ms until > 0 or 30s timeout
- **Timeout handling**: After 30s, closes dhtReady anyway (operations may fail but are logged)
- **No DHT case**: If peer created without DHT, dhtReady is closed immediately (operations won't queue)
- **processQueuedDHTOperations**: Extracts all queued operations under lock, then executes without lock
- **advertiseTopic and discoverTopicPeers**: Both use enqueueDHTOperation wrapper to queue if DHT not ready
- **Synchronization hygiene**: Follows lock → extract data → unlock → process pattern to minimize lock duration
- **Operation transparency**: Browser receives immediate success response, queuing is transparent
- **Execution guarantees**: All queued operations execute once bootstrap completes (or times out)
