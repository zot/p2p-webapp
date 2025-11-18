# CommandRouter

**Source Spec:** main.md

## Responsibilities

### Knows
- args: Command-line arguments
- flags: Parsed command flags (--dir, --noopen, -v, -p)
- subcommand: Identified subcommand or default (server)

### Does
- parseArgs: Parse command-line arguments
- routeCommand: Route to appropriate command handler
- handleServer: Start server (default behavior)
- handleExtract: Extract bundled site to current directory
- handleBundle: Bundle site directory into binary
- handleLs: List files in bundled site
- handleCp: Copy files from bundle to destination
- handlePs: List running instance PIDs
- handleKill: Kill specific instance
- handleKillAll: Kill all instances
- handleVersion: Display version

## Collaborators

- Server: Starts server for default command
- BundleManager: Executes bundle operations
- ProcessTracker: Manages process tracking commands

## Sequences

- seq-command-route.md: Command routing and execution
