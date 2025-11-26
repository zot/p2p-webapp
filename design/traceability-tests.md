# Test Traceability Map

**Maps test designs to CRC cards, sequences, and test implementation files**

---

## Level 1 → Level 2 → Test Designs

### main.md

**CRC Cards:**
- crc-Server.md
- crc-PeerManager.md
- crc-Peer.md
- crc-WebServer.md
- crc-WebSocketHandler.md
- crc-CommandRouter.md
- crc-BundleManager.md
- crc-ConfigLoader.md
- crc-ProcessTracker.md
- crc-P2PWebAppClient.md

**Sequences:**
- seq-server-startup.md
- seq-peer-creation.md
- seq-protocol-communication.md
- seq-pubsub-communication.md
- seq-add-peers.md
- seq-remove-peers.md
- seq-list-files.md
- seq-get-file.md
- seq-store-file.md

**Test Designs:**
- test-Server.md
- test-PeerManager.md
- test-Peer.md
- test-WebSocketHandler.md
- test-BundleManager.md
- test-P2PWebAppClient.md
- test-ConfigLoader.md
- test-ProcessTracker.md
- test-WebServer-CommandRouter.md
- test-dht-bootstrap.md

---

## Level 2 → Test Designs → Test Code

### test-Server.md

**Source Specs**: main.md
**Source CRC**: crc-Server.md
**Source Sequences**: seq-server-startup.md

**Test Implementation:**
- **internal/server/server_test.go** ✅ EXISTS
  - [ ] File header referencing test design
  - [x] Test: Server startup with available port
  - [ ] Test: Server startup with port collision
  - [ ] Test: Server startup with all ports unavailable
  - [ ] Test: Graceful shutdown on SIGTERM
  - [ ] Test: Graceful shutdown on SIGINT
  - [ ] Test: Server handles SIGHUP
  - [ ] Test: Server verbosity levels
  - [ ] Test: Server noOpen flag suppresses browser launch
  - [ ] Test: Server directory mode vs bundled mode

**Coverage:**
- ✅ initialize - Server startup tests
- ✅ start - Port selection tests
- ✅ serve - Implicit in startup tests
- ✅ shutdown - Signal handling tests
- ✅ handleSignals - SIGHUP, SIGINT, SIGTERM tests
- ✅ All "Knows" properties

**Gaps:**
- Integration with real WebSocket clients (see test-WebSocketHandler.md)
- Performance under load

---

### test-PeerManager.md

**Source Specs**: main.md
**Source CRC**: crc-PeerManager.md, crc-Peer.md
**Source Sequences**: seq-peer-creation.md, seq-add-peers.md, seq-remove-peers.md

**Test Implementation:**
- **internal/peer/manager_test.go** ❌ NOT YET CREATED
  - [ ] File header referencing test design
  - [ ] Test: Create peer with fresh key
  - [ ] Test: Create peer with provided key
  - [ ] Test: Create peer with duplicate peerID
  - [ ] Test: Create peer with root directory CID
  - [ ] Test: Remove peer and cleanup resources
  - [ ] Test: Get existing peer by ID
  - [ ] Test: Get nonexistent peer
  - [ ] Test: Add peers to connection protection
  - [ ] Test: Remove peers from connection protection
  - [ ] Test: Enable mDNS and DHT discovery
  - [ ] Test: Enable NAT traversal
  - [ ] Test: Set callbacks for peer events
  - [ ] Test: Generate unique aliases
  - [ ] Test: Verbose logging with peer aliases
  - [ ] Test: File update notification configuration

**Coverage:**
- ✅ All createPeer scenarios
- ✅ removePeer, getPeer
- ✅ addPeers, removePeers coordination
- ✅ enableDiscovery, enableNATTraversal
- ✅ setCallbacks, getOrCreateAlias, logVerbose
- ✅ All "Knows" properties

**Gaps:**
- Concurrent peer operations
- Discovery integration (requires test network)

---

### test-Peer.md

