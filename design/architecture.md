# Architecture

**Entry point to the design - shows how design elements are organized into logical systems**

**Sources**: All CRC cards and sequences created from main.md

---

## Systems

### Server System

**Purpose**: Main server orchestration and lifecycle management

**Design Elements**:
- crc-Server.md
- seq-server-startup.md

### WebSocket Protocol System

**Purpose**: JSON-RPC communication between browser and server

**Design Elements**:
- crc-WebSocketHandler.md
- seq-peer-creation.md

### Peer-to-Peer Networking System

**Purpose**: libp2p networking, peer discovery (mDNS/DHT), protocol messaging, and PubSub

**Design Elements**:
- crc-PeerManager.md
- seq-protocol-communication.md
- seq-pubsub-communication.md

### HTTP Server System

**Purpose**: Web file serving with SPA routing support

**Design Elements**:
- crc-WebServer.md

### Bundle Management System

**Purpose**: ZIP bundling and extraction for self-contained executables

**Design Elements**:
- crc-BundleManager.md

### Process Management System

**Purpose**: PID tracking for running instances

**Design Elements**:
- crc-ProcessTracker.md

### Client Library System

**Purpose**: TypeScript browser library for P2P communication

**Design Elements**:
- crc-P2PWebAppClient.md

### IPFS File Management System

**Purpose**: Per-peer IPFS file storage using HAMTDirectory, with ownership enforcement and cross-peer file list queries

**Design Elements**:
- crc-PeerManager.md (HAMTDirectory management, CID tracking)
- crc-WebSocketHandler.md (ownership enforcement, request routing)
- crc-P2PWebAppClient.md (client file API)
- seq-list-files.md
- seq-store-file.md

---

## Cross-Cutting Concerns

**Design elements that span multiple systems**

**Design Elements**:
- crc-CommandRouter.md

---

*This file serves as the architectural "main program" - start here to understand the design structure*
