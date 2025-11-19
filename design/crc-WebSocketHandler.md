# WebSocketHandler

**Source Spec:** main.md

## Responsibilities

### Knows
- connections: Active WebSocket connections mapped to peers
- requestID: Current request ID counter for protocol messages
- messageQueue: Queue for sequential server-initiated message processing

### Does
- acceptConnection: Accept new WebSocket connection from browser
- receiveMessage: Receive and parse JSON-RPC messages from client
- sendMessage: Send JSON-RPC responses and server-initiated messages to client
- routeRequest: Route client request to appropriate handler
- routeFileOperations: Route listFiles/getFile/storeFile/removeFile to PeerManager with connection's peerID
- enforceFileOwnership: Ensure storeFile/removeFile operate only on connection's own peer
- queueServerMessage: Queue server-initiated messages for sequential processing
- closeConnection: Clean up connection and associated peer

## Collaborators

- PeerManager: Creates peer when client sends Peer() request, manages protocol streams and file operations
- Server: Reports connection lifecycle events

## Sequences

- seq-websocket-connect.md: WebSocket connection establishment
- seq-client-request.md: Client request/response flow
- seq-server-notification.md: Server-initiated message delivery
- seq-list-files.md: File list request routing with peerID
- seq-store-file.md: File storage with ownership enforcement
