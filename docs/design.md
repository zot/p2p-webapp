# Design Documentation

<!-- CRC Cards: crc-PeerManager.md, crc-Server.md, crc-WebSocketHandler.md -->
<!-- Sequences: seq-peer-creation.md, seq-protocol-communication.md, seq-file-operations.md -->

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [System Components](#system-components)
- [Design Patterns](#design-patterns)
- [Data Flow](#data-flow)
- [Key Design Decisions](#key-design-decisions)

## Architecture Overview

<!-- CRC: crc-PeerManager.md, crc-Server.md, crc-WebSocketHandler.md -->

**Architecture Style**: Layered Architecture with Manager-based Orchestration

**Layers**:
```
┌─────────────────────────────────────────┐
│         Browser (TypeScript)            │
│      Client Library (ES Modules)        │
└──────────────┬──────────────────────────┘
               │ WebSocket (JSON-RPC)
┌──────────────▼──────────────────────────┐
│         Command Router                  │
│  (extract, bundle, ps, kill, etc.)     │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         Server Orchestrator             │
│  (HTTP + WebSocket coordination)        │
└────┬────────────────────┬────────────────┘
     │                    │
┌────▼────────┐    ┌─────▼──────────────┐
│ WebServer   │    │ WebSocketHandler   │
│ (HTTP/SPA)  │    │ (JSON-RPC Protocol)│
└─────────────┘    └─────┬──────────────┘
                         │
                  ┌──────▼──────────┐
                  │  PeerManager    │
                  │ (libp2p + IPFS) │
                  └─────────────────┘
```

**Component Diagram**:
```
┌──────────────────────────────────────────────────┐
│                   p2p-webapp                     │
├──────────────────────────────────────────────────┤
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │ Bundle   │  │ Process  │  │ Command  │      │
│  │ Manager  │  │ Tracker  │  │ Router   │      │
│  └──────────┘  └──────────┘  └────┬─────┘      │
│                                    │             │
│  ┌────────────────────────────────▼──────────┐  │
│  │           Server Orchestrator             │  │
│  └───────┬──────────────────────┬────────────┘  │
│          │                      │                │
│  ┌───────▼──────┐       ┌──────▼────────┐      │
│  │  WebServer   │       │  WebSocket    │      │
│  │  (HTTP/SPA)  │       │  Handler      │      │
│  └──────────────┘       └───────┬───────┘      │
│                                  │               │
│                         ┌────────▼─────────┐    │
│                         │   PeerManager    │    │
│                         │                  │    │
│                         │  ┌────────────┐  │    │
│                         │  │   Peer     │  │    │
│                         │  │ (libp2p +  │  │    │
│                         │  │   IPFS)    │  │    │
│                         │  └────────────┘  │    │
│                         └──────────────────┘    │
└──────────────────────────────────────────────────┘
```

## System Components

### Server Orchestrator

<!-- CRC: crc-Server.md -->

**Purpose**: Coordinates HTTP server and WebSocket handler lifecycle

**Responsibilities**:
- Initialize HTTP server with static file serving
- Set up WebSocket upgrade endpoint
- Manage graceful shutdown
- Handle signal termination (SIGINT, SIGTERM, SIGHUP)
- Auto-exit when no connections remain

**Collaborates With**: WebServer, WebSocketHandler, PeerManager

**Key Methods**:
- `Start()` - Initialize and start server
- `Shutdown()` - Graceful cleanup of all resources

**Design Pattern**: Facade Pattern (simplifies subsystem coordination)

### WebServer

<!-- CRC: crc-Server.md -->

**Purpose**: Serve static files with SPA routing support

**Responsibilities**:
- Serve files from bundle or directory
- SPA fallback to index.html for routes without extensions
- Apply cache control and security headers
- CORS support (optional)

**Collaborates With**: BundleManager (if bundled mode), OS filesystem (if directory mode)

**Key Methods**:
- `spaHandler()` - Route detection and SPA fallback logic

**Design Pattern**: Handler Chain Pattern

### WebSocketHandler

<!-- CRC: crc-WebSocketHandler.md -->

**Purpose**: Implement JSON-RPC protocol over WebSocket

**Responsibilities**:
- Upgrade HTTP connections to WebSocket
- Parse and dispatch JSON-RPC messages
- Route requests to appropriate handlers
- Send responses and server-initiated messages
- Manage per-connection state

**Collaborates With**: PeerManager, Server

**Key Methods**:
- `handleWebSocket()` - Connection lifecycle
- `handleMessage()` - Message routing
- `sendResponse()` - Reply to client requests

**Design Pattern**: Command Pattern (message routing)

### PeerManager

<!-- CRC: crc-PeerManager.md -->

**Purpose**: Manage libp2p peers and IPFS operations

**Responsibilities**:
- Create peers with identity management
- Initialize libp2p host with DHT, mDNS, pubsub
- Track active peers per WebSocket connection
- Coordinate file operations via IPFS-lite
- Manage verbose logging with peer aliases
- Publish file availability notifications (optional)

**Collaborates With**: libp2p, IPFS-lite, Peer

**Key Methods**:
- `CreatePeer()` - Initialize new peer
- `GetPeer()` - Lookup peer by ID
- `NewManager()` - Setup with configuration

**Design Pattern**: Factory Pattern (peer creation), Observer Pattern (callbacks)

**Configuration**:
- `fileUpdateNotifyTopic` - Optional topic for file update notifications

### Peer

<!-- CRC: crc-PeerManager.md -->

**Purpose**: Represent individual peer with P2P operations

**Responsibilities**:
- Protocol message sending/receiving
- PubSub subscribe/publish/unsubscribe
- File operations (store, list, get, remove)
- HAMTDirectory management with pinning
- Stream lifecycle management
- Publish file update notifications when configured

**Collaborates With**: libp2p host, IPFS peer, PeerManager

**Key Methods**:
- `Start()` - Register protocol handler
- `SendToPeer()` - Send message to remote peer
- `Subscribe()` - Join pubsub topic
- `StoreFile()` - Add file to IPFS and directory
- `RemoveFile()` - Remove file from IPFS and directory
- `publishFileUpdateNotification()` - Notify subscribers of file changes

**Design Pattern**: Active Record Pattern (encapsulates data + operations)

**File Notification Flow**:
1. Check if `fileUpdateNotifyTopic` is configured
2. Check if peer is subscribed to notification topic
3. If both true, publish `{"type":"p2p-webapp-file-update","peer":"<peerID>"}`
4. Called after successful `StoreFile()` and `RemoveFile()` operations

### BundleManager

<!-- CRC: crc-BundleManager.md -->

**Purpose**: Create and extract ZIP bundles appended to executables

**Responsibilities**:
- Append ZIP archive to binary with footer
- Extract bundled content to directory
- List files in bundle
- Copy files from bundle to destination
- Validate bundle integrity

**Collaborates With**: OS filesystem

**Key Methods**:
- `CreateBundle()` - Build bundled executable
- `ExtractBundle()` - Extract to directory
- `ListFiles()` - List bundled files
- `CopyFiles()` - Copy files matching patterns

**Design Pattern**: Adapter Pattern (abstracts bundle vs. filesystem)

### ProcessTracker

<!-- CRC: crc-ProcessTracker.md -->

**Purpose**: Track running p2p-webapp instances for management

**Responsibilities**:
- Register PIDs on startup
- List all running instances
- Validate PIDs are actual p2p-webapp processes
- Clean up stale entries
- File locking for concurrent safety

**Collaborates With**: OS process system

**Key Methods**:
- `Register()` - Add current PID
- `List()` - Get all running instances
- `Remove()` - Remove specific PID
- `Kill()` - Terminate with graceful/force pattern

**Design Pattern**: Singleton Pattern (shared PID file), Registry Pattern

## Design Patterns

### Virtual Connection Model

<!-- CRC: crc-PeerManager.md -->

**Where Used**: PeerManager, Peer, WebSocketHandler

**Why**: Simplify client API by abstracting stream lifecycle

**Implementation**:
- Client uses (peer, protocol) tuple for addressing
- Server maintains `"peerID:protocol"` stream map
- Streams created on-demand and reused
- Automatic retry and reconnection

**Trade-offs**:
- Gained: Simple client API, no connection state management
- Lost: Fine-grained control over individual streams

### Sequential Message Processing

<!-- CRC: crc-WebSocketHandler.md -->

**Where Used**: TypeScript client library

**Why**: Guarantee message ordering and prevent race conditions

**Implementation**:
- Server-initiated messages queued in array
- `processMessageQueue()` processes one message at a time with async/await
- Response messages bypass queue for immediate handling

**Trade-offs**:
- Gained: Ordering guarantees, no race conditions
- Lost: Maximum throughput (serialization overhead)

### Observer Pattern

<!-- CRC: crc-PeerManager.md -->

**Where Used**: PeerManager callbacks

**Why**: Decouple peer operations from WebSocket communication

**Implementation**:
- PeerManager accepts callbacks: `onPeerData`, `onTopicData`, `onPeerChange`, etc.
- Callbacks invoked when events occur in peers
- WebSocketHandler registers callbacks to send messages to client

**Trade-offs**:
- Gained: Decoupling, testability, flexibility
- Lost: Slight complexity with callback management

### Facade Pattern

<!-- CRC: crc-Server.md, crc-PeerManager.md -->

**Where Used**: Server (orchestrates subsystems), PeerManager (simplifies libp2p)

**Why**: Hide complexity of subsystem coordination

**Implementation**:
- Server exposes simple `Start()` and `Shutdown()`
- Internally coordinates WebServer, WebSocketHandler, PeerManager
- PeerManager hides libp2p, DHT, mDNS, GossipSub complexity

### Factory Pattern

<!-- CRC: crc-PeerManager.md -->

**Where Used**: PeerManager.CreatePeer()

**Why**: Complex peer initialization with many dependencies

**Implementation**:
- `CreatePeer()` constructs Peer with:
  - libp2p host
  - IPFS peer
  - Protocol handlers
  - HAMTDirectory
  - Callbacks
- Validates peer key and prevents duplicates

## Data Flow

### Peer Creation Flow

<!-- Sequence: seq-peer-creation.md -->

**Flow Description**: Browser requests peer creation, server initializes libp2p peer with optional identity restoration

**Sequence**:
```
Browser → WebSocketHandler: Peer(peerKey?, rootDirectory?)
WebSocketHandler → PeerManager: CreatePeer(peerKey, rootCID)
PeerManager → libp2p: Create host with key
PeerManager → IPFS: Initialize peer
PeerManager → Peer: New instance
Peer → IPFS: Restore/create HAMTDirectory
Peer → IPFS: Pin directory
PeerManager → WebSocketHandler: [peerID, peerKey]
WebSocketHandler → Browser: Response [peerID, peerKey]
```

**Error Handling**:
- Duplicate peer ID returns error
- Invalid peer key returns error
- Directory CID validation failure returns error

### Protocol Communication Flow

<!-- Sequence: seq-protocol-communication.md -->

**Flow Description**: Browser sends message to remote peer via protocol

**Sequence**:
```
Browser → WebSocketHandler: send(peer, protocol, data, ack)
WebSocketHandler → PeerManager: GetPeer(localPeerID)
PeerManager → Peer: SendToPeer(targetPeer, protocol, data)
Peer → libp2p: Open/reuse stream to target
Peer → Stream: Write message
Stream → TargetPeer: Deliver message
TargetPeer → TargetPeer: Invoke protocol handler
TargetPeer → Stream: Write acknowledgment
Stream → Peer: Receive acknowledgment
Peer → WebSocketHandler: Delivery confirmed
WebSocketHandler → Browser: ack(ack)
```

**Error Handling**:
- Protocol not started: return error before sending
- Target peer offline: stream error returned
- Send timeout: error after retry attempts

### File Storage Flow

<!-- Sequence: seq-file-operations.md -->

**Flow Description**: Browser stores file, updates directory, optionally notifies subscribers

**Sequence**:
```
Browser → WebSocketHandler: storeFile(path, content)
WebSocketHandler → PeerManager: GetPeer(peerID)
PeerManager → Peer: StoreFile(path, content)
Peer → IPFS: Add file node
IPFS → Peer: File CID
Peer → IPFS: Get parent directory from HAMTDirectory
Peer → IPFS: Add entry to parent directory
IPFS → Peer: Updated parent CID
Peer → IPFS: Pin updated directory tree
Peer → Peer: Check if fileUpdateNotifyTopic configured
Peer → Peer: Check if subscribed to notification topic
Peer → Peer: Publish notification (if both checks pass)
Peer → WebSocketHandler: File CID
WebSocketHandler → Browser: Response with CID
```

**Notification Message**: `{"type":"p2p-webapp-file-update","peer":"<peerID>"}`

**Error Handling**:
- Invalid path: return error
- IPFS add failure: return error
- Directory update failure: rollback and return error
- Notification publish failure: logged but doesn't affect operation

### File Listing Flow

<!-- Sequence: seq-file-operations.md -->

**Flow Description**: Browser queries peer's file list via reserved protocol or local lookup

**Sequence (Local)**:
```
Browser → WebSocketHandler: listFiles(peerID)
WebSocketHandler → PeerManager: GetPeer(peerID)
PeerManager → Peer: ListFiles(peerID == local)
Peer → IPFS: Walk HAMTDirectory tree
Peer → WebSocketHandler: peerFiles(peerID, rootCID, entries)
WebSocketHandler → Browser: peerFiles server message
Browser → Browser: Resolve promise with {rootCID, entries}
```

**Sequence (Remote)**:
```
Browser → WebSocketHandler: listFiles(remotePeerID)
WebSocketHandler → PeerManager: GetPeer(localPeerID)
PeerManager → Peer: ListFiles(remotePeerID)
Peer → RemotePeer: getFileList() via P2PWebAppProtocol
RemotePeer → RemotePeer: Walk HAMTDirectory tree
RemotePeer → Peer: fileList(rootCID, entries)
Peer → WebSocketHandler: peerFiles(remotePeerID, rootCID, entries)
WebSocketHandler → Browser: peerFiles server message
Browser → Browser: Resolve promise with {rootCID, entries}
```

**Error Handling**:
- Remote peer offline: timeout and return error
- Invalid peer ID: return error immediately

### File Update Notification Flow

<!-- CRC: crc-PeerManager.md -->

**Flow Description**: After file storage/removal, notify subscribed peers of availability change

**Sequence**:
```
Peer → Peer: publishFileUpdateNotification()
Peer → Peer: Check fileUpdateNotifyTopic != ""
Peer → Peer: Check subscribed to fileUpdateNotifyTopic
Peer → Peer: Create notification message
Peer → PubSub: Publish to fileUpdateNotifyTopic
PubSub → SubscribedPeers: Deliver notification
SubscribedPeer → Client: topicData with notification
Client → Client: Check message type == "p2p-webapp-file-update"
Client → Client: Check if viewing that peer's files
Client → Client: Call listFiles() to refresh (if viewing)
```

**Privacy Design**:
- No notification if `fileUpdateNotifyTopic` not configured
- No notification if peer not subscribed to topic
- Opt-in mechanism ensures no unintended broadcasts

**Example (Demo)**:
- Configuration: `fileUpdateNotifyTopic = "chatroom"`
- Peer subscribes to "chatroom" for chat
- File operations automatically notify chatroom subscribers
- Demo refreshes file list when viewing that peer's files

## Key Design Decisions

### Decision: Virtual Connection Model

<!-- CRC: crc-PeerManager.md -->

**Context**: libp2p uses stream-based communication requiring explicit lifecycle management

**Decision**: Abstract streams behind (peer, protocol) addressing

**Rationale**:
- Simplifies client API dramatically
- Server better positioned to manage connections
- Aligns with typical application messaging patterns

**Alternatives Considered**:
- Expose connection IDs to client: too complex, error-prone
- One stream per message: performance overhead
- Manual stream management by client: leaky abstraction

**Trade-offs**:
- Gained: Simple API, automatic reconnection, stream reuse
- Lost: Per-stream flow control, parallel streams per protocol

### Decision: Sequential Message Processing

<!-- CRC: crc-WebSocketHandler.md -->

**Context**: Concurrent message processing could cause race conditions and ordering issues

**Decision**: Queue and serialize server-initiated messages

**Rationale**:
- Guarantees message order (critical for protocols like chat)
- Prevents race conditions in callbacks
- Simplifies client state management

**Alternatives Considered**:
- Full concurrency: requires complex synchronization
- Per-protocol queues: partial ordering, more complex
- No ordering guarantee: breaks many use cases

**Trade-offs**:
- Gained: Ordering, simplicity, correctness
- Lost: Maximum throughput under high message load

### Decision: Bundle-by-Append

<!-- CRC: crc-BundleManager.md -->

**Context**: Need single-executable distribution without go:embed limitations

**Decision**: Append ZIP archive to binary with footer metadata

**Rationale**:
- Go binaries can have trailing data
- ZIP files can be opened from arbitrary offset
- No compilation needed to create bundles
- Users can bundle their own applications

**Alternatives Considered**:
- go:embed: requires recompilation, not user-accessible
- Separate bundle file: loses single-executable benefit
- Custom archive format: reinventing the wheel

**Trade-offs**:
- Gained: User bundling, no recompilation, standard format
- Lost: Slightly larger binaries (ZIP overhead)

### Decision: Opt-In File Notifications

<!-- CRC: crc-PeerManager.md -->

**Context**: Peers may want to know when other peers update their files for automatic UI refresh

**Decision**: File update notifications only published if topic configured AND peer subscribed

**Rationale**:
- Privacy-first: no unintended broadcasts
- Flexible: applications choose notification topic (can be separate or shared with other features)
- Efficient: no notifications when nobody listening
- Simple: reuses existing pubsub infrastructure

**Alternatives Considered**:
- Always notify: violates privacy, wasted bandwidth
- Separate notification protocol: more complexity
- Poll-based refresh: inefficient, delayed updates
- Dedicated notification topic: less flexible

**Trade-offs**:
- Gained: Privacy, flexibility, efficiency
- Lost: Requires explicit configuration (not automatic)

**Implementation Notes**:
- Demo uses "chatroom" topic for both chat and file notifications
- Single topic subscription serves dual purpose
- Message type field (`"p2p-webapp-file-update"`) distinguishes notification from chat
- Applications can use separate topics if preferred

### Decision: HAMTDirectory for Peer Files

<!-- CRC: crc-PeerManager.md -->

**Context**: Need efficient directory structure for peer file storage

**Decision**: Use IPFS HAMTDirectory with automatic pinning

**Rationale**:
- Scalable to large directories
- Content-addressed (CID tracks version)
- Built-in IPFS integration
- Pinning ensures persistence

**Alternatives Considered**:
- Flat file list: doesn't scale, no hierarchy
- Custom tree structure: reinventing IPFS
- No pinning: files could be garbage collected

**Trade-offs**:
- Gained: Scalability, IPFS compatibility, persistence
- Lost: Directory operations slightly more complex

---

*Last updated: 2025-11-20 - Added File Update Notification Flow*
