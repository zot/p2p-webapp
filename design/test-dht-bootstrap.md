# Test Design: DHT Bootstrap and Operation Queuing

**Source Design:** crc-Peer.md, seq-dht-bootstrap.md
**Component:** Peer (DHT bootstrap and operation queuing system)

## Overview

Test the DHT bootstrap process and operation queuing system that prevents "no peers in table" errors by queuing DHT operations until the DHT routing table is populated.

## Test Scenarios

### 1. DHT Bootstrap Success

**Objective**: Verify DHT bootstraps successfully and signals readiness

**Setup**:
- Create peer with DHT enabled
- Mock DHT with controllable routing table

**Test Steps**:
1. Create peer with DHT
2. Verify dhtReady channel is open (not closed)
3. Wait for bootstrap goroutine to start
4. Mock DHT routing table to have 1+ peers
5. Verify dhtReady channel closes within expected time
6. Verify no queued operations remain

**Expected Results**:
- dhtReady channel closes when routing table populated
- Bootstrap completes without errors
- Logging shows successful bootstrap

**Error Cases**:
- Bootstrap peer connection failures (should continue if 1+ connects)
- DHT.Bootstrap() returns error (should log warning and continue)

### 2. DHT Bootstrap Timeout

**Objective**: Verify system handles DHT bootstrap timeout gracefully

**Setup**:
- Create peer with DHT enabled
- Mock DHT with empty routing table (never populates)

**Test Steps**:
1. Create peer with DHT
2. Verify dhtReady channel is open
3. Wait 30+ seconds
4. Verify dhtReady channel closes despite empty routing table
5. Verify timeout logged

