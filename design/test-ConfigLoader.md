# Test Design: ConfigLoader

**Source Specs**: main.md - Configuration section
**CRC Cards**: crc-ConfigLoader.md

## Overview

Test suite for ConfigLoader component covering configuration loading from filesystem and ZIP, TOML parsing, validation, flag merging, and default values.

## Test Cases

### Test: Load configuration from filesystem directory

**Purpose**: Verify that ConfigLoader can load configuration from p2p-webapp.toml in directory mode.

**Motivation**: Core functionality for directory-based deployments.

**Input**:
- Directory containing p2p-webapp.toml with custom values:
  ```toml
  [server]
  port = 8080

  [behavior]
  auto_open_browser = false
  ```
- Call LoadFromDir("/path/to/dir")

**References**:
- CRC: crc-ConfigLoader.md - "Does: LoadFromDir"

**Expected Results**:
- Configuration file read successfully
- TOML parsed into Config struct
- Custom values loaded:
  - ServerConfig.Port = 8080
  - BehaviorConfig.AutoOpenBrowser = false
- Default values used for unspecified settings
- No errors

**References**:
- CRC: crc-ConfigLoader.md - "Does: Parse TOML configuration file into Config struct"

---

### Test: Load configuration from ZIP bundle

**Purpose**: Verify that ConfigLoader can load configuration from bundled ZIP archive.

**Motivation**: Core functionality for bundled deployments.

**Input**:
- ZIP bundle containing config/p2p-webapp.toml
- Call LoadFromZIP(zipReader)

**References**:
- CRC: crc-ConfigLoader.md - "Does: LoadFromZIP"

**Expected Results**:
- Configuration file read from ZIP
- TOML parsed successfully
- Configuration values loaded
- No errors

**References**:
- CRC: crc-ConfigLoader.md - "Collaborators: zipFileSystem"

---

### Test: Handle missing configuration file gracefully

**Purpose**: Verify that ConfigLoader uses default configuration when file not present.

**Motivation**: Enables zero-configuration deployments.

**Input**:
- Directory without p2p-webapp.toml file
- Call LoadFromDir("/path/to/dir")

**References**:
- CRC: crc-ConfigLoader.md - "Does: Handle missing configuration file gracefully (use defaults)"

**Expected Results**:
- No error returned
- DefaultConfig() values used for all settings
- Configuration valid and usable
- Server can start with defaults

**References**:
- CRC: crc-ConfigLoader.md - "Does: DefaultConfig"

---

### Test: Provide default configuration values

**Purpose**: Verify that ConfigLoader provides correct default values for all settings.

**Motivation**: Documents expected default behavior.

**Input**:
- Call DefaultConfig()

**References**:
- CRC: crc-ConfigLoader.md - "Does: DefaultConfig"
- CRC: crc-ConfigLoader.md - "Knows: Default configuration values for all server settings"

**Expected Results**:
- Returns complete Config with defaults:
  - ServerConfig.Port = 10000
  - ServerConfig.PortRange = 100
  - ServerConfig.ReadTimeout = 15s
  - ServerConfig.WriteTimeout = 15s
  - ServerConfig.IdleTimeout = 60s
  - ServerConfig.ReadHeaderTimeout = 5s
  - HTTPConfig.CacheControl = "no-cache, no-store, must-revalidate"
  - BehaviorConfig.AutoExitTimeout = 5s
  - BehaviorConfig.AutoOpenBrowser = true
  - BehaviorConfig.Linger = false
  - FilesConfig.IndexFile = "index.html"
  - FilesConfig.SPAFallback = true
  - P2PConfig.ProtocolName = "/p2p-webapp/1.0.0"
  - P2PConfig.FileUpdateNotifyTopic = "" (empty, disabled)

**References**:
- CRC: crc-ConfigLoader.md - "Knows: Default configuration values"

---

### Test: Merge command-line flags into configuration

**Purpose**: Verify that ConfigLoader merges CLI flags with precedence over config file values.

**Motivation**: Enables runtime configuration overrides.

**Input**:
- Config from file with Port = 8080
- CLI flags: --port 9000, --noopen
- Call Merge(config, flags)

**References**:
- CRC: crc-ConfigLoader.md - "Does: Merge"
- CRC: crc-ConfigLoader.md - "Knows: Configuration precedence (defaults → file → flags)"

