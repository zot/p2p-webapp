# Peer

**Source Spec:** main.md

## Responsibilities

### Knows
- host: libp2p host instance for networking (includes BasicConnMgr for connection management via ConnManager())
- peerID: Unique peer identifier
- alias: Human-readable alias (peer-a, peer-b, ...)
- pubsub: GossipSub instance for topic-based messaging
- dht: Distributed Hash Table for peer discovery
- dhtReady: Channel signaling when DHT is bootstrapped and ready for operations (closed when ready)
- dhtOperations: Queue of pending DHT operations (executed when DHT ready)
- dhtOpMu: Mutex protecting dhtOperations queue
- mdnsService: mDNS service for local peer discovery
- protocols: Map of protocol ID to ProtocolHandler
- topics: Map of topic name to TopicHandler
- monitoredTopics: Map of topic name to TopicMonitor
- vcm: VirtualConnectionManager for stream lifecycle
- directory: HAMTDirectory for file storage
- directoryCID: Current CID of the peer's directory
- fileListHandler: Handler for pending listFiles request
- addedPeers: Map tracking peers added via AddPeers (for retry attempts)

### Does
- start: Register protocol listener for incoming streams
- stop: Remove protocol listener
- sendToPeer: Send data to peer on protocol (create/reuse stream)
- subscribe: Subscribe to GossipSub topic, wait for mesh formation, monitor peer join/leave, advertise topic to DHT for global discovery, discover and connect to peers via DHT
- publish: Publish message to GossipSub topic
- unsubscribe: Unsubscribe from topic, stop DHT advertisement
- listPeers: Get list of peers subscribed to topic
- bootstrapDHT: Connect to bootstrap peers, run DHT.Bootstrap(), wait for routing table to populate (max 30s), signal readiness via dhtReady channel, process queued DHT operations
- enqueueDHTOperation: Queue DHT operation if DHT not ready, execute immediately if ready (non-blocking check)
- processQueuedDHTOperations: Execute all queued DHT operations (called after DHT ready)
- advertiseTopic: Advertise topic subscription to DHT, re-advertise periodically (runs continuously until topic unsubscribed), queues operation if DHT not ready
- discoverTopicPeers: Discover peers subscribed to topic via DHT, connect to discovered peers (runs once per subscription), queues operation if DHT not ready
- addPeers: Protect and tag peer connections using ConnManager().Protect(peerID, "connected") and TagPeer(peerID, "connected", 100), attempt connection if not connected, track for retry
- removePeers: Unprotect and untag peer connections using ConnManager().Unprotect(peerID, "connected") and UntagPeer(peerID, "connected")
- retryAddedPeersLoop: Background goroutine that periodically retries connecting to added peers via DHT lookup (every 30s)
- monitor: Start monitoring topic for peer join/leave events
- stopMonitor: Stop monitoring topic
- listFiles: Request file list from target peer (local or remote via p2p-webapp protocol)
- getFile: Retrieve IPFS content by CID, with optional fallback peer to request from if not found locally (uses p2p-webapp protocol)
- storeFile: Create file/directory node in IPFS, update HAMTDirectory at path, return file CID and root CID (StoreFileResponse), publish file update notification if configured (handles both storeFile and createDirectory operations)
- removeFile: Remove file or directory from HAMTDirectory at path, publish file update notification if configured
- publishFileUpdateNotification: Publish file change notification to configured topic (if subscribed)
- handleGetFileList: Handle incoming getFileList() message (type 0) on p2p-webapp protocol
- handleFileList: Handle incoming fileList() message (type 1) on p2p-webapp protocol
- handleGetFile: Handle incoming getFile() message (type 2) on p2p-webapp protocol - retrieve and send file content
- handleFileContent: Handle incoming fileContent() message (type 3) on p2p-webapp protocol - receive file from fallback peer

## Collaborators

- PeerManager: Provides callbacks for events (onPeerData, onTopicData, onPeerChange, onPeerFiles, onGotFile)
- libp2p Host: Manages P2P networking and streams
- GossipSub: Manages topic-based pub/sub messaging
- DHT: Manages peer discovery
- VirtualConnectionManager: Manages stream lifecycle and reliability
- HAMTDirectory: IPFS data structure for file storage
- ipfs-lite Peer: Manages IPFS connection and operations

## Sequences

- seq-protocol-communication.md: Protocol-based peer-to-peer messaging
- seq-pubsub-communication.md: Topic subscribe/publish flow
- seq-dht-bootstrap.md: DHT bootstrap and operation queuing flow
- seq-add-peers.md: Connection protection and tagging flow
- seq-remove-peers.md: Connection unprotection and untagging flow
- seq-list-files.md: File list retrieval (local and remote)
- seq-get-file.md: File retrieval with optional fallback peer
- seq-store-file.md: File/directory storage in peer directory
