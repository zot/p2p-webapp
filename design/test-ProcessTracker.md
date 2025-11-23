# Test Design: ProcessTracker

**Source Specs**: main.md
**CRC Cards**: crc-ProcessTracker.md

## Overview

Test suite for ProcessTracker component covering PID registration, file locking, process verification, and graceful termination.

## Test Cases

### Test: Register PID on startup

**Purpose**: Verify that ProcessTracker registers current process PID in tracking file.

**Motivation**: Foundation for multi-instance management.

**Input**:
- ProcessTracker instance
- Current process PID: 12345
- Call registerPID()

**References**:
- CRC: crc-ProcessTracker.md - "Does: registerPID"

**Expected Results**:
- PID file created at /tmp/.p2p-webapp if not exists
- Current PID (12345) added to file
- File lock acquired during write
- File lock released after write
- PID persisted for other instances to see

**References**:
- CRC: crc-ProcessTracker.md - "Knows: pidFile"
- CRC: crc-ProcessTracker.md - "Does: lockFile, unlockFile"

---

### Test: Unregister PID on shutdown

**Purpose**: Verify that ProcessTracker removes PID from tracking file on shutdown.

**Motivation**: Prevents stale PID accumulation in tracking file.

**Input**:
- ProcessTracker with PID 12345 registered
- Server shutting down
- Call unregisterPID()

**References**:
- CRC: crc-ProcessTracker.md - "Does: unregisterPID"

**Expected Results**:
- File lock acquired
- PID 12345 removed from file
- Other PIDs remain in file
- File lock released
- File updated successfully

**References**:
- CRC: crc-ProcessTracker.md - "Does: lockFile, unlockFile"

---

### Test: List all tracked PIDs

**Purpose**: Verify that ProcessTracker can list all registered PIDs.

**Motivation**: Enables ps command to show running instances.

**Input**:
- Tracking file contains PIDs: 12345, 12346, 12347
- Call listPIDs()

**References**:
- CRC: crc-ProcessTracker.md - "Does: listPIDs"

**Expected Results**:
- Returns array: [12345, 12346, 12347]
- All PIDs from file included
- No duplicates

**References**:
- CRC: crc-ProcessTracker.md - "Knows: pids"

---

### Test: Verify PID is actual p2p-webapp process

**Purpose**: Verify that ProcessTracker checks if PID is actually a p2p-webapp instance.

**Motivation**: Security: prevents killing unrelated processes.

**Input**:
- Test 1: PID 12345 is running p2p-webapp binary
- Test 2: PID 99999 is running different program
- Call verifyPID() for each

**References**:
- CRC: crc-ProcessTracker.md - "Does: verifyPID"

**Expected Results**:
- Test 1: verifyPID(12345) returns true
- Test 2: verifyPID(99999) returns false
- Process name verified using gopsutil library

**References**:
- CRC: crc-ProcessTracker.md - "Collaborators: gopsutil library"

---

### Test: Clean stale PIDs

**Purpose**: Verify that ProcessTracker removes PIDs for processes that are no longer running.

**Motivation**: Prevents tracking file from accumulating dead entries.

**Input**:
- Tracking file contains:
  - PID 12345 (running p2p-webapp)
  - PID 12346 (not running)
  - PID 12347 (running different program)
- Call cleanStale()

**References**:
- CRC: crc-ProcessTracker.md - "Does: cleanStale"

**Expected Results**:
- File lock acquired
- PID 12346 removed (not running)
- PID 12347 removed (wrong program)
- PID 12345 remains (valid p2p-webapp)
- File lock released
- File updated

**References**:
- CRC: crc-ProcessTracker.md - "Does: verifyPID"
- CRC: crc-ProcessTracker.md - "Does: lockFile, unlockFile"

---

### Test: Kill specific instance by PID

**Purpose**: Verify that ProcessTracker can gracefully terminate specific instance.

**Motivation**: Enables kill command for managing specific instances.

**Input**:
- Running p2p-webapp instance with PID 12345
- Call killPID(12345)

**References**:
- CRC: crc-ProcessTracker.md - "Does: killPID"

**Expected Results**:
- SIGTERM (signal 15) sent to PID 12345
- Wait up to 5 seconds for graceful shutdown
- If process still running after 5s:
  - SIGKILL (signal 9) sent
- Process terminated
- PID removed from tracking file

**References**:
- CRC: crc-ProcessTracker.md - "Does: killPID"

---

### Test: Kill instance with graceful shutdown

**Purpose**: Verify that ProcessTracker allows graceful shutdown before forcing kill.

**Motivation**: Prevents data loss by allowing clean shutdown.

**Input**:
- Running p2p-webapp instance that responds to SIGTERM within 2 seconds
- Call killPID()

