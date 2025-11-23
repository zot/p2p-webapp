# Test Design: WebServer and CommandRouter

**Source Specs**: main.md
**CRC Cards**: crc-WebServer.md, crc-CommandRouter.md

## Overview

Test suite for WebServer (HTTP file serving and SPA routing) and CommandRouter (command-line argument parsing and routing) components.

## WebServer Test Cases

### Test: Serve static file

**Purpose**: Verify that WebServer serves static files from content root.

**Motivation**: Core HTTP file serving functionality.

**Input**:
- WebServer with contentRoot pointing to directory containing index.html
- HTTP GET request to /index.html

**References**:
- CRC: crc-WebServer.md - "Does: serveFile"

**Expected Results**:
- File content returned in response body
- Correct Content-Type header (text/html)
- 200 OK status code
- File served successfully

**References**:
- CRC: crc-WebServer.md - "Does: detectFileType"

---

### Test: Serve JavaScript file with correct Content-Type

**Purpose**: Verify that WebServer sets correct Content-Type for JavaScript files.

**Motivation**: Browser needs correct MIME type for script execution.

**Input**:
- File: app.js
- HTTP GET request to /app.js

**References**:
- CRC: crc-WebServer.md - "Does: detectFileType"

**Expected Results**:
- Content-Type: application/javascript
- File content returned
- 200 OK status code

**References**:
- CRC: crc-WebServer.md - "Does: detectFileType"

---

### Test: Serve CSS file with correct Content-Type

**Purpose**: Verify that WebServer sets correct Content-Type for CSS files.

**Motivation**: Browser needs correct MIME type for stylesheet.

**Input**:
- File: styles.css
- HTTP GET request to /styles.css

**References**:
- CRC: crc-WebServer.md - "Does: detectFileType"

**Expected Results**:
- Content-Type: text/css
- File content returned
- 200 OK status code

**References**:
- CRC: crc-WebServer.md - "Does: detectFileType"

---

### Test: SPA route fallback to index.html

**Purpose**: Verify that WebServer serves index.html for routes without file extensions.

**Motivation**: Enables client-side routing for single-page applications.

**Input**:
- WebServer with SPA routing enabled
- HTTP GET request to /friends (no extension)
- No file named "friends" exists

**References**:
- CRC: crc-WebServer.md - "Does: handleSPARoute"

**Expected Results**:
- index.html content returned
- URL preserved in browser (/friends)
- 200 OK status code
- Content-Type: text/html

**References**:
- CRC: crc-WebServer.md - "Knows: indexHTML"

---

### Test: SPA route with nested path

**Purpose**: Verify that SPA routing works for nested paths.

**Motivation**: Real-world SPA routes often have nested structure.

**Input**:
- HTTP GET request to /worlds/123/edit

**References**:
- CRC: crc-WebServer.md - "Does: handleSPARoute"

**Expected Results**:
- index.html returned
- URL preserved: /worlds/123/edit
- Browser can parse URL for client-side routing

**References**:
- CRC: crc-WebServer.md - "Does: handleSPARoute"

---

### Test: Return 404 for missing file with extension

**Purpose**: Verify that WebServer returns 404 for missing files with extensions.

**Motivation**: Files with extensions are static resources, not SPA routes.

**Input**:
- HTTP GET request to /nonexistent.js
- File does not exist

**References**:
- CRC: crc-WebServer.md - "Does: return404"

**Expected Results**:
- 404 Not Found status code
- Error message in response body
- No fallback to index.html

**References**:
- CRC: crc-WebServer.md - "Does: return404"

---

### Test: Serve files from BundleManager in bundled mode

**Purpose**: Verify that WebServer serves files from bundle instead of filesystem.

**Motivation**: Bundled mode uses ZIP archive, not filesystem.

**Input**:
- WebServer with BundleManager (bundled mode)
- HTTP GET request to /index.html

**References**:
- CRC: crc-WebServer.md - "Knows: contentRoot"
- CRC: crc-WebServer.md - "Collaborators: BundleManager"

**Expected Results**:
- File read from BundleManager.readFile()
- File content returned
- 200 OK status code
- No filesystem access

**References**:
- CRC: crc-WebServer.md - "Does: serveFile"

---

### Test: Cache index.html for SPA routing

**Purpose**: Verify that WebServer caches index.html to avoid repeated reads.

