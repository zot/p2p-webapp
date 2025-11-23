# p2p-webapp Developer Guide

**Complete guide for developers working with the p2p-webapp codebase**

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Project Structure](#project-structure)
3. [Building](#building)
4. [Testing](#testing)
5. [Development Workflow](#development-workflow)
6. [Configuration](#configuration)
7. [Design Methodology](#design-methodology)
8. [Contributing](#contributing)
9. [Debugging](#debugging)
10. [Best Practices](#best-practices)

---

## Getting Started

### Prerequisites

- **Go 1.21+** - Server implementation
- **Node.js 18+** - TypeScript client library compilation
- **Make** - Build automation
- **Git** - Version control

### Clone and Build

```bash
git clone https://github.com/your-org/p2p-webapp.git
cd p2p-webapp

# Build everything (installs dependencies, compiles TypeScript, builds binary)
make build

# Or just run the demo
make demo
```

### First Run

```bash
# Run server with bundled demo
./p2p-webapp

# Or extract demo and run from directory
mkdir /tmp/demo
cd /tmp/demo
/path/to/p2p-webapp extract
./p2p-webapp --dir .

# With verbose logging
./p2p-webapp -vv
```

---

## Project Structure

```
p2p-webapp/
├── cmd/
│   └── p2p-webapp/         # Main entry point, command router
│       └── main.go
├── internal/               # Internal packages (not exported)
│   ├── server/             # Server orchestrator, WebServer
│   │   ├── server.go
│   │   └── webserver.go
│   ├── websocket/          # WebSocket handler, JSON-RPC
│   │   └── handler.go
│   ├── peer/               # PeerManager, libp2p integration
│   │   └── manager.go
│   ├── bundle/             # Bundle manager, ZIP operations
│   │   └── manager.go
│   └── process/            # Process tracker, PID management
│       └── tracker.go
├── pkg/                    # Public packages (exportable)
│   └── client/             # TypeScript client library
│       └── src/
│           ├── index.ts    # Main client implementation
│           └── index.test.ts
├── internal/commands/
│   └── demo/               # Bundled demo application
│       ├── index.html
│       ├── styles.css
│       └── app.js
├── design/                 # Level 2 CRC design documents
│   ├── architecture.md
│   ├── crc-*.md           # CRC cards for each class
│   ├── seq-*.md           # Sequence diagrams
│   ├── traceability.md
│   └── gaps.md
├── specs/                  # Level 1 specifications
│   └── main.md
├── docs/                   # Documentation
│   ├── architecture.md
│   ├── api-reference.md
│   └── developer-guide.md
├── Makefile               # Build automation
├── go.mod                 # Go dependencies
├── go.sum
└── README.md              # Project overview
```

---

## Building

### Build Targets

```bash
# Build everything (default)
make build
# - Installs npm dependencies if needed
# - Compiles TypeScript client
# - Builds Go binary
# - Bundles demo into binary

# Clean all build artifacts
make clean
# - Removes node_modules, build/, p2p-webapp binary

# Run demo (builds if needed)
make demo
# - Extracts demo to temp directory
# - Runs p2p-webapp
# - Opens browser

# Build client library only
make client
# - Compiles TypeScript to ES modules
# - Outputs to pkg/client/build/

# Build Go binary only (without bundling)
go build -o p2p-webapp cmd/p2p-webapp/main.go

# Create custom bundle
./p2p-webapp bundle my-site/ -o my-app
```

### Build Process Details

**Step 1: TypeScript Compilation**
```bash
cd pkg/client
npm install  # Only if node_modules missing
npm run build
# → Outputs to pkg/client/build/
```

**Step 2: Copy Client to Demo**
```bash
cp pkg/client/build/* internal/commands/demo/
```

**Step 3: Build Go Binary**
```bash
go build -o p2p-webapp-temp cmd/p2p-webapp/main.go
```

**Step 4: Bundle Demo**
```bash
./p2p-webapp-temp bundle internal/commands/demo -o p2p-webapp
```

Result: Single executable with demo bundled, ready to distribute.

---

## Testing

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/peer

# With coverage
go test -cover ./...

# With verbose output
go test -v ./...

# Client library tests
cd pkg/client
npm test
```

### Writing Tests

**Go Tests** (follow `*_test.go` convention):

```go
// internal/peer/manager_test.go
package peer

import "testing"

func TestPeerCreation(t *testing.T) {
    // Arrange
    manager := NewPeerManager()

    // Act
    peer, err := manager.CreatePeer(nil)

    // Assert
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if peer == nil {
        t.Error("Expected peer, got nil")
    }
}
```

**TypeScript Tests** (Jest):

```typescript
// pkg/client/src/index.test.ts
import { connect } from './index';

test('connect returns peerID and peerKey', async () => {
  // Mock WebSocket
  global.WebSocket = jest.fn();

  const [peerID, peerKey] = await connect();

  expect(peerID).toBeTruthy();
  expect(peerKey).toBeTruthy();
});
```

### Integration Testing

**With Playwright MCP**:

```bash
# Start server in background
./p2p-webapp -p 8080 &
PID=$!

# Run tests
# (use Playwright to test browser interactions)

# Cleanup
kill $PID
```

**Important**: Always check for running instances before testing:

```bash
# Check for running instances
pgrep -a p2p-webapp

# Kill all instances
./p2p-webapp killall

# Or kill specific PID
./p2p-webapp kill <PID>
```

---

## Development Workflow

### Working on a Feature

**1. Check Design First**

Before coding, check if design exists:
```bash
ls design/crc-*.md   # CRC cards
ls design/seq-*.md   # Sequences
cat design/architecture.md
```

If no design, create it first:
```bash
# Add spec to specs/
echo "# My Feature" > specs/my-feature.md

# Generate design (if CRC framework installed)
# Use designer agent or create manually
```

**2. Find Implementation Location**

Check traceability:
```bash
cat design/traceability.md
# Find which file implements the component
```

**3. Add Traceability Comments**

```go
// internal/peer/manager.go

// CRC: design/crc-PeerManager.md
// Spec: specs/main.md
// Sequences: design/seq-peer-creation.md, design/seq-protocol-communication.md

package peer

// PeerManager manages libp2p peers and discovery.
// CRC: design/crc-PeerManager.md
type PeerManager struct {
    // ...
}

// CreatePeer creates new libp2p peer with given or fresh key.
// Sequence: design/seq-peer-creation.md
func (pm *PeerManager) CreatePeer(peerKey string) (*Peer, error) {
    // ...
}
```

**4. Implement with Tests**

```bash
# Write test first (TDD)
vim internal/peer/manager_test.go

# Implement feature
vim internal/peer/manager.go

# Run tests
go test ./internal/peer

# Check coverage
go test -cover ./internal/peer
```

**5. Update Design if Needed**

If implementation differs from design:
```bash
# Update CRC card
vim design/crc-PeerManager.md

# Update gaps.md with discovered issues
vim design/gaps.md
```

### Local Development Loop

**Terminal 1: Server**
```bash
# Run with high verbosity
./p2p-webapp -vvv

# Or with auto-rebuild (using entr or similar)
ls **/*.go | entr -r make run
```

**Terminal 2: Client Development**
```bash
cd pkg/client

# Watch mode
npm run build -- --watch

# In another terminal, run tests
npm test -- --watch
```

**Terminal 3: Testing**
```bash
# Manual testing
curl http://localhost:10000/

# Or use browser DevTools console
# Access client library in demo
```

---

## Configuration

### Configuration File

p2p-webapp supports optional configuration via `p2p-webapp.toml` file:

**Location**:
- **Directory mode**: Place in base directory (same level as html/, ipfs/, storage/)
- **Bundle mode**: Include in root before bundling

**Precedence**: Command-line flags → Config file → Defaults

### File Update Notifications

<!-- Spec: main.md (FR15: File Update Notifications) -->

**Purpose**: Automatically notify subscribed peers when files change

**Configuration** (`p2p-webapp.toml`):
```toml
[p2p]
# Optional topic for file availability notifications
# Disabled by default (empty string = disabled)
fileUpdateNotifyTopic = "chatroom"
```

**Behavior**:
- When configured and peer subscribed to topic: publishes notification after `storeFile()` / `removeFile()`
- Message format: `{"type":"p2p-webapp-file-update","peer":"<peerID>"}`
- Privacy-friendly: only publishes if BOTH conditions met:
  1. `fileUpdateNotifyTopic` is set in config
  2. Peer is subscribed to that topic

**Use Cases**:
- **Automatic file list refresh**: When viewing a peer's files, refresh list when they update
- **Collaboration awareness**: Notify team members of file changes
- **Sync triggers**: Use as signal to synchronize content

**Demo Example**:

The bundled demo uses this feature for automatic file list updates:

```javascript
// Subscribe to room topic (doubles as file notification topic)
await client.subscribe(ROOM_TOPIC, (senderPeerID, data) => {
    // Check for file update notification
    if (data.type === 'p2p-webapp-file-update' && data.peer) {
        // If viewing this peer's files, refresh the list
        if (currentFilePeerID === data.peer && currentTab === 'files') {
            client.listFiles(data.peer).catch(err => {
                console.error('Error refreshing file list:', err);
            });
        }
        return; // Don't process as chat message
    }

    // Handle normal chat message...
});
```

**Implementation Notes**:
- Notification published in `publishFileUpdateNotification()` method (internal/peer/manager.go)
- Called after successful file storage/removal
- No error if publish fails (non-critical operation)
- Applications can use separate topics for chat and notifications if desired

**Configuration in Code**:

When creating PeerManager:
```go
// internal/server/server.go
fileUpdateNotifyTopic := settings.P2P.FileUpdateNotifyTopic
manager, err := peer.NewManager(ctx, bootstrapHost, ipfsPeer,
    verbosity, fileUpdateNotifyTopic)
```

### Other Configuration Options

See `docs/examples/p2p-webapp.toml` for complete configuration reference:

- **[server]**: Port, port range, header size, timeouts
- **[http]**: Cache control, security headers, CORS
- **[websocket]**: Origin validation, buffer sizes
- **[behavior]**: Auto-exit, auto-open browser, linger, verbosity
- **[files]**: Index file, SPA fallback
- **[p2p]**: Protocol name, file update notifications

**Example Configurations**:

Development (disable caching):
```toml
[http]
cacheControl = "no-cache, no-store, must-revalidate"

[behavior]
verbosity = 2  # Show WebSocket details
```

Production (enable caching):
```toml
[http]
cacheControl = "public, max-age=3600, immutable"

[behavior]
verbosity = 0  # Minimal output
autoOpenBrowser = false
```

### Connection Management

<!-- Spec: main.md (Connection Management) -->
<!-- CRC: crc-Peer.md, crc-PeerManager.md -->
<!-- Sequence: seq-add-peers.md, seq-remove-peers.md -->

p2p-webapp provides explicit control over peer connection priorities through libp2p's **BasicConnMgr** (Connection Manager).

#### Overview

While peer discovery (mDNS/DHT) and connections happen automatically, applications can protect critical connections and assign priority values using the connection management API:

- **`addPeers(peerIds)`**: Protect and tag peer connections
- **`removePeers(peerIds)`**: Unprotect and untag peer connections

#### BasicConnMgr

The libp2p host includes a **BasicConnMgr** from `github.com/libp2p/go-libp2p/p2p/net/connmgr`, accessed via `host.ConnManager()`:

- **Protection**: Prevents the connection manager from closing specific peer connections
  - `ConnManager().Protect(peerID, tag)` - Mark connection as protected
  - `ConnManager().Unprotect(peerID, tag)` - Remove protection

- **Tagging**: Assigns priority values to peers for connection management decisions
  - `ConnManager().TagPeer(peerID, tag, value)` - Assign priority (higher = more important)
  - `ConnManager().UntagPeer(peerID, tag)` - Remove priority tag

p2p-webapp uses tag name **"connected"** and priority value **100** for protected peers.

#### Implementation

**Client API** (TypeScript):
```typescript
// Protect important peer connections
await addPeers(['12D3KooW...', '12D3KooX...']);

// Later: remove protection
await removePeers(['12D3KooW...']);
```

**Server Implementation** (Go):
```go
// internal/peer/manager.go
func (p *Peer) AddPeers(targetPeerIDs []string) error {
    connMgr := p.host.ConnManager()

    for _, peerIDStr := range targetPeerIDs {
        targetPeerID, _ := peer.Decode(peerIDStr)

        // Protect connection from being closed
        connMgr.Protect(targetPeerID, "connected")

        // Tag peer with priority value
        connMgr.TagPeer(targetPeerID, "connected", 100)

        // Attempt connection (best-effort)
        // ...
    }
    return nil
}
```

#### Use Cases

- **Relay nodes**: Protect connections to circuit relay servers
- **Application peers**: Ensure critical app peers stay connected
- **Known servers**: Maintain connections to trusted infrastructure
- **Testing**: Control peer connections in test scenarios

#### Notes

- Protection and tagging work on peer IDs whether connected or not
- Connection attempts are best-effort (failures ignored)
- `removePeers` does NOT disconnect peers, only removes protection/priority
- Silently skips invalid peer IDs
- Default BasicConnMgr has unlimited connections (no pruning unless configured)

---

## Design Methodology

### CRC Methodology (3-Level Spec-Driven Development)

This project uses the CRC (Class-Responsibility-Collaboration) framework:

```
Level 1: Human Specs (specs/*.md)
   ↓ designer agent
Level 2: Design Models (design/*.md)
   ↓ implementation
Level 3: Code (internal/**, pkg/**)
```

### Design Files

**CRC Cards** (`design/crc-*.md`):
- One per class/component
- Knows: Data/attributes
- Does: Behaviors/actions
- Collaborators: What it works with

**Sequence Diagrams** (`design/seq-*.md`):
- PlantUML ASCII art diagrams
- Show interactions over time
- One per scenario/use case

**Architecture** (`design/architecture.md`):
- Entry point to design
- Organizes components into systems
- Shows cross-cutting concerns

**Traceability** (`design/traceability.md`):
- Maps specs → design → code
- Checkboxes for tracking implementation
- Ensures complete coverage

### Adding Traceability Comments

**File Header**:
```go
// CRC: design/crc-ClassName.md
// Spec: specs/spec-name.md
// Sequences: design/seq-scenario1.md, design/seq-scenario2.md
```

**Class/Struct**:
```go
// ClassName does X.
// CRC: design/crc-ClassName.md
type ClassName struct { ... }
```

**Methods**:
```go
// MethodName does Y.
// Sequence: design/seq-scenario.md
func (c *ClassName) MethodName() { ... }
```

---

## Contributing

### Code Style

**Go**:
- Follow standard Go conventions
- Use `gofmt` for formatting
- Document all exported symbols
- Add traceability comments

**TypeScript**:
- Use Prettier for formatting (configured in package.json)
- Prefer explicit types over `any`
- Document public API functions

### Commit Messages

Follow conventional commit format:

```
<type>: <description>

[optional body]

[optional footer]
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Build/tooling changes

**Examples**:
```
feat: add message acknowledgment with callbacks

Implements optional delivery confirmation for send().
Client library manages ack numbers automatically.

CRC: design/crc-PeerManager.md
Sequence: design/seq-protocol-communication.md

fix: prevent duplicate peer IDs across tabs

Checks for existing peerID before registering.
Returns error if duplicate detected.

docs: add API reference for TypeScript client

Complete documentation of connect, start, send, subscribe APIs
with examples and error handling patterns.
```

### Pull Request Process

1. Create feature branch from `main`
2. Add/update design documents if needed
3. Implement with traceability comments
4. Add/update tests
5. Run full test suite
6. Update documentation
7. Create PR with clear description
8. Link to related design documents

---

## Debugging

### Verbose Logging

**Level 1** (`-v`):
```bash
./p2p-webapp -v
```
Logs:
- Peer creation
- Connections
- Messages sent/received

**Level 2** (`-vv`):
```bash
./p2p-webapp -vv
```
Adds:
- WebSocket message details
- Request IDs
- Protocol operations

**Level 3** (`-vvv`):
```bash
./p2p-webapp -vvv
```
Adds:
- Stream operations
- Discovery details
- Internal state changes

**Output Format**:
```
[peer-a] Connected to peer-b
[peer-a] Sent message on protocol 'chat'
[peer-b] Received message from peer-a
```

### Common Issues

**Issue: Port already in use**
```bash
# Check what's using port
ss -tuln | grep 10000

# Use different port
./p2p-webapp -p 10001
```

**Issue: Peer discovery not working**
```bash
# Check firewall
# On Linux:
sudo ufw allow 10000

# Check if peers on same network (for mDNS)
ping <other-device>

# Check verbose logs for DHT bootstrap
./p2p-webapp -vvv | grep DHT
```

**Issue: Multiple tabs, duplicate peer error**
```bash
# Solution: Use unique peer keys per tab
# Or: Close other tabs
# Or: Clear localStorage
```

**Issue: Bundle not found**
```bash
# Check if binary is bundled
./p2p-webapp ls

# If not, create bundle
make build
# Or: ./p2p-webapp bundle my-site/ -o p2p-webapp
```

**Issue: WebSocket connection refused**
```bash
# Check server is running
./p2p-webapp ps

# Check port
netstat -tuln | grep 10000

# Check browser console for errors
# (Right-click → Inspect → Console)
```

### Debugging Tools

**Process Management**:
```bash
# List running instances
./p2p-webapp ps -v

# Kill stuck instance (graceful shutdown with SIGTERM, force kill after 5s if needed)
./p2p-webapp kill <PID>

# Kill all instances (graceful shutdown with SIGTERM, force kill after 5s if needed)
./p2p-webapp killall
```

**Note**: The `kill` and `killall` commands use graceful shutdown:
1. First send SIGTERM for clean shutdown
2. Wait up to 5 seconds for process to terminate
3. If still running, force kill with SIGKILL

**Network Inspection**:
```bash
# Check libp2p connections
# (No built-in tool, use verbose logging)
./p2p-webapp -vvv 2>&1 | grep "connection"

# Check WebSocket in browser
# DevTools → Network → WS → Click connection → Messages
```

**State Inspection**:
```bash
# Check storage directory
ls -la .p2p-webapp-storage/

# Check PID file
cat /tmp/.p2p-webapp

# Check bundle contents
./p2p-webapp ls
```

---

## Best Practices

### Architecture

1. **Follow SOLID principles**
   - Each component has single responsibility
   - Depend on abstractions, not implementations

2. **Maintain traceability**
   - Add design references to all code
   - Update design when implementation changes

3. **Keep virtual connection model**
   - Don't expose stream details to client
   - Server manages all lifecycle

### Code Quality

1. **Test coverage**
   - Unit tests for all components
   - Integration tests for workflows
   - Aim for 80%+ coverage

2. **Error handling**
   - Return errors, don't panic
   - Provide context in error messages
   - Log at appropriate levels

3. **Documentation**
   - Document all exported symbols
   - Add examples for complex APIs
   - Keep docs in sync with code

### Synchronization Hygiene

**CRITICAL**: Proper synchronization is essential to prevent deadlocks, race conditions, and performance issues in concurrent Go code.

#### Core Principles

##### 1. Centralize Locking Around Resources

**Rule**: Only the object that owns a resource should lock and unlock access to it.

```go
// ✅ GOOD: Manager owns peers map and controls its lock
func (m *Manager) GetPeer(id string) (*Peer, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    peer, exists := m.peers[id]
    if !exists {
        return nil, fmt.Errorf("peer not found")
    }
    return peer, nil
}

// ❌ BAD: Method leaves resource locked
func (m *Manager) LockPeers() {
    m.mu.Lock()  // Caller must remember to unlock!
}
```

**Exception**: Component lock systems (like `pidfile_unix.go` providing primitives for `pidfile.go`) are acceptable.

##### 2. Never Hold Locks While Calling Other Objects

```go
// ✅ GOOD: Copy data under lock, then call external method
func (m *Manager) NotifyPeers(msg string) {
    m.mu.RLock()
    peersCopy := make([]*Peer, 0, len(m.peers))
    for _, p := range m.peers {
        peersCopy = append(peersCopy, p)
    }
    m.mu.RUnlock()

    // Now safe to call methods on peers
    for _, p := range peersCopy {
        p.Notify(msg)
    }
}

// ❌ BAD: Holding lock while calling peer methods
func (m *Manager) NotifyPeers(msg string) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    for _, p := range m.peers {
        p.Notify(msg)  // What if Notify() tries to lock something?
    }
}
```

##### 3. Minimize Lock Duration

**Pattern**: Lock → Extract Data → Unlock → Process

```go
// ✅ GOOD: Minimal lock time for queue processing
func (q *Queue) ProcessMessages() {
    for {
        // Lock only to dequeue
        q.mu.Lock()
        if len(q.messages) == 0 {
            q.mu.Unlock()
            break
        }
        msg := q.messages[0]
        q.messages = q.messages[1:]
        q.mu.Unlock()

        // Process without holding lock
        processMessage(msg)
    }
}

// ❌ BAD: Holding lock during processing
func (q *Queue) ProcessMessages() {
    q.mu.Lock()
    defer q.mu.Unlock()
    for _, msg := range q.messages {
        processMessage(msg)  // Lock held entire time!
    }
    q.messages = nil
}
```

#### Anti-Patterns to Avoid

##### ❌ Holding Locks During I/O Operations

```go
// ❌ BAD: Network I/O under lock
func (p *Peer) StoreFile(path string, content []byte) error {
    p.mu.Lock()
    defer p.mu.Unlock()

    // IPFS AddFile does network I/O!
    node, err := p.ipfs.AddFile(content)  // DEADLOCK RISK
    if err != nil {
        return err
    }
    p.files[path] = node
    return nil
}

// ✅ GOOD: I/O without lock, then update under lock
func (p *Peer) StoreFile(path string, content []byte) error {
    // Do I/O without holding lock
    node, err := p.ipfs.AddFile(content)
    if err != nil {
        return err
    }

    // Lock only to update map
    p.mu.Lock()
    p.files[path] = node
    p.mu.Unlock()
    return nil
}
```

##### ❌ Nested Locking

```go
// ❌ BAD: Acquiring another lock while holding one
func (m *Manager) ConnectPeers(id1, id2 string) error {
    m.mu.RLock()  // Manager lock
    p1 := m.peers[id1]

    p1.mu.Lock()  // Peer lock while holding manager lock
    // ... DEADLOCK RISK if another goroutine locks in different order
    p1.mu.Unlock()

    m.mu.RUnlock()
    return nil
}

// ✅ GOOD: Get references, then lock individually
func (m *Manager) ConnectPeers(id1, id2 string) error {
    m.mu.RLock()
    p1 := m.peers[id1]
    p2 := m.peers[id2]
    m.mu.RUnlock()

    // Now safe to work with peers
    if err := p1.Connect(p2); err != nil {
        return err
    }
    return nil
}
```

##### ❌ Long Critical Sections

```go
// ❌ BAD: Lock held during entire peer creation
func (m *Manager) CreatePeer() (*Peer, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    // libp2p host creation can take seconds!
    host, err := libp2p.New(...)

    peer := &Peer{host: host}
    m.peers[peer.ID()] = peer
    return peer, nil
}

// ✅ GOOD: Lock only for map update
func (m *Manager) CreatePeer() (*Peer, error) {
    // Do expensive work without lock
    host, err := libp2p.New(...)
    if err != nil {
        return nil, err
    }

    peer := &Peer{host: host}

    // Lock only to add to map
    m.mu.Lock()
    m.peers[peer.ID()] = peer
    m.mu.Unlock()

    return peer, nil
}
```

#### When to Use `withResource` Patterns

Avoid `withResource(func)` patterns when possible, as they maximize lock duration:

```go
// ⚠️ ACCEPTABLE BUT NOT PREFERRED
func (m *Manager) WithPeers(fn func(map[string]*Peer)) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    fn(m.peers)  // Caller's function runs under lock
}

// ✅ BETTER: Return a copy
func (m *Manager) GetPeersCopy() map[string]*Peer {
    m.mu.RLock()
    defer m.mu.RUnlock()

    copy := make(map[string]*Peer, len(m.peers))
    for k, v := range m.peers {
        copy[k] = v
    }
    return copy
}
```

Use `withResource` only when:
1. Copying data is too expensive
2. Caller needs transactional semantics
3. The callback is guaranteed to be fast (no I/O, no external calls)

#### Code Review Checklist

When reviewing code with locks, check:

- [ ] No I/O operations while holding lock
- [ ] No network calls while holding lock
- [ ] No calls to methods on other objects while holding lock
- [ ] No nested lock acquisitions
- [ ] Lock held for minimal time
- [ ] `defer unlock()` used appropriately
- [ ] No locks left held on method exit (except component systems)
- [ ] Consider: could this deadlock with another goroutine?

#### Common Patterns

**Pattern 1: Two-Phase Update**
```go
// Phase 1: Prepare data (no lock)
newData := prepareExpensiveData()

// Phase 2: Update (minimal lock)
m.mu.Lock()
m.data = newData
m.mu.Unlock()
```

**Pattern 2: Copy-Modify-Replace**
```go
m.mu.RLock()
oldSlice := m.items
m.mu.RUnlock()

// Work with copy
newSlice := append([]Item{}, oldSlice...)
newSlice = append(newSlice, newItem)

m.mu.Lock()
m.items = newSlice
m.mu.Unlock()
```

**Pattern 3: Extract-Process**
```go
// Extract work under lock
m.mu.Lock()
workItems := m.queue
m.queue = nil
m.mu.Unlock()

// Process without lock
for _, item := range workItems {
    process(item)
}
```

### Performance

1. **Sequential processing trade-off**
   - Accept serialization for ordering guarantees
   - Profile before optimizing

2. **Stream reuse**
   - Virtual connections automatically reuse streams
   - Don't create streams manually

3. **Memory management**
   - Close connections properly
   - Clean up listeners on disconnect

### Security

1. **Never skip validation**
   - Validate all client inputs
   - Check protocol started before sending

2. **Rate limiting** (future)
   - Consider implementing rate limits
   - Prevent abuse and resource exhaustion

3. **Access control** (future)
   - Consider adding authentication
   - Define threat model clearly

---

## References

- **Requirements**: `docs/requirements.md`
- **Design**: `docs/design.md`
- **Architecture**: `docs/architecture.md`
- **API Reference**: `docs/api-reference.md`
- **Design Documents**: `design/` directory
- **Specifications**: `specs/main.md`
- **CRC Methodology**: `.claude/doc/crc.md`

---

## Quick Reference

### Common Commands

```bash
# Build
make build
make clean

# Run
./p2p-webapp              # Bundled mode
./p2p-webapp --dir .      # Directory mode
./p2p-webapp -vvv         # With verbose logging

# Test
go test ./...
cd pkg/client && npm test

# Bundle
./p2p-webapp extract
./p2p-webapp bundle site/ -o app
./p2p-webapp ls
./p2p-webapp cp client.* dest/

# Process
./p2p-webapp ps
./p2p-webapp kill <PID>
./p2p-webapp killall

# Info
./p2p-webapp version
./p2p-webapp --help
```

### File Locations

- Source: `cmd/`, `internal/`, `pkg/`
- Design: `design/`
- Specs: `specs/`
- Docs: `docs/`
- Demo: `internal/commands/demo/`
- Tests: `*_test.go`, `*.test.ts`

---

*Last updated: Initial developer guide from CRC design*
