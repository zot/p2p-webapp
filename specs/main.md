# Form: a single Go executable for the server and a TypeScript library for the browser
- the TypeScript library (`pkg/client/`) provides a clean API over the WebSocket protocol
  - located at `pkg/client/src/`
  - compiled to ES modules and bundled with the demo
  - promise-based API for all operations
  - `connect(peerKey?)` connects to server and initializes peer in one call
    - merges WebSocket connection and peer initialization for simplicity
    - returns `[peerID, peerKey]` tuple
    - internally performs WebSocket connect followed by peer protocol message
  - uses protocol-based addressing with (peer, protocol) tuples instead of connection IDs
  - `start(protocol, onData)` registers listener that receives (peer, data) for all messages on that protocol
  - `send(peer, protocol, data): Promise<void>` sends data directly using peer+protocol addressing
    - returns Promise that resolves when delivery is confirmed
    - client library manages ack numbers internally (upward counting)
    - consumers don't see ack numbers, only get delivery confirmation via promise resolution
  - server manages all connection lifecycle, retry, and reliability concerns transparently
  - listeners automatically removed on `stop()` or WebSocket disconnect
  - the demo (`internal/commands/demo/index.html`) demonstrates how to use the library
  - users can get the library from the demo directory after building
- implements a web server to launch a site on a random port on localhost
  - **SPA routing support**: automatically serves `index.html` for routes without file extensions
  - preserves URL path for client-side routing (e.g., `/settings`, `/adventure/world/123`)
  - real files (with extensions) served normally or return 404 if not found
