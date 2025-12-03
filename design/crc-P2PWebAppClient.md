# P2PWebAppClient

**Source Spec:** main.md

## Responsibilities

### Knows
- websocket: WebSocket connection to server
- connected: Boolean indicating if fully connected (after peer response succeeds)
- peerID: This client's peer ID (null until connected)
- peerKey: This client's peer key (null until connected)
- version: Server version received during connection (null until connected)
- onCloseCallback: Optional callback invoked when connection closes
- requestID: Current request ID counter
- pendingRequests: Map of requestID to Promise resolvers
- protocolListeners: Map of protocol to callback functions
- topicListeners: Map of topic to callback functions
- ackCallbacks: Map of ack number to delivery confirmation callbacks
- nextAckNumber: Next ack number to assign (auto-incrementing from 0)
- messageQueue: Queue for sequential server-initiated message processing
- fileListHandlers: Map of peerID to pending listFiles request handlers

### Does
- connect(options?): Connect to server and initialize peer, accepts {peerKey?, onClose?}, returns this
- connected: Getter returning true if fully connected
- start: Register protocol listener to receive (peer, data) messages
- stop: Remove protocol listener
- send: Send data to peer on protocol with optional delivery confirmation callback
- subscribe: Subscribe to topic, receive messages and peer join/leave events
- publish: Publish message to topic
- unsubscribe: Unsubscribe from topic
- listPeers: Get peers subscribed to topic
- addPeers: Protect and tag peer connections to ensure they remain active (sends addPeers request to server)
- removePeers: Unprotect and untag peer connections (sends removePeers request to server)
- listFiles: Request file list from peer (returns promise, manages deduplication and handler pattern for async peerFiles server message)
- getFile: Request IPFS content by CID with optional fallbackPeerID (triggers gotFile server message with {success, content})
- storeFile: Store file with signature storeFile(path, content) where content is string or Uint8Array, returns promise resolving to StoreFileResponse {fileCid, rootCid}
- createDirectory: Create directory with signature createDirectory(path), returns promise resolving to StoreFileResponse {fileCid, rootCid}
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
- seq-add-peers.md: Protect and tag peer connections
- seq-remove-peers.md: Unprotect and untag peer connections
- seq-list-files.md: File list request with deduplication
- seq-get-file.md: File retrieval with optional fallback peer
- seq-store-file.md: File storage on own peer
