# Requirements Documentation

<!-- Source: main.md -->

## Table of Contents

- [Overview](#overview)
- [Business Requirements](#business-requirements)
- [Functional Requirements](#functional-requirements)
- [Non-Functional Requirements](#non-functional-requirements)
- [Technical Constraints](#technical-constraints)

## Overview

**Purpose**: p2p-webapp is a local backend server that enables peer-to-peer web applications with IPFS file storage. It eliminates the need for traditional hosting by embedding a complete peer-to-peer networking stack into a single executable.

**Target Users**: Web application developers building decentralized, peer-to-peer applications

**Key Goals**:
- Provide simple TypeScript API for P2P communication
- Enable IPFS-based file storage and sharing between peers
- Deliver zero-configuration deployment model
- Support automatic peer discovery and NAT traversal

<!-- Source: main.md -->

## Business Requirements

### BR1: Local Backend Model

<!-- Source: main.md -->

**Description**: Replace traditional client-server architecture with local backend that runs on user's machine

**Success Criteria**:
- Single executable runs web server and P2P networking
- No external server infrastructure required
- Applications work offline after initial peer discovery

**Priority**: High

### BR2: Developer Experience

<!-- Source: main.md -->

**Description**: Provide simple, promise-based API that abstracts P2P complexity

**Success Criteria**:
- TypeScript library with type definitions
- Minimal boilerplate code required
- Clear documentation and examples

**Priority**: High

## Functional Requirements

### FR1: Peer Management

<!-- Source: main.md -->

**Description**: Create and manage libp2p peers with persistent identity

**Acceptance Criteria**:
- Generate new peer keys or restore from existing key
- Maintain peer identity across sessions
- Detect and prevent duplicate peer IDs

**Related Requirements**: FR2, FR7

### FR2: Protocol-Based Messaging

<!-- Source: main.md -->

**Description**: Support custom protocol handlers for peer-to-peer communication

**Acceptance Criteria**:
- Register protocol listeners with callbacks
- Send messages using (peer, protocol) addressing
- Deliver messages with ordering guarantees
- Provide delivery confirmation via promises

**Related Requirements**: FR1, NFR2

### FR3: PubSub Messaging

<!-- Source: main.md -->

**Description**: Enable group communication via topic-based publish/subscribe

**Acceptance Criteria**:
- Subscribe to topics with message callbacks
- Publish messages to all topic subscribers
- Monitor peer join/leave events automatically
- List peers subscribed to a topic

**Related Requirements**: FR1, FR15

### FR4: IPFS File Storage

<!-- Source: main.md -->

**Description**: Store and retrieve files using IPFS with HAMTDirectory structure

**Acceptance Criteria**:
- Store files at specified paths
- Create directory hierarchies
- Remove files and empty directories
- Return CIDs for all stored content
- Pin content automatically

**Related Requirements**: FR5, FR6, FR15

### FR5: File Listing

<!-- Source: main.md -->

**Description**: Query file directory structure from local or remote peers

**Acceptance Criteria**:
- List own files with full pathname tree
- Query remote peer's files via reserved protocol
- Return hierarchical structure with CIDs
- Include MIME types for files

**Related Requirements**: FR4, FR6

### FR6: File Retrieval

<!-- Source: main.md -->

**Description**: Retrieve file content by CID from IPFS network

**Acceptance Criteria**:
- Fetch files using CID
- Support both text and binary files
- Return MIME type with content
- Handle directory retrieval

**Related Requirements**: FR4, FR5

### FR7: Configuration System

<!-- Source: main.md -->

**Description**: Support optional TOML configuration for server behavior

**Acceptance Criteria**:
- Load configuration from p2p-webapp.toml
- Support server, HTTP, WebSocket, and behavior settings
- Command-line flags override config file
- Provide sensible defaults for all settings

**Related Requirements**: FR1, FR15

### FR8: Bundle Management

<!-- Source: main.md -->

**Description**: Bundle web applications into single executable

**Acceptance Criteria**:
- Append ZIP archive to binary without compilation
- Extract bundled content to directory
- List bundled files
- Copy bundled files to destination
- Serve content directly from bundle

**Related Requirements**: FR9

### FR9: Web Server

<!-- Source: main.md -->

**Description**: Host web applications with SPA routing support

**Acceptance Criteria**:
- Serve static files with configurable cache headers
- Automatic SPA fallback to index.html
- WebSocket endpoint for client communication
- Security headers (X-Content-Type-Options, X-Frame-Options)
- CORS support (optional)

**Related Requirements**: FR8, FR10

### FR10: Process Management

<!-- Source: main.md -->

**Description**: Track and manage running p2p-webapp instances

**Acceptance Criteria**:
- Register PIDs in shared tracking file
- List all running instances
- Terminate specific instance by PID
- Terminate all instances
- Graceful shutdown with SIGTERM, force with SIGKILL after timeout

**Related Requirements**: NFR4

### FR11: Peer Discovery

<!-- Source: main.md -->
<!-- CRC: crc-Peer.md -->
<!-- Sequence: seq-dht-bootstrap.md -->

**Description**: Automatically discover peers using mDNS and DHT with topic-based advertisement

**Acceptance Criteria**:
- mDNS discovery for local network peers (automatic, zero-config)
- DHT discovery for global peer connectivity
- **DHT bootstrap with operation queuing**: DHT operations queue automatically until routing table populates
- **Topic subscription advertisement via DHT**: Peers advertise topic subscriptions to DHT for global discovery
- **Continuous DHT advertisement**: Re-advertise topic subscriptions periodically to maintain discoverability
- **Automatic topic peer discovery**: Query DHT to discover other peers subscribed to same topics
- Bootstrap using public IPFS DHT nodes

**Implementation Details**:
- **DHT Bootstrap Process**:
  1. Connect to 3+ bootstrap peers from public IPFS DHT nodes
  2. Run DHT.Bootstrap() to populate routing table
  3. Wait up to 30 seconds for routing table to have peers (poll every 500ms)
  4. Signal readiness via channel close when routing table populated
  5. Process any queued DHT operations (advertisement, discovery)
- **DHT Operation Queuing**:
  - Operations (Advertise, FindPeers) queue automatically if DHT not ready
  - Subscribe returns immediately (queuing happens transparently)
  - Operations execute when DHT routing table populated
  - Prevents "failed to find any peer in table" errors
  - Typical bootstrap time: 5-15 seconds, max wait: 30 seconds
- When peer subscribes to a topic (e.g., "chatroom"), automatically:
  1. Queue advertiseTopic() if DHT not ready, execute immediately if ready
  2. Queue discoverTopicPeers() if DHT not ready, execute immediately if ready
  3. Advertise subscription to DHT with periodic re-advertisement (every TTL/2)
  4. Query DHT to discover other peers subscribed to same topic
  5. Attempt connections to discovered peers
- Enables geographically distant peers to find each other via topic interest
- Complements mDNS (local) with DHT (global) for complete coverage

**Timing Expectations**:
- **First operation after peer creation**: May take 10-30 seconds (queued during bootstrap)
- **Subsequent operations**: Immediate (DHT already bootstrapped)
- **Bootstrap success logging**: Visible at verbosity level 2 (`-vv`)

**Related Requirements**: FR3, FR12

### FR12: NAT Traversal

<!-- Source: main.md -->

**Description**: Enable connections between peers behind NATs/firewalls

**Acceptance Criteria**:
- Circuit Relay v2 for relay-based connections
- Hole punching for direct NAT traversal
- AutoRelay to find public relay nodes
- NAT port mapping via UPnP/NAT-PMP

**Related Requirements**: FR11

### FR13: Verbose Logging

<!-- Source: main.md -->

**Description**: Provide detailed logging at multiple verbosity levels

**Acceptance Criteria**:
- Level 1: peer creation, connections, messages
- Level 2: WebSocket details with request IDs
- Level 3: stream operations, discovery, internal state
- Peer aliases for readable output

**Related Requirements**: NFR3

### FR14: Signal Handling

<!-- Source: main.md -->

**Description**: Graceful shutdown on SIGHUP, SIGINT, SIGTERM

**Acceptance Criteria**:
- Stop accepting new connections
- Close active WebSocket connections
- Stop all peers and close streams
- Unregister PID from tracking
- Clean exit without data loss

**Related Requirements**: FR10, NFR4

### FR15: File Availability Notifications

<!-- Source: main.md (FR15: File Update Notifications) -->

**Description**: Notify subscribed peers when file availability changes

**Acceptance Criteria**:
- Publish notification to configured topic after storeFile/removeFile
- Only publish if peer is subscribed to the notification topic
- Include peer ID in notification message
- Applications can refresh file lists automatically when peer's files change

**Related Requirements**: FR3, FR4, FR7

**Message Format**: `{"type":"p2p-webapp-file-update","peer":"<peerID>"}`

**Privacy Design**: Opt-in by default - notifications only published when both conditions met:
1. `fileUpdateNotifyTopic` is configured in settings
2. Peer is subscribed to that topic

### FR16: Connection Management

<!-- Source: main.md (Connection Management) -->

**Description**: Provide explicit control over peer connection priorities using libp2p's BasicConnMgr

**Acceptance Criteria**:
- `addPeers(peerIds)` method to protect and tag peer connections
- `removePeers(peerIds)` method to unprotect and untag peer connections
- Protection prevents connection manager from closing connections
- Tagging assigns priority value (100) to protected peers
- Best-effort connection attempts for peers not already connected
- Silently skip invalid peer IDs or connection failures

**Related Requirements**: FR11, FR12

**Use Cases**:
- Maintain connections to critical relay nodes
- Ensure application-specific peers stay connected
- Control peer connections in testing scenarios

**Implementation Details**:
- Uses libp2p BasicConnMgr (`github.com/libp2p/go-libp2p/p2p/net/connmgr`)
- Accessed via `host.ConnManager()` methods
- Tag name: "connected", priority value: 100
- `removePeers` does NOT disconnect peers, only removes protection/priority

## Non-Functional Requirements

### NFR1: Zero Configuration

<!-- Source: main.md -->

**Description**: Work out of the box with sensible defaults

**Metric**: Steps required to run application

**Target**:
- Extract/build: 1 command
- Run: 1 command with no required flags

### NFR2: Message Ordering

<!-- Source: main.md -->

**Description**: Guarantee message delivery order within protocols

**Metric**: Message sequence preservation

**Target**: 100% ordering guarantee for sequential processing

**Trade-off**: Serialized processing reduces maximum throughput but ensures correctness

### NFR3: Developer Experience

<!-- Source: main.md -->

**Description**: Simple, intuitive API with clear error messages

**Metric**:
- API complexity (method count, parameter count)
- Error message clarity

**Target**:
- Core API < 15 methods
- All errors include context and suggested fixes

### NFR4: Reliability

<!-- Source: main.md -->

**Description**: Robust error handling and graceful degradation

**Metric**:
- Error recovery success rate
- Data loss on abnormal termination

**Target**:
- 100% recovery from network errors
- Zero data loss on SIGTERM shutdown
- Stale process cleanup on next start

### NFR5: Performance

<!-- Source: main.md -->

**Description**: Efficient resource usage for local backend

**Metric**:
- Memory footprint
- CPU usage
- Startup time

**Target**:
- < 100MB memory per peer
- < 5% CPU idle
- < 2 second startup

### NFR6: Security

<!-- Source: main.md -->

**Description**: Secure by default with minimal attack surface

**Metric**: Security headers, input validation

**Target**:
- All inputs validated
- Security headers on all responses
- Private address support gated behind explicit option

## Technical Constraints

<!-- Source: main.md -->

**Technology Stack**:
- **Go 1.21+**: Server implementation, libp2p/IPFS integration
  - Reason: Mature libp2p and IPFS ecosystem
- **TypeScript**: Client library with type safety
  - Reason: Developer experience, IDE support
- **WebSocket**: Client-server communication
  - Reason: Bidirectional, event-driven messaging
- **libp2p**: P2P networking foundation
  - Reason: Production-ready P2P stack with discovery, NAT traversal
- **IPFS-lite**: Content-addressed storage
  - Reason: Lightweight IPFS implementation without full daemon

**Limitations**:
- Requires Go 1.21+ for build
- Client library requires modern browser with ES modules support
- mDNS discovery only works on same local network
- DHT bootstrap requires initial internet connectivity
- Circuit relay performance limited by relay node bandwidth

**Configuration**:
- Optional TOML configuration file at site root
- Command-line flags take precedence over config file
- All settings have sensible defaults
- File update notifications disabled by default (opt-in via configuration)

---

*Last updated: 2025-11-24 - Added DHT topic discovery details to FR11*
