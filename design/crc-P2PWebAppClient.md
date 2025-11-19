# P2PWebAppClient

**Source Spec:** main.md

## Responsibilities

### Knows
- websocket: WebSocket connection to server
- peerID: This client's peer ID (returned from connect(), null until connected)
- peerKey: This client's peer key (returned from connect(), null until connected)
- requestID: Current request ID counter
- pendingRequests: Map of requestID to Promise resolvers
- protocolListeners: Map of protocol to callback functions
- topicListeners: Map of topic to callback functions
- ackCallbacks: Map of ack number to delivery confirmation callbacks
- nextAckNumber: Next ack number to assign (auto-incrementing from 0)
- messageQueue: Queue for sequential server-initiated message processing
- fileListHandlers: Map of peerID to pending listFiles request handlers

### Does
- connect: Connect to server and initialize peer (WebSocket + Peer command)
- start: Register protocol listener to receive (peer, data) messages
- stop: Remove protocol listener
- send: Send data to peer on protocol with optional delivery confirmation callback
- subscribe: Subscribe to topic, receive messages and peer join/leave events
- publish: Publish message to topic
- unsubscribe: Unsubscribe from topic
- listPeers: Get peers subscribed to topic
- listFiles: Request file list from peer (returns promise, manages deduplication and handler pattern for async peerFiles server message)
- getFile: Request IPFS content by CID (triggers gotFile server message with {success, content})
- storeFile: Store file or directory with signature storeFile(path, content, directory) where content is null for directories
- removeFile: Remove file or directory from this peer's directory (implicit peerID)
- sendRequest: Send JSON-RPC request and return Promise
- handleResponse: Process response messages (resolve pending Promises)
- handleServerMessage: Queue and process server-initiated messages sequentially
- routePeerData: Route peerData to protocol listener
- routeTopicData: Route topicData to topic listener
- routePeerChange: Route peerChange to topic listener
- routePeerFiles: Route peerFiles(peerid, CID, entries) to pending listFiles handlers for that peerID
- routeGotFile: Route gotFile to pending getFile handlers
- routeAck: Invoke ack callback and remove from map

## Collaborators

- WebSocketHandler: Communicates via WebSocket JSON-RPC protocol
- Application: Provides callbacks for protocol and topic messages

## Sequences

- seq-client-connect.md: Connection and peer initialization
- seq-client-protocol.md: Protocol-based messaging flow
- seq-client-pubsub.md: Topic subscription and publishing
- seq-client-ack.md: Message delivery confirmation flow
- seq-list-files.md: File list request with deduplication
- seq-store-file.md: File storage on own peer