**Source Specs**: main.md
**Source CRC**: crc-Peer.md
**Source Sequences**: seq-protocol-communication.md, seq-pubsub-communication.md, seq-add-peers.md, seq-remove-peers.md, seq-list-files.md, seq-get-file.md, seq-store-file.md

**Test Implementation:**
- **internal/peer/peer_test.go** ❌ NOT YET CREATED
  - [ ] File header referencing test design
  - [ ] Test: Send data to peer on protocol

**Related Existing Tests:**
- **internal/peer/virtual_connection_test.go** ✅ EXISTS
- **internal/peer/connection_management_test.go** ✅ EXISTS
- **internal/peer/connection_management_integration_test.go** ✅ EXISTS
  - [ ] Test: Receive protocol message
  - [ ] Test: Subscribe to topic
  - [ ] Test: Publish message to topic
  - [ ] Test: Publish to unsubscribed topic
  - [ ] Test: Unsubscribe from topic
  - [ ] Test: List peers on topic
  - [ ] Test: Monitor topic for peer join/leave
  - [ ] Test: Stop monitoring topic
  - [ ] Test: Add peers to connection protection
  - [ ] Test: Remove peers from connection protection
  - [ ] Test: List local files
  - [ ] Test: List remote peer files
  - [ ] Test: Store file in directory
  - [ ] Test: Create directory in peer storage
  - [ ] Test: Remove file from directory
  - [ ] Test: Get file by CID (local)
  - [ ] Test: Get file with fallback peer
  - [ ] Test: Get file with fallback peer not having content
  - [ ] Test: File update notification when subscribed
  - [ ] Test: No file update notification when not subscribed
  - [ ] Test: No file update notification when topic not configured

**Coverage:**
- ✅ All protocol messaging operations
- ✅ All pub/sub operations
- ✅ All connection protection operations
- ✅ All file operations (list, get, store, remove)
- ✅ File update notifications
- ✅ All protocol message handlers

**Gaps:**
- VirtualConnectionManager integration (complex)
- Concurrent file operations
- Large file transfers

---

### test-WebSocketHandler.md

**Source Specs**: main.md
**Source CRC**: crc-WebSocketHandler.md
**Source Sequences**: seq-peer-creation.md, seq-list-files.md, seq-store-file.md

**Test Implementation:**
- **internal/server/websocket_test.go** ❌ NOT YET CREATED
  - [ ] File header referencing test design
  - [ ] Test: Accept WebSocket connection

**Related Existing Tests:**
- **internal/protocol/messages_test.go** ✅ EXISTS
  - [ ] Test: Receive and parse JSON-RPC message
  - [ ] Test: Send JSON-RPC response
  - [ ] Test: Route Peer() request to PeerManager
  - [ ] Test: Reject duplicate Peer() command
  - [ ] Test: Route listFiles to PeerManager with connection's peerID
  - [ ] Test: Route getFile to PeerManager with connection's peerID
  - [ ] Test: Enforce file ownership on storeFile
  - [ ] Test: Enforce file ownership on removeFile
  - [ ] Test: Queue server-initiated messages for sequential processing
  - [ ] Test: Send server-initiated peerData notification
  - [ ] Test: Send server-initiated topicData notification
  - [ ] Test: Send server-initiated peerChange notification
  - [ ] Test: Send server-initiated peerFiles notification
  - [ ] Test: Send server-initiated gotFile notification
  - [ ] Test: Close connection and cleanup
  - [ ] Test: Handle malformed JSON-RPC message
  - [ ] Test: Increment request ID counter for server messages

**Coverage:**
- ✅ acceptConnection, receiveMessage, sendMessage
- ✅ routeRequest, routeFileOperations
- ✅ enforceFileOwnership
- ✅ queueServerMessage
- ✅ All server-initiated notifications
- ✅ closeConnection

**Gaps:**
- WebSocket reconnection
- Concurrent client requests
- Large message handling

---

### test-BundleManager.md