**Motivation**: Performance optimization for SPA routes.

**Input**:
- WebServer starts
- Multiple requests to SPA routes: /friends, /worlds, /settings

**References**:
- CRC: crc-WebServer.md - "Knows: indexHTML"

**Expected Results**:
- index.html read from disk once on first SPA route
- Subsequent SPA routes serve from cache
- No repeated disk/bundle reads

**References**:
- CRC: crc-WebServer.md - "Knows: indexHTML"

---

### Test: Detect route vs static file

**Purpose**: Verify that WebServer correctly distinguishes routes from files.

**Motivation**: Core logic for SPA routing decision.

**Input**:
- Test 1: /friends (no extension, route)
- Test 2: /app.js (has extension, file)
- Test 3: / (root, route)

**References**:
- CRC: crc-WebServer.md - "Does: handleSPARoute"

**Expected Results**:
- Test 1: SPA route, serves index.html
- Test 2: Static file, serves app.js
- Test 3: SPA route (root), serves index.html

**References**:
- CRC: crc-WebServer.md - "Does: handleSPARoute"

---

## CommandRouter Test Cases

### Test: Route to default server command

**Purpose**: Verify that CommandRouter starts server when no subcommand specified.

**Motivation**: Default behavior is to run server.

**Input**:
- Command-line: ./p2p-webapp (no arguments)

**References**:
- CRC: crc-CommandRouter.md - "Does: handleServer"

**Expected Results**:
- Server started
- Default port 10000
- Browser opens automatically

**References**:
- CRC: crc-CommandRouter.md - "Collaborators: Server"

---

### Test: Parse command-line flags

**Purpose**: Verify that CommandRouter parses common flags.

**Motivation**: Enables server customization via CLI.

**Input**:
- Command-line: ./p2p-webapp --dir /path/to/site --noopen -v -p 8080

**References**:
- CRC: crc-CommandRouter.md - "Does: parseArgs"
- CRC: crc-CommandRouter.md - "Knows: flags"

**Expected Results**:
- Flags parsed correctly:
  - dir: "/path/to/site"
  - noopen: true
  - verbose: 1
  - port: 8080
- Server started with these settings

**References**:
- CRC: crc-CommandRouter.md - "Knows: flags (--dir, --noopen, -v, -p)"

---

### Test: Route to extract command

**Purpose**: Verify that CommandRouter routes to extract subcommand.

**Motivation**: Enables extracting bundled site to directory.

**Input**:
- Command-line: ./p2p-webapp extract

**References**:
- CRC: crc-CommandRouter.md - "Does: handleExtract"

**Expected Results**:
- BundleManager.extractAll() called
- Bundled site extracted to current directory
- Success message displayed

**References**:
- CRC: crc-CommandRouter.md - "Collaborators: BundleManager"

---

### Test: Route to bundle command

**Purpose**: Verify that CommandRouter routes to bundle subcommand.

**Motivation**: Enables creating bundled binary.

**Input**:
- Command-line: ./p2p-webapp bundle /path/to/site

**References**:
- CRC: crc-CommandRouter.md - "Does: handleBundle"

**Expected Results**:
- BundleManager.appendBundle() called
- ZIP created from directory
- Bundled binary created

**References**:
- CRC: crc-CommandRouter.md - "Collaborators: BundleManager"

---

### Test: Route to ls command

**Purpose**: Verify that CommandRouter routes to ls subcommand.

**Motivation**: Enables listing bundled files.

**Input**:
- Command-line: ./p2p-webapp ls

**References**:
- CRC: crc-CommandRouter.md - "Does: handleLs"

**Expected Results**:
- BundleManager.listFiles() called
- File list displayed to console

**References**:
- CRC: crc-CommandRouter.md - "Collaborators: BundleManager"

---

### Test: Route to cp command

**Purpose**: Verify that CommandRouter routes to cp subcommand.

**Motivation**: Enables copying files from bundle.

**Input**:
- Command-line: ./p2p-webapp cp "html/*.js" /tmp/output

**References**:
- CRC: crc-CommandRouter.md - "Does: handleCp"

**Expected Results**:
- BundleManager.copyFiles() called with glob pattern
- Files copied to destination

**References**:
- CRC: crc-CommandRouter.md - "Collaborators: BundleManager"

