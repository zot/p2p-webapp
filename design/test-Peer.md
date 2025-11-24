# Test Design: Peer

**Source Specs**: main.md
**CRC Cards**: crc-Peer.md
**Sequences**: seq-protocol-communication.md, seq-pubsub-communication.md, seq-add-peers.md, seq-remove-peers.md, seq-list-files.md, seq-get-file.md, seq-store-file.md

## Overview

Test suite for Peer component covering protocol messaging, pub/sub communication, file operations, connection management, and stream lifecycle.

## Test Cases

### Test: Send data to peer on protocol

**Purpose**: Verify that Peer can send protocol messages to another peer, creating or reusing streams.

**Motivation**: Core peer-to-peer messaging functionality. Enables direct communication between peers.

**Input**:
- Peer A initialized
- Peer B initialized and discoverable
- Peer A calls sendToPeer(peerB_ID, "/app/chat/1.0.0", "Hello")

**References**:
- CRC: crc-Peer.md - "Does: sendToPeer"
- Sequence: seq-protocol-communication.md

**Expected Results**:
- Peer A creates or reuses stream to Peer B on protocol "/app/chat/1.0.0"
- Data "Hello" sent to Peer B
- Peer B's protocol handler receives (peerA_ID, "Hello")
- Stream managed by VirtualConnectionManager for reliability

**References**:
- CRC: crc-Peer.md - "Knows: protocols, vcm"
- CRC: crc-Peer.md - "Collaborators: VirtualConnectionManager"

---

### Test: Receive protocol message

**Purpose**: Verify that Peer receives and routes incoming protocol messages to callback.

**Motivation**: Completes bidirectional protocol messaging.

**Input**:
- Peer A with protocol listener registered for "/app/chat/1.0.0"
- Peer B sends message "Hello" on protocol to Peer A

**References**:
- CRC: crc-Peer.md - "Does: start"
- Sequence: seq-protocol-communication.md

**Expected Results**:
- Peer A's protocol handler invoked
- Callback receives (peerB_ID, "/app/chat/1.0.0", "Hello")
- Callback invoked exactly once

**References**:
- CRC: crc-Peer.md - "Knows: protocols"
- CRC: crc-PeerManager.md - "Knows: onPeerData"

---

### Test: Subscribe to topic

**Purpose**: Verify that Peer can subscribe to GossipSub topic, advertise to DHT, discover peers, and wait for mesh formation.

**Motivation**: Foundation for pub/sub messaging with global peer discovery. Must handle mesh formation and DHT integration.

**Input**:
- Peer A initialized with DHT enabled
- Call subscribe("general-chat")

**References**:
- CRC: crc-Peer.md - "Does: subscribe"
- Sequence: seq-pubsub-communication.md

**Expected Results**:
- Peer A subscribes to topic "general-chat"
- Topic entry created in topics map
- advertiseTopic goroutine launched (starts DHT advertisement)
- discoverTopicPeers goroutine launched (queries DHT for peers)
- Waits for mesh formation before returning
- TopicHandler created for message routing
- Peer A visible to other peers via listPeers()

**References**:
- CRC: crc-Peer.md - "Knows: topics, pubsub, dht"
- CRC: crc-Peer.md - "Collaborators: GossipSub, DHT"

---

### Test: DHT topic advertisement

**Purpose**: Verify that Peer continuously advertises topic subscription to DHT.

**Motivation**: Enables global peer discovery beyond local networks.

**Input**:
- Peer A subscribed to "general-chat"
- DHT enabled
- Wait for initial advertisement and re-advertisement

**References**:
- CRC: crc-Peer.md - "Does: advertiseTopic"
- Sequence: seq-pubsub-communication.md

**Expected Results**:
- Initial advertisement succeeds with TTL returned
- Re-advertisement occurs at TTL/2 interval
- Advertisement stops when topic unsubscribed (handler.ctx.Done())
- No advertisement attempted if DHT is nil

**References**:
- CRC: crc-Peer.md - "Knows: dht"
- CRC: crc-Peer.md - "Collaborators: DHT"

---

### Test: DHT topic peer discovery

**Purpose**: Verify that Peer discovers other peers subscribed to same topic via DHT.

**Motivation**: Enables geographically distant peers to find each other.

**Input**:
- Peer A subscribed to "general-chat" with DHT advertisement
- Peer B subscribed to "general-chat" with DHT advertisement
- DHT network functioning

