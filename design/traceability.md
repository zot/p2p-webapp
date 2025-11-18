# Traceability Map

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

**Architecture:**
- architecture.md

---

## Level 2 ↔ Level 3 (Design to Implementation)

### crc-Server.md

**Source Spec:** main.md

**Implementation:**
- **internal/server/server.go**
  - [ ] File header (CRC + Spec)
  - [ ] Server struct comment → crc-Server.md
  - [ ] initialize() comment → seq-server-startup.md
  - [ ] start() comment → seq-server-startup.md
  - [ ] shutdown() comment

**Tests:**
- **internal/server/server_test.go**
  - [ ] File header referencing CRC card

---

### crc-WebSocketHandler.md

**Source Spec:** main.md

**Implementation:**
- **internal/websocket/handler.go**
  - [ ] File header (CRC + Spec + Sequences)
  - [ ] WebSocketHandler struct comment → crc-WebSocketHandler.md
  - [ ] acceptConnection() comment → seq-peer-creation.md
  - [ ] receiveMessage() comment
  - [ ] sendMessage() comment
  - [ ] routeRequest() comment
  - [ ] queueServerMessage() comment

**Tests:**
- **internal/websocket/handler_test.go**
  - [ ] File header referencing CRC card

---

### crc-PeerManager.md

**Source Spec:** main.md

**Implementation:**
- **internal/peer/manager.go**
  - [ ] File header (CRC + Spec + Sequences)
  - [ ] PeerManager struct comment → crc-PeerManager.md
  - [ ] createPeer() comment → seq-peer-creation.md
  - [ ] enableDiscovery() comment → seq-server-startup.md
  - [ ] enableNATTraversal() comment → seq-server-startup.md
  - [ ] startProtocol() comment → seq-protocol-communication.md
  - [ ] stopProtocol() comment
  - [ ] sendToPeer() comment → seq-protocol-communication.md
  - [ ] subscribeTopic() comment → seq-pubsub-communication.md
  - [ ] publishToTopic() comment → seq-pubsub-communication.md
  - [ ] unsubscribeTopic() comment
  - [ ] listTopicPeers() comment
  - [ ] allowPrivateGater type comment

**Tests:**
- **internal/peer/manager_test.go**
  - [ ] File header referencing CRC card

---

### crc-WebServer.md

**Source Spec:** main.md

**Implementation:**
- **internal/server/webserver.go**
  - [ ] File header (CRC + Spec)
  - [ ] WebServer struct comment → crc-WebServer.md
  - [ ] serveFile() comment
  - [ ] handleSPARoute() comment (spaHandler function)
  - [ ] detectFileType() comment

**Tests:**
- **internal/server/webserver_test.go**
  - [ ] File header referencing CRC card

---

### crc-BundleManager.md

**Source Spec:** main.md

**Implementation:**
- **internal/bundle/manager.go**
  - [ ] File header (CRC + Spec)
  - [ ] BundleManager struct comment → crc-BundleManager.md
  - [ ] checkBundled() comment
  - [ ] readFile() comment
  - [ ] listFiles() comment
  - [ ] copyFiles() comment
  - [ ] extractAll() comment
  - [ ] appendBundle() comment

**Tests:**
- **internal/bundle/manager_test.go**
  - [ ] File header referencing CRC card

---

### crc-ProcessTracker.md

**Source Spec:** main.md

**Implementation:**
- **internal/process/tracker.go**
  - [ ] File header (CRC + Spec)
  - [ ] ProcessTracker struct comment → crc-ProcessTracker.md
  - [ ] registerPID() comment → seq-server-startup.md
  - [ ] unregisterPID() comment
  - [ ] listPIDs() comment
  - [ ] verifyPID() comment
  - [ ] cleanStale() comment
  - [ ] killPID() comment
  - [ ] killAll() comment
  - [ ] lockFile() comment
  - [ ] unlockFile() comment

**Tests:**
- **internal/process/tracker_test.go**
  - [ ] File header referencing CRC card

---

### crc-P2PWebAppClient.md

**Source Spec:** main.md

**Implementation:**
- **pkg/client/src/index.ts**
  - [ ] File header (CRC + Spec)
  - [ ] P2PWebAppClient class comment → crc-P2PWebAppClient.md
  - [ ] connect() comment
  - [ ] start() comment
  - [ ] stop() comment
  - [ ] send() comment
  - [ ] subscribe() comment
  - [ ] publish() comment
  - [ ] unsubscribe() comment
  - [ ] listPeers() comment
  - [ ] handleResponse() comment
  - [ ] handleServerMessage() comment

**Tests:**
- **pkg/client/src/index.test.ts**
  - [ ] File header referencing CRC card

---

### crc-CommandRouter.md

**Source Spec:** main.md

**Implementation:**
- **cmd/p2p-webapp/main.go**
  - [ ] File header (CRC + Spec)
  - [ ] CommandRouter struct comment → crc-CommandRouter.md
  - [ ] parseArgs() comment
  - [ ] routeCommand() comment
  - [ ] handleServer() comment
  - [ ] handleExtract() comment
  - [ ] handleBundle() comment
  - [ ] handleLs() comment
  - [ ] handleCp() comment
  - [ ] handlePs() comment
  - [ ] handleKill() comment
  - [ ] handleKillAll() comment
  - [ ] handleVersion() comment

**Tests:**
- **cmd/p2p-webapp/main_test.go**
  - [ ] File header referencing CRC card
