# PeerManager

**Source Spec:** main.md

## Responsibilities

### Knows
- peers: Map of peerID to Peer instances
- peerAliases: Map of peerID to human-readable alias (peer-a, peer-b, ...)
- aliasCounter: Counter for generating unique aliases
- verbosity: Logging verbosity level
- virtualConnections: Map of "peerID:protocol" to libp2p streams
- peerDirectories: Map of peerID to HAMTDirectory (each peer owns its directory, pins it)
- peerDirectoryCIDs: Map of peerID to current directory CID
- fileListHandlers: Map of peerID to pending getFileList/fileList protocol handlers

### Does
- createPeer: Create new libp2p peer with given or fresh peer key, accepts optional rootDirectory CID to restore state
- enableDiscovery: Configure mDNS and DHT discovery
- enableNATTraversal: Configure Circuit Relay, hole punching, AutoRelay, port mapping
- startProtocol: Register protocol listener for a peer
- stopProtocol: Remove protocol listener
- sendToPeer: Send data to peer on protocol (create stream on-demand, reuse existing)
- subscribeTopic: Subscribe peer to GossipSub topic, monitor peer join/leave
- publishToTopic: Publish message to GossipSub topic
- unsubscribeTopic: Unsubscribe peer from topic
- listTopicPeers: Get list of peers subscribed to topic
- listFiles: Request file list from peer (local or remote via reserved "p2p-webapp" libp2p protocol)
- getFile: Retrieve IPFS content by CID
- storeFile: Create file/directory node in IPFS, update peer's HAMTDirectory at path, update and pin CID
- removeFile: Remove file or directory from peer's HAMTDirectory at path, update CID
- handleGetFileList: Handle incoming getFileList() libp2p message on reserved "p2p-webapp" protocol
- handleFileList: Handle incoming fileList(CID, directory) libp2p message on reserved "p2p-webapp" protocol
- logVerbose: Log with peer alias prefix at appropriate verbosity level
- generateAlias: Generate human-readable alias for new peer

## Collaborators

- WebSocketHandler: Notifies of peer data, topic data, peer changes, acks
- ipfs-lite library: Manages IPFS connection and DHT operations
- libp2p library: Manages P2P networking, streams, protocols, GossipSub
- allowPrivateGater: Allows connections on localhost for development

## Sequences

- seq-peer-creation.md: Peer initialization and discovery setup
- seq-protocol-communication.md: Protocol-based peer-to-peer messaging
- seq-pubsub-communication.md: Topic subscribe/publish flow
- seq-peer-discovery.md: mDNS and DHT peer discovery
- seq-list-files.md: File list retrieval (local and remote)
- seq-store-file.md: File/directory storage in peer directory
