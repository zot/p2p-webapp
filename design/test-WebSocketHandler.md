# Test Design: WebSocketHandler

**Source Specs**: main.md
**CRC Cards**: crc-WebSocketHandler.md
**Sequences**: seq-peer-creation.md, seq-list-files.md, seq-store-file.md

## Overview

Test suite for WebSocketHandler component covering WebSocket connections, JSON-RPC protocol, request routing, file operation routing, ownership enforcement, and message queuing.

## Test Cases

### Test: Accept WebSocket connection

**Purpose**: Verify that WebSocketHandler accepts and establishes WebSocket connection from browser.

**Motivation**: Foundation for client-server communication.

**Input**:
- WebSocketHandler listening on port 10000
- Browser initiates WebSocket connection to ws://localhost:10000

**References**:
- CRC: crc-WebSocketHandler.md - "Does: acceptConnection"

**Expected Results**:
- WebSocket connection established
- Connection stored in connections map (without peer yet)
- Connection ready to receive messages
- No peer created yet (requires Peer() command)

**References**:
- CRC: crc-WebSocketHandler.md - "Knows: connections"

---

### Test: Receive and parse JSON-RPC message

**Purpose**: Verify that WebSocketHandler receives and parses JSON-RPC formatted messages.

**Motivation**: Core protocol compliance. Must handle standard JSON-RPC structure.

**Input**:
- Established WebSocket connection
- Client sends: `{"jsonrpc": "2.0", "id": 1, "method": "Peer", "params": {}}`

**References**:
- CRC: crc-WebSocketHandler.md - "Does: receiveMessage"

**Expected Results**:
- Message received and parsed successfully
- Extracted: method="Peer", id=1, params={}
- Message routed to appropriate handler
- No parsing errors

**References**:
- CRC: crc-WebSocketHandler.md - "Does: routeRequest"

---

### Test: Send JSON-RPC response

**Purpose**: Verify that WebSocketHandler sends properly formatted JSON-RPC responses.

**Motivation**: Protocol compliance for client responses.

**Input**:
- Client request with id=1
- Handler generates response data: {peerID: "12D3...", peerKey: "CAA..."}

**References**:
- CRC: crc-WebSocketHandler.md - "Does: sendMessage"

**Expected Results**:
- Response sent: `{"jsonrpc": "2.0", "id": 1, "result": {peerID: "12D3...", peerKey: "CAA..."}}`
- Message properly JSON-formatted
- Client receives response

**References**:
- CRC: crc-WebSocketHandler.md - "Does: sendMessage"

---

### Test: Route Peer() request to PeerManager

**Purpose**: Verify that WebSocketHandler routes Peer() command to PeerManager for peer creation.

**Motivation**: First command after connection must create peer.

**Input**:
- New WebSocket connection
- Client sends: `{"jsonrpc": "2.0", "id": 1, "method": "Peer", "params": {"peerKey": null}}`

**References**:
- CRC: crc-WebSocketHandler.md - "Does: routeRequest"
- CRC: crc-WebSocketHandler.md - "Collaborators: PeerManager"

**Expected Results**:
- Handler calls PeerManager.createPeer(nil)
- Peer created and associated with connection
- Connection stored in connections map with peerID
- Response sent: {peerID: "12D3...", peerKey: "CAA..."}

**References**:
- CRC: crc-WebSocketHandler.md - "Knows: connections"

---

### Test: Reject duplicate Peer() command

**Purpose**: Verify that WebSocketHandler rejects Peer() command if peer already created for connection.

**Motivation**: Security: prevents connection from creating multiple peers.

**Input**:
- WebSocket connection with peer already created
- Client sends Peer() command again

**References**:
- CRC: crc-WebSocketHandler.md - "Does: routeRequest"

**Expected Results**:
- Request rejected with error response
- Error message indicates peer already created
- Existing peer unchanged
- Connection remains active

**References**:
- CRC: crc-WebSocketHandler.md - "Knows: connections"

---

### Test: Route listFiles to PeerManager with connection's peerID

**Purpose**: Verify that WebSocketHandler routes listFiles request with connection's peerID for ownership context.

**Motivation**: File operations need peer context even though client doesn't specify it.

**Input**:
- Connection with peerID "12D3KooWABC..."
- Client sends: `{"method": "listFiles", "params": {"peerID": "12D3KooWXYZ..."}}`

**References**:
- CRC: crc-WebSocketHandler.md - "Does: routeFileOperations"
- Sequence: seq-list-files.md

**Expected Results**:
- Handler extracts connection's peerID ("12D3KooWABC...")
- Handler calls PeerManager with both peerIDs:
  - Connection's peerID: "12D3KooWABC..." (for ownership context)
  - Target peerID: "12D3KooWXYZ..." (for file listing)
- PeerManager returns file list
- Response sent to client

**References**:
- CRC: crc-WebSocketHandler.md - "Does: routeFileOperations"

---

### Test: Route getFile to PeerManager with connection's peerID

**Purpose**: Verify that WebSocketHandler routes getFile request with ownership context.

