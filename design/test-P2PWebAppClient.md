# Test Design: P2PWebAppClient

**Source Specs**: main.md
**CRC Cards**: crc-P2PWebAppClient.md
**Sequences**: seq-add-peers.md, seq-remove-peers.md, seq-list-files.md, seq-get-file.md, seq-store-file.md

## Overview

Test suite for P2PWebAppClient (TypeScript client library) covering WebSocket connection, peer initialization, protocol messaging, pub/sub, file operations, and delivery confirmations.

## Test Cases

### Test: Connect to server and initialize peer

**Purpose**: Verify that client connects to server via WebSocket and initializes peer.

**Motivation**: Foundation for all client operations.

**Input**:
- Server running on ws://localhost:10000
- Client calls connect()

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: connect"

**Expected Results**:
- WebSocket connection established
- Peer() request sent automatically
- Response received with peerID and peerKey
- Client stores peerID and peerKey
- Client ready for operations

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: websocket, peerID, peerKey"

---

### Test: Connect with existing peer key

**Purpose**: Verify that client can restore peer identity using saved peer key.

**Motivation**: Enables persistent peer identity across sessions.

**Input**:
- Server running
- Saved peer key from previous session
- Client calls connect(existingPeerKey)

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: connect"

**Expected Results**:
- Peer() request includes existing peer key
- Server creates peer with provided key
- Same peerID returned as previous session
- Client can access previous peer's file storage

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: peerKey"

---

### Test: Start protocol listener

**Purpose**: Verify that client can register listener to receive protocol messages.

**Motivation**: Enables bidirectional protocol communication.

**Input**:
- Connected client
- Protocol listener: (peerID, data) => { handleMessage(peerID, data) }
- Client calls start("/app/chat/1.0.0", listener)

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: start"

**Expected Results**:
- Listener registered in protocolListeners map for "/app/chat/1.0.0"
- Server-initiated peerData messages routed to listener
- Listener receives (peerID, data) when message arrives

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: protocolListeners"
- CRC: crc-P2PWebAppClient.md - "Does: routePeerData"

---

### Test: Send protocol message to peer

**Purpose**: Verify that client can send protocol messages to remote peers.

**Motivation**: Core peer-to-peer messaging functionality.

**Input**:
- Connected client with peerID
- Remote peer ID: "12D3KooWRemote..."
- Client calls send("12D3KooWRemote...", "/app/chat/1.0.0", "Hello")

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: send"

**Expected Results**:
- JSON-RPC request sent to server:
  - method: "send"
  - params: {peerID: "12D3...", protocol: "/app/chat/1.0.0", data: "Hello"}
- Server routes message to remote peer
- Response received (success)

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: sendRequest"

---

### Test: Send protocol message with delivery confirmation

**Purpose**: Verify that client can request delivery confirmation callback.

**Motivation**: Enables reliable messaging with acknowledgment.

**Input**:
- Connected client
- Callback: (success) => { handleAck(success) }
- Client calls send(peerID, protocol, data, callback)

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: send"
- CRC: crc-P2PWebAppClient.md - "Knows: ackCallbacks, nextAckNumber"

**Expected Results**:
- Ack number assigned (auto-incrementing from 0)
- Callback stored in ackCallbacks map
- Request includes ack number
- When server responds with ack message:
  - routeAck() invoked
  - Callback called with success status
  - Callback removed from map

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: routeAck"

---

### Test: Receive protocol message

**Purpose**: Verify that client receives incoming protocol messages from server.

**Motivation**: Completes bidirectional protocol communication.

**Input**:
- Connected client with protocol listener registered
- Server sends peerData notification: {method: "peerData", params: {peerID: "12D3...", protocol: "/app/chat", data: "Hello"}}

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: handleServerMessage"
- CRC: crc-P2PWebAppClient.md - "Does: routePeerData"

**Expected Results**:
- Message queued in messageQueue
- Message processed sequentially
- routePeerData() invoked
- Protocol listener callback receives (peerID, data)

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: messageQueue, protocolListeners"

---

### Test: Subscribe to topic

**Purpose**: Verify that client can subscribe to GossipSub topic and receive messages.

