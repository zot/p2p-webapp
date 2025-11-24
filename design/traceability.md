# Traceability Map

## Implementation Status Summary

**Last Updated:** After implementing IPFS file operations (listFiles, getFile, storeFile, removeFile)

**Overall Status:**
- ✅ Core functionality implemented (server, WebSocket, peer management, bundle management, process tracking, client library, commands)
- ✅ **File operations implemented** (IPFS file management system with listFiles, getFile, storeFile, removeFile)
- ✅ **Traceability comments added** to all major components (file headers and key methods)

**File Structure Notes:**
- Implementation uses package-level functions instead of manager structs in some cases (bundle, pidfile)
- Some design files reference non-existent paths (corrected in this document)
- WebSocket handling split between `internal/server/websocket.go` and `internal/protocol/handler.go`
- Web server functionality integrated into `internal/server/server.go`

**Testing Coverage:**
- ✅ Existing: server_test.go, virtual_connection_test.go, messages_test.go, ps_integration_test.go
- ❌ Missing: Most component test files, all TypeScript tests

---

## Level 1 ↔ Level 2 (Human Specs to Models)

### main.md

**CRC Cards:**
- crc-Server.md
- crc-WebSocketHandler.md
- crc-PeerManager.md
- crc-WebServer.md
- crc-BundleManager.md
- crc-ProcessTracker.md
- crc-P2PWebAppClient.md
- crc-CommandRouter.md

**Sequence Diagrams:**
- seq-server-startup.md
- seq-peer-creation.md
- seq-protocol-communication.md
- seq-pubsub-communication.md
- seq-dht-bootstrap.md
- seq-add-peers.md
- seq-remove-peers.md
- seq-list-files.md
- seq-store-file.md
- seq-get-file.md

**Test Designs:**
- test-dht-bootstrap.md

**Architecture:**
- architecture.md

---

## Level 2 ↔ Level 3 (Design to Implementation)

### crc-Server.md

**Source Spec:** main.md

**Implementation:**
- **internal/server/server.go** ✅ EXISTS
  - [x] File header (CRC + Spec)
  - [x] Server struct comment → crc-Server.md
  - [x] NewServerFromDir() comment → seq-server-startup.md
  - [x] NewServerFromBundle() comment → seq-server-startup.md
  - [x] Start() comment → seq-server-startup.md
  - [x] Stop() comment
  - [ ] startExitTimer() comment

**Tests:**
- **internal/server/server_test.go** ✅ EXISTS
  - [ ] File header referencing CRC card

---

### crc-WebSocketHandler.md

**Source Spec:** main.md

**Implementation:**
- **internal/server/websocket.go** ✅ EXISTS (not internal/websocket/handler.go)
  - [x] File header (CRC + Spec + Sequences)
  - [x] WSConnection struct comment → crc-WebSocketHandler.md
  - [x] NewWSConnection() comment → seq-peer-creation.md
  - [x] Start() comment
  - [ ] readPump() comment
  - [ ] writePump() comment
  - [ ] handleRequest() comment (routes to protocol.Handler)

- **internal/protocol/handler.go** ✅ EXISTS (protocol routing/handling)
  - [x] File header (CRC + Spec + Sequences)
  - [x] Handler struct comment
  - [x] NewHandler() comment
  - [x] HandleClientMessage() comment
  - [ ] handlePeer() comment → seq-peer-creation.md
  - [ ] handleStart() comment
  - [ ] handleSend() comment
  - [ ] handleListPeers() comment

- **internal/protocol/messages.go** ✅ EXISTS (message type definitions)
  - [x] File operation message types added:
    - ListFilesResponse
    - GetFileRequest, GetFileResponse
    - StoreFileRequest, StoreFileResponse
    - RemoveFileRequest

