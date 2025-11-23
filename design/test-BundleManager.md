# Test Design: BundleManager

**Source Specs**: main.md
**CRC Cards**: crc-BundleManager.md

## Overview

Test suite for BundleManager component covering bundled content detection, file reading, ZIP operations, and bundle creation.

## Test Cases

### Test: Detect bundled content in binary

**Purpose**: Verify that BundleManager correctly detects when binary has bundled ZIP content.

**Motivation**: Foundation for bundled mode operation. Must reliably detect presence of bundle.

**Input**:
- Binary file with ZIP archive appended
- Footer with magic marker "P2PWEBAPPBUNDLE" + offset + size (24 bytes)

**References**:
- CRC: crc-BundleManager.md - "Does: checkBundled"
- CRC: crc-BundleManager.md - "Knows: footer"

**Expected Results**:
- checkBundled() returns true
- Footer parsed correctly
- zipOffset and zipSize extracted
- zipReader initialized

**References**:
- CRC: crc-BundleManager.md - "Knows: zipOffset, zipSize, zipReader"

---

### Test: Detect absence of bundled content

**Purpose**: Verify that BundleManager correctly detects when binary has no bundled content.

**Motivation**: Ensures clean error handling for unbundled binaries.

**Input**:
- Binary file without ZIP archive
- No footer marker present

**References**:
- CRC: crc-BundleManager.md - "Does: checkBundled"

**Expected Results**:
- checkBundled() returns false
- No ZIP reader initialized
- No errors or crashes

**References**:
- CRC: crc-BundleManager.md - "Does: checkBundled"

---

### Test: Read file from bundle

**Purpose**: Verify that BundleManager can read individual files from bundled ZIP archive.

**Motivation**: Core bundled mode functionality for serving files.

**Input**:
- Bundled binary with ZIP containing:
  - html/index.html
  - html/app.js
  - config/p2p-webapp.toml
- Call readFile("html/index.html")

**References**:
- CRC: crc-BundleManager.md - "Does: readFile"

**Expected Results**:
- File content returned successfully
- Content matches original file
- No errors

**References**:
- CRC: crc-BundleManager.md - "Knows: zipReader"
- CRC: crc-BundleManager.md - "Collaborators: WebServer"

---

### Test: Read nonexistent file from bundle

**Purpose**: Verify that BundleManager handles requests for files not in bundle.

**Motivation**: Edge case for missing files. Must fail gracefully.

**Input**:
- Bundled binary with ZIP archive
- Call readFile("html/nonexistent.html")

**References**:
- CRC: crc-BundleManager.md - "Does: readFile"

**Expected Results**:
- Returns error indicating file not found
- Error message clear and actionable
- No crash

**References**:
- CRC: crc-BundleManager.md - "Does: readFile"

---

### Test: List all files in bundle

**Purpose**: Verify that BundleManager can enumerate all files in bundled archive.

**Motivation**: Enables ls command and bundle inspection.

**Input**:
- Bundled binary with ZIP containing:
  - html/index.html
  - html/app.js
  - html/styles.css
  - config/p2p-webapp.toml

**References**:
- CRC: crc-BundleManager.md - "Does: listFiles"

**Expected Results**:
- Returns complete list of files
- File paths match ZIP structure
- All files enumerated correctly

**References**:
- CRC: crc-BundleManager.md - "Does: listFiles"
- CRC: crc-BundleManager.md - "Collaborators: LsCommand"

---

### Test: Copy files from bundle with glob pattern

**Purpose**: Verify that BundleManager can copy files matching glob patterns from bundle to filesystem.

**Motivation**: Enables selective file extraction via cp command.

**Input**:
- Bundled binary with various files
- Glob pattern: "html/*.js"
- Destination directory: "/tmp/output"

**References**:
- CRC: crc-BundleManager.md - "Does: copyFiles"

**Expected Results**:
- All .js files from html/ directory copied to destination
- Files created with correct content
- Directory structure preserved if needed
- Non-matching files not copied

**References**:
- CRC: crc-BundleManager.md - "Does: copyFiles"
- CRC: crc-BundleManager.md - "Collaborators: CpCommand"

---

### Test: Extract entire bundle to directory

**Purpose**: Verify that BundleManager can extract complete bundle to filesystem.

**Motivation**: Enables extract command for converting bundled binary to directory mode.

**Input**:
- Bundled binary with full site structure
- Destination directory: "./extracted"

**References**:
- CRC: crc-BundleManager.md - "Does: extractAll"

**Expected Results**:
- All files extracted to destination
- Directory structure recreated:
  - extracted/html/index.html
  - extracted/html/app.js
  - extracted/config/p2p-webapp.toml
- File contents identical to bundle
- Executable bit preserved for binaries if applicable

**References**:
- CRC: crc-BundleManager.md - "Does: extractAll"
- CRC: crc-BundleManager.md - "Collaborators: ExtractCommand"

