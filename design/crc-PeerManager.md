# PeerManager

**Source Spec:** main.md

## Responsibilities

### Knows
- peers: Map of peerID to Peer instances
- peerAliases: Map of peerID to human-readable alias (peer-a, peer-b, ...)
- aliasCounter: Counter for generating unique aliases
- verbosity: Logging verbosity level
- ipfsPeer: IPFS peer for file storage operations
- fileUpdateNotifyTopic: Optional topic for publishing file change notifications (from config)
- onPeerData: Callback for protocol data events
- onTopicData: Callback for topic data events
- onPeerChange: Callback for topic peer join/leave events
- onPeerFiles: Callback for file list responses
- onGotFile: Callback for file retrieval responses

### Does
- createPeer: Create new libp2p peer with given or fresh peer key, accepts optional rootDirectory CID to restore state
- removePeer: Remove peer and clean up resources
- getPeer: Return Peer instance by peerID
- enableDiscovery: Configure mDNS and DHT discovery for peer
- enableNATTraversal: Configure Circuit Relay, hole punching, AutoRelay, port mapping for peer
- setCallbacks: Set callback functions for events
- logVerbose: Log with peer alias prefix at appropriate verbosity level
- getOrCreateAlias: Generate human-readable alias for peer (or return existing)

## Collaborators

- Peer: Creates and manages Peer instances, provides callbacks
- WebSocketHandler: Receives callbacks with peer events (data, topic messages, peer changes, file responses)
- ipfs-lite Peer: Provides IPFS connection for file operations

## Sequences

- seq-peer-creation.md: Peer initialization and discovery setup