- **internal/protocol/handler.go** ✅ EXISTS (file operation handlers)
  - ✅ handleListFiles() implemented
  - ✅ handleGetFile() implemented
  - ✅ handleStoreFile() implemented
  - ✅ handleRemoveFile() implemented
  - [ ] handleListFiles() comment → seq-list-files.md
  - [ ] handleGetFile() comment
  - [ ] handleStoreFile() comment → seq-store-file.md
  - [ ] handleRemoveFile() comment

**Tests:**
- **internal/server/websocket_test.go** ❌ DOES NOT EXIST
  - [ ] File header referencing CRC card
- **internal/protocol/messages_test.go** ✅ EXISTS (partial coverage)
  - [ ] File header referencing CRC card

---

### crc-PeerManager.md

**Source Spec:** main.md

**Implementation:**
- **internal/peer/manager.go** ✅ EXISTS
  - [x] File header (CRC + Spec + Sequences)
  - [x] Manager struct comment → crc-PeerManager.md
  - [x] Peer struct comment → crc-PeerManager.md
  - [x] NewManager() comment → seq-server-startup.md
  - [x] CreatePeer() comment → seq-peer-creation.md
  - [x] Start() comment → seq-protocol-communication.md
  - [ ] Stop() comment
  - [x] Send() comment → seq-protocol-communication.md
  - [x] Subscribe() comment → seq-pubsub-communication.md
  - [x] Publish() comment → seq-pubsub-communication.md
  - [ ] Unsubscribe() comment
  - [ ] ListPeers() comment
  - [ ] Monitor() comment
  - [ ] StopMonitor() comment
  - [ ] RemovePeer() comment
  - [ ] Bootstrap() comment
  - [ ] allowPrivateGater type comment
  - [x] advertiseTopic() comment → seq-pubsub-communication.md (DHT topic advertisement)
  - [x] discoverTopicPeers() comment → seq-pubsub-communication.md (DHT peer discovery)

**DHT bootstrap and queuing implemented:**
  - ✅ bootstrapDHT() → seq-dht-bootstrap.md
  - ✅ enqueueDHTOperation() → seq-dht-bootstrap.md
  - ✅ processQueuedDHTOperations() → seq-dht-bootstrap.md
  - ✅ retryAddedPeersLoop() (uses DHT for retry)
  - [ ] bootstrapDHT() comment → seq-dht-bootstrap.md
  - [ ] enqueueDHTOperation() comment → seq-dht-bootstrap.md
  - [ ] processQueuedDHTOperations() comment → seq-dht-bootstrap.md
  - [ ] retryAddedPeersLoop() comment

**File operations implemented:**
  - ✅ ListFiles() → seq-list-files.md
  - ✅ GetFile()
  - ✅ StoreFile() → seq-store-file.md
  - ✅ RemoveFile()
  - [ ] ListFiles() comment → seq-list-files.md
  - [ ] GetFile() comment
  - [ ] StoreFile() comment → seq-store-file.md
  - [ ] RemoveFile() comment

**Tests:**
- **internal/peer/manager_test.go** ❌ DOES NOT EXIST
  - [ ] File header referencing CRC card
- **internal/peer/virtual_connection_test.go** ✅ EXISTS (tests VirtualConnectionManager)
  - [ ] File header referencing CRC card
- **tests/peer_dht_test.go** ❌ DOES NOT EXIST (DHT bootstrap tests)
  - [ ] File header → test-dht-bootstrap.md
  - [ ] Test DHT bootstrap success
  - [ ] Test DHT bootstrap timeout
  - [ ] Test operation queuing before bootstrap
  - [ ] Test operation immediate execution after bootstrap
  - [ ] Test multiple queued operations
  - [ ] Test no DHT case
  - [ ] Test enqueueDHTOperation thread safety
  - [ ] Test processQueuedDHTOperations synchronization
  - [ ] Test bootstrap peer connection
  - [ ] Test bootstrap routing table polling

---

### crc-WebServer.md

**Source Spec:** main.md

