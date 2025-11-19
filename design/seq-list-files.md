# Sequence: File List Request

**Source Spec:** main.md - listFiles(peerid) (lines 258-281, 343-346)

**Participants:**
- Client A: P2PWebAppClient instance
- WebSocketHandler: Routes file list requests to PeerManager
- PeerManager: Handles both local and remote file list retrieval via p2p-webapp protocol
- Local Peer: The peer associated with Client A
- Remote Peer: The peer being queried for its file list

**Flow:**

This sequence demonstrates the file list retrieval flow with request deduplication. The client checks if a handler already exists for the target peerID before sending the request. For local peers, the HAMTDirectory is read directly. For remote peers, the reserved "p2p-webapp" libp2p protocol is used to exchange `getFileList()` and `fileList(CID, directory)` messages. The server sends `peerFiles(peerid, CID, entries)` to the client with structured entries containing full pathname tree and metadata. All pending promises for the same peerID are resolved when the response arrives.

```
                              ┌────────┐                                           ┌────────────────┐             ┌───────────┐                                         ┌──────────┐           ┌───────────┐
                              │Client A│                                           │WebSocketHandler│             │PeerManager│                                         │Local Peer│           │Remote Peer│
                              └────┬───┘                                           └────────┬───────┘             └─────┬─────┘                                         └─────┬────┘           └─────┬─────┘
                                   │                                                        │                           │                                                     │                      │
                                   │                                                        │                       ╔═══╧═══════════════╗                                     │                      │
═══════════════════════════════════╪════════════════════════════════════════════════════════╪═══════════════════════╣ Request File List ╠═════════════════════════════════════╪══════════════════════╪═══════════════════════════════════════════════════════
                                   │                                                        │                       ╚═══╤═══════════════╝                                     │                      │
                                   │                                                        │                           │                                                     │                      │
                                   │────┐                                                   │                           │                                                     │                      │
                                   │    │ listFiles(peerID)                                 │                           │                                                     │                      │
                                   │<───┘                                                   │                           │                                                     │                      │
                                   │                                                        │                           │                                                     │                      │
                                   │────┐                                                   │                           │                                                     │                      │
                                   │    │ Check fileListHandlers[peerID]                    │                           │                                                     │                      │
                                   │<───┘                                                   │                           │                                                     │                      │
                                   │                                                        │                           │                                                     │                      │
                                   │                                                        │                           │                                                     │                      │
          ╔══════╤═════════════════╪════════════════════════════════════════════════════════╪═══════════════════════════╪═════════════════════════════════════════════════════╪══════════════════════╪════════════════════════════════════════════╗
          ║ ALT  │  No existing handler for peerID                                          │                           │                                                     │                      │                                            ║
          ╟──────┘                 │                                                        │                           │                                                     │                      │                                            ║
          ║                        │────┐                                                   │                           │                                                     │                      │                                            ║
          ║                        │    │ Create handler, store in fileListHandlers[peerID] │                           │                                                     │                      │                                            ║
          ║                        │<───┘                                                   │                           │                                                     │                      │                                            ║
          ║                        │                                                        │                           │                                                     │                      │                                            ║
          ║                        │                   listFiles(peerID)                    │                           │                                                     │                      │                                            ║
          ║                        │───────────────────────────────────────────────────────>│                           │                                                     │                      │                                            ║
          ║                        │                                                        │                           │                                                     │                      │                                            ║
          ║                        │                                                        │    listFiles(peerID)      │                                                     │                      │                                            ║
          ║                        │                                                        │──────────────────────────>│                                                     │                      │                                            ║
          ║                        │                                                        │                           │                                                     │                      │                                            ║
          ║                        │                                                        │                           │                                                     │                      │                                            ║
          ║         ╔══════╤═══════╪════════════════════════════════════════════════════════╪═══════════════════════════╪═════════════════════════════════════════════════════╪══════════════════════╪══════════════════════════════════╗         ║
          ║         ║ ALT  │  peerID is local peer                                          │                           │                                                     │                      │                                  ║         ║
          ║         ╟──────┘       │                                                        │                           │                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │                           │────┐                                                │                      │                                  ║         ║
          ║         ║              │                                                        │                           │    │ Get HAMTDirectory from peerDirectories[peerID] │                      │                                  ║         ║
          ║         ║              │                                                        │                           │<───┘                                                │                      │                                  ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │                           │────┐                                                │                      │                                  ║         ║
          ║         ║              │                                                        │                           │    │ Spawn goroutine                                │                      │                                  ║         ║
          ║         ║              │                                                        │                           │<───┘                                                │                      │                                  ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │peerFiles(peerID, CID, entries) │                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │<───────────────────────────────│                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │                                │                                                     │                      │                                  ║         ║
          ║         ║              │         peerFiles(peerID, CID, entries)                │                                │                                                     │                      │                                  ║         ║
          ║         ║              │<───────────────────────────────────────────────────────│                                │                                                     │                      │                                  ║         ║
          ║         ╠══════════════╪════════════════════════════════════════════════════════╪═══════════════════════════╪═════════════════════════════════════════════════════╪══════════════════════╪══════════════════════════════════╣         ║
          ║         ║ [peerID is remote peer]                                               │                           │                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │                           │                     send("p2p-webapp", getFileList())                      │                                  ║         ║
          ║         ║              │                                                        │                           │───────────────────────────────────────────────────────────────────────────>│                                  ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │────┐                             ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │    │ handleGetFileList()         ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │<───┘                             ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │────┐                             ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │    │ Get local HAMTDirectory     ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │<───┘                             ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │                           │                send("p2p-webapp", fileList(CID, directory))                │                                  ║         ║
          ║         ║              │                                                        │                           │<───────────────────────────────────────────────────────────────────────────│                                  ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │                           │────┐                                                │                      │                                  ║         ║
          ║         ║              │                                                        │                           │    │ handleFileList(CID, directory)                 │                      │                                  ║         ║
          ║         ║              │                                                        │                           │<───┘                                                │                      │                                  ║         ║
          ║         ║              │                                                        │                           │                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │peerFiles(peerID, CID, entries) │                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │<───────────────────────────────│                                                     │                      │                                  ║         ║
          ║         ║              │                                                        │                                │                                                     │                      │                                  ║         ║
          ║         ║              │         peerFiles(peerID, CID, entries)                │                                │                                                     │                      │                                  ║         ║
          ║         ║              │<───────────────────────────────────────────────────────│                                │                                                     │                      │                                  ║         ║
          ║         ╚══════════════╪════════════════════════════════════════════════════════╪═══════════════════════════╪═════════════════════════════════════════════════════╪══════════════════════╪══════════════════════════════════╝         ║
          ║                        │                                                        │                           │                                                     │                      │                                            ║
          ║                        │────┐                                                   │                           │                                                     │                      │                                            ║
          ║                        │    │ Route to routePeerFiles()                         │                           │                                                     │                      │                                            ║
          ║                        │<───┘                                                   │                           │                                                     │                      │                                            ║
          ║                        │                                                        │                           │                                                     │                      │                                            ║
          ║                        │────┐                                                   │                           │                                                     │                      │                                            ║
          ║                        │    │ Get handlers from fileListHandlers[peerID]        │                           │                                                     │                      │                                            ║
          ║                        │<───┘                                                   │                           │                                                     │                      │                                            ║
          ║                        │                                                        │                           │                                                     │                      │                                            ║
          ║                        │────┐                                                   │                           │                                                     │                      │                                            ║
          ║                        │    │ Resolve all pending promises                      │                           │                                                     │                      │                                            ║
          ║                        │<───┘                                                   │                           │                                                     │                      │                                            ║
          ║                        │                                                        │                           │                                                     │                      │                                            ║
          ║                        │────┐                                                   │                           │                                                     │                      │                                            ║
          ║                        │    │ Remove fileListHandlers[peerID]                   │                           │                                                     │                      │                                            ║
          ║                        │<───┘                                                   │                           │                                                     │                      │                                            ║
          ╠════════════════════════╪════════════════════════════════════════════════════════╪═══════════════════════════╪═════════════════════════════════════════════════════╪══════════════════════╪════════════════════════════════════════════╣
          ║ [Handler already exists for peerID]                                             │                           │                                                     │                      │                                            ║
          ║                        │────┐                                               ╔═══╧═══════════════════════════╧═══════╗                                             │                      │                                            ║
          ║                        │    │ Add new handler to fileListHandlers[peerID]   ║Deduplication - reuse pending request ░║                                             │                      │                                            ║
          ║                        │<───┘                                               ╚═══╤═══════════════════════════╤═══════╝                                             │                      │                                            ║
          ╚════════════════════════╪════════════════════════════════════════════════════════╪═══════════════════════════╪═════════════════════════════════════════════════════╪══════════════════════╪════════════════════════════════════════════╝
                              ┌────┴───┐                                           ┌────────┴───────┐             ┌─────┴─────┐                                         ┌─────┴────┐           ┌─────┴─────┐
                              │Client A│                                           │WebSocketHandler│             │PeerManager│                                         │Local Peer│           │Remote Peer│
                              └────────┘                                           └────────────────┘             └───────────┘                                         └──────────┘           └───────────┘
```

