# Sequence: Store File Operation

**Source Spec:** main.md - storeFile(path, content, directory) (lines 316-320)

**Participants:**
- Client A: P2PWebAppClient instance storing file or directory
- WebSocketHandler: Enforces file ownership security
- PeerManager: Manages peer HAMTDirectories and CID tracking
- HAMTDirectory: IPFS/boxo data structure for peer file storage

**Flow:**

This sequence demonstrates the file storage operation with ownership enforcement. The client calls `storeFile(path, content, directory)` where content is null for directories or a string for files. The operation implicitly operates on the peer associated with the WebSocket connection. The WebSocketHandler enforces this security model by extracting the peerID from the connection context. The PeerManager creates a file or directory node in IPFS, updates the peer's HAMTDirectory at the specified path, and updates the peer's CID. The peer also pins its directory for persistence.

```
     ┌────────┐                             ┌────────────────┐                              ┌───────────┐                          ┌─────────────┐
     │Client A│                             │WebSocketHandler│                              │PeerManager│                          │HAMTDirectory│
     └────┬───┘                             └────────┬───────┘                              └─────┬─────┘                          └──────┬──────┘
          │                                          │                                            │                                       │
          │                                          │                          ╔═════════════════╧════╗                                  │
══════════╪══════════════════════════════════════════╪══════════════════════════╣ Store File Operation ╠══════════════════════════════════╪═════════════════════════════════════════════
          │                                          │                          ╚═════════════════╤════╝                                  │
          │                                          │                                            │                                       │
          │────┐                                     │ ╔═════════════════════════════════════╗    │                                       │
          │    │ storeFile(path, content, directory) │ ║Implicit peerID (connection's peer) ░║    │                                       │
          │<───┘                                     │ ╚═════════════════════════════════════╝    │                                       │
          │                                          │                                            │                                       │
          │   storeFile(path, content, directory)    │                                            │                                       │
          │─────────────────────────────────────────>│                                            │                                       │
          │                                          │                                            │                                       │
          │                                          │────┐                          ╔════════════╧════════════════════════╗              │
          │                                          │    │ enforceFileOwnership()   ║Get peerID from connection          ░║              │
          │                                          │<───┘                          ║Ensure request operates on own peer  ║              │
          │                                          │                               ╚════════════╤════════════════════════╝              │
          │                                          │storeFile(peerID, path, content, directory) │                                       │
          │                                          │───────────────────────────────────────────>│                                       │
          │                                          │                                            │                                       │
          │                                          │                                            │────┐                                  │
          │                                          │                                            │    │ Get HAMTDirectory from           │
          │                                          │                                            │<───┘ peerDirectories[peerID]          │
          │                                          │                                            │                                       │
          │                                          │                                            │                                       │
          │                                          │                                            │                                       │
          │                                          │                            ╔══════╤════════╪═══════════════════════════════════════╪══════════════════════════════════╗
          │                                          │                            ║ ALT  │  content is not null (file)                    │                                  ║
          │                                          │                            ╟──────┘        │                                       │                                  ║
          │                                          │                            ║               │        Add file entry at path         │ ╔═════════════════════╗          ║
          │                                          │                            ║               │──────────────────────────────────────>│ ║Store file content  ░║          ║
          │                                          │                            ║               │                                       │ ║Set directory=false  ║          ║
          │                                          │                            ╠═══════════════╪═══════════════════════════════════════╪══════════════════════════════════╣
          │                                          │                            ║ [content is null (directory)]                         │                                  ║
          │                                          │                            ║               │     Add directory entry at path       │ ╔════════════════════╗           ║
          │                                          │                            ║               │──────────────────────────────────────>│ ║No content stored  ░║           ║
          │                                          │                            ║               │                                       │ ║Set directory=true  ║           ║
          │                                          │                            ╚═══════════════╪═══════════════════════════════════════╪═╚════════════════════╝═══════════╝
          │                                          │                                            │                                       │
          │                                          │                                            │────┐                                  │
          │                                          │                                            │    │ Calculate new root CID           │
          │                                          │                                            │<───┘                                  │
          │                                          │                                            │                                       │
          │                                          │                                            │────┐                                  │
          │                                          │                                            │    │ Update peerDirectoryCIDs[peerID] │
          │                                          │                                            │<───┘                                  │
          │                                          │                                            │                                       │
          │                                          │            Return new root CID             │                                       │
          │                                          │<───────────────────────────────────────────│                                       │
          │                                          │                                            │                                       │
          │      storeFile response (root CID)       │                                            │                                       │
          │<─────────────────────────────────────────│                                            │                                       │
          │                                          │                                            │                                       │
          │────┐                                  ╔══╧══════════════════════════════════╗         │                                       │
          │    │ Update local rootDirectory CID   ║Client persists CID for restoration ░║         │                                       │
          │<───┘                                  ╚══╤══════════════════════════════════╝         │                                       │
     ┌────┴───┐                             ┌────────┴───────┐                              ┌─────┴─────┐                          ┌──────┴──────┐
     │Client A│                             │WebSocketHandler│                              │PeerManager│                          │HAMTDirectory│
     └────────┘                             └────────────────┘                              └───────────┘                          └─────────────┘
```

**Key Points:**

1. **New Signature**: `storeFile(path: string, content: string | null, directory: bool)`
   - `path`: Full pathname in HAMTDirectory tree (e.g., "docs/readme.txt")
   - `content`: String for files, null for directories
   - `directory`: Boolean flag indicating if this is a directory node
   - Error if `directory` is false but `content` is null

2. **Implicit Peer Context**: The storeFile() operation has no peerID parameter. The client can only modify its own peer's directory.

3. **Ownership Enforcement**: WebSocketHandler's enforceFileOwnership() extracts the peerID from the connection context and ensures the request operates only on that peer.

4. **IPFS Node Creation**: PeerManager creates a file or directory node in IPFS and stores it via ipfs-lite, which returns the new node with CID.

5. **Path-based Update**: Uses the path to find the correct subdirectory in the peer's HAMTDirectory and adds the new node there.

6. **CID Management**: After modifying the HAMTDirectory, PeerManager updates the peer's CID (stored in peerDirectoryCIDs[peerID]).

7. **Pinning**: The peer pins its directory for persistence across sessions.

8. **Security Model**: The implicit peerID design prevents clients from modifying other peers' directories. Only listFiles() has an explicit peerID parameter because it's a read-only query operation.

**Related:**
- crc-P2PWebAppClient.md: Client-side file storage
- crc-PeerManager.md: Server-side HAMTDirectory management
- crc-WebSocketHandler.md: Ownership enforcement
- seq-list-files.md: Related file list retrieval sequence