---

### Test: Route to ps command

**Purpose**: Verify that CommandRouter routes to ps subcommand.

**Motivation**: Enables listing running instances.

**Input**:
- Command-line: ./p2p-webapp ps

**References**:
- CRC: crc-CommandRouter.md - "Does: handlePs"

**Expected Results**:
- ProcessTracker.listPIDs() called
- Running instance PIDs displayed

**References**:
- CRC: crc-CommandRouter.md - "Collaborators: ProcessTracker"

---

### Test: Route to kill command

**Purpose**: Verify that CommandRouter routes to kill subcommand.

**Motivation**: Enables stopping specific instance.

**Input**:
- Command-line: ./p2p-webapp kill 12345

**References**:
- CRC: crc-CommandRouter.md - "Does: handleKill"

**Expected Results**:
- ProcessTracker.killPID(12345) called
- Instance terminated
- Success message displayed

**References**:
- CRC: crc-CommandRouter.md - "Collaborators: ProcessTracker"

---

### Test: Route to killall command

**Purpose**: Verify that CommandRouter routes to killall subcommand.

**Motivation**: Enables stopping all instances.

**Input**:
- Command-line: ./p2p-webapp killall

**References**:
- CRC: crc-CommandRouter.md - "Does: handleKillAll"

**Expected Results**:
- ProcessTracker.killAll() called
- All instances terminated

**References**:
- CRC: crc-CommandRouter.md - "Collaborators: ProcessTracker"

---

### Test: Route to version command

**Purpose**: Verify that CommandRouter routes to version subcommand.

**Motivation**: Enables displaying version information.

**Input**:
- Command-line: ./p2p-webapp version

**References**:
- CRC: crc-CommandRouter.md - "Does: handleVersion"

**Expected Results**:
- Version string displayed
- Program exits

**References**:
- CRC: crc-CommandRouter.md - "Does: handleVersion"

---

### Test: Handle unknown command

**Purpose**: Verify that CommandRouter handles unknown subcommands gracefully.

**Motivation**: User-friendly error handling.

**Input**:
- Command-line: ./p2p-webapp unknown

**References**:
- CRC: crc-CommandRouter.md - "Does: routeCommand"

**Expected Results**:
- Error message: "Unknown command: unknown"
- Help text displayed
- Exit with error code

**References**:
- CRC: crc-CommandRouter.md - "Does: routeCommand"

---

### Test: Multiple verbosity flags

**Purpose**: Verify that CommandRouter handles -v, -vv, -vvv for verbosity levels.

**Motivation**: Enables progressive logging detail.

**Input**:
- Test 1: -v (level 1)
- Test 2: -vv (level 2)
- Test 3: -vvv (level 3)

**References**:
- CRC: crc-CommandRouter.md - "Knows: flags (-v)"

**Expected Results**:
- Test 1: verbosity = 1
- Test 2: verbosity = 2
- Test 3: verbosity = 3
- Server receives correct verbosity level

**References**:
- CRC: crc-CommandRouter.md - "Does: parseArgs"

## Coverage Summary

### WebServer Responsibilities Covered:
- ✅ serveFile - Static file serving tests
- ✅ handleSPARoute - SPA routing tests (simple and nested)
- ✅ detectFileType - Content-Type detection tests
- ✅ return404 - Missing file with extension test
- ✅ indexHTML caching - Cache optimization test
- ✅ BundleManager integration - Bundled mode serving test

### CommandRouter Responsibilities Covered:
- ✅ parseArgs - Flag parsing test
- ✅ routeCommand - Routing tests for all commands
- ✅ handleServer - Default server start test
- ✅ handleExtract - Extract command test
- ✅ handleBundle - Bundle command test
- ✅ handleLs - Ls command test
- ✅ handleCp - Cp command test
- ✅ handlePs - Ps command test
- ✅ handleKill - Kill command test
- ✅ handleKillAll - Killall command test
- ✅ handleVersion - Version command test
- ✅ Unknown command handling - Error case test

### Scenarios Covered:
- ⚠️ No dedicated sequences for WebServer or CommandRouter
- Components tested through unit tests

### Gaps:
- Large file serving performance not tested
- Concurrent HTTP requests not tested
- HTTP header validation not tested
- Configuration integration with CommandRouter not fully tested
- Command-line help text not tested