**Expected Results**:
- dhtReady channel closes after 30s timeout
- Timeout warning logged
- System continues operating (doesn't hang)

**Performance**:
- Timeout occurs at ~30 seconds (±500ms polling tolerance)

### 3. Operation Queuing Before Bootstrap

**Objective**: Verify DHT operations queue when DHT not ready

**Setup**:
- Create peer with DHT enabled
- Block DHT bootstrap (don't populate routing table yet)
- Mock advertiseTopic and discoverTopicPeers

**Test Steps**:
1. Create peer with DHT
2. Verify dhtReady is open
3. Call subscribe(topic) (triggers advertiseTopic)
4. Verify operation queued (dhtOperations contains 1 function)
5. Verify DHT.Advertise() NOT called yet
6. Populate DHT routing table
7. Wait for dhtReady to close
8. Verify operation executed
9. Verify DHT.Advertise() called

**Expected Results**:
- Operation queues before bootstrap complete
- Operation executes after bootstrap complete
- Subscribe returns immediately (doesn't block)
- Logging shows "Queued DHT operation"

### 4. Operation Immediate Execution After Bootstrap

**Objective**: Verify DHT operations execute immediately when DHT ready

**Setup**:
- Create peer with DHT enabled
- Complete DHT bootstrap
- Mock advertiseTopic

**Test Steps**:
1. Create peer with DHT
2. Wait for bootstrap to complete (dhtReady closed)
3. Call subscribe(topic) (triggers advertiseTopic)
4. Verify operation NOT queued (dhtOperations empty)
5. Verify DHT.Advertise() called immediately

**Expected Results**:
- Operation executes immediately
- No queuing occurs
- No "Queued DHT operation" log message

### 5. Multiple Queued Operations

**Objective**: Verify multiple operations queue and execute in order

**Setup**:
- Create peer with DHT enabled
- Block DHT bootstrap
- Mock multiple DHT operations

**Test Steps**:
1. Create peer with DHT
2. Call subscribe(topic1) - queues operation 1
3. Call subscribe(topic2) - queues operation 2
4. Call subscribe(topic3) - queues operation 3
5. Verify dhtOperations has 3 functions
6. Complete DHT bootstrap (close dhtReady)
7. Verify all 3 operations execute
8. Verify DHT.Advertise() called 3 times
9. Verify operations executed in queue order

**Expected Results**:
- All operations queue successfully
- All operations execute after bootstrap
- Execution order matches queue order
- Queue cleared after execution

### 6. No DHT Case

**Objective**: Verify system handles peer without DHT

**Setup**:
- Create peer with DHT disabled (nil DHT)

**Test Steps**:
1. Create peer with DHT=nil
2. Verify dhtReady closes immediately
3. Call subscribe(topic) (triggers advertiseTopic)
4. Verify operation executes (doesn't queue)
5. Verify no DHT calls made

**Expected Results**:
- dhtReady closes immediately if no DHT
- Operations don't queue
- No DHT operations attempted
- No errors or panics

### 7. enqueueDHTOperation Thread Safety

**Objective**: Verify operation queuing is thread-safe

**Setup**:
- Create peer with DHT enabled
- Block DHT bootstrap
- Create multiple goroutines

**Test Steps**:
1. Create peer with DHT
2. Launch 100 goroutines simultaneously
3. Each goroutine calls enqueueDHTOperation with unique operation
4. Verify no race conditions detected
5. Complete DHT bootstrap
6. Verify all 100 operations execute exactly once

**Expected Results**:
- No race conditions with concurrent enqueuing
- All operations queued successfully
- All operations execute exactly once
- dhtOpMu protects queue correctly

**Performance**:
- Run with `-race` flag
- No data races reported

### 8. processQueuedDHTOperations Synchronization

**Objective**: Verify queue processing follows synchronization hygiene

**Setup**:
- Create peer with DHT enabled
- Queue multiple operations
- Mock operation execution timing

**Test Steps**:
1. Create peer with DHT
2. Queue 5 operations
3. Verify dhtOpMu locked only during queue extraction
4. Trigger processQueuedDHTOperations
5. Verify lock held only for extraction, not execution
6. Verify all operations execute without lock held
7. Verify queue cleared under lock

**Expected Results**:
- Lock → extract → unlock → process pattern followed
- Lock duration minimized
- No lock held during operation execution
- Queue properly cleared

**Pattern Verification**:
- Follows synchronization hygiene principles
- Minimal lock duration
- No nested locks

### 9. Bootstrap Peer Connection

**Objective**: Verify bootstrap peer connection logic

**Setup**:
- Create peer with DHT enabled
- Mock bootstrap peer connections

**Test Steps**:
1. Create peer with DHT
2. Verify attempts to connect to bootstrap peers
3. Verify stops after 3 successful connections
4. Test with all connections failing
5. Test with partial failures (2/5 succeed)
6. Verify bootstrap continues even if 0 connections

**Expected Results**:
- Connects to 3 bootstrap peers if available
- Stops trying after 3 successes
- Continues with fewer connections if failures
- Logs warning if 0 connections
- Bootstrap doesn't fail if connections fail

### 10. Bootstrap Routing Table Polling

**Objective**: Verify routing table polling logic

**Setup**:
- Create peer with DHT enabled
- Mock DHT routing table with controllable size

**Test Steps**:
1. Create peer with DHT
2. Set routing table size to 0
3. Verify polls every 500ms
4. After 2 seconds, set size to 1
5. Verify dhtReady closes immediately after next poll
6. Verify polling stops after readiness signaled

**Expected Results**:
- Polls routing table every 500ms
- Closes dhtReady when size > 0
- Stops polling after ready
- Doesn't wait full 30s if table populates early

**Performance**:
- Polling interval ~500ms (±50ms tolerance)
- Readiness detected within one polling cycle

## Integration Tests

### DHT Bootstrap with Real Subscribe/Advertise

**Objective**: End-to-end test with real operations

**Setup**:
- Create 2 peers with DHT enabled
- Real bootstrap peers
- Real topic subscription

**Test Steps**:
1. Create peer1 (bootstrap starts)
2. Immediately call peer1.Subscribe(topic)
3. Verify subscribe succeeds
4. Wait for bootstrap to complete
5. Verify advertiseTopic executed
6. Create peer2 (bootstrap starts)
7. Call peer2.Subscribe(topic)
8. Verify peers discover each other via DHT

**Expected Results**:
- Early subscribe operations queue automatically
- Operations execute after bootstrap
- Peers discover each other successfully
- No "no peers in table" errors

### Retry Added Peers with DHT

**Objective**: Verify added peers retry via DHT after bootstrap

**Setup**:
- Create 2 peers
- Add peer2 to peer1 before connection possible
- Complete DHT bootstrap

**Test Steps**:
1. Create peer1
2. Call peer1.AddPeers(peer2.ID) with unreachable address
3. Verify connection fails initially
4. Wait for DHT bootstrap
5. Verify retry loop uses DHT.FindPeer()
6. Make peer2 reachable via DHT
7. Verify connection succeeds via DHT retry

**Expected Results**:
- Added peers tracked for retry
- Retry uses DHT lookup
- Connection succeeds via DHT discovery

## Test Implementation Guidelines

**File**: `tests/peer_dht_test.go` or `internal/peer/dht_test.go`

**Mock Requirements**:
- Mock DHT with controllable routing table size
- Mock DHT.Bootstrap() return value
- Mock DHT.Advertise() and FindPeers()
- Mock bootstrap peer connections
- Controllable time (for timeout testing)

**Test Utilities**:
- Helper to create test peer with mock DHT
- Helper to wait for channel close with timeout
- Helper to verify operation queue state
- Race detector enabled for concurrency tests

**Coverage Goals**:
- 100% coverage of bootstrapDHT()
- 100% coverage of enqueueDHTOperation()
- 100% coverage of processQueuedDHTOperations()
- All error paths tested
- All timeout paths tested

## Traceability

**Design Specs**:
- crc-Peer.md (responsibilities: bootstrapDHT, enqueueDHTOperation, processQueuedDHTOperations)
- seq-dht-bootstrap.md (bootstrap flow and queuing behavior)

**Implementation**:
- internal/peer/manager.go (Peer struct, bootstrap methods)

**Related Tests**:
- Connection management tests (test-connection-management.md)
- PubSub tests (existing tests for subscribe/advertise)