**Expected Results**:
- Merged config:
  - Port = 9000 (flag overrides file)
  - AutoOpenBrowser = false (noopen flag)
  - Other values from config file unchanged
- Flag precedence enforced

**References**:
- CRC: crc-ConfigLoader.md - "Does: Merge"

---

### Test: Validate configuration values

**Purpose**: Verify that ConfigLoader validates configuration for correctness and safety.

**Motivation**: Prevents invalid configurations from causing runtime errors.

**Input**:
- Test 1: Valid configuration
- Test 2: Invalid port number (negative)
- Test 3: Invalid timeout (negative duration)

**References**:
- CRC: crc-ConfigLoader.md - "Does: Validate"
- CRC: crc-ConfigLoader.md - "Knows: Validation rules for configuration values"

**Expected Results**:
- Test 1: Validation passes, no errors
- Test 2: Validation fails with error about invalid port
- Test 3: Validation fails with error about invalid timeout
- Clear error messages for each validation failure

**References**:
- CRC: crc-ConfigLoader.md - "Does: Validate"

---

### Test: Parse duration strings

**Purpose**: Verify that ConfigLoader correctly parses duration strings from TOML.

**Motivation**: Enables human-readable duration configuration.

**Input**:
- TOML with duration values:
  ```toml
  [server]
  read_timeout = "30s"
  write_timeout = "1m"
  idle_timeout = "2m30s"
  ```

**References**:
- CRC: crc-ConfigLoader.md - "Does: Convert duration strings to time.Duration"

**Expected Results**:
- ReadTimeout = 30 * time.Second
- WriteTimeout = 60 * time.Second
- IdleTimeout = 150 * time.Second
- All durations parsed correctly

**References**:
- CRC: crc-ConfigLoader.md - "Does: Convert duration strings to time.Duration (e.g., \"5s\" → 5 seconds)"

---

### Test: Parse ServerConfig section

**Purpose**: Verify that ConfigLoader parses HTTP server configuration correctly.

**Motivation**: Ensures server runs with correct settings.

**Input**:
- TOML:
  ```toml
  [server]
  port = 8080
  port_range = 50
  read_timeout = "20s"
  write_timeout = "20s"
  idle_timeout = "120s"
  read_header_timeout = "10s"
  max_header_bytes = 2097152
  ```

**References**:
- CRC: crc-ConfigLoader.md - "Configuration Structure: ServerConfig"

**Expected Results**:
- All fields parsed into ServerConfig struct
- Values match TOML specification
- Durations converted correctly

**References**:
- CRC: crc-ConfigLoader.md - "Configuration Structure: ServerConfig"

---

### Test: Parse HTTPConfig section with CORS

**Purpose**: Verify that ConfigLoader parses HTTP headers and CORS configuration.

**Motivation**: Enables security and caching configuration.

**Input**:
- TOML:
  ```toml
  [http]
  cache_control = "max-age=3600"

  [http.cors]
  enabled = true
  allow_origin = "*"
  allow_methods = "GET, POST, OPTIONS"
  ```

**References**:
- CRC: crc-ConfigLoader.md - "Configuration Structure: HTTPConfig"

**Expected Results**:
- HTTPConfig.CacheControl = "max-age=3600"
- HTTPConfig.CORS.Enabled = true
- HTTPConfig.CORS.AllowOrigin = "*"
- HTTPConfig.CORS.AllowMethods = "GET, POST, OPTIONS"

**References**:
- CRC: crc-ConfigLoader.md - "Configuration Structure: HTTPConfig"

---

### Test: Parse WebSocketConfig section

**Purpose**: Verify that ConfigLoader parses WebSocket configuration.

**Motivation**: Enables WebSocket security and buffer configuration.

**Input**:
- TOML:
  ```toml
  [websocket]
  check_origin = true
  allowed_origins = ["http://localhost:10000", "https://example.com"]
  read_buffer_size = 2048
  write_buffer_size = 2048
  ```

**References**:
- CRC: crc-ConfigLoader.md - "Configuration Structure: WebSocketConfig"

**Expected Results**:
- WebSocketConfig.CheckOrigin = true
- WebSocketConfig.AllowedOrigins = ["http://localhost:10000", "https://example.com"]
- WebSocketConfig.ReadBufferSize = 2048
- WebSocketConfig.WriteBufferSize = 2048

**References**:
- CRC: crc-ConfigLoader.md - "Configuration Structure: WebSocketConfig"

