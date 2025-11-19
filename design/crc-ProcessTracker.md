# ProcessTracker

**Source Spec:** main.md

## Responsibilities

### Knows
- pidFile: Path to PID tracking file (/tmp/.p2p-webapp)
- pids: List of running instance PIDs

### Does
- registerPID: Add current process PID to tracking file
- unregisterPID: Remove PID from tracking file
- listPIDs: Get all tracked PIDs
- verifyPID: Check if PID is actual running p2p-webapp instance
- cleanStale: Remove stale entries (PIDs no longer running)
- killPID: Terminate specific instance by PID (SIGTERM first, wait 5s, then SIGKILL if needed)
- killAll: Terminate all tracked instances (SIGTERM first, wait 5s, then SIGKILL if needed)
- lockFile: Acquire file lock for safe concurrent access
- unlockFile: Release file lock

## Collaborators

- Server: Registers PID on startup, unregisters on shutdown
- ExtractCommand: Registers PID while running
- gopsutil library: Verifies PIDs are actual p2p-webapp processes

## Sequences

- seq-process-register.md: PID registration on startup
- seq-process-list.md: Listing and cleaning PIDs
- seq-process-kill.md: Terminating instance by PID