**Motivation**: Foundation for pub/sub messaging.

**Input**:
- Connected client
- Topic listener: (peerID, data) => { handleMessage(peerID, data) }
- Client calls subscribe("general-chat", listener)

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: subscribe"

**Expected Results**:
- JSON-RPC request sent: {method: "subscribe", params: {topic: "general-chat"}}
- Listener registered in topicListeners map
- Response received (success)
- Client ready to receive topic messages

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: topicListeners"

---

### Test: Publish message to topic

**Purpose**: Verify that client can publish messages to subscribed topics.

**Motivation**: Core pub/sub functionality.

**Input**:
- Client subscribed to "general-chat"
- Client calls publish("general-chat", "Hello everyone")

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: publish"

**Expected Results**:
- JSON-RPC request sent: {method: "publish", params: {topic: "general-chat", data: "Hello everyone"}}
- Message published to topic on server
- Response received (success)

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: sendRequest"

---

### Test: Receive topic message

**Purpose**: Verify that client receives topic messages from other peers.

**Motivation**: Completes pub/sub messaging.

**Input**:
- Client subscribed to "general-chat" with listener
- Server sends topicData notification: {method: "topicData", params: {peerID: "12D3...", topic: "general-chat", data: "Hello"}}

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: handleServerMessage"
- CRC: crc-P2PWebAppClient.md - "Does: routeTopicData"

**Expected Results**:
- Message queued and processed
- routeTopicData() invoked
- Topic listener callback receives (peerID, data)

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: topicListeners"

---

### Test: Receive peer change notification

**Purpose**: Verify that client receives notifications when peers join/leave topics.

**Motivation**: Enables presence awareness in applications.

**Input**:
- Client subscribed to "general-chat" with listener that handles peer changes
- Server sends peerChange notification: {method: "peerChange", params: {topic: "general-chat", peerID: "12D3...", joined: true}}

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: handleServerMessage"
- CRC: crc-P2PWebAppClient.md - "Does: routePeerChange"

**Expected Results**:
- Message queued and processed
- routePeerChange() invoked
- Topic listener callback receives peer change event
- Application can update presence UI

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: routePeerChange"

---

### Test: Unsubscribe from topic

**Purpose**: Verify that client can unsubscribe from topics.

**Motivation**: Allows dynamic topic management.

**Input**:
- Client subscribed to "general-chat"
- Client calls unsubscribe("general-chat")

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: unsubscribe"

**Expected Results**:
- JSON-RPC request sent: {method: "unsubscribe", params: {topic: "general-chat"}}
- Listener removed from topicListeners map
- Response received (success)
- Client no longer receives messages on topic

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: topicListeners"

---

### Test: List peers on topic

**Purpose**: Verify that client can query which peers are subscribed to a topic.

**Motivation**: Enables presence awareness.

**Input**:
- Client subscribed to "general-chat"
- Other peers subscribed to same topic
- Client calls listPeers("general-chat")

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: listPeers"

**Expected Results**:
- JSON-RPC request sent
- Response contains array of peer IDs
- Promise resolved with peer list

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: sendRequest"

---

### Test: Add peers to connection protection

**Purpose**: Verify that client can protect peer connections to prevent pruning.

**Motivation**: Ensures important connections persist (friends, collaborators).

**Input**:
- Connected client
- List of peer IDs to protect: ["12D3KooWFriend1...", "12D3KooWFriend2..."]
- Client calls addPeers(peerIDs)

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: addPeers"
- Sequence: seq-add-peers.md

**Expected Results**:
- JSON-RPC request sent: {method: "addPeers", params: {peerIDs: [...]}}
- Server protects and tags connections
- Response received (success)
- Protected connections persist

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: sendRequest"

---

### Test: Remove peers from connection protection

**Purpose**: Verify that client can unprotect peer connections.

**Motivation**: Allows connection manager to prune unneeded connections.

**Input**:
- Client with protected peer connections
- Client calls removePeers(["12D3KooWFriend1..."])

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: removePeers"
- Sequence: seq-remove-peers.md

**Expected Results**:
- JSON-RPC request sent: {method: "removePeers", params: {peerIDs: [...]}}
- Server unprotects and untags connections
- Response received (success)

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: sendRequest"