---

### Test: Parse BehaviorConfig section

**Purpose**: Verify that ConfigLoader parses application behavior configuration.

**Motivation**: Enables customization of server behavior.

**Input**:
- TOML:
  ```toml
  [behavior]
  auto_exit_timeout = "10s"
  auto_open_browser = false
  linger = true
  verbosity = 2
  ```

**References**:
- CRC: crc-ConfigLoader.md - "Configuration Structure: BehaviorConfig"

**Expected Results**:
- BehaviorConfig.AutoExitTimeout = 10 * time.Second
- BehaviorConfig.AutoOpenBrowser = false
- BehaviorConfig.Linger = true
- BehaviorConfig.Verbosity = 2

**References**:
- CRC: crc-ConfigLoader.md - "Configuration Structure: BehaviorConfig"

---

### Test: Parse FilesConfig section

**Purpose**: Verify that ConfigLoader parses file serving configuration.

**Motivation**: Enables SPA routing configuration.

**Input**:
- TOML:
  ```toml
  [files]
  index_file = "app.html"
  spa_fallback = false
  ```

**References**:
- CRC: crc-ConfigLoader.md - "Configuration Structure: FilesConfig"

**Expected Results**:
- FilesConfig.IndexFile = "app.html"
- FilesConfig.SPAFallback = false

**References**:
- CRC: crc-ConfigLoader.md - "Configuration Structure: FilesConfig"

---

### Test: Parse P2PConfig section

**Purpose**: Verify that ConfigLoader parses P2P protocol configuration.

**Motivation**: Enables customization of P2P protocol and notifications.

**Input**:
- TOML:
  ```toml
  [p2p]
  protocol_name = "/myapp/1.0.0"
  file_update_notify_topic = "file-updates"
  ```

**References**:
- CRC: crc-ConfigLoader.md - "Configuration Structure: P2PConfig"

**Expected Results**:
- P2PConfig.ProtocolName = "/myapp/1.0.0"
- P2PConfig.FileUpdateNotifyTopic = "file-updates"

**References**:
- CRC: crc-ConfigLoader.md - "Configuration Structure: P2PConfig"

---

### Test: Handle malformed TOML file

**Purpose**: Verify that ConfigLoader reports clear errors for malformed TOML.

**Motivation**: Helps users diagnose configuration errors.

**Input**:
- Malformed TOML file:
  ```toml
  [server
  port = 8080
  ```
  (Missing closing bracket)

**References**:
- CRC: crc-ConfigLoader.md - "Does: Parse TOML configuration file into Config struct"

**Expected Results**:
- Parsing fails with clear error message
- Error indicates line number and problem
- No panic or crash

**References**:
- CRC: crc-ConfigLoader.md - "Collaborators: BurntSushi/toml"

---

### Test: Configuration precedence order

**Purpose**: Verify that ConfigLoader applies configuration in correct precedence order.

**Motivation**: Ensures predictable configuration behavior.

**Input**:
- Defaults: Port = 10000
- Config file: Port = 8080
- CLI flag: --port 9000

**References**:
- CRC: crc-ConfigLoader.md - "Knows: Configuration precedence (defaults → file → flags)"

**Expected Results**:
- Final Port = 9000 (flags win)
- Precedence order enforced: defaults → file → flags
- Each layer overrides previous

**References**:
- CRC: crc-ConfigLoader.md - "Knows: Configuration precedence (defaults → file → flags)"

## Coverage Summary

**Responsibilities Covered**:
- ✅ LoadFromDir - Filesystem loading test
- ✅ LoadFromZIP - ZIP bundle loading test
- ✅ DefaultConfig - Default values test
- ✅ Merge - Flag merging test with precedence
- ✅ Validate - Validation tests (valid and invalid cases)
- ✅ Parse TOML - Parsing tests for all config sections
- ✅ Handle missing config file - Graceful handling test
- ✅ Convert duration strings - Duration parsing test
- ✅ All configuration structures - Individual section tests

**Scenarios Covered**:
- ✅ Configuration loaded before server creation (implicit in Server tests)
- ⚠️ No dedicated sequences for configuration loading

**Gaps**:
- Partial configuration files (some sections missing) not explicitly tested
- Configuration reload during runtime not tested (may not be implemented)
- Invalid duration strings not tested
- Edge cases for numeric limits not tested