- uses [ipfs-lite](https://github.com/hsanjuan/ipfs-lite) to manage the IPFS connection
- manages peers for browser connections

# Peer Discovery
p2p-webapp uses two complementary discovery mechanisms to find and connect to peers:

## mDNS (Local Discovery)
- Automatically discovers peers on the same local network
- Zero configuration - works out of the box on LANs
- Ideal for local development and same-network collaboration
- Low latency peer discovery (typically sub-second)

## DHT (Global Discovery)
- Distributed Hash Table for discovering peers across the internet
- Enables global peer connectivity beyond local networks
- Bootstraps using well-known IPFS DHT nodes
- **Integrated with GossipSub**: peers advertise and discover topic subscriptions via DHT
- Provides resilient, decentralized peer routing

Both mechanisms work simultaneously and automatically - peers discovered via mDNS are used for fast local connections while DHT provides global reach.

## NAT Traversal
For peers behind NATs/firewalls (typical for home networks):
- **Circuit Relay v2**: enables connections via relay servers when direct connection isn't possible
- **Hole Punching**: attempts direct NAT traversal when both peers support it
- **AutoRelay**: automatically finds and uses public relay nodes from the IPFS network
- **NAT Port Mapping**: attempts UPnP/NAT-PMP for automatic port forwarding

# Configuration
p2p-webapp supports optional configuration via a `p2p-webapp.toml` file placed at the site root (same level as html/, ipfs/, storage/ directories).

## Configuration File Location
- **Directory mode (--dir)**: Place `p2p-webapp.toml` in the base directory
- **Bundle mode**: Include `p2p-webapp.toml` in the root before bundling

## Configuration Precedence
1. Command-line flags (highest priority)
2. Configuration file (`p2p-webapp.toml`)
3. Default values (lowest priority)

## Configuration Options

### [server]
- `port`: Starting port number (default: 10000)
- `portRange`: Number of ports to try if starting port unavailable (default: 100)
- `maxHeaderBytes`: Maximum size of request headers in bytes (default: 1048576 = 1MB)

### [server.timeouts]
- `read`: Maximum duration for reading entire request (default: "15s")
- `write`: Maximum duration for writing response (default: "15s")
- `idle`: Maximum duration to wait for next request (default: "60s")
- `readHeader`: Maximum duration for reading request headers (default: "5s")

### [http]
- `cacheControl`: Cache-Control header for all files (default: "no-cache, no-store, must-revalidate")
  - Development: `"no-cache, no-store, must-revalidate"` (prevents caching)
  - Production: `"public, max-age=3600, immutable"` (enables caching)

### [http.security]
- `xContentTypeOptions`: X-Content-Type-Options header (default: "nosniff")
- `xFrameOptions`: X-Frame-Options header (default: "DENY")
- `contentSecurityPolicy`: Content-Security-Policy header (default: "" = not set)

### [http.cors]
- `enabled`: Enable CORS headers (default: false)
- `allowOrigin`: Access-Control-Allow-Origin header (default: "")
- `allowMethods`: Access-Control-Allow-Methods header (default: [])
- `allowHeaders`: Access-Control-Allow-Headers header (default: [])

### [websocket]
- `checkOrigin`: Validate WebSocket origin (default: false = allow all)
- `allowedOrigins`: List of allowed origins (requires checkOrigin = true)
- `readBufferSize`: WebSocket read buffer in bytes (default: 1024)
- `writeBufferSize`: WebSocket write buffer in bytes (default: 1024)

### [behavior]
- `autoExitTimeout`: Auto-exit timeout when no connections remain (default: "5s")
- `autoOpenBrowser`: Automatically open browser on startup (default: true)
- `linger`: Keep server running after all connections close (default: false)
- `verbosity`: Verbosity level 0-3 (default: 0)

### [files]
- `indexFile`: File to serve for SPA routes (default: "index.html")
- `spaFallback`: Enable SPA routing fallback (default: true)

### [p2p]
- `protocolName`: Reserved libp2p protocol name for file list queries (default: "/p2p-webapp/1.0.0")
- `fileUpdateNotifyTopic`: Optional topic for file availability notifications (default: "" = disabled)
  - When configured and peer is subscribed to the topic, file changes trigger notifications
  - Message format: `{"type":"p2p-webapp-file-update","peer":"<peerID>"}`
  - Applications can use this to automatically refresh file lists when peers update their files
  - Privacy-friendly: only publishes when explicitly subscribed to the topic

## Example Configuration

See `p2p-webapp.example.toml` for a fully documented example configuration file.

# Commands
- **Default behavior (no subcommand)**
  - running `./p2p-webapp` without a subcommand starts the server
  - operates in one of two modes
    - **default mode** (no --dir flag): serves directly from bundled site without filesystem extraction
      - requires binary to be bundled (use `bundle` command)
      - creates `.p2p-webapp-storage` directory in current directory for peer keys
      - efficient - no extraction needed, serves directly from ZIP
    - **directory mode** (with --dir flag): serves from filesystem directory
      - directory must contain these subdirectories
        - html: website to serve, must contain index.html
          - launches this in a browser by default
        - ipfs: content to make available in IPFS (optional)
        - storage: server storage (peer keys, etc.)
      - use after extracting with `extract` command
      - example: `./p2p-webapp --dir .`
  - hosts WebSocket-based JSON-RPC protocol service for the site
    - use text-based JSON for commands
    - each side pushes commands to the other
  - flags
    - --dir DIR: directory to serve from (if not specified, serves from bundled site)
    - --noopen: do not open browser automatically
    - -v, --verbose: verbose output (can be specified multiple times: -v, -vv, -vvv)
      - level 1: log peer creation, connections, and messages
      - level 2+: additional debug information
    - -p, --port PORT: specify port to listen on (default: auto-select starting from 10000)
      - if port not available, automatically tries the next port (up to 100 attempts)
      - example: `./p2p-webapp -p 8080`
- **extract**
  - extracts the bundled site from the binary to the current directory
  - current dir must be empty
  - extracts html/, ipfs/, and storage/ subdirectories
  - does NOT start the server - run `p2p-webapp --dir .` afterwards to serve
  - the default bundled site is a chatroom application that demonstrates key features
    - uses listPeers to discover peers subscribed to the chat topic
    - displays a list of peers on the right side
      - first entry: "Chat room" (room chat mode)
      - remaining entries: individual users
    - two communication modes indicated at the top of the view
      - room chat (default)
        - uses pubsub for group messaging
        - selected by clicking "Chat room" in the peer list
      - direct messaging
        - uses peer-to-peer protocol for private messages
        - selected by clicking a user in the peer list
    - demo listener architecture
      - single protocol listener registered at startup with `start(protocol, onData)`
      - **room chat**: `subscribe` callback always stores messages via `storeRoomMessage()`
        - stores messages even when in DM mode (never lost)
        - only displays messages if currently in room mode
        - when switching to room mode, `displayStoredRoomMessages()` shows all stored messages
      - **direct messages**: protocol listener callback receives `(peer, data)` and:
        - always stores messages via `storeDMMessage(peer, ...)` (never lost)
        - only displays via `addMessage()` if currently viewing that DM
      - `switchToDM()` simply switches UI view - no connection management needed
      - server transparently manages all stream lifecycle and reliability
    - file browser feature
      - **Browse Files button**: Located in the top-right header
        - Context-aware label changes based on current selection:
          - "Browse My Files" when chatroom is selected (default)
          - "Browse Peer's Files" when a peer is selected
      - **File Browser Modal**: Displays IPFS files in a hierarchical tree structure
        - **Peer selector dropdown**: Switch between viewing own files and peer files
        - **Hierarchical tree display**:
          - Folders and files displayed with icons (📁 for directories, 📄 for files)
          - Expandable/collapsible folders to navigate nested structure
          - Indentation shows hierarchy depth
          - Files display name and CID
        - **Folder selection**:
          - Click a folder to select it (visual highlight)
          - Selected folder becomes the target for uploads and new directories
          - When a folder is created, it is automatically selected
          - Root (no selection) is the default target
        - **File operations** (context-dependent):
          - **Own files** (full access):
            - Upload files: drag-and-drop or file picker button
              - Files uploaded to currently selected folder (prefixed with folder path)
            - Create new directory: button with name prompt
              - Directory created in currently selected folder (prefixed with folder path)
              - Newly created directory is automatically selected
            - Download files: click to retrieve by CID
          - **Peer files** (read-only):
            - Download only: retrieve files by CID
            - Upload and create directory buttons disabled
        - **Implementation details**:
          - Uses `listFiles(peerid): Promise<{rootCID, entries}>` to fetch directory tree
          - Converts flat pathname entries to hierarchical tree structure
          - Stores expanded/collapsed state per directory
          - Uses `storeFile(path, content)` for file uploads and `createDirectory(path)` for directory creation
          - Uses `getFile(cid): Promise<FileContent>` for downloads
          - Modal overlay with close button to return to chat
- **bundle**
  - creates a standalone binary with a site bundled into it
  - works with both bundled and unbundled source binaries (replaces existing bundle if present)
  - no compilation tools needed - works out of the box
  - usage: `./p2p-webapp bundle [site-directory] -o [output-binary]`
  - site directory must contain
    - html/: website files (must contain index.html)
    - ipfs/: IPFS content (optional)
    - storage/: storage directory (optional, created if missing)
  - example: `./p2p-webapp bundle my-site -o my-app`
  - the output binary can be distributed and will extract/serve the bundled site
  - uses ZIP append technique - appends ZIP archive + footer to binary
    - footer contains: magic marker (8 bytes) + offset (8 bytes) + size (8 bytes)
    - Go binaries can have trailing data without affecting execution
    - on startup, checks for bundled content and extracts if found
- **version**
  - displays the current version
- **ls**
  - lists files available in the bundled site
  - shows files that can be copied with the cp command
  - usage: `./p2p-webapp ls`
  - displays file names in a clean list format
  - useful for discovering what files are available before using cp
  - reads directly from the bundled content (no extraction needed)
- **cp**
  - copies files from the bundled site to a target location
  - supports glob patterns for source selection (e.g., `*.js`, `client.*`)
  - similar to UNIX cp command but operates on bundled site files
  - usage: `./p2p-webapp cp SOURCE... DEST`
    - SOURCE: one or more glob patterns or file names from the bundled site
    - DEST: destination directory (must exist or will be created)
  - examples
    - `./p2p-webapp cp client.js my-project/` - copy single file
    - `./p2p-webapp cp client.* my-project/` - copy client.js and client.d.ts
    - `./p2p-webapp cp *.js *.html my-project/` - copy multiple patterns
  - creates destination directory if it doesn't exist
  - preserves file names (no renaming)
  - validates that at least one file matches the patterns
  - reads directly from the bundled content (no extraction needed)
  - error handling
    - fails if no files match the patterns
    - fails if destination is not a directory (when copying multiple files)
    - fails if binary is not bundled
- **ps**
  - lists process IDs for all running p2p-webapp instances
  - usage: `./p2p-webapp ps`
  - flags
    - -v, --verbose: also shows command line arguments for each instance
  - displays PID and optionally command line args in a clean table format
  - automatically cleans stale entries from the tracking file
- **kill**
  - terminates a specific running instance by PID
  - usage: `./p2p-webapp kill PID`
  - validates that PID is actually an p2p-webapp instance
  - uses graceful shutdown procedure:
    1. First sends SIGTERM (15) for clean shutdown
    2. Waits up to 5 seconds for process to terminate
    3. If still running after 5 seconds, sends SIGKILL (9) to force termination
  - removes the instance from the tracking file
  - returns error if PID is not found or not an p2p-webapp process
- **killall**
  - terminates all running p2p-webapp instances
  - usage: `./p2p-webapp killall`
  - kills all instances tracked in the PID file using graceful shutdown:
    1. Sends SIGTERM (15) to all instances for clean shutdown
    2. Waits up to 5 seconds for each process to terminate
    3. For any processes still running, sends SIGKILL (9) to force termination
  - automatically validates and cleans up stale entries
  - reports how many instances were killed

# Process Tracking
p2p-webapp maintains a JSON list of running instance PIDs for process management:
- **PID file location**: `/tmp/.p2p-webapp` on systems with `/tmp`
- **File locking**: Uses file locking to prevent race conditions during read/write operations
- **Process verification**: When reading the file, automatically verifies PIDs are actual running p2p-webapp instances
- **Auto-correction**: Removes stale entries for processes that are no longer running
- **Library**: Uses `github.com/shirou/gopsutil` for cross-platform process management
- **Registration**: The server (when started without a subcommand) and `extract` command automatically register their PIDs on startup
- **Cleanup**: PIDs are removed when the process exits normally or via `kill`/`killall` commands

The tracking system ensures:
1. File is locked during read/write operations to prevent corruption
2. Only verified p2p-webapp processes are listed
3. Stale entries are automatically cleaned up
4. Safe concurrent access from multiple instances

## Signal Handling
p2p-webapp responds to the following signals with graceful shutdown:
- **SIGHUP (1)**: Graceful shutdown, cleanup resources, close connections
- **SIGINT (2)**: Graceful shutdown (typically sent by Ctrl+C)
- **SIGTERM (15)**: Graceful shutdown (default kill signal)

**Graceful shutdown procedure:**
1. Stop accepting new WebSocket connections
2. Close all active WebSocket connections gracefully
3. Stop all libp2p peers and close streams
4. Close DHT and other libp2p resources
5. Unregister PID from process tracking file
6. Exit cleanly

**Why these signals?**
- SIGHUP: Traditional "hangup" signal, often used for reload/restart in daemons
- SIGINT: Interactive interrupt (Ctrl+C), allows user to stop server cleanly
- SIGTERM: Standard termination signal, allows orchestrators/scripts to stop server cleanly

# Message format
- a request has a requestID
  - starts at 0 and counts up
- a response contains the ID of its request
- message definitions are name(param...) params are strings unless otherwise indicated
  - Each message corresponds to a JSON-encodable Go struct in the server and a TypeScript struct in the browser
  - JSON properties are lowercase

# Peer directories
Each peer has a [HAMTDirectory](https://pkg.go.dev/github.com/ipfs/boxo@v0.35.2/ipld/unixfs/io#HAMTDirectory), populated at creation time by the `Peer()` message. After that, the `storeFile()` message can add or remove entries.
The peer also keeps the CID of its directory which it updates after any change.

**IMPORTANT**: each peer pins its directory.

## Peer Lifecycle
Peers and their WebSocket connections are ephemeral. The client provides peerKey and rootDirectory CID to restore a peer's identity and directory state across sessions. The storeFile() and removeFile() operations implicitly operate on the peer associated with the WebSocket connection sending the request.

# Client Request messages

## Peer(peerkey?, rootDirectory?: CID)
- Create a new peer for this websocket connection with peerkey. If none given, use a fresh peerkey.
- Must be the first command from the browser for a websocket connection
- Cannot be sent more than once
- on the server, the new peer is associated with this WebSocket connection
- If a peer with the resulting peer ID is already registered (e.g., duplicate browser tab with same peer key), returns an error
  - This prevents multiple WebSocket connections from using the same peer identity
  - Common cause: user opens the same app in multiple browser tabs with the same stored peer key
- rootDirectory is an optional string representation of the peer directory's CID
  - if present, initialize the peer's directory
  - if absent, the peer's directory remains nil
### Response: [peerid, peerkey] or error

## start(protocol)
- Start a protocol and register to receive messages
- Must be called before sending on the protocol
- Protocol validation: cannot send unless protocol is started
### Response: null or error

## stop(protocol)
- Stop a protocol and remove message listener
### Response: null or error

## send(peer, protocol, data: any, ack: number = -1)
- Send data to a peer on a protocol
- Uses (peer, protocol) addressing instead of connection IDs
- Server manages stream lifecycle transparently
- Validation: protocol must be started first
- `ack` parameter: if non-negative (>= 0), server will send an `ack` message back to this client when delivery to the peer is confirmed
  - If `ack` is -1 or not provided, no acknowledgment is sent
  - Ack numbers must be non-negative integers (0, 1, 2, ...)
  - This is an internal protocol detail - client library manages ack numbers automatically
### Response: null or error

## subscribe(topic: string)
- Subscribe to a topic to receive published messages
- Automatically monitors the topic for peer join/leave events
- Server will send `peerChange` notifications when peers join or leave
### Response: null or error

## publish(topic, data: any)
### Response: null or error

## unsubscribe(topic)
- Unsubscribe from a topic
- Stops monitoring peer join/leave events for this topic
### Response: null or error

## listPeers(topic: string)
- Get list of peers subscribed to a topic
### Response: array of peer IDs or error

## listFiles(peerid: string): Promise<{rootCID: string, entries: FileEntries}>
### Client TS code
1. Create a promise and add the resolve/reject pair to the listFiles handler for peerid (create if needed), this will be called later on when the client receives the `peerFiles` server message
2. If the handler list was just created, send this client message to Go, otherwise do not send because one is already pending
3. Return the promise that will resolve with `{rootCID, entries}` when the `peerFiles` message is received

### Go code
Requests a list of files for a peer. The response will go to the client as a `peerFiles` server message which the client library uses to resolve the promise.

Libp2p messaging in this section uses a reserved libp2p peer messaging protocol named `p2p-webapp`.

If peerid is the local peer, returns null and spawn a goroutine to send the `peerFiles` server message
Otherwise ask the requested peer for its files
1. If there is already a fileList handler for that peer return null because there is already a pending message
2. Otherwise register a fileList handler for the peer
3. Send `getFileList()` libp2p message to the requested peer using the reserved `p2p-webapp` protocol
   - if there is an error
     - remove the handler
     - return the error
   - otherwise spawn a goroutine for the rest of this operation and return null
4. When the requested peer receives a `getFileList` libp2p message on the reserved protocol, it will send a `fileList(CID, directory)` libp2p message back to this peer, also in the reserved protocol.
5. Upon receiving a `fileList` libp2p message on the reserved protocol (see step 4), send the `peerFiles` server message to the client (see response)
### Response: null or error (will also send a server `peerFiles` message)
Also generates a server message `peerFiles(peerid, CID, entries)` where entries contains JSON object with an entry for each item in the peer's entire HAMTDirectory tree: `{PATHNAME: entry}`. PATHNAME is the unix-style relative path for a tree entry, starting at the top of the tree.

Entries:
  - `{type: "directory", cid: CID}`
  - `{type: "file", cid: CID, mimeType: MIMETYPE}`

Example entries object:
```json
{
  "docs/readme.txt": {
    "type": "file",
    "cid": "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V",
    "mimeType": "text/plain"
  },
  "images": {
    "type": "directory",
    "cid": "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG"
  },
  "images/photo.jpg": {
    "type": "file",
    "cid": "QmPZ9gcCEpqKTo6aq61g2nXGUhM4iCL3ewB6LDXZCtioEB",
    "mimeType": "image/jpeg"
  }
}
```

## getFile(cid: string): Promise<FileContent>
- Get IPFS content by CID
- Returns Promise that resolves with file content or rejects on error
- Content format for files:
  - `{type: "file", mimeType: string, content: string}` (content is base64-encoded)
  - **Why base64?** Binary files (images, PDFs, executables) contain arbitrary bytes that aren't valid UTF-8. JSON can only safely encode UTF-8 strings, so base64 encoding is required to transmit binary data without corruption.
- Content format for directories:
  - `{type: "directory", entries: {PATHNAME: CID, ...}}`
- Internally sends a server message `gotFile(cid, {success: bool, content})` which the client library uses to resolve/reject the promise
### Response: null or error (promise resolution handled by client library)

## storeFile(path: string, content: string | Uint8Array)
Make a file node and store it in ipfs-lite, which will return the new node.
Content can be either a string (which will be UTF-8 encoded) or binary data as Uint8Array.
Use path to find the correct subdirectory in the peer's directory and add the new node there.
Update the peer's CID after the change.

**File Availability Notifications**: If `fileUpdateNotifyTopic` is configured in settings and the peer is subscribed to that topic, the server publishes a notification message after successfully storing the file. This allows other peers to be notified of file changes and refresh their file lists automatically.

### Response: CID string of the stored file node, or error

## createDirectory(path: string)
Make a directory node and store it in ipfs-lite, which will return the new node.
Use path to find the correct subdirectory in the peer's directory and add the new node there.
Update the peer's CID after the change.

**File Availability Notifications**: If `fileUpdateNotifyTopic` is configured in settings and the peer is subscribed to that topic, the server publishes a notification message after successfully creating the directory. This allows other peers to be notified of directory changes and refresh their file lists automatically.

### Response: CID string of the stored directory node, or error

## removeFile(path: string)
Use path to find the correct directory and remove the element from it.

**File Availability Notifications**: If `fileUpdateNotifyTopic` is configured in settings and the peer is subscribed to that topic, the server publishes a notification message after successfully removing the file. This allows other peers to be notified of file changes and refresh their file lists automatically.

### Response: null or error

# Server Request messages

## peerData(peer, protocol, data: any)
- Notifies client of data received from a peer on a protocol
- Sent to the protocol listener registered with `start(protocol, onData)`
- Listener receives `(peer, data)` for all messages on the protocol
### Response: null or error

## topicData(topic, peerId, data: any)
### Response: null or error

## peerChange(topic: string, peerId: string, joined: boolean)
- Notifies client that a peer joined or left a subscribed topic
- `joined` is true when peer joins, false when peer leaves
- Automatically sent for all subscribed topics (no separate monitoring needed)
### Response: null or error

## peerFiles(peerid, CID, fileObj)
- Notifies the client of the current files in the given peer. This is sent to the client whenever the peer receives a `peerFiles` libp2p message on the reserved `p2p-webapp` protocol.
- See the listFiles response section for the format of fileObj
### Response: null or error

## ack(ack: number, optionalData: any)
- Notifies client that a message with the given ack number was successfully delivered to the peer
  - optionalData is the data the peer responded with -- not present unless the client message specifies it
- Only sent in response to `send` requests with a non-negative `ack` parameter (>= 0)
- If a `send` request has `ack` = -1 or no ack parameter, no `ack` message is sent back
- Client library manages ack numbers internally and notifies consumers via callback
### Response: null or error

# Implementation Details

## Virtual Connection Model
The server manages all connection lifecycle transparently using a virtual connection model:
- Client uses (peer, protocol) tuples for addressing instead of connection IDs
- Server internally manages libp2p streams using `"peerID:protocol"` keys
- Streams are created on-demand when sending data
- Streams are automatically reused for multiple messages
- Server handles all retry logic, buffering, and reliability concerns
- Client API is simplified - no connection state management needed

## Private Address Support
libp2p by default blocks connections on private/localhost addresses. To enable local development and testing:
- `internal/peer/manager.go` implements `allowPrivateGater` type
- This ConnectionGater returns `true` for all connection gating methods
- Passed to `libp2p.New()` via `libp2p.ConnectionGater(&allowPrivateGater{})`
- Allows peers on the same machine to connect via localhost addresses

## Verbose Logging
Complete verbose logging with peer aliases:
- Manager tracks: `peerAliases` map (peerID→alias), `aliasCounter`, `verbosity` level
- Peer struct includes `alias` field
- Aliases generated as "peer-a", "peer-b", etc.
- Verbose output preceded by peer alias: `[peer-a] message`
- Logging at different verbosity levels:
  - Level 1: peer creation, connections, messages
  - Level 2: WebSocket message send/receive with request IDs
  - Level 3+: additional debug information

## Client Library Message Handling
The TypeScript client library implements robust message handling with proper ordering:

### Sequential Message Processing
- Server-initiated messages are queued and processed strictly sequentially
- Messages processed one at a time via `processMessageQueue()` with async/await
- Ensures message ordering is preserved (critical for peerData)
- Each message fully processed before the next begins
- Error handling wraps each message processing to continue on failure
- Response messages (replies to client requests) are NOT queued - processed immediately

### Protocol-Based Message Routing
- `start(protocol, onData)` registers a listener for a protocol
- Listener receives `(peer, data)` for all messages on that protocol
- `peerData` messages routed directly to the protocol listener
- No buffering needed - messages delivered immediately to registered listener
- `stop(protocol)` removes the listener and prevents further message delivery
- Simple, straightforward routing with no connection state to manage

### Async Callback Support
- All callbacks (ProtocolDataCallback, TopicDataCallback) support both sync and async
- Type signature: `(peer: string, data: any) => void | Promise<void>`
- Client library awaits callback completion with try/catch error handling
- Errors logged to console but don't stop message processing

### Protocol Lifecycle
- `start(protocol, onData)`: Register protocol listener, enable sending
- `peerData(peer, protocol, data)`: Route to protocol listener
- `send(peer, protocol, data)`: Validate protocol started, send message
- `stop(protocol)`: Remove listener, disable sending

### Message Acknowledgment
The client library provides delivery confirmation for all sent messages via Promise resolution:
- **Client API**: `send(peer, protocol, data): Promise<void>`
  - Returns a Promise that resolves when the message is successfully delivered to the peer
  - Client library automatically:
    - Assigns an internal ack number (starting from 0, incrementing)
    - Includes ack number in the send request to server
    - Stores the promise's resolve/reject functions in an internal map keyed by ack number
- **Server behavior**: Sends `ack(ack: number)` message when delivery confirmed
- **Client handling**: On receiving `ack` message:
  - Looks up promise resolve function in internal map by ack number
  - Calls resolve to fulfill the promise
  - Removes entry from map
- **Consumer perspective**: Simple promise-based interface, no ack number management needed

## SPA Routing Support
The server implements automatic SPA (Single Page Application) routing fallback:
- **Route detection**: Paths without file extensions are treated as SPA routes
- **Fallback behavior**: Non-existent routes serve `index.html` while preserving the URL
- **File serving**: Real files (with extensions) are served normally
- **404 handling**: Missing files with extensions return proper 404 errors
- **Cache prevention**: All files served with `Cache-Control: no-cache, no-store` headers to prevent browser caching
  - Ensures fresh content after code changes without manual cache clearing
  - Browser revalidates with server on every request
  - No need to clear browser cache during development
- **Implementation**: `internal/server/server.go:spaHandler()`

Examples:
- `/` → serves `html/index.html`
- `/settings` → serves `html/index.html` (URL stays `/settings`)
- `/adventure/world/123` → serves `html/index.html` (URL stays `/adventure/world/123`)
- `/main.js` → serves `html/main.js` (actual file)
- `/nonexistent.js` → returns 404 (file with extension not found)

## Demo Application

The bundled demo (`internal/commands/demo/index.html`) demonstrates the TypeScript client library and includes a P2P chatroom with file sharing capabilities.

### Connection Lifecycle

The demo displays a connection status indicator that accurately reflects the peer's connection state:

**Connection States:**
1. **"Connecting to server..."** (initial state)
   - Displayed while WebSocket connection is being established

2. **"Connecting to network..."** (with spinning indicator)
   - WebSocket is connected and peer identity is established
   - Protocol handlers are being registered
   - **Pubsub subscription is in progress** - joining the gossipsub topic
   - Indicator shows spinning animation (⟳) to indicate ongoing connection

3. **"Connected"** (final state)
   - All connection phases complete:
     - WebSocket connected ✓
     - Peer identity created ✓
     - Protocol handlers started ✓
     - **Pubsub topic joined** ✓
   - **Peer is now a member of the pubsub group** and can send/receive group messages
   - Ready for peer-to-peer messaging and file operations

**Implementation Details:**
- The `subscribe()` call blocks until the peer successfully joins the gossipsub topic on the server
- Server waits for `pubsub.Join(topic)` and `topic.Subscribe()` to complete before responding
- "Connected" status is only shown **after** `subscribe()` resolves, ensuring pubsub group membership
- This guarantees that when "Connected" is displayed, the peer is fully ready for all P2P operations

**Status Transitions:**
```javascript
// WebSocket connection
client = await connect();
setStatus('Connecting to network...', 'connecting');  // Show spinning

// Protocol registration
await client.start(PROTOCOL, onData);

// Pubsub subscription - BLOCKS until peer joins group
await client.subscribe(ROOM_TOPIC, onMessage, onPeerChange);

// Now fully connected to pubsub group
setStatus('Connected', 'connected');  // Stop spinning, show connected
```