---

### Test: List files from local peer

**Purpose**: Verify that client can list its own peer's files.

**Motivation**: Enables file browsing UI.

**Input**:
- Connected client with files in storage
- Client calls listFiles(client.peerID)

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: listFiles"
- Sequence: seq-list-files.md

**Expected Results**:
- JSON-RPC request sent
- Immediate response with rootCID and entries array
- Promise resolved with file list
- Handler stored in fileListHandlers (for consistency with remote case)

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: fileListHandlers"

---

### Test: List files from remote peer

**Purpose**: Verify that client can request file list from remote peer.

**Motivation**: Enables browsing remote peer files.

**Input**:
- Connected client
- Remote peer ID: "12D3KooWRemote..."
- Client calls listFiles("12D3KooWRemote...")

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: listFiles"
- Sequence: seq-list-files.md

**Expected Results**:
- JSON-RPC request sent
- Promise created and handler stored in fileListHandlers
- Server sends peerFiles notification asynchronously
- routePeerFiles() invoked
- Handler retrieves promise resolver from fileListHandlers
- Promise resolved with (rootCID, entries)
- Handler removed from map (deduplication)

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: fileListHandlers"
- CRC: crc-P2PWebAppClient.md - "Does: routePeerFiles"

---

### Test: List files deduplication

**Purpose**: Verify that client prevents duplicate listFiles requests for same peer.

**Motivation**: Efficiency: reuse pending request instead of creating duplicates.

**Input**:
- Client calls listFiles("12D3KooWRemote...") twice rapidly

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: listFiles"

**Expected Results**:
- First call: Request sent, handler stored
- Second call: No new request sent, returns same pending promise
- When response arrives, both callers receive same result

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: fileListHandlers"

---

### Test: Store file with string content

**Purpose**: Verify that client can store text files in peer directory.

**Motivation**: Core file storage functionality for text data.

**Input**:
- Connected client
- Client calls storeFile("/readme.txt", "Hello World")

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: storeFile"
- Sequence: seq-store-file.md

**Expected Results**:
- JSON-RPC request sent: {method: "storeFile", params: {path: "/readme.txt", content: "Hello World"}}
- Server stores file in client's peer directory
- Response received: {fileCid: "Qm123...", rootCid: "Qm456..."}
- Promise resolved with StoreFileResponse

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: sendRequest"

---

### Test: Store file with binary content

**Purpose**: Verify that client can store binary files using Uint8Array.

**Motivation**: Enables storing images, videos, arbitrary binary data.

**Input**:
- Connected client
- Binary data: new Uint8Array([0x89, 0x50, 0x4E, 0x47, ...])
- Client calls storeFile("/image.png", binaryData)

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: storeFile"

**Expected Results**:
- Binary data serialized appropriately for JSON-RPC
- Request sent successfully
- Server stores binary content
- Response received with CIDs

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: storeFile"

---

### Test: Create directory

**Purpose**: Verify that client can create directories in peer storage.

**Motivation**: Enables hierarchical file organization.

**Input**:
- Connected client
- Client calls createDirectory("/docs")

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: createDirectory"

**Expected Results**:
- JSON-RPC request sent with directory=true flag
- Server creates directory node
- Response received with fileCid and rootCid
- Promise resolved

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: sendRequest"

---

### Test: Remove file

**Purpose**: Verify that client can remove files from peer directory.

**Motivation**: Enables file management.

**Input**:
- Connected client with file "/test.txt" in storage
- Client calls removeFile("/test.txt")

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: removeFile"

**Expected Results**:
- JSON-RPC request sent: {method: "removeFile", params: {path: "/test.txt"}}
- Server removes file from client's peer directory
- Response received (success)
- Promise resolved

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: sendRequest"

---

### Test: Get file by CID (local)

**Purpose**: Verify that client can retrieve file content by CID.

**Motivation**: Core file retrieval functionality.

**Input**:
- Connected client
- CID of stored file: "Qm123..."
- Client calls getFile("Qm123...")

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: getFile"
- Sequence: seq-get-file.md

