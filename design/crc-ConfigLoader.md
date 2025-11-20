# ConfigLoader

**Source Spec:** main.md - Configuration section

## Responsibilities

### Knows
- Default configuration values for all server settings
- TOML file format structure and parsing
- Configuration file name (`p2p-webapp.toml`)
- Validation rules for configuration values
- Configuration precedence (defaults → file → flags)

### Does
- LoadFromDir: Load configuration from filesystem directory
- LoadFromZIP: Load configuration from bundled ZIP archive
- DefaultConfig: Provide default configuration values
- Merge: Merge command-line flags into configuration (flags take precedence)
- Validate: Check configuration values are valid and within acceptable ranges
- Parse TOML configuration file into Config struct
- Handle missing configuration file gracefully (use defaults)
- Convert duration strings to time.Duration (e.g., "5s" → 5 seconds)

## Collaborators

- **Server**: Uses Config to configure HTTP server, timeouts, and headers
- **CommandRouter**: Loads config before creating server, merges with CLI flags
- **BurntSushi/toml**: TOML parsing library
- **zipFileSystem**: Reads config file from bundled ZIP

## Configuration Structure

### Config
- ServerConfig: HTTP server settings
- HTTPConfig: HTTP headers and caching
- WebSocketConfig: WebSocket settings
- BehaviorConfig: Application behavior
- FilesConfig: File serving settings
- P2PConfig: P2P protocol settings

### ServerConfig
- Port: Starting port number
- PortRange: Number of ports to try
- Timeouts: ReadTimeout, WriteTimeout, IdleTimeout, ReadHeaderTimeout
- MaxHeaderBytes: Maximum request header size

### HTTPConfig
- CacheControl: Cache-Control header value
- Security: Security headers (X-Content-Type-Options, X-Frame-Options, CSP)
- CORS: CORS settings (enabled, allowOrigin, allowMethods, allowHeaders)

### WebSocketConfig
- CheckOrigin: Validate WebSocket origin
- AllowedOrigins: List of allowed origins
- ReadBufferSize: WebSocket read buffer size
- WriteBufferSize: WebSocket write buffer size

### BehaviorConfig
- AutoExitTimeout: Auto-exit timer duration
- AutoOpenBrowser: Whether to open browser on startup
- Linger: Keep server running after connections close
- Verbosity: Logging verbosity level

### FilesConfig
- IndexFile: File to serve for SPA routes
- SPAFallback: Enable SPA routing fallback

### P2PConfig
- ProtocolName: Reserved libp2p protocol name for file list queries
- FileUpdateNotifyTopic: Optional topic for file availability notifications

## Key Points

1. **Optional Configuration**: If no config file exists, uses default values without error
2. **Dual Loading**: Supports loading from both filesystem (--dir mode) and ZIP (bundle mode)
3. **Flag Precedence**: Command-line flags override config file values
4. **Validation**: Ensures configuration values are sensible before use
5. **Duration Parsing**: Custom Duration type wraps time.Duration for TOML unmarshaling
6. **Backward Compatible**: All existing behavior preserved when no config file present
7. **TOML Format**: Human-readable, commented, supports sections and lists

## Default Values

- Port: 10000, range 100
- Cache-Control: "no-cache, no-store, must-revalidate" (development-friendly)
- Timeouts: Read/Write 15s, Idle 60s, ReadHeader 5s
- Auto-exit: 5 seconds
- Security: nosniff, DENY frame options
- No CORS by default
- Auto-open browser enabled
- SPA fallback enabled
- Protocol name: "/p2p-webapp/1.0.0"
- File update notifications: disabled (empty topic)

## Sequences

Related to server startup:
- seq-server-startup.md: Configuration loaded before server creation

## Related

- crc-Server.md: Uses Config for server settings
- crc-CommandRouter.md: Loads and merges configuration
- specs/main.md: Configuration options documentation
