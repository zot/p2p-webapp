# PeerManager

**Source Spec:** main.md

## Responsibilities

### Knows
- peers: Map of peerID to Peer instances
- peerAliases: Map of peerID to human-readable alias (peer-a, peer-b, ...)
- aliasCounter: Counter for generating unique aliases
- verbosity: Logging verbosity level
- virtualConnections: Map of "peerID:protocol" to libp2p streams

### Does
- createPeer: Create new libp2p peer with given or fresh peer key
- enableDiscovery: Configure mDNS and DHT discovery
- enableNATTraversal: Configure Circuit Relay, hole punching, AutoRelay, port mapping
- startProtocol: Register protocol listener for a peer
- stopProtocol: Remove protocol listener
- sendToPeer: Send data to peer on protocol (create stream on-demand, reuse existing)
- subscribeTopic: Subscribe peer to GossipSub topic, monitor peer join/leave
- publishToTopic: Publish message to GossipSub topic
- unsubscribeTopic: Unsubscribe peer from topic
- listTopicPeers: Get list of peers subscribed to topic
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
