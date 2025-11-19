# Peer

**Source Spec:** main.md

## Responsibilities

### Knows
- host: libp2p host instance for networking
- peerID: Unique peer identifier
- alias: Human-readable alias (peer-a, peer-b, ...)
- pubsub: GossipSub instance for topic-based messaging
- dht: Distributed Hash Table for peer discovery
- mdnsService: mDNS service for local peer discovery
- protocols: Map of protocol ID to ProtocolHandler
- topics: Map of topic name to TopicHandler
- monitoredTopics: Map of topic name to TopicMonitor
- vcm: VirtualConnectionManager for stream lifecycle
- directory: HAMTDirectory for file storage
- directoryCID: Current CID of the peer's directory
- fileListHandler: Handler for pending listFiles request

### Does
- start: Register protocol listener for incoming streams
- stop: Remove protocol listener
- sendToPeer: Send data to peer on protocol (create/reuse stream)
- subscribe: Subscribe to GossipSub topic, monitor peer join/leave
- publish: Publish message to GossipSub topic
- unsubscribe: Unsubscribe from topic
- listPeers: Get list of peers subscribed to topic
- monitor: Start monitoring topic for peer join/leave events
- stopMonitor: Stop monitoring topic
- listFiles: Request file list from target peer (local or remote via p2p-webapp protocol)
- getFile: Retrieve IPFS content by CID
- storeFile: Create file/directory node in IPFS, update HAMTDirectory at path, return CID of stored node
- removeFile: Remove file or directory from HAMTDirectory at path
- handleGetFileList: Handle incoming getFileList() message on p2p-webapp protocol
- handleFileList: Handle incoming fileList(CID, directory) message on p2p-webapp protocol

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
- seq-list-files.md: File list retrieval (local and remote)
- seq-store-file.md: File/directory storage in peer directory