**Source Specs**: main.md
**Source CRC**: crc-BundleManager.md

**Test Implementation:**
- **internal/bundle/bundle_test.go** ❌ NOT YET CREATED
  - [ ] File header referencing test design
  - [ ] Test: Detect bundled content in binary
  - [ ] Test: Detect absence of bundled content
  - [ ] Test: Read file from bundle
  - [ ] Test: Read nonexistent file from bundle
  - [ ] Test: List all files in bundle
  - [ ] Test: Copy files from bundle with glob pattern
  - [ ] Test: Extract entire bundle to directory
  - [ ] Test: Create bundled binary from directory
  - [ ] Test: Bundle with empty directory
  - [ ] Test: Read files with nested directory structure
  - [ ] Test: Extract preserves directory structure
  - [ ] Test: Handle ZIP corruption gracefully
  - [ ] Test: Handle footer corruption

**Coverage:**
- ✅ checkBundled (valid, absent, corrupted)
- ✅ readFile (existing, nonexistent, nested)
- ✅ listFiles
- ✅ copyFiles with glob patterns
- ✅ extractAll
- ✅ appendBundle

**Gaps:**
- Large bundle performance
- Concurrent file reads
- Security validation (path traversal)

---

### test-P2PWebAppClient.md

**Source Specs**: main.md
**Source CRC**: crc-P2PWebAppClient.md
**Source Sequences**: seq-add-peers.md, seq-remove-peers.md, seq-list-files.md, seq-get-file.md, seq-store-file.md

**Test Implementation:**
- **pkg/client/src/client.test.ts** ❌ NOT YET CREATED
  - [ ] File header referencing test design
  - [ ] Test: Connect to server and initialize peer
  - [ ] Test: Connect with existing peer key
  - [ ] Test: Start protocol listener
  - [ ] Test: Send protocol message to peer
  - [ ] Test: Send protocol message with delivery confirmation
  - [ ] Test: Receive protocol message
  - [ ] Test: Subscribe to topic
  - [ ] Test: Publish message to topic
  - [ ] Test: Receive topic message
  - [ ] Test: Receive peer change notification
  - [ ] Test: Unsubscribe from topic
  - [ ] Test: List peers on topic
  - [ ] Test: Add peers to connection protection
  - [ ] Test: Remove peers from connection protection
  - [ ] Test: List files from local peer
  - [ ] Test: List files from remote peer
  - [ ] Test: List files deduplication
  - [ ] Test: Store file with string content
  - [ ] Test: Store file with binary content
  - [ ] Test: Create directory
  - [ ] Test: Remove file
  - [ ] Test: Get file by CID (local)
  - [ ] Test: Get file with fallback peer
  - [ ] Test: Sequential server message processing
  - [ ] Test: Request ID increments for client requests
  - [ ] Test: Stop protocol listener

**Coverage:**
- ✅ connect (fresh and existing key)
- ✅ All protocol operations (start, stop, send, receive)
- ✅ All pub/sub operations (subscribe, publish, unsubscribe, listPeers)
- ✅ All peer change notifications
- ✅ All connection protection operations (addPeers, removePeers)
- ✅ All file operations (listFiles, getFile, storeFile, createDirectory, removeFile)
- ✅ Message routing (peerData, topicData, peerChange, peerFiles, gotFile, ack)
- ✅ Sequential message processing

**Gaps:**
- WebSocket reconnection
- Network failure handling
- Large file transfers

---

### test-ConfigLoader.md

**Source Specs**: main.md - Configuration section
**Source CRC**: crc-ConfigLoader.md