**Implementation:**
- **internal/server/server.go** ✅ EXISTS (integrated with Server, not separate file)
  - [x] File header (CRC + Spec) - shared with Server
  - [x] zipFileSystem struct comment → crc-WebServer.md
  - [x] zipFile struct comment → crc-WebServer.md
  - [x] Open() comment (implements http.FileSystem)
  - [x] Read() comment
  - [ ] Stat() comment
  - [x] spaHandler() comment (SPA routing)

**Tests:**
- **internal/server/webserver_test.go** ❌ DOES NOT EXIST (covered by server_test.go)
  - [ ] File header referencing CRC card

---

### crc-BundleManager.md

**Source Spec:** main.md

**Implementation:**
- **internal/bundle/bundle.go** ✅ EXISTS (package-level functions, not struct-based)
  - [x] File header (CRC + Spec)
  - [x] Footer struct comment → crc-BundleManager.md
  - [x] IsBundled() comment (checks if binary is bundled)
  - [x] GetBundleReader() comment (reads bundled content)
  - [x] ExtractBundle() comment (extracts to filesystem)
  - [x] CreateBundle() comment (creates bundled binary)
  - [ ] GetBinarySize() comment (gets executable size)
  - [ ] addDirToZip() comment (helper for zipping)
  - [ ] extractZipFile() comment (helper for extraction)

**Tests:**
- **internal/bundle/bundle_test.go** ❌ DOES NOT EXIST
  - [ ] File header referencing CRC card

---

### crc-ProcessTracker.md

**Source Spec:** main.md

**Implementation:**
- **internal/pidfile/pidfile.go** ✅ EXISTS (package-level functions, not internal/process/tracker.go)
  - [x] File header (CRC + Spec)
  - [x] PIDFile struct comment → crc-ProcessTracker.md
  - [x] Register() comment → seq-server-startup.md
  - [ ] Unregister() comment
  - [x] List() comment
  - [ ] GetProcessInfo() comment
  - [ ] verifyPIDs() comment (stale cleanup)
  - [x] Kill() comment
  - [ ] KillAll() comment
  - [ ] withLockedPIDFile() comment (file locking)
  - [ ] writePIDFile() comment
  - [ ] isIPFSWebappProcess() comment

- **internal/pidfile/pidfile_unix.go** ✅ EXISTS (Unix-specific file locking)
  - [ ] lockFile() comment
  - [ ] unlockFile() comment

- **internal/pidfile/pidfile_windows.go** ✅ EXISTS (Windows-specific file locking)
  - [ ] lockFile() comment
  - [ ] unlockFile() comment

**Tests:**
- **internal/pidfile/pidfile_test.go** ❌ DOES NOT EXIST
  - [ ] File header referencing CRC card

---

### crc-P2PWebAppClient.md

**Source Spec:** main.md

**Implementation:**
- **pkg/client/src/client.ts** ✅ EXISTS (main implementation, index.ts just re-exports)
  - [x] File header (CRC + Spec)
  - [x] P2PWebAppClient class comment → crc-P2PWebAppClient.md
  - [x] connect() comment
  - [ ] close() comment
  - [ ] start() comment
  - [ ] stop() comment
  - [ ] send() comment
  - [ ] subscribe() comment
  - [ ] publish() comment
  - [ ] unsubscribe() comment
  - [ ] listPeers() comment
  - [ ] peerID getter comment
  - [ ] peerKey getter comment
  - [ ] handleMessage() comment
  - [ ] handleServerRequest() comment
  - [ ] handleClose() comment
  - [ ] sendRequest() comment
  - [ ] processMessageQueue() comment

- **pkg/client/src/index.ts** ✅ EXISTS (re-exports from client.ts)
  - [x] File header (CRC + Spec)

- **pkg/client/src/types.ts** ✅ EXISTS (type definitions)
  - [ ] File header
  - [x] File operation types added:
    - ListFilesResponse
    - GetFileRequest, GetFileResponse
    - StoreFileRequest, StoreFileResponse
    - RemoveFileRequest

