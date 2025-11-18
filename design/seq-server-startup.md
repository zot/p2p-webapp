# Sequence: Server Startup

**Source Spec:** main.md
**Use Case:** User starts p2p-webapp server

## Participants

- User: Runs the p2p-webapp executable
- CommandRouter: Parses command-line arguments and routes to server
- Server: Main server orchestrator
- WebSocketHandler: Manages WebSocket connections
- PeerManager: Manages libp2p peers and discovery
- WebServer: HTTP file server with SPA routing
- ProcessTracker: Tracks running instance PIDs

## Sequence

       ┌─┐
       ║"│
       └┬┘
       ┌┼┐
        │             ┌─────────────┐          ┌──────┐           ┌────────────────┐           ┌───────────┐                   ┌─────────┐          ┌──────────────┐
       ┌┴┐            │CommandRouter│          │Server│           │WebSocketHandler│           │PeerManager│                   │WebServer│          │ProcessTracker│
      User            └──────┬──────┘          └───┬──┘           └────────┬───────┘           └─────┬─────┘                   └────┬────┘          └───────┬──────┘
        │   ./p2p-webapp     │                     │                       │                         │                              │                       │
        │───────────────────>│                     │                       │                         │                              │                       │
        │                    │                     │                       │                         │                              │                       │
        │                    │      start()        │                       │                         │                              │                       │
        │                    │────────────────────>│                       │                         │                              │                       │
        │                    │                     │                       │                         │                              │                       │
        │                    │                     │                     new()                       │                              │                       │
        │                    │                     │────────────────────────────────────────────────>│                              │                       │
        │                    │                     │                       │                         │                              │                       │
        │                    │                     │                       │                         │────┐                         │                       │
        │                    │                     │                       │                         │    │ configure discovery     │                       │
        │                    │                     │                       │                         │<───┘                         │                       │
        │                    │                     │                       │                         │                              │                       │
        │                    │                     │                       │                         │ ╔═══════════════════╗        │                       │
        │                    │                     │                       │                         │ ║Enable mDNS + DHT ░║        │                       │
        │                    │                     │                       │                         │ ╚═══════════════════╝        │                       │
        │                    │                     │                       │                         │────┐                         │                       │
        │                    │                     │                       │                         │    │ configure NAT traversal │                       │
        │                    │                     │                       │                         │<───┘                         │                       │
        │                    │                     │                       │                         │                              │                       │
        │                    │                     │                       │                         │ ╔════════════════════════════╧═╗                     │
        │                    │                     │                       │                         │ ║Circuit Relay, hole punching ░║                     │
        │                    │                     │                       │                         │ ╚════════════════════════════╤═╝                     │
        │                    │                     │        new()          │                         │                              │                       │
        │                    │                     │──────────────────────>│                         │                              │                       │
        │                    │                     │                       │                         │                              │                       │
        │                    │                     │                       │             new()       │                              │                       │
        │                    │                     │───────────────────────────────────────────────────────────────────────────────>│                       │
        │                    │                     │                       │                         │                              │                       │
        │                    │                     │                       │                     registerPID()                      │                       │
        │                    │                     │───────────────────────────────────────────────────────────────────────────────────────────────────────>│
        │                    │                     │                       │                         │                              │                       │
        │                    │                     │                       │         listen(port)    │                              │                       │
        │                    │                     │───────────────────────────────────────────────────────────────────────────────>│                       │
        │                    │                     │                       │                         │                              │                       │
        │                    │                     │     listen(port)      │                         │                              │                       │
        │                    │                     │──────────────────────>│                         │                              │                       │
        │                    │                     │                       │                         │                              │                       │
        │                    │                     │            Open browser                         │                              │                       │
        │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │                       │
      User            ┌──────┴──────┐          ┌───┴──┐           ┌────────┴───────┐           ┌─────┴─────┐                   ┌────┴────┐          ┌───────┴──────┐
       ┌─┐            │CommandRouter│          │Server│           │WebSocketHandler│           │PeerManager│                   │WebServer│          │ProcessTracker│
       ║"│            └─────────────┘          └──────┘           └────────────────┘           └───────────┘                   └─────────┘          └──────────────┘
       └┬┘
       ┌┼┐
        │
       ┌┴┐

## Notes

- Server automatically selects port starting from 10000
- If port unavailable, tries next port (up to 100 attempts)
- PeerManager configures both mDNS (local) and DHT (global) discovery
- NAT traversal includes Circuit Relay v2, hole punching, AutoRelay, and UPnP
- ProcessTracker uses file locking for safe concurrent PID registration
- Browser opens automatically unless --noopen flag specified
