# p2p-webapp Architecture

**Complete architectural documentation for p2p-webapp**

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Architectural Principles](#architectural-principles)
3. [System Components](#system-components)
4. [Key Workflows](#key-workflows)
5. [Design Patterns](#design-patterns)
6. [Technology Stack](#technology-stack)

---

## System Overview

p2p-webapp is a local backend that eliminates the need for hosting by embedding a complete peer-to-peer networking stack with IPFS file storage into a single executable. The architecture is organized into seven interconnected systems:

```
┌─────────────────────────────────────────────────────────────────┐
│                      p2p-webapp Binary                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│  │   Server     │  │  WebSocket   │  │   HTTP       │        │
│  │  Orchestrator│◄─┤   Handler    │  │  Server      │        │
│  └──────────────┘  └──────┬───────┘  └──────────────┘        │
│                            │                                    │
│                    ┌───────▼────────┐                          │
│                    │  Peer Manager  │                          │
│                    │  (libp2p/IPFS) │                          │
│                    └────────────────┘                          │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│  │   Bundle     │  │   Process    │  │   Command    │        │
│  │   Manager    │  │   Tracker    │  │   Router     │        │
│  └──────────────┘  └──────────────┘  └──────────────┘        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ WebSocket
                              ▼
                    ┌─────────────────┐
                    │    Browser      │
                    │  (TypeScript    │
                    │   Client Lib)   │
                    └─────────────────┘
```

### Core Philosophy

1. **Single Executable Distribution** - Everything bundled into one binary
2. **Zero Configuration** - Works out of the box with sensible defaults
3. **Transparent P2P** - Application developers don't see networking complexity
4. **Local Backend Model** - Replaces traditional client-server architecture

---

## Architectural Principles

### SOLID Principles Applied

**Single Responsibility**
- Each component has one clear purpose
- Server orchestrates, PeerManager handles P2P, WebSocketHandler manages protocol

**Open/Closed**
- Protocol handlers extensible without modifying core
- New commands can be added to CommandRouter

**Liskov Substitution**
- BundleManager can be replaced with filesystem implementation
- PeerManager abstracts libp2p implementation details

**Interface Segregation**
- Client sees simple protocol API (start, send, subscribe)
- Internal components use specific interfaces they need

**Dependency Inversion**
- Server depends on abstractions (handlers, managers) not concrete implementations
- Allows testing and modular development

### Key Design Decisions

**Virtual Connection Model**
- Client uses (peer, protocol) addressing instead of managing connections
- Server transparently handles stream lifecycle
- Trade-off: Simplicity vs. fine-grained control

**Sequential Message Processing**
- Server-initiated messages queued and processed one at a time
- Guarantees ordering, prevents race conditions
- Trade-off: Ordering guarantees vs. maximum throughput

**Bundle-by-Append**
- ZIP appended to executable with footer
- No compilation tools needed
- Trade-off: Single-file distribution vs. binary size

---

## System Components

### 1. Server System

**Purpose**: Main orchestrator for server lifecycle

**Responsibilities**:
- Initialize all subsystems
- Start/stop services coordinately
- Manage port allocation (auto-select from 10000)
- Register/unregister with ProcessTracker
- Handle graceful shutdown

**Key Classes**: `Server`

**Configuration**:
- `--dir DIR` - Serve from directory vs. bundled mode
- `--noopen` - Suppress browser launch
- `-v, -vv, -vvv` - Verbosity levels (1-3)
- `-p PORT` - Specify port (default: auto-select)

---

### 2. WebSocket Protocol System

**Purpose**: JSON-RPC communication between browser and server

**Responsibilities**:
- Accept WebSocket connections
- Parse/validate JSON-RPC messages
- Route requests to PeerManager
- Queue server-initiated messages for sequential delivery
- Manage connection lifecycle

**Key Classes**: `WebSocketHandler`

**Message Flow**:
```
Browser → WebSocket → WebSocketHandler → PeerManager
Browser ← WebSocket ← WebSocketHandler ← PeerManager
```

**Protocol Features**:
- Request/response with requestID
- Server push notifications (peerData, topicData, peerChange, ack)
- Error propagation with error responses

---

### 3. Peer-to-Peer Networking System

**Purpose**: libp2p networking, discovery, and messaging

**Responsibilities**:
- Create and manage libp2p peers
- Enable discovery (mDNS + DHT)
- Configure NAT traversal (Circuit Relay, hole punching, AutoRelay, UPnP)
- Manage virtual connections (peerID:protocol → stream)
- Handle protocol-based messaging
- Manage GossipSub topics
- Generate peer aliases for logging

**Key Classes**: `PeerManager`, `Peer`, `allowPrivateGater`

**Discovery Mechanisms**:

**mDNS (Local Discovery)**:
- Zero-config local network discovery
- Sub-second peer discovery on LAN
- Perfect for development and same-network collaboration

**DHT (Global Discovery)**:
- Distributed Hash Table for internet-wide discovery
- Bootstraps via well-known IPFS DHT nodes
- Integrated with GossipSub for topic-based discovery

**NAT Traversal**:
- Circuit Relay v2: Connect via relay when direct connection blocked
- Hole Punching: Attempt direct NAT traversal
- AutoRelay: Automatically find public relay nodes
- UPnP/NAT-PMP: Automatic port forwarding

**Message Delivery**:
- Virtual connections keyed by "peerID:protocol"
- Streams created on-demand, reused for efficiency
- Optional delivery acknowledgment with callbacks
- All retry logic and buffering handled server-side

**IPFS File Storage**:
- Each peer maintains a HAMTDirectory (Hash Array Mapped Trie Directory)
- Directory structure stored in IPFS, identified by root CID
- Supports hierarchical file organization with Unix-style paths
- Files and directories are content-addressed (immutable)
- Peer pins its root directory to prevent garbage collection
- Directory state can be restored across sessions using root CID

**File Operations**:
- `listFiles(peerID)` - List files in peer's directory (local or remote)
- `getFile(cid)` - Retrieve content by CID from IPFS network
- `storeFile(path, content, directory)` - Add files/directories to peer's tree, returns CID of stored node
- `removeFile(path)` - Remove entries from peer's directory
- Reserved `p2p-webapp` protocol for peer-to-peer file list requests

---

### 4. HTTP Server System

**Purpose**: Serve web application files with SPA routing

**Responsibilities**:
- Serve static files from bundled content or directory
- Implement SPA routing fallback
- Detect content types
- Return appropriate 404 errors

**Key Classes**: `WebServer`

**SPA Routing Logic**:
```
Request → Has extension?
            ├─ Yes → Serve file or 404
            └─ No  → Serve index.html, preserve URL
```

**Examples**:
- `/` → `html/index.html`
- `/settings` → `html/index.html` (URL preserved)
- `/main.js` → `html/main.js`
- `/missing.js` → 404 error

---

### 5. Bundle Management System

**Purpose**: Self-contained executable with embedded site

**Responsibilities**:
- Detect bundled content (magic marker)
- Read files from ZIP archive
- List bundled files (`ls` command)
- Copy files with glob patterns (`cp` command)
- Extract entire bundle (`extract` command)
- Create new bundled executables (`bundle` command)

**Key Classes**: `BundleManager`

**Bundle Format**:
```
[Go Binary Executable]
[ZIP Archive: html/, ipfs/, storage/]
[Footer: Magic(8) + Offset(8) + Size(8)]
```

**Advantages**:
- Single-file distribution
- No extraction required for serving
- Works cross-platform
- No compilation tools needed

---

### 6. Process Management System

**Purpose**: Track and manage running instances

**Responsibilities**:
- Register PIDs on startup
- Maintain PID file with file locking
- Verify PIDs are actual p2p-webapp instances
- Clean stale entries automatically
- Support ps/kill/killall commands with graceful shutdown (SIGTERM→SIGKILL)

**Key Classes**: `ProcessTracker`

**PID File**: `/tmp/.p2p-webapp` (JSON format)

**Concurrency Safety**:
- File locking during read/write operations
- Prevents corruption from multiple simultaneous accesses
- Automatic stale entry cleanup

---

### 7. Client Library System

**Purpose**: TypeScript API for browser applications

**Responsibilities**:
- Connect to server and initialize peer
- Manage protocol listeners
- Queue and process server messages sequentially
- Handle message acknowledgments
- Subscribe to topics
- Provide promise-based API

**Key Classes**: `P2PWebAppClient`

**API Design**:
- Promise-based for all async operations
- Callback-based for incoming messages
- Internal ack number management (transparent to user)
- Sequential message processing guarantees ordering

---

### 8. Cross-Cutting: Command Router

**Purpose**: CLI command routing and parsing

**Responsibilities**:
- Parse command-line arguments
- Identify subcommand or default (server)
- Route to appropriate handler
- Handle all CLI commands

**Key Classes**: `CommandRouter`

**Commands**:
- Default: Start server
- `extract`: Extract bundle to current directory
- `bundle`: Create bundled executable
- `ls`: List bundled files
- `cp`: Copy files from bundle
- `ps`: List running instances
- `kill PID`: Terminate specific instance (SIGTERM first, SIGKILL after 5s if needed)
- `killall`: Terminate all instances (SIGTERM first, SIGKILL after 5s if needed)
- `version`: Display version

---

## Key Workflows

### Workflow 1: Server Startup

**Trigger**: User runs `./p2p-webapp`

**Steps**:
1. CommandRouter parses arguments
2. Server initializes PeerManager
   - PeerManager configures mDNS + DHT discovery
   - PeerManager enables NAT traversal features
3. Server initializes WebSocketHandler
4. Server initializes WebServer
5. Server registers PID with ProcessTracker
6. WebServer starts listening on port
7. WebSocketHandler starts listening on same port
8. Browser opens automatically (unless --noopen)

**Key Design Points**:
- Port auto-selection from 10000 (tries up to 100 ports)
- All services share same port
- PID tracking for process management
- Discovery enabled immediately for fast peer connections

---

### Workflow 2: Peer Creation

**Trigger**: Browser connects and sends `Peer(peerKey?)` command

**Steps**:
1. Browser establishes WebSocket connection
2. Browser sends `Peer(peerKey?)` as first message
3. WebSocketHandler routes to PeerManager
4. PeerManager creates new Peer with key (or generates fresh key)
5. Peer initializes libp2p
6. Peer enables discovery mechanisms
7. PeerManager checks for duplicate peerID
   - If duplicate: Return error (prevents multi-tab issues)
   - If unique: Store peer and generate alias (peer-a, peer-b, ...)
8. PeerManager returns `[peerID, peerKey]` to browser

**Key Design Points**:
- Peer() must be first command (cannot be sent twice)
- Duplicate detection prevents multi-tab conflicts
- Aliases improve log readability
- Discovery automatic (no separate command needed)

---

### Workflow 3: Protocol-Based Communication

**Trigger**: Browser wants to send data to another peer

**Steps**:
1. Browser calls `start(protocol)` to register listener
2. PeerManager registers protocol listener callback
3. Browser calls `send(peer, protocol, data, onAck?)`
4. PeerManager validates protocol is started
5. PeerManager gets or creates stream for "peerID:protocol"
   - If new: Opens libp2p stream to peer
   - If exists: Reuses existing stream
6. PeerManager writes data to stream
7. If ack requested: PeerManager waits for confirmation
8. Remote peer receives data
9. Remote PeerManager routes to protocol listener
10. Remote PeerManager sends `peerData(peer, protocol, data)` to browser
11. Remote browser's protocol listener callback receives `(peer, data)`

**Key Design Points**:
- Virtual connection model: Client addresses by (peer, protocol)
- Server manages stream lifecycle transparently
- Streams created on-demand, reused for efficiency
- Optional acknowledgment provides delivery confirmation
- Sequential processing guarantees message order

---

### Workflow 4: PubSub Communication

**Trigger**: Browser wants to communicate via topic

**Steps**:
1. Browser1 calls `subscribe(topic)`
2. PeerManager1 joins GossipSub topic
3. GossipSub advertises topic to DHT
4. PeerManager1 monitors peer join/leave events
5. Browser2 calls `subscribe(topic)`
6. GossipSub2 joins topic and advertises to DHT
7. DHT notifies GossipSub1 that peer2 joined
8. PeerManager1 sends `peerChange(topic, peer2, true)` to Browser1
9. Browser1 calls `publish(topic, data)`
10. GossipSub broadcasts to all topic subscribers
11. PeerManager2 receives message
12. PeerManager2 sends `topicData(topic, peer1, data)` to Browser2

**Key Design Points**:
- GossipSub integrated with DHT for discovery
- Automatic peer join/leave monitoring
- No separate monitoring command needed
- Messages include sender peerID for identification

---

## Design Patterns

### Pattern 1: Virtual Connection Model

**Problem**: Client code is complex when managing connection state

**Solution**: Client addresses by (peer, protocol) tuple; server manages streams

**Implementation**:
```
Client: send(peerID, protocol, data)
Server: Map[peerID:protocol] → libp2p stream
```

**Benefits**:
- Simplified client API
- Server handles retry, buffering, reconnection
- Transparent stream reuse

**Trade-offs**:
- Less fine-grained control
- Potential hidden resource usage

---

### Pattern 2: Sequential Message Processing

**Problem**: Message ordering critical, race conditions in callbacks

**Solution**: Queue server-initiated messages, process sequentially with async/await

**Implementation**:
```typescript
messageQueue: Message[]
async processMessageQueue() {
  for (const message of queue) {
    await handleMessage(message)  // Process one at a time
  }
}
```

**Benefits**:
- Guaranteed ordering
- No race conditions
- Simpler concurrency model

**Trade-offs**:
- Lower throughput for independent messages
- Head-of-line blocking risk

---

### Pattern 3: Protocol-Based Routing

**Problem**: Managing multiple protocols over connections

**Solution**: Route by protocol instead of connection

**Implementation**:
```
Client: Map[protocol] → callback(peer, data)
Server: Map[peerID:protocol] → stream
```

**Benefits**:
- Decouples protocol from connection
- Multiple protocols per peer
- Easy to add new protocols

---

### Pattern 4: Alias Generation

**Problem**: Peer IDs are long hashes, hard to read in logs

**Solution**: Generate sequential human-readable aliases

**Implementation**:
```
peer-a, peer-b, peer-c, ...
[peer-a] Connected to peer-b
[peer-b] Received message from peer-a
```

**Benefits**:
- Logs are much easier to follow
- Especially helpful with verbose output
- Consistent across session

---

### Pattern 5: Bundle-by-Append

**Problem**: Want single executable with embedded content

**Solution**: Append ZIP + footer to Go binary

**Implementation**:
```
[Go Binary][ZIP Archive][Footer: Magic + Offset + Size]
```

**Benefits**:
- No compilation tools needed
- Cross-platform compatible
- Go runtime ignores trailing data

**Trade-offs**:
- Slightly larger binary
- Can't partially extract

---

## Technology Stack

### Backend (Go)

**Core Libraries**:
- **libp2p** - P2P networking, streams, protocols
- **ipfs-lite** - IPFS integration, DHT
- **gopsutil** - Cross-platform process management
- **gorilla/websocket** - WebSocket implementation

**P2P Features**:
- **mDNS** - Local peer discovery
- **Kademlia DHT** - Global peer discovery
- **GossipSub** - Efficient pub/sub messaging
- **Circuit Relay v2** - NAT traversal via relays
- **Hole Punching** - Direct NAT traversal
- **AutoRelay** - Automatic relay discovery
- **UPnP/NAT-PMP** - Port forwarding

### Frontend (TypeScript)

**Client Library**:
- ES modules (compiled from TypeScript)
- Promise-based API
- Sequential message processing
- Automatic ack number management

**Distribution**:
- Bundled with demo in executable
- Can be copied out with `cp` command
- Type definitions included

### Build System

**Make-based**:
- TypeScript compilation
- Go build
- Bundle creation
- Demo extraction

**Dependencies**:
- Node.js (for TypeScript compilation)
- Go 1.21+ (for server build)

---

## Performance Characteristics

### Message Latency

**Local Network (mDNS)**:
- Peer discovery: < 1 second
- Protocol message: < 10ms
- PubSub message: < 50ms

**Internet (DHT)**:
- Peer discovery: 2-10 seconds
- Protocol message: 50-500ms (depending on path)
- PubSub message: 100-1000ms

### Scalability

**Per Instance**:
- 100s of concurrent WebSocket connections
- 1000s of peers in DHT
- 100s of active protocol streams
- 10s of active topics

**Limitations**:
- Single process (no horizontal scaling)
- Sequential message processing (serialization point)
- File descriptor limits (open streams)

---

## Security Considerations

### Current Security Model

**Peer Authentication**: None (fully open)
**Access Control**: None (all peers can communicate)
**Encryption**: libp2p provides transport encryption
**Trust Model**: All peers trusted

### Future Security Enhancements

**Recommended Improvements**:
1. Peer authentication mechanism
2. Protocol-level access control
3. Topic subscription authorization
4. Rate limiting per peer
5. Defined threat model

---

## References

- **Design Documents**: `design/` directory
  - CRC cards for each component
  - Sequence diagrams for key workflows
  - Traceability mappings
- **Specifications**: `specs/main.md`
- **API Reference**: `docs/api-reference.md`
- **Developer Guide**: `docs/developer-guide.md`

---

*Last updated: Initial architecture documentation from CRC design*