**References**:
- CRC: crc-Peer.md - "Does: discoverTopicPeers"
- Sequence: seq-pubsub-communication.md

**Expected Results**:
- Peer B discovers Peer A via DHT query (or vice versa)
- Discovered peer's addresses added to peerstore with TempAddrTTL
- Automatic connection attempt to discovered peer
- Connection succeeds (or fails gracefully if unreachable)
- Discovered peer excluded if it's self (skip own peer ID)

**References**:
- CRC: crc-Peer.md - "Knows: dht, host"
- CRC: crc-Peer.md - "Collaborators: DHT, libp2p Host"

---

### Test: Subscribe without DHT

**Purpose**: Verify that Peer handles subscription when DHT is disabled/nil.

**Motivation**: DHT is optional - must work without it.

**Input**:
- Peer A initialized without DHT (dht == nil)
- Call subscribe("general-chat")

**References**:
- CRC: crc-Peer.md - "Does: subscribe"

**Expected Results**:
- Peer A subscribes to topic "general-chat"
- No DHT advertisement attempted
- No DHT peer discovery attempted
- Waits for mesh formation via other mechanisms (mDNS, direct connections)
- TopicHandler created normally

**References**:
- CRC: crc-Peer.md - "Knows: topics, pubsub, dht"

---

### Test: Publish message to topic

**Purpose**: Verify that Peer can publish messages to subscribed topics.

**Motivation**: Core pub/sub functionality for broadcasting messages.

**Input**:
- Peer A subscribed to "general-chat"
- Peer B subscribed to "general-chat"
- Peer A calls publish("general-chat", "Hello everyone")

**References**:
- CRC: crc-Peer.md - "Does: publish"
- Sequence: seq-pubsub-communication.md

**Expected Results**:
- Message "Hello everyone" published to topic
- Peer B receives message via topic callback
- Callback receives (peerA_ID, "general-chat", "Hello everyone")
- Message delivered to all subscribed peers

**References**:
- CRC: crc-Peer.md - "Knows: pubsub"
- CRC: crc-PeerManager.md - "Knows: onTopicData"

---

### Test: Publish to unsubscribed topic

**Purpose**: Verify that Peer handles attempt to publish to topic it's not subscribed to.

**Motivation**: Edge case for application error. Must fail gracefully.

**Input**:
- Peer A not subscribed to any topics
- Call publish("general-chat", "Hello")

**References**:
- CRC: crc-Peer.md - "Does: publish"

**Expected Results**:
- Operation fails with error indicating not subscribed
- No message sent to network
- No crash

**References**:
- CRC: crc-Peer.md - "Knows: topics"

---

### Test: Unsubscribe from topic

**Purpose**: Verify that Peer can unsubscribe from topic and clean up resources.

**Motivation**: Allows dynamic topic management and prevents resource leaks.

**Input**:
- Peer A subscribed to "general-chat"
- Call unsubscribe("general-chat")

**References**:
- CRC: crc-Peer.md - "Does: unsubscribe"

**Expected Results**:
- Peer A unsubscribes from topic
- Topic entry removed from topics map
- Peer A no longer receives messages on topic
- Peer A no longer visible via listPeers()

**References**:
- CRC: crc-Peer.md - "Knows: topics"

---

### Test: List peers on topic

**Purpose**: Verify that Peer can query which peers are subscribed to a topic.

**Motivation**: Enables presence awareness in pub/sub applications.

**Input**:
- Peer A, B, C all subscribed to "general-chat"
- Peer A calls listPeers("general-chat")

**References**:
- CRC: crc-Peer.md - "Does: listPeers"

**Expected Results**:
- Returns list containing Peer B and Peer C IDs
- List may or may not include Peer A (depending on implementation)
- List accurate at time of query

**References**:
- CRC: crc-Peer.md - "Knows: pubsub"

---

### Test: Monitor topic for peer join/leave

**Purpose**: Verify that Peer can monitor topic for peer presence changes.

**Motivation**: Enables presence notifications for collaborative features.

**Input**:
- Peer A subscribed to "general-chat"
- Call monitor("general-chat")
- Peer B subscribes to "general-chat"
- Peer C subscribes then unsubscribes

**References**:
- CRC: crc-Peer.md - "Does: monitor"

**Expected Results**:
- Monitor entry created in monitoredTopics map
- Callback invoked when Peer B joins: (peerB_ID, "general-chat", joined=true)
- Callback invoked when Peer C joins: (peerC_ID, "general-chat", joined=true)
- Callback invoked when Peer C leaves: (peerC_ID, "general-chat", joined=false)