**References**:
- CRC: crc-ProcessTracker.md - "Does: killPID"

**Expected Results**:
- SIGTERM sent
- Process shuts down gracefully within 2 seconds
- SIGKILL never sent (not needed)
- PID removed from tracking file
- No forced termination

**References**:
- CRC: crc-ProcessTracker.md - "Does: killPID"

---

### Test: Kill instance requiring SIGKILL

**Purpose**: Verify that ProcessTracker forces termination if graceful shutdown fails.

**Motivation**: Ensures process can always be stopped even if hung.

**Input**:
- Running p2p-webapp instance that ignores SIGTERM
- Call killPID()

**References**:
- CRC: crc-ProcessTracker.md - "Does: killPID"

**Expected Results**:
- SIGTERM sent
- Wait 5 seconds
- Process still running
- SIGKILL sent
- Process forcefully terminated
- PID removed from tracking file

**References**:
- CRC: crc-ProcessTracker.md - "Does: killPID"

---

### Test: Kill nonexistent PID

**Purpose**: Verify that ProcessTracker handles attempt to kill nonexistent PID.

**Motivation**: Edge case for invalid PID or already-terminated process.

**Input**:
- PID 99999 not running
- Call killPID(99999)

**References**:
- CRC: crc-ProcessTracker.md - "Does: killPID"

**Expected Results**:
- Returns error indicating PID not found or not running
- No signals sent
- No crash
- Clear error message

**References**:
- CRC: crc-ProcessTracker.md - "Does: killPID"

---

### Test: Kill all tracked instances

**Purpose**: Verify that ProcessTracker can terminate all tracked instances.

**Motivation**: Enables killall command for cleanup.

**Input**:
- Multiple running p2p-webapp instances:
  - PID 12345
  - PID 12346
  - PID 12347
- Call killAll()

**References**:
- CRC: crc-ProcessTracker.md - "Does: killAll"

**Expected Results**:
- SIGTERM sent to all three instances
- Wait 5 seconds for all to shutdown
- Any still running receive SIGKILL
- All PIDs removed from tracking file
- File cleaned up

**References**:
- CRC: crc-ProcessTracker.md - "Does: killAll"

---

### Test: File locking prevents concurrent access corruption

**Purpose**: Verify that ProcessTracker uses file locking to prevent race conditions.

**Motivation**: Critical for correctness when multiple instances start/stop simultaneously.

**Input**:
- Two processes attempting to register PIDs simultaneously

**References**:
- CRC: crc-ProcessTracker.md - "Does: lockFile"

**Expected Results**:
- First process acquires lock
- Second process blocks until lock released
- Both PIDs registered correctly
- No data corruption
- No lost PIDs

**References**:
- CRC: crc-ProcessTracker.md - "Does: lockFile, unlockFile"

---

### Test: PID file location

**Purpose**: Verify that ProcessTracker uses correct PID file location.

**Motivation**: Ensures consistent tracking across all instances.

**Input**:
- ProcessTracker instance

**References**:
- CRC: crc-ProcessTracker.md - "Knows: pidFile"

**Expected Results**:
- PID file path: /tmp/.p2p-webapp
- File created if not exists
- File readable by all instances
- Location consistent across platform (Unix/Linux)

**References**:
- CRC: crc-ProcessTracker.md - "Knows: pidFile"

---

### Test: Register PID while extracting bundle

**Purpose**: Verify that ExtractCommand registers PID during long-running extract operation.

**Motivation**: Ensures extract command is visible to ps and killable.

**Input**:
- ExtractCommand running
- Call registerPID() before extraction

**References**:
- CRC: crc-ProcessTracker.md - "Collaborators: ExtractCommand"

**Expected Results**:
- PID registered before extraction begins
- PID visible in ps output during extraction
- PID can be killed during extraction
- PID unregistered after extraction completes

**References**:
- CRC: crc-ProcessTracker.md - "Collaborators: ExtractCommand"

## Coverage Summary

**Responsibilities Covered**:
- ✅ registerPID - Registration test
- ✅ unregisterPID - Unregistration test
- ✅ listPIDs - List all PIDs test
- ✅ verifyPID - Verification tests (valid and invalid)
- ✅ cleanStale - Stale PID cleanup test
- ✅ killPID - Kill specific instance tests (graceful, forced, nonexistent)
- ✅ killAll - Kill all instances test
- ✅ lockFile - File locking test
- ✅ unlockFile - Implicit in locking test
- ✅ pidFile location - PID file path test

**Scenarios Covered**:
- ⚠️ No dedicated sequences for ProcessTracker operations
- Process lifecycle tested through unit tests

**Gaps**:
- File lock timeout behavior not tested
- File corruption recovery not tested
- Platform-specific file locking differences not tested
- Extract command PID registration tested minimally (integration test would be better)