---

### Test: Create bundled binary from directory

**Purpose**: Verify that BundleManager can create a bundled binary by appending ZIP + footer to base binary.

**Motivation**: Enables bundle command for distribution.

**Input**:
- Base binary (p2p-webapp executable)
- Source directory with:
  - html/index.html
  - html/app.js
  - config/p2p-webapp.toml
- Output: p2p-webapp-bundled

**References**:
- CRC: crc-BundleManager.md - "Does: appendBundle"

**Expected Results**:
- ZIP archive created from source directory
- ZIP appended to copy of base binary
- Footer appended with:
  - Magic marker: "P2PWEBAPPBUNDLE"
  - ZIP offset (8 bytes)
  - ZIP size (8 bytes)
- Output binary is executable
- Output binary checkBundled() returns true
- Output binary can serve bundled content

**References**:
- CRC: crc-BundleManager.md - "Does: appendBundle"
- CRC: crc-BundleManager.md - "Knows: footer"
- CRC: crc-BundleManager.md - "Collaborators: BundleCommand"

---

### Test: Bundle with empty directory

**Purpose**: Verify that BundleManager handles bundling of empty directory.

**Motivation**: Edge case for minimal deployments.

**Input**:
- Base binary
- Empty source directory

**References**:
- CRC: crc-BundleManager.md - "Does: appendBundle"

**Expected Results**:
- Empty ZIP created and appended
- Footer appended correctly
- Output binary valid but contains no files
- checkBundled() returns true
- listFiles() returns empty list

**References**:
- CRC: crc-BundleManager.md - "Does: appendBundle"

---

### Test: Read files with nested directory structure

**Purpose**: Verify that BundleManager correctly handles nested directories in bundle.

**Motivation**: Real-world bundles have complex directory structures.

**Input**:
- Bundle with nested structure:
  - html/index.html
  - html/js/app.js
  - html/js/lib/framework.js
  - html/css/styles.css
- Call readFile("html/js/lib/framework.js")

**References**:
- CRC: crc-BundleManager.md - "Does: readFile"

**Expected Results**:
- File read successfully from nested path
- Content matches original
- Path separators handled correctly (cross-platform)

**References**:
- CRC: crc-BundleManager.md - "Does: readFile"

---

### Test: Extract preserves directory structure

**Purpose**: Verify that extractAll recreates exact directory structure from bundle.

**Motivation**: Ensures fidelity of extracted files.

**Input**:
- Bundle with nested directories (as above)
- Extract to "./output"

**References**:
- CRC: crc-BundleManager.md - "Does: extractAll"

**Expected Results**:
- All directories created:
  - output/html/
  - output/html/js/
  - output/html/js/lib/
  - output/html/css/
- Files placed in correct locations
- Directory structure identical to source

**References**:
- CRC: crc-BundleManager.md - "Does: extractAll"

---

### Test: Handle ZIP corruption gracefully

**Purpose**: Verify that BundleManager detects and reports corrupted ZIP data.

**Motivation**: Robustness against file corruption or incomplete downloads.

**Input**:
- Binary with corrupted ZIP data
- Footer present but ZIP content damaged

**References**:
- CRC: crc-BundleManager.md - "Does: checkBundled"

**Expected Results**:
- checkBundled() detects corruption
- Clear error message returned
- No crash or panic
- Binary doesn't attempt to serve corrupted content

**References**:
- CRC: crc-BundleManager.md - "Does: checkBundled"

---

### Test: Handle footer corruption

**Purpose**: Verify that BundleManager detects corrupted or invalid footer.

**Motivation**: Ensures reliable bundle detection.

**Input**:
- Binary with corrupted footer (wrong magic marker, invalid offset/size)

**References**:
- CRC: crc-BundleManager.md - "Does: checkBundled"
- CRC: crc-BundleManager.md - "Knows: footer"

**Expected Results**:
- checkBundled() returns false
- No attempt to read ZIP
- No errors or crashes

**References**:
- CRC: crc-BundleManager.md - "Does: checkBundled"

## Coverage Summary

**Responsibilities Covered**:
- ✅ checkBundled - Detection tests (valid, absent, corrupted)
- ✅ readFile - File reading tests (existing, nonexistent, nested paths)
- ✅ listFiles - File enumeration test
- ✅ copyFiles - Selective file copy with glob patterns test
- ✅ extractAll - Complete extraction tests (full bundle, preserving structure)
- ✅ appendBundle - Bundle creation tests (normal, empty directory)
- ✅ All "Knows" properties - Covered through various tests

**Scenarios Covered**:
- ⚠️ Bundle-related sequences not present in provided sequence diagrams
- Bundle operations tested through unit tests

**Gaps**:
- Large bundle performance not tested
- Concurrent file reads from bundle not tested
- Bundle compression levels not tested (may not be configurable)
- Security validation of ZIP contents not tested (e.g., path traversal)