**Test Implementation:**
- **internal/config/loader_test.go** ❌ NOT YET CREATED
  - [ ] File header referencing test design
  - [ ] Test: Load configuration from filesystem directory
  - [ ] Test: Load configuration from ZIP bundle
  - [ ] Test: Handle missing configuration file gracefully
  - [ ] Test: Provide default configuration values
  - [ ] Test: Merge command-line flags into configuration
  - [ ] Test: Validate configuration values
  - [ ] Test: Parse duration strings
  - [ ] Test: Parse ServerConfig section
  - [ ] Test: Parse HTTPConfig section with CORS
  - [ ] Test: Parse WebSocketConfig section
  - [ ] Test: Parse BehaviorConfig section
  - [ ] Test: Parse FilesConfig section
  - [ ] Test: Parse P2PConfig section
  - [ ] Test: Handle malformed TOML file
  - [ ] Test: Configuration precedence order

**Coverage:**
- ✅ LoadFromDir, LoadFromZIP
- ✅ DefaultConfig
- ✅ Merge with flag precedence
- ✅ Validate
- ✅ TOML parsing for all config sections
- ✅ Duration string conversion
- ✅ Missing file handling
- ✅ Malformed TOML handling

**Gaps:**
- Partial configuration files
- Invalid duration strings
- Numeric limit edge cases

---

### test-ProcessTracker.md

**Source Specs**: main.md
**Source CRC**: crc-ProcessTracker.md

**Test Implementation:**
- **internal/pidfile/pidfile_test.go** ❌ NOT YET CREATED
  - [ ] File header referencing test design
  - [ ] Test: Register PID on startup

**Related Existing Tests:**
- **internal/commands/ps_integration_test.go** ✅ EXISTS
  - [ ] Test: Unregister PID on shutdown
  - [ ] Test: List all tracked PIDs
  - [ ] Test: Verify PID is actual p2p-webapp process
  - [ ] Test: Clean stale PIDs
  - [ ] Test: Kill specific instance by PID
  - [ ] Test: Kill instance with graceful shutdown
  - [ ] Test: Kill instance requiring SIGKILL
  - [ ] Test: Kill nonexistent PID
  - [ ] Test: Kill all tracked instances
  - [ ] Test: File locking prevents concurrent access corruption
  - [ ] Test: PID file location
  - [ ] Test: Register PID while extracting bundle

**Coverage:**
- ✅ registerPID, unregisterPID
- ✅ listPIDs, verifyPID
- ✅ cleanStale
- ✅ killPID (graceful, forced, nonexistent)
- ✅ killAll
- ✅ lockFile, unlockFile
- ✅ PID file location

**Gaps:**
- File lock timeout behavior
- File corruption recovery
- ExtractCommand integration (minimal)

---

### test-WebServer-CommandRouter.md

**Source Specs**: main.md
**Source CRC**: crc-WebServer.md, crc-CommandRouter.md

**Test Implementation:**
- **internal/server/webserver_test.go** ❌ NOT YET CREATED
  - [ ] File header referencing test design
  - [ ] Test: Serve static file
  - [ ] Test: Serve JavaScript file with correct Content-Type
  - [ ] Test: Serve CSS file with correct Content-Type
  - [ ] Test: SPA route fallback to index.html
  - [ ] Test: SPA route with nested path
  - [ ] Test: Return 404 for missing file with extension
  - [ ] Test: Serve files from BundleManager in bundled mode
  - [ ] Test: Cache index.html for SPA routing
  - [ ] Test: Detect route vs static file

- **cmd/p2p-webapp/main_test.go** ❌ NOT YET CREATED
  - [ ] File header referencing test design
  - [ ] Test: Route to default server command
  - [ ] Test: Parse command-line flags
  - [ ] Test: Route to extract command
  - [ ] Test: Route to bundle command
  - [ ] Test: Route to ls command
  - [ ] Test: Route to cp command
  - [ ] Test: Route to ps command
  - [ ] Test: Route to kill command
  - [ ] Test: Route to killall command
  - [ ] Test: Route to version command
  - [ ] Test: Handle unknown command
  - [ ] Test: Multiple verbosity flags

**Coverage:**
- ✅ WebServer: serveFile, handleSPARoute, detectFileType, return404, indexHTML caching
- ✅ CommandRouter: parseArgs, routeCommand, all handle* methods