**File operations implemented:**
  - ✅ listFiles() → seq-list-files.md
  - ✅ getFile()
  - ✅ storeFile() → seq-store-file.md
  - ✅ removeFile()
  - [ ] listFiles() comment → seq-list-files.md
  - [ ] getFile() comment
  - [ ] storeFile() comment → seq-store-file.md
  - [ ] removeFile() comment

**Tests:**
- **pkg/client/src/client.test.ts** ❌ DOES NOT EXIST
  - [ ] File header referencing CRC card

---

### crc-CommandRouter.md

**Source Spec:** main.md

**Implementation:**
- **cmd/p2p-webapp/main.go** ✅ EXISTS (uses cobra.Command, not a struct-based router)
  - [x] File header (CRC + Spec)
  - [x] rootCmd variable comment → crc-CommandRouter.md
  - [ ] init() comment (sets up flags and subcommands)
  - [ ] runServe() comment (default serve command)
  - [ ] validateDirectoryStructure() comment

- **internal/commands/extract.go** ✅ EXISTS
  - [x] File header (CRC + Spec)
  - [x] ExtractCmd variable comment

- **internal/commands/bundle.go** ✅ EXISTS
  - [x] File header (CRC + Spec)
  - [ ] BundleCmd variable comment

- **internal/commands/ls.go** ✅ EXISTS
  - [x] File header (CRC + Spec)
  - [ ] LsCmd variable comment

- **internal/commands/cp.go** ✅ EXISTS
  - [x] File header (CRC + Spec)
  - [ ] CpCmd variable comment

- **internal/commands/ps.go** ✅ EXISTS
  - [x] File header (CRC + Spec)
  - [x] PsCmd variable comment

- **internal/commands/kill.go** ✅ EXISTS
  - [x] File header (CRC + Spec)
  - [ ] KillCmd variable comment

- **internal/commands/killall.go** ✅ EXISTS
  - [x] File header (CRC + Spec)
  - [ ] KillAllCmd variable comment

- **internal/commands/version.go** ✅ EXISTS
  - [x] File header (CRC + Spec)
  - [ ] VersionCmd variable comment

**Tests:**
- **cmd/p2p-webapp/main_test.go** ❌ DOES NOT EXIST
  - [ ] File header referencing CRC card
- **internal/commands/ps_integration_test.go** ✅ EXISTS (integration test for ps command)
  - [ ] File header referencing CRC card

---

## Additional Implementation Files

**These files exist but are not directly mapped to CRC cards (infrastructure/support code):**

### IPFS Node Infrastructure
- **internal/ipfs/node.go** ✅ EXISTS
  - [ ] File header (CRC + Spec)
  - [ ] Node struct comment
  - [ ] NewNode() comment
  - [ ] Close() comment
  - [ ] Host() getter comment
  - [ ] PeerID() getter comment
  - [ ] loadOrGenerateKey() comment

### Virtual Connection Infrastructure
- **internal/peer/virtual_connection.go** ✅ EXISTS
  - [ ] File header
  - [ ] VirtualConnectionManager struct comment
  - [ ] NewVirtualConnectionManager() comment
  - [ ] Related test: virtual_connection_test.go ✅ EXISTS

---

## Missing Traceability Comments

**ACTION REQUIRED:** None of the implementation files currently have traceability comments.

**To add traceability comments:**
1. Use the file header format: `// CRC: crc-ComponentName.md, Spec: main.md`
2. For functions/methods referenced in sequences: `// Sequence: seq-operation-name.md`
3. Verify with: `python .claude/scripts/trace-verify.py`

**Example:**
```go
// CRC: crc-Server.md, Spec: main.md
package server

// Server manages the HTTP server and WebSocket connections
// CRC: crc-Server.md
type Server struct { ... }

// Start initializes and starts the server
// Sequence: seq-server-startup.md
func (s *Server) Start() error { ... }
```
