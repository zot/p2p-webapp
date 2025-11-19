# Server

**Source Spec:** main.md

## Responsibilities

### Knows
- serverPort: Port to listen on (auto-selected starting from 10000)
- verbose: Verbosity level (0-3)
- noOpen: Whether to suppress browser auto-launch
- dirMode: Whether serving from directory or bundled site

### Does
- initialize: Create WebSocketHandler, PeerManager, WebServer instances
- start: Begin listening on port, register PID with ProcessTracker
- serve: Coordinate between WebSocket, peer, and HTTP services
- shutdown: Clean shutdown of all services, unregister PID
- handleSignals: Listen for SIGHUP (1), SIGINT (2), SIGTERM (15) and trigger graceful shutdown

## Collaborators

- WebSocketHandler: Manages WebSocket connections and JSON-RPC protocol
- PeerManager: Creates and manages libp2p peers for browser connections
- WebServer: Serves HTML/JS files with SPA routing support
- ProcessTracker: Registers/unregisters this server instance's PID
- BundleManager: Reads bundled content if in bundled mode

## Sequences

- seq-server-startup.md: Server initialization and startup flow
- seq-server-shutdown.md: Clean shutdown sequence