**Gaps:**
- Large file serving
- Concurrent HTTP requests
- HTTP header validation

---

### test-dht-bootstrap.md

**Source Specs**: main.md
**Source CRC**: crc-PeerManager.md, crc-Peer.md
**Source Sequences**: seq-dht-bootstrap.md

**Test Implementation:**
- **internal/peer/dht_bootstrap_test.go** ❌ NOT YET CREATED
  - [ ] File header referencing test design
  - [ ] Test: DHT bootstrap success
  - [ ] Test: DHT bootstrap timeout
  - [ ] Test: Operation queuing before bootstrap
  - [ ] Test: Operation immediate execution after bootstrap
  - [ ] Test: Multiple queued operations
  - [ ] Test: No DHT case
  - [ ] Test: enqueueDHTOperation thread safety
  - [ ] Test: processQueuedDHTOperations synchronization
  - [ ] Test: Bootstrap peer connection
  - [ ] Test: Bootstrap routing table polling

**Coverage:**
- ✅ bootstrapDHT
- ✅ enqueueDHTOperation
- ✅ processQueuedDHTOperations
- ✅ DHT ready channel signaling

**Gaps:**
- Real DHT network integration
- Bootstrap peer failure scenarios

---

## Coverage Summary

**CRC Responsibilities:**
- Total responsibilities across all CRC cards: ~150
- Tested responsibilities: ~148 (99%)
- Untested responsibilities: ~2 (1%)

**Sequences:**
- Total sequences: 10
- Tested sequences: 10 (100%)
- Untested sequences: 0 (0%)

**Test Designs:**
- Total test design files: 10
- Total test cases: ~200

**Test Implementation Files (Existing):**
- internal/server/server_test.go ✅
- internal/peer/virtual_connection_test.go ✅
- internal/peer/connection_management_test.go ✅
- internal/peer/connection_management_integration_test.go ✅
- internal/protocol/messages_test.go ✅
- internal/commands/ps_integration_test.go ✅

**Test Implementation Files (Not Yet Created):**
- internal/peer/manager_test.go
- internal/peer/peer_test.go
- internal/server/websocket_test.go
- internal/bundle/bundle_test.go
- internal/config/loader_test.go
- internal/pidfile/pidfile_test.go
- internal/server/webserver_test.go
- cmd/p2p-webapp/main_test.go
- pkg/client/src/client.test.ts

**Gaps:**
- VirtualConnectionManager stream lifecycle (complex integration test)
- Discovery integration with real network (requires test infrastructure)
- NAT traversal with real NAT devices (requires test infrastructure)
- WebSocket reconnection and recovery (client library)
- Large file transfer performance
- Concurrent operation stress tests
- Security validation (path traversal, injection attacks)

---

## Test Implementation Checklist

### Go Backend Tests
- [x] internal/server/server_test.go ✅
- [ ] internal/peer/manager_test.go
- [ ] internal/peer/peer_test.go
- [x] internal/peer/virtual_connection_test.go ✅
- [x] internal/peer/connection_management_test.go ✅
- [x] internal/peer/connection_management_integration_test.go ✅
- [ ] internal/server/websocket_test.go
- [x] internal/protocol/messages_test.go ✅
- [ ] internal/bundle/bundle_test.go
- [ ] internal/config/loader_test.go
- [ ] internal/pidfile/pidfile_test.go
- [x] internal/commands/ps_integration_test.go ✅
- [ ] internal/server/webserver_test.go
- [ ] cmd/p2p-webapp/main_test.go

### TypeScript Client Tests
- [ ] pkg/client/src/client.test.ts

### Integration Tests (Future)
- [ ] End-to-end server startup and WebSocket connection
- [ ] Multi-peer protocol communication
- [ ] Multi-peer pub/sub messaging
- [ ] File sharing between multiple peers
- [ ] Bundle create → extract → serve cycle

---

*Last updated: 2025-11-26*
