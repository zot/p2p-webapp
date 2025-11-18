# Gap Analysis

**Analysis of differences between specification and design**

## Type A: Spec-Required But Missing (Critical)

*No critical gaps identified*

All requirements from main.md are covered by the design:
- ✓ Server startup and lifecycle
- ✓ WebSocket JSON-RPC protocol
- ✓ Peer creation and management
- ✓ Protocol-based messaging with virtual connections
- ✓ PubSub with GossipSub
- ✓ Peer discovery (mDNS + DHT)
- ✓ NAT traversal (Circuit Relay, hole punching, AutoRelay, UPnP)
- ✓ Bundle management (create, extract, ls, cp)
- ✓ Process tracking (ps, kill, killall)
- ✓ SPA routing support
- ✓ Client library with message queueing
- ✓ Message acknowledgment with callbacks

---

## Type B: Design Improvements (Code Quality)

### B1: Error Handling Patterns

**Issue**: Spec describes error responses but doesn't specify error handling strategy

**Impact**: Medium - affects reliability and user experience

**Recommendation**:
- Define error codes and error types
- Specify retry strategies for transient failures
- Document error propagation patterns
- Specify timeout values for operations

### B2: Connection Lifecycle Details

**Issue**: Virtual connection model described but cleanup strategy not specified

**Impact**: Medium - could lead to resource leaks

**Recommendation**:
- Define stream timeout and cleanup policy
- Specify max concurrent streams per peer
- Document connection pool size limits
- Define idle connection timeout

### B3: Security Considerations

**Issue**: Spec doesn't address authentication or authorization

**Impact**: High - current design is fully open

**Recommendation**:
- Consider peer authentication mechanism
- Define access control for protocols/topics
- Specify encryption requirements
- Document threat model

### B4: Rate Limiting

**Issue**: No rate limiting mentioned for message sending

**Impact**: Medium - potential for abuse or resource exhaustion

**Recommendation**:
- Define rate limits per peer/protocol
- Specify backpressure handling
- Document queue size limits

### B5: Verbose Logging Consistency

**Issue**: Verbosity levels described but log format not specified

**Impact**: Low - affects debugging experience

**Recommendation**:
- Define structured logging format
- Specify log levels consistently
- Document what each verbosity level logs

### B6: IPFS File Sharing Not Implemented

**Issue**: ipfs-lite imported and `ipfs/` directory structure exists, but no file sharing functionality

**Impact**: Medium - library over-specified for actual usage, confusing to users

**Current State**:
- ipfs-lite only used for DHT operations and relay discovery
- `ipfs/` directory in bundle format but no API to add/get/pin files
- No CID-based content addressing in WebSocket protocol
- Missing commands: `addFile()`, `getFile()`, `pinFile()`, `listFiles()`

**Recommendation**:
- **Option A**: Implement full IPFS file sharing API
  - Add WebSocket commands for file operations
  - Implement CID-based content retrieval
  - Add pinning and unpinning operations
  - Provide progress callbacks for large files
- **Option B**: Remove ipfs-lite dependency
  - Replace with direct libp2p Kademlia DHT usage
  - Remove `ipfs/` directory from bundle format
  - Simplify dependencies and reduce confusion
  - Keep only the DHT functionality actually being used

**Decision Required**: Is file sharing a planned feature or unnecessary complexity?

---

## Type C: Enhancements (Nice-to-Have)

### C1: Metrics and Monitoring

**Enhancement**: Add metrics collection for operational visibility

**Benefit**: Better production observability

**Suggestions**:
- Message counts per protocol/topic
- Connection duration metrics
- Peer discovery latency
- Bundle operation metrics

### C2: Configuration Management

**Enhancement**: Allow configuration via file in addition to flags

**Benefit**: Easier deployment and customization

**Suggestions**:
- YAML/JSON config file support
- Environment variable overrides
- Configuration validation
- Default configuration documentation

### C3: Health Checks

**Enhancement**: Add health check endpoint

**Benefit**: Better integration with orchestration systems

**Suggestions**:
- HTTP /health endpoint
- Peer connectivity status
- WebSocket connection count
- Discovery status

### C4: Performance Optimization

**Enhancement**: Optimize for high-throughput scenarios

**Benefit**: Better scalability

**Suggestions**:
- Stream multiplexing optimization
- Message batching for PubSub
- Zero-copy buffer management
- Connection pooling tuning

### C5: Developer Experience

**Enhancement**: Improve debugging and development tools

**Benefit**: Easier development and troubleshooting

**Suggestions**:
- Interactive protocol testing tool
- Connection visualizer
- Message inspector/debugger
- Performance profiling integration

---

## Design Decisions

### D1: Virtual Connection Model

**Decision**: Use (peer, protocol) addressing instead of explicit connection management

**Rationale**: Simplifies client API, server manages complexity

**Trade-offs**:
- Pro: Cleaner client code, easier to use
- Con: Less control over connection lifecycle
- Con: Potential for hidden resource usage

### D2: Sequential Message Processing

**Decision**: Queue server-initiated messages, process sequentially

**Rationale**: Preserves message ordering, simpler concurrency model

**Trade-offs**:
- Pro: Guaranteed ordering for peerData
- Pro: No race conditions in callbacks
- Con: Slower processing for independent messages
- Con: Head-of-line blocking risk

### D3: Bundled Distribution

**Decision**: ZIP append technique for bundling

**Rationale**: No compilation tools needed, works cross-platform

**Trade-offs**:
- Pro: Simple deployment (single executable)
- Pro: No external dependencies
- Con: Slightly larger binary size
- Con: Can't partially extract

---

## Implementation Patterns Observed

### Pattern 1: Alias Generation

**Purpose**: Human-readable peer identifiers for logging

**Implementation**: Sequential naming (peer-a, peer-b, ...)

**Benefits**: Easy to follow in logs, especially with verbose output

### Pattern 2: Protocol-Based Routing

**Purpose**: Route messages by protocol instead of connection

**Implementation**: Map[protocol]callback in client, Map[peerID:protocol]stream in server

**Benefits**: Decouples protocol from connection management

### Pattern 3: File Locking for PID Tracking

**Purpose**: Prevent race conditions in multi-instance scenarios

**Implementation**: File lock during read/write operations

**Benefits**: Safe concurrent access without database

---

*Last updated: Added B6 - IPFS file sharing gap identified*
