# Test Design: Server

**Source Specs**: main.md
**CRC Cards**: crc-Server.md
**Sequences**: seq-server-startup.md

## Overview

Test suite for Server component covering server lifecycle, port selection, signal handling, and service coordination.

## Test Cases

### Test: Server startup with available port

**Purpose**: Verify that Server successfully starts on the first available port and initializes all services.

**Motivation**: Core happy path for server operation. Ensures all services start correctly and are properly coordinated.

**Input**:
- Server instance configured with default port 10000
- Port 10000 is available
- No other instances running

**References**:
- CRC: crc-Server.md - "Does: initialize"
- CRC: crc-Server.md - "Does: start"
- Sequence: seq-server-startup.md

**Expected Results**:
- Server listens on port 10000
- WebSocketHandler instance created and listening
- PeerManager instance created with discovery enabled
- WebServer instance created and serving files
- PID registered with ProcessTracker
- Browser opens automatically (unless --noopen flag set)
- Server ready to accept connections

**References**:
- CRC: crc-Server.md - "Knows: serverPort"
- CRC: crc-Server.md - "Collaborators: WebSocketHandler, PeerManager, WebServer, ProcessTracker"

---

### Test: Server startup with port collision

**Purpose**: Verify that Server tries subsequent ports when the default port is unavailable.

**Motivation**: Ensures multiple instances can run simultaneously and handles common port collision scenario.

**Input**:
- Server instance configured with starting port 10000
- Port 10000 already in use
- Port 10001 available

**References**:
- CRC: crc-Server.md - "Does: start"
- CRC: crc-Server.md - "Knows: serverPort"

**Expected Results**:
- Server attempts port 10000, fails
- Server automatically tries port 10001
- Server successfully starts on port 10001
- All services initialized correctly
- PID registered with ProcessTracker

**References**:
- CRC: crc-Server.md - "Does: start"

---

### Test: Server startup with all ports unavailable

**Purpose**: Verify that Server fails gracefully when no ports are available in the range.

**Motivation**: Edge case for resource exhaustion. Ensures proper error handling and cleanup.

**Input**:
- Server instance configured with starting port 10000
- All ports 10000-10099 already in use (100 port range)

**References**:
- CRC: crc-Server.md - "Does: start"

**Expected Results**:
- Server attempts all 100 ports in range
- Server exits with clear error message indicating port exhaustion
- No partial service initialization
- No PID registered with ProcessTracker

**References**:
- CRC: crc-Server.md - "Does: start"

---

### Test: Graceful shutdown on SIGTERM

**Purpose**: Verify that Server performs clean shutdown when receiving SIGTERM signal.

**Motivation**: Ensures proper cleanup of resources and prevents data loss or corruption on shutdown.

**Input**:
- Running Server instance with active services
- PID registered with ProcessTracker
- Send SIGTERM (signal 15) to server process

**References**:
- CRC: crc-Server.md - "Does: handleSignals"
- CRC: crc-Server.md - "Does: shutdown"

**Expected Results**:
- Signal handler catches SIGTERM
- Server initiates graceful shutdown:
  - WebSocketHandler stops accepting new connections
  - Active WebSocket connections closed cleanly
  - PeerManager stops all peers
  - WebServer stops HTTP server
  - PID unregistered from ProcessTracker
- Server exits with success code

**References**:
- CRC: crc-Server.md - "Does: shutdown"
- CRC: crc-Server.md - "Collaborators: ProcessTracker"

---

### Test: Graceful shutdown on SIGINT

**Purpose**: Verify that Server performs clean shutdown when user presses Ctrl+C (SIGINT).

**Motivation**: Common user-initiated shutdown method. Must handle same as SIGTERM.

**Input**:
- Running Server instance
- User sends SIGINT (signal 2, Ctrl+C)

**References**:
- CRC: crc-Server.md - "Does: handleSignals"

**Expected Results**:
- Signal handler catches SIGINT
- Same graceful shutdown sequence as SIGTERM test
- Clean exit with all resources released

**References**:
- CRC: crc-Server.md - "Does: shutdown"

---

### Test: Server handles SIGHUP

**Purpose**: Verify that Server handles SIGHUP signal appropriately.

**Motivation**: SIGHUP is mentioned in CRC card as handled signal.

**Input**:
- Running Server instance
- Send SIGHUP (signal 1) to server process

**References**:
- CRC: crc-Server.md - "Does: handleSignals"

**Expected Results**:
- Signal handler catches SIGHUP
- Server performs appropriate action (graceful shutdown or reload configuration)

**References**:
- CRC: crc-Server.md - "Does: handleSignals"

---

### Test: Server verbosity levels

**Purpose**: Verify that Server correctly applies verbosity levels 0-3 to logging output.

**Motivation**: Ensures debugging capability and appropriate log filtering.

**Input**:
- Server instances with verbosity levels 0, 1, 2, 3
- Various operations that generate logs

**References**:
- CRC: crc-Server.md - "Knows: verbose"

**Expected Results**:
- Level 0: Minimal/no output
- Level 1: Basic startup/shutdown messages
- Level 2: Service lifecycle events
- Level 3: Detailed operation logs including peer events
- Higher verbosity includes all lower level messages

**References**:
- CRC: crc-Server.md - "Knows: verbose"

---

### Test: Server noOpen flag suppresses browser launch

**Purpose**: Verify that --noopen flag prevents automatic browser opening.

**Motivation**: Headless/testing scenarios require server without browser UI.

**Input**:
- Server instance with noOpen=true flag
- Server starts successfully

**References**:
- CRC: crc-Server.md - "Knows: noOpen"

**Expected Results**:
- Server starts normally
- Browser does NOT open automatically
- Server remains running and accessible
- All services function normally

**References**:
- CRC: crc-Server.md - "Knows: noOpen"

---

### Test: Server directory mode vs bundled mode

**Purpose**: Verify that Server correctly switches between directory and bundled serving modes.

**Motivation**: Ensures both deployment modes work correctly with appropriate file serving.

**Input**:
- Test 1: Server with --dir flag pointing to filesystem directory
- Test 2: Server running from bundled binary without --dir flag

**References**:
- CRC: crc-Server.md - "Knows: dirMode"
- CRC: crc-Server.md - "Collaborators: BundleManager"

**Expected Results**:
- Directory mode: WebServer serves files from specified directory
- Bundled mode: WebServer serves files from BundleManager's ZIP archive
- Configuration loaded from appropriate source (filesystem vs ZIP)
- Both modes function identically from client perspective

**References**:
- CRC: crc-Server.md - "Knows: dirMode"

## Coverage Summary

**Responsibilities Covered**:
- ✅ initialize - Server startup tests
- ✅ start - Port selection and service startup tests
- ✅ serve - Implicit in startup tests (service coordination)
- ✅ shutdown - Signal handling and shutdown tests
- ✅ handleSignals - SIGHUP, SIGINT, SIGTERM tests
- ✅ serverPort - Port collision and selection tests
- ✅ verbose - Verbosity level tests
- ✅ noOpen - Browser launch suppression test
- ✅ dirMode - Directory vs bundled mode test

**Scenarios Covered**:
- ✅ seq-server-startup.md - Happy path and port collision tests
- ⚠️ seq-server-shutdown.md - Shutdown tests (sequence file may not exist)

**Gaps**:
- Integration tests with actual WebSocket clients not covered here (see test-WebSocketHandler.md)
- Performance tests under load not covered (future enhancement)
- Multiple concurrent server instances coordination not tested here