**Motivation**: File retrieval may use connection's peer for fallback.

**Input**:
- Connection with peerID "12D3KooWABC..."
- Client sends: `{"method": "getFile", "params": {"cid": "Qm123...", "fallbackPeerID": "12D3KooWXYZ..."}}`

**References**:
- CRC: crc-WebSocketHandler.md - "Does: routeFileOperations"

**Expected Results**:
- Handler routes to PeerManager with connection's peerID
- PeerManager retrieves file (local or from fallback)
- Response sent with file content

**References**:
- CRC: crc-WebSocketHandler.md - "Does: routeFileOperations"

---

### Test: Enforce file ownership on storeFile

**Purpose**: Verify that WebSocketHandler enforces that storeFile operates only on connection's own peer.

**Motivation**: Security: prevent clients from modifying other peers' directories.

**Input**:
- Connection with peerID "12D3KooWABC..."
- Client sends: `{"method": "storeFile", "params": {"path": "/test.txt", "content": "data"}}`
- (Note: no peerID in params, implicit ownership)

**References**:
- CRC: crc-WebSocketHandler.md - "Does: enforceFileOwnership"
- Sequence: seq-store-file.md

**Expected Results**:
- Handler extracts connection's peerID "12D3KooWABC..."
- Handler calls PeerManager.GetPeer("12D3KooWABC...")
- Handler calls peer.StoreFile() directly on connection's peer
- File stored successfully
- Response sent with fileCid and rootCid

**References**:
- CRC: crc-WebSocketHandler.md - "Does: enforceFileOwnership"

---

### Test: Enforce file ownership on removeFile

**Purpose**: Verify that WebSocketHandler enforces that removeFile operates only on connection's own peer.

**Motivation**: Security: prevent clients from deleting other peers' files.

**Input**:
- Connection with peerID "12D3KooWABC..."
- Client sends: `{"method": "removeFile", "params": {"path": "/test.txt"}}`

**References**:
- CRC: crc-WebSocketHandler.md - "Does: enforceFileOwnership"

**Expected Results**:
- Handler extracts connection's peerID
- Handler ensures operation only affects connection's peer
- File removed from connection's peer directory
- No other peers affected

**References**:
- CRC: crc-WebSocketHandler.md - "Does: enforceFileOwnership"

---

### Test: Queue server-initiated messages for sequential processing

**Purpose**: Verify that WebSocketHandler queues server-initiated messages (peerData, topicData, etc.) for sequential delivery.

**Motivation**: Prevents message reordering and race conditions in client.

**Input**:
- Active WebSocket connection
- Multiple rapid server-initiated messages:
  - peerData from peer A
  - topicData from topic "chat"
  - peerChange for topic "chat"

**References**:
- CRC: crc-WebSocketHandler.md - "Does: queueServerMessage"
- CRC: crc-WebSocketHandler.md - "Knows: messageQueue"

**Expected Results**:
- Messages queued in order received
- Messages sent to client sequentially (FIFO)
- Each message fully sent before next begins
- Client receives messages in correct order
- No message loss or reordering

**References**:
- CRC: crc-WebSocketHandler.md - "Knows: messageQueue"

---

### Test: Send server-initiated peerData notification

**Purpose**: Verify that WebSocketHandler sends peerData notifications from protocol messages.

**Motivation**: Enables asynchronous protocol message delivery to client.

**Input**:
- Connection with peer A
- Peer B sends protocol message to peer A
- PeerManager callback: onPeerData(peerB_ID, "/app/chat", "Hello")

**References**:
- CRC: crc-WebSocketHandler.md - "Does: queueServerMessage"

**Expected Results**:
- WebSocketHandler queues peerData message
- Message sent to client: `{"method": "peerData", "params": {peerID: "12D3...", protocol: "/app/chat", data: "Hello"}}`
- Client application receives notification

**References**:
- CRC: crc-WebSocketHandler.md - "Does: queueServerMessage"

---

### Test: Send server-initiated topicData notification

**Purpose**: Verify that WebSocketHandler sends topicData notifications from pub/sub messages.

**Motivation**: Enables asynchronous topic message delivery to client.

**Input**:
- Connection with peer A subscribed to "general-chat"
- Peer B publishes to "general-chat": "Hello everyone"
- PeerManager callback: onTopicData(peerB_ID, "general-chat", "Hello everyone")

**References**:
- CRC: crc-WebSocketHandler.md - "Does: queueServerMessage"

**Expected Results**:
- WebSocketHandler queues topicData message
- Message sent to client: `{"method": "topicData", "params": {peerID: "12D3...", topic: "general-chat", data: "Hello everyone"}}`
- Client application receives notification

**References**:
- CRC: crc-WebSocketHandler.md - "Does: queueServerMessage"

---

### Test: Send server-initiated peerChange notification

**Purpose**: Verify that WebSocketHandler sends peerChange notifications for topic join/leave events.

**Motivation**: Enables presence awareness in client applications.

**Input**:
- Connection with peer A monitoring topic "general-chat"
- Peer B joins topic
- PeerManager callback: onPeerChange("general-chat", peerB_ID, joined=true)