**References**:
- CRC: crc-Peer.md - "Knows: monitoredTopics"
- CRC: crc-PeerManager.md - "Knows: onPeerChange"

---

### Test: Stop monitoring topic

**Purpose**: Verify that Peer can stop monitoring topic for peer changes.

**Motivation**: Prevents callback overhead when presence not needed.

**Input**:
- Peer A monitoring "general-chat"
- Call stopMonitor("general-chat")
- Peer B joins topic

**References**:
- CRC: crc-Peer.md - "Does: stopMonitor"

**Expected Results**:
- Monitor entry removed from monitoredTopics map
- No callback invoked when Peer B joins
- Topic subscription remains active (can still publish/receive)

**References**:
- CRC: crc-Peer.md - "Knows: monitoredTopics"

---

### Test: Add peers to connection protection

**Purpose**: Verify that Peer protects and tags peer connections to prevent pruning.

**Motivation**: Ensures important connections (friends, collaborators) persist.

**Input**:
- Peer A initialized
- List of peer IDs: ["12D3KooWFriend1...", "12D3KooWFriend2..."]
- Call addPeers(peerIDs)

**References**:
- CRC: crc-Peer.md - "Does: addPeers"
- Sequence: seq-add-peers.md

**Expected Results**:
- For each peer ID:
  - ConnManager().Protect(peerID, "connected") called
  - TagPeer(peerID, "connected", 100) called
  - Connection attempted if not already connected
- Protected connections persist even under connection pressure
- Tagged connections prioritized by connection manager

**References**:
- CRC: crc-Peer.md - "Knows: host"
- CRC: crc-Peer.md - "Collaborators: libp2p Host"

---

### Test: Remove peers from connection protection

**Purpose**: Verify that Peer unprotects and untags peer connections.

**Motivation**: Allows connection manager to prune connections when no longer needed.

**Input**:
- Peer A with protected connections to Friend1 and Friend2
- Call removePeers(["12D3KooWFriend1..."])

**References**:
- CRC: crc-Peer.md - "Does: removePeers"
- Sequence: seq-remove-peers.md

**Expected Results**:
- ConnManager().Unprotect(peerID, "connected") called for Friend1
- UntagPeer(peerID, "connected") called for Friend1
- Friend1 connection eligible for pruning
- Friend2 remains protected

**References**:
- CRC: crc-Peer.md - "Knows: host"

---

### Test: List local files

**Purpose**: Verify that Peer can list its own directory contents.

**Motivation**: Enables file browsing for local peer.

**Input**:
- Peer A with files in directory:
  - /readme.txt (CID: Qm123...)
  - /docs/guide.md (CID: Qm456...)
- Call listFiles(peerA_ID)

**References**:
- CRC: crc-Peer.md - "Does: listFiles"
- Sequence: seq-list-files.md

**Expected Results**:
- Returns root directory CID
- Returns entries array:
  - {name: "readme.txt", cid: "Qm123...", size: 1234, directory: false}
  - {name: "docs/guide.md", cid: "Qm456...", size: 5678, directory: false}
- Immediate response (no network request)

**References**:
- CRC: crc-Peer.md - "Knows: directory, directoryCID"

---

### Test: List remote peer files

**Purpose**: Verify that Peer can request file list from remote peer via p2p-webapp protocol.

**Motivation**: Enables file browsing on remote peers.

**Input**:
- Peer A initialized
- Peer B initialized with files in directory
- Peer A calls listFiles(peerB_ID)

**References**:
- CRC: crc-Peer.md - "Does: listFiles"
- CRC: crc-Peer.md - "Does: handleGetFileList"
- Sequence: seq-list-files.md

**Expected Results**:
- Peer A sends getFileList() message (type 0) to Peer B on p2p-webapp protocol
- Peer B receives message, invokes handleGetFileList()
- Peer B responds with fileList() message (type 1) containing root CID and entries
- Peer A receives response via handleFileList()
- Callback invoked: onPeerFiles(peerB_ID, rootCID, entries)

**References**:
- CRC: crc-Peer.md - "Does: listFiles, handleGetFileList, handleFileList"
- CRC: crc-PeerManager.md - "Knows: onPeerFiles"

---

### Test: Store file in directory

**Purpose**: Verify that Peer can store file content in its HAMTDirectory.

**Motivation**: Core file storage functionality.