**Key Points:**

1. **Request Deduplication**: The client checks fileListHandlers before sending a new request. If a request is already pending for the same peerID, the new handler is added to the list and the existing request is reused.

2. **Local vs Remote**: PeerManager determines if the peerID is local (in peerDirectories) or remote, and handles accordingly.

3. **Local Path**: Direct HAMTDirectory access from peerDirectories map, response sent via goroutine to avoid blocking.

4. **Remote Path**: Uses the reserved "p2p-webapp" libp2p protocol to exchange `getFileList()` and `fileList(CID, directory)` messages between peers.

5. **Promise Resolution**: When peerFiles arrives, all pending promises for that peerID are resolved simultaneously, then the handlers are removed.

6. **Entry Structure**: The `peerFiles(peerid, CID, entries)` server message contains:
   - `CID`: The peer's current directory root CID
   - `entries`: JSON object with full pathname tree structure: `{PATHNAME: entry}`
   - File entries: `{type: "file", cid: CID, mimeType: MIMETYPE}`
   - Directory entries: `{type: "directory", cid: CID}`
   - Example pathnames: "docs/readme.txt", "images", "images/photo.jpg"

**Related:**
- crc-P2PWebAppClient.md: Client-side file list handling
- crc-PeerManager.md: Server-side file list management
- crc-WebSocketHandler.md: Request routing
