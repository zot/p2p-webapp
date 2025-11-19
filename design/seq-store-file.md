# Sequence: Store File Operation

**Source Spec:** main.md - storeFile(path, content, directory)

**Participants:**
- Client A: P2PWebAppClient instance storing file or directory
- WebSocketHandler: Enforces file ownership security
- PeerManager: Manages peer HAMTDirectories and CID tracking
- HAMTDirectory: IPFS/boxo data structure for peer file storage

**Flow:**

This sequence demonstrates the file storage operation with ownership enforcement. The client calls storeFile() without specifying a peerID - the operation implicitly operates on the peer associated with the WebSocket connection. The WebSocketHandler enforces this security model by extracting the peerID from the connection context. The PeerManager updates the HAMTDirectory and returns the new root CID, which the client persists for future restoration.

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

1. **Implicit Peer Context**: The storeFile() operation has no peerID parameter. The client can only modify its own peer's directory.

2. **Ownership Enforcement**: WebSocketHandler's enforceFileOwnership() extracts the peerID from the connection context and ensures the request operates only on that peer.

3. **File vs Directory**: The content parameter distinguishes between files (content is a string) and directories (content is null). Both are stored in the HAMTDirectory with appropriate metadata.

4. **CID Management**: After modifying the HAMTDirectory, PeerManager calculates the new root CID and updates peerDirectoryCIDs[peerID]. This CID is returned to the client.

5. **Client Persistence**: The client receives the new root CID and persists it locally. This allows the client to restore the peer's directory state across sessions by providing the peerKey and rootDirectory CID to connect().

6. **Security Model**: The implicit peerID design prevents clients from modifying other peers' directories. Only listFiles() has an explicit peerID parameter because it's a read-only query operation.

**Related:**
- crc-P2PWebAppClient.md: Client-side file storage
- crc-PeerManager.md: Server-side HAMTDirectory management
- crc-WebSocketHandler.md: Ownership enforcement
- seq-list-files.md: Related file list retrieval sequence