**Input**:
- Peer A with empty directory
- Call storeFile("/readme.txt", "Hello World", directory=false)

**References**:
- CRC: crc-Peer.md - "Does: storeFile"
- Sequence: seq-store-file.md

**Expected Results**:
- File node created in IPFS with content "Hello World"
- Entry added to HAMTDirectory at path "/readme.txt"
- directoryCID updated with new root CID
- Returns StoreFileResponse with:
  - fileCid: CID of file content
  - rootCid: Updated directory CID
- If fileUpdateNotifyTopic configured and peer subscribed:
  - Publishes notification: {"type":"p2p-webapp-file-update","peer":"<peerID>"}

**References**:
- CRC: crc-Peer.md - "Knows: directory, directoryCID"
- CRC: crc-Peer.md - "Does: publishFileUpdateNotification"

---

### Test: Create directory in peer storage

**Purpose**: Verify that Peer can create directory nodes in HAMTDirectory.

**Motivation**: Enables hierarchical file organization.

**Input**:
- Peer A with directory
- Call storeFile("/docs", nil, directory=true)

**References**:
- CRC: crc-Peer.md - "Does: storeFile"
- Sequence: seq-store-file.md

**Expected Results**:
- Directory node created in HAMTDirectory at "/docs"
- No content stored (directory=true flag)
- directoryCID updated
- Returns StoreFileResponse with fileCid and rootCid
- File update notification published if configured

**References**:
- CRC: crc-Peer.md - "Knows: directory, directoryCID"

---

### Test: Remove file from directory

**Purpose**: Verify that Peer can remove files from HAMTDirectory.

**Motivation**: Enables file management and cleanup.

**Input**:
- Peer A with file "/readme.txt" in directory
- Call removeFile("/readme.txt")

**References**:
- CRC: crc-Peer.md - "Does: removeFile"

**Expected Results**:
- Entry removed from HAMTDirectory at path "/readme.txt"
- directoryCID updated to new root CID
- File content remains in IPFS (unpinned but not deleted)
- File update notification published if configured

**References**:
- CRC: crc-Peer.md - "Knows: directory, directoryCID"

---

### Test: Get file by CID (local)

**Purpose**: Verify that Peer can retrieve IPFS content by CID when available locally.

**Motivation**: Core file retrieval from local IPFS storage.

**Input**:
- Peer A with file stored, CID: Qm123...
- Call getFile(Qm123..., nil)

**References**:
- CRC: crc-Peer.md - "Does: getFile"
- Sequence: seq-get-file.md

**Expected Results**:
- Content retrieved from local IPFS storage
- Callback invoked: onGotFile(success=true, content="file content")
- No network request made

**References**:
- CRC: crc-Peer.md - "Collaborators: ipfs-lite Peer"
- CRC: crc-PeerManager.md - "Knows: onGotFile"

---

### Test: Get file by CID with fallback peer

**Purpose**: Verify that Peer can request file from fallback peer when not available locally.

**Motivation**: Enables file retrieval from other peers when content not cached.

**Input**:
- Peer A with empty IPFS storage
- Peer B has file with CID: Qm123...
- Peer A calls getFile(Qm123..., peerB_ID)

**References**:
- CRC: crc-Peer.md - "Does: getFile, handleGetFile, handleFileContent"
- Sequence: seq-get-file.md

**Expected Results**:
- Peer A attempts local retrieval, fails
- Peer A sends getFile() message (type 2) to Peer B on p2p-webapp protocol
- Peer B receives message, invokes handleGetFile()
- Peer B retrieves content and sends fileContent() message (type 3)
- Peer A receives message, invokes handleFileContent()
- Peer A stores content in local IPFS (caching)
- Callback invoked: onGotFile(success=true, content="file content")

**References**:
- CRC: crc-Peer.md - "Does: getFile, handleGetFile, handleFileContent"

---

### Test: Get file with fallback peer not having content

**Purpose**: Verify that Peer handles case when fallback peer doesn't have requested content.

**Motivation**: Edge case for incomplete file distribution.

**Input**:
- Peer A with empty IPFS storage
- Peer B also doesn't have CID: Qm123...
- Peer A calls getFile(Qm123..., peerB_ID)

**References**:
- CRC: crc-Peer.md - "Does: getFile, handleGetFile"

**Expected Results**:
- Peer A attempts local retrieval, fails
- Peer A requests from Peer B
- Peer B attempts retrieval, fails
- Peer B sends error response or empty content
- Callback invoked: onGotFile(success=false, content=nil)