**Expected Results**:
- JSON-RPC request sent
- Server retrieves content locally
- Server sends gotFile notification: {success: true, content: "file data"}
- routeGotFile() invoked
- Callback receives (success, content)

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: routeGotFile"

---

### Test: Get file with fallback peer

**Purpose**: Verify that client can request file from fallback peer when not available locally.

**Motivation**: Enables collaborative file sharing.

**Input**:
- Connected client
- CID not available locally: "Qm123..."
- Fallback peer ID: "12D3KooWRemote..."
- Client calls getFile("Qm123...", "12D3KooWRemote...")

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: getFile"
- Sequence: seq-get-file.md

**Expected Results**:
- Request includes fallback peer ID
- Server attempts local retrieval, fails
- Server requests from fallback peer
- Server caches content locally
- gotFile notification sent with content
- Callback receives (success=true, content)

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: getFile"

---

### Test: Sequential server message processing

**Purpose**: Verify that client processes server-initiated messages sequentially to prevent reordering.

**Motivation**: Ensures message order integrity for application logic.

**Input**:
- Connected client
- Multiple rapid server notifications:
  - peerData message 1
  - topicData message 2
  - peerChange message 3

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: handleServerMessage"
- CRC: crc-P2PWebAppClient.md - "Knows: messageQueue"

**Expected Results**:
- All messages queued in order
- Message 1 processed completely before message 2
- Message 2 processed completely before message 3
- Callbacks invoked in correct order
- No race conditions or reordering

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: messageQueue"

---

### Test: Request ID increments for client requests

**Purpose**: Verify that client assigns unique incrementing request IDs.

**Motivation**: Enables correlation of requests and responses.

**Input**:
- Connected client
- Send three requests sequentially

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: requestID"

**Expected Results**:
- First request: id = 1 (initial Peer() request is id 0)
- Second request: id = 2
- Third request: id = 3
- Each response correlates to correct request

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: requestID, pendingRequests"

---

### Test: Stop protocol listener

**Purpose**: Verify that client can remove protocol listener.

**Motivation**: Allows dynamic protocol management.

**Input**:
- Client with protocol listener registered for "/app/chat"
- Client calls stop("/app/chat")

**References**:
- CRC: crc-P2PWebAppClient.md - "Does: stop"

**Expected Results**:
- Listener removed from protocolListeners map
- Incoming peerData messages for protocol no longer routed to callback
- No errors

**References**:
- CRC: crc-P2PWebAppClient.md - "Knows: protocolListeners"

## Coverage Summary

**Responsibilities Covered**:
- ✅ connect - Connection and peer initialization tests (fresh and existing key)
- ✅ start - Protocol listener registration test
- ✅ stop - Protocol listener removal test
- ✅ send - Protocol message sending with and without ack callback
- ✅ subscribe - Topic subscription test
- ✅ publish - Topic publishing test
- ✅ unsubscribe - Topic unsubscription test
- ✅ listPeers - Topic peer listing test
- ✅ addPeers - Connection protection test
- ✅ removePeers - Connection unprotection test
- ✅ listFiles - Local and remote file listing with deduplication
- ✅ getFile - File retrieval with fallback peer test
- ✅ storeFile - Text and binary file storage tests
- ✅ createDirectory - Directory creation test
- ✅ removeFile - File removal test
- ✅ sendRequest - Implicit in all request tests
- ✅ handleResponse - Implicit in all request tests
- ✅ handleServerMessage - Sequential message processing test
- ✅ routePeerData - Protocol message routing test
- ✅ routeTopicData - Topic message routing test
- ✅ routePeerChange - Peer change routing test
- ✅ routePeerFiles - File list response routing test
- ✅ routeGotFile - File content routing test
- ✅ routeAck - Delivery confirmation routing test

**Scenarios Covered**:
- ✅ seq-add-peers.md - Connection protection test
- ✅ seq-remove-peers.md - Connection unprotection test
- ✅ seq-list-files.md - File listing and deduplication tests
- ✅ seq-get-file.md - File retrieval with fallback test
- ✅ seq-store-file.md - File storage tests

**Gaps**:
- WebSocket reconnection and recovery not tested
- Error handling for network failures not tested
- Large file transfers and performance not tested
- Memory management for message queue not tested