**References**:
- CRC: crc-WebSocketHandler.md - "Does: queueServerMessage"

**Expected Results**:
- WebSocketHandler queues peerChange message
- Message sent: `{"method": "peerChange", "params": {topic: "general-chat", peerID: "12D3...", joined: true}}`
- Client receives presence notification

**References**:
- CRC: crc-WebSocketHandler.md - "Does: queueServerMessage"

---

### Test: Send server-initiated peerFiles notification

**Purpose**: Verify that WebSocketHandler sends peerFiles notifications from listFiles responses.

**Motivation**: Enables asynchronous file list delivery (remote peer response).

**Input**:
- Connection with peer A
- Peer A requests listFiles(peerB_ID)
- Peer B responds with file list
- PeerManager callback: onPeerFiles(peerB_ID, "Qm123...", [{name: "test.txt", ...}])

**References**:
- CRC: crc-WebSocketHandler.md - "Does: queueServerMessage"

**Expected Results**:
- WebSocketHandler queues peerFiles message
- Message sent: `{"method": "peerFiles", "params": {peerID: "12D3...", rootCID: "Qm123...", entries: [...]}}`
- Client resolves pending listFiles promise

**References**:
- CRC: crc-WebSocketHandler.md - "Does: queueServerMessage"

---

### Test: Send server-initiated gotFile notification

**Purpose**: Verify that WebSocketHandler sends gotFile notifications from getFile responses.

**Motivation**: Enables asynchronous file content delivery.

**Input**:
- Connection with peer A
- Peer A requests getFile(cid, fallbackPeerID)
- File retrieved successfully
- PeerManager callback: onGotFile(success=true, content="file data")

**References**:
- CRC: crc-WebSocketHandler.md - "Does: queueServerMessage"

**Expected Results**:
- WebSocketHandler queues gotFile message
- Message sent: `{"method": "gotFile", "params": {success: true, content: "file data"}}`
- Client resolves pending getFile promise

**References**:
- CRC: crc-WebSocketHandler.md - "Does: queueServerMessage"

---

### Test: Close connection and cleanup

**Purpose**: Verify that WebSocketHandler properly cleans up connection and associated peer.

**Motivation**: Prevents resource leaks and ensures proper lifecycle management.

**Input**:
- Active WebSocket connection with associated peer
- Connection closed (browser closes tab, network error, etc.)

**References**:
- CRC: crc-WebSocketHandler.md - "Does: closeConnection"

**Expected Results**:
- Connection removed from connections map
- Associated peer removed via PeerManager
- WebSocket connection closed cleanly
- Server notified of connection lifecycle event
- No resource leaks

**References**:
- CRC: crc-WebSocketHandler.md - "Knows: connections"
- CRC: crc-WebSocketHandler.md - "Collaborators: Server"

---

### Test: Handle malformed JSON-RPC message

**Purpose**: Verify that WebSocketHandler handles malformed JSON gracefully.

**Motivation**: Robustness against client errors or attacks.

**Input**:
- Active connection
- Client sends invalid JSON: `{invalid json}`

**References**:
- CRC: crc-WebSocketHandler.md - "Does: receiveMessage"

**Expected Results**:
- Parsing error caught
- Error response sent to client
- Connection remains active
- No crash or panic

**References**:
- CRC: crc-WebSocketHandler.md - "Does: receiveMessage"

---

### Test: Increment request ID counter for server messages

**Purpose**: Verify that WebSocketHandler increments requestID counter for server-initiated messages.

**Motivation**: Enables correlation of requests/responses if needed.

**Input**:
- Active connection
- Multiple server-initiated messages sent

**References**:
- CRC: crc-WebSocketHandler.md - "Knows: requestID"

**Expected Results**:
- First message: requestID = 1
- Second message: requestID = 2
- Third message: requestID = 3
- Counter increments monotonically

**References**:
- CRC: crc-WebSocketHandler.md - "Knows: requestID"

## Coverage Summary

**Responsibilities Covered**:
- ✅ acceptConnection - Connection establishment test
- ✅ receiveMessage - JSON-RPC parsing and malformed message tests
- ✅ sendMessage - JSON-RPC response formatting test
- ✅ routeRequest - Peer() routing and duplicate rejection tests
- ✅ routeFileOperations - listFiles and getFile routing tests
- ✅ enforceFileOwnership - storeFile and removeFile ownership tests
- ✅ queueServerMessage - Sequential message delivery test
- ✅ closeConnection - Connection cleanup test
- ✅ All server-initiated notifications - peerData, topicData, peerChange, peerFiles, gotFile tests
- ✅ requestID counter - Incremental ID test

**Scenarios Covered**:
- ✅ seq-peer-creation.md - Peer() command and duplicate rejection tests
- ✅ seq-list-files.md - File list routing test
- ✅ seq-store-file.md - File storage ownership enforcement test

**Gaps**:
- WebSocket reconnection and recovery not tested
- Concurrent client requests not tested
- Large message handling not tested
- Rate limiting not tested (may not be implemented)