**References**:
- CRC: crc-Peer.md - "Does: getFile"

---

### Test: File update notification when subscribed

**Purpose**: Verify that Peer publishes file update notifications when configured and subscribed.

**Motivation**: Enables automatic UI refresh on file changes.

**Input**:
- Peer A with fileUpdateNotifyTopic = "file-updates"
- Peer A subscribed to "file-updates" topic
- Peer A calls storeFile("/test.txt", "content", false)

**References**:
- CRC: crc-Peer.md - "Does: publishFileUpdateNotification"

**Expected Results**:
- After successful storeFile, notification published to "file-updates" topic
- Message: {"type":"p2p-webapp-file-update","peer":"<peerA_ID>"}
- Other peers subscribed to topic receive notification

**References**:
- CRC: crc-PeerManager.md - "Knows: fileUpdateNotifyTopic"

---

### Test: No file update notification when not subscribed

**Purpose**: Verify that Peer doesn't publish notifications when not subscribed to notify topic.

**Motivation**: Privacy-friendly: only publish when explicitly opted in.

**Input**:
- Peer A with fileUpdateNotifyTopic = "file-updates"
- Peer A NOT subscribed to "file-updates" topic
- Peer A calls storeFile("/test.txt", "content", false)

**References**:
- CRC: crc-Peer.md - "Does: publishFileUpdateNotification"

**Expected Results**:
- storeFile succeeds
- No notification published to topic
- No network traffic for notification

**References**:
- CRC: crc-Peer.md - "Does: publishFileUpdateNotification"

---

### Test: No file update notification when topic not configured

**Purpose**: Verify that Peer doesn't publish notifications when fileUpdateNotifyTopic is empty.

**Motivation**: Allows disabling notification feature globally.

**Input**:
- Peer A with fileUpdateNotifyTopic = "" (empty)
- Peer A calls storeFile("/test.txt", "content", false)

**References**:
- CRC: crc-Peer.md - "Does: publishFileUpdateNotification"

**Expected Results**:
- storeFile succeeds
- No notification attempted
- No network traffic

**References**:
- CRC: crc-PeerManager.md - "Knows: fileUpdateNotifyTopic"

## Coverage Summary

**Responsibilities Covered**:
- ✅ start - Protocol listener registration (implicit in receive test)
- ✅ stop - Protocol listener removal (not explicitly tested, low priority)
- ✅ sendToPeer - Protocol messaging send test
- ✅ subscribe - Topic subscription test with DHT advertisement and discovery
- ✅ publish - Topic publishing tests (success and failure)
- ✅ unsubscribe - Topic unsubscription test
- ✅ listPeers - Topic peer listing test
- ✅ advertiseTopic - DHT topic advertisement test (continuous re-advertisement)
- ✅ discoverTopicPeers - DHT peer discovery test
- ✅ addPeers - Connection protection test
- ✅ removePeers - Connection unprotection test
- ✅ monitor - Topic monitoring test
- ✅ stopMonitor - Stop monitoring test
- ✅ listFiles - Local and remote file listing tests
- ✅ getFile - Local retrieval and fallback peer tests
- ✅ storeFile - File and directory storage tests
- ✅ removeFile - File removal test
- ✅ publishFileUpdateNotification - Notification tests (subscribed, not subscribed, disabled)
- ✅ handleGetFileList - Remote file list request handling (implicit in listFiles remote test)
- ✅ handleFileList - File list response handling (implicit in listFiles remote test)
- ✅ handleGetFile - File request handling (implicit in getFile fallback test)
- ✅ handleFileContent - File response handling (implicit in getFile fallback test)

**Scenarios Covered**:
- ✅ seq-protocol-communication.md - Send and receive protocol message tests
- ✅ seq-pubsub-communication.md - Subscribe, publish, monitor tests with DHT integration
- ✅ seq-add-peers.md - Connection protection test
- ✅ seq-remove-peers.md - Connection unprotection test
- ✅ seq-list-files.md - Local and remote file listing tests
- ✅ seq-get-file.md - Local retrieval and fallback peer tests
- ✅ seq-store-file.md - File and directory storage tests

**Gaps**:
- VirtualConnectionManager stream lifecycle not explicitly tested (complex integration test)
- Concurrent file operations not tested (future enhancement)
- Large file transfers and performance not tested
- Network partition and recovery scenarios not tested
- DHT bootstrap and connectivity edge cases not fully tested
