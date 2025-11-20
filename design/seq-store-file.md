# Sequence: Store File Operation

**Source Spec:** main.md - storeFile(path, content) and createDirectory(path)

**Participants:**
- Client A: P2PWebAppClient instance storing file or directory
- WebSocketHandler: Enforces file ownership security, routes to Peer via PeerManager
- PeerManager: Returns Peer instance by peerID
- Peer: Manages HAMTDirectory and CID tracking
- HAMTDirectory: IPFS/boxo data structure for peer file storage

**Flow:**

This sequence demonstrates the file storage operation with ownership enforcement. The client calls either `storeFile(path, content)` for files or `createDirectory(path)` for directories. Both methods internally use the underlying storefile protocol message with appropriate parameters (directory=false for files, directory=true for directories). The operation implicitly operates on the peer associated with the WebSocket connection. The WebSocketHandler enforces this security model by extracting the peerID from the connection context, gets the Peer from PeerManager, then calls storeFile() on the Peer. The Peer creates a file or directory node in IPFS, updates its HAMTDirectory at the specified path, and updates its directory CID. The peer also pins its directory for persistence.

```
     ┌────────┐                             ┌────────────────┐                              ┌───────────┐                      ┌────┐                ┌─────────────┐
     │Client A│                             │WebSocketHandler│                              │PeerManager│                      │Peer│                │HAMTDirectory│
     └────┬───┘                             └────────┬───────┘                              └─────┬─────┘                      └─┬──┘                └──────┬──────┘
          │                                          │                                            │                              │                          │
          │                                          │                          ╔═════════════════╧════╗                         │                          │
══════════╪══════════════════════════════════════════╪══════════════════════════╣ Store File Operation ╠═════════════════════════╪══════════════════════════╪═════════════════════
          │                                          │                          ╚═════════════════╤════╝                         │                          │
          │                                          │                                            │                              │                          │
          │────┐                                     │ ╔═════════════════════════════════════╗    │                              │                          │
          │    │ storeFile(path, content, directory) │ ║Implicit peerID (connection's peer) ░║    │                              │                          │
          │<───┘                                     │ ╚═════════════════════════════════════╝    │                              │                          │
          │                                          │                                            │                              │                          │
          │   storeFile(path, content, directory)    │                                            │                              │                          │
          │─────────────────────────────────────────>│                                            │                              │                          │
          │                                          │                                            │                              │                          │
          │                                          │────┐                          ╔════════════╧════════════════════════╗     │                          │
          │                                          │    │ enforceFileOwnership()   ║Get peerID from connection          ░║     │                          │
          │                                          │<───┘                          ║Ensure request operates on own peer  ║     │                          │
          │                                          │                               ╚════════════╤════════════════════════╝     │                          │
          │                                          │          getPeer(peerID)                   │                              │                          │
          │                                          │───────────────────────────────────────────>│                              │                          │
          │                                          │                                            │                              │                          │
          │                                          │             peer instance                  │                              │                          │
          │                                          │<───────────────────────────────────────────│                              │                          │
          │                                          │                                            │                              │                          │
          │                                          │             storeFile(path, content, directory)                           │                          │
          │                                          │──────────────────────────────────────────────────────────────────────────>│                          │
          │                                          │                                            │                              │                          │
          │                                          │                                            │                              │────┐                     │
          │                                          │                                            │                              │    │ Get own HAMTDirectory│
          │                                          │                                            │                              │<───┘                     │
          │                                          │                                            │                              │                          │
          │                                          │                                            │                              │                          │
          │                                          │                                            │                              │                          │
          │                                          │                            ╔══════╤════════╪══════════════════════════════╪══════════════════════════╪══════════════════════════╗
          │                                          │                            ║ ALT  │  content is not null (file)           │                          │                          ║
          │                                          │                            ╟──────┘        │                              │                          │                          ║
          │                                          │                            ║               │                              │ Add file entry at path   │ ╔═════════════════════╗ ║
          │                                          │                            ║               │                              │─────────────────────────>│ ║Store file content  ░║ ║
          │                                          │                            ║               │                              │                          │ ║Set directory=false  ║ ║
          │                                          │                            ╠═══════════════╪══════════════════════════════╪══════════════════════════╪══════════════════════════╣
          │                                          │                            ║ [content is null (directory)]                │                          │                          ║
          │                                          │                            ║               │                              │ Add directory entry      │ ╔════════════════════╗  ║
          │                                          │                            ║               │                              │─────────────────────────>│ ║No content stored  ░║  ║
          │                                          │                            ║               │                              │                          │ ║Set directory=true  ║  ║
          │                                          │                            ╚═══════════════╪══════════════════════════════╪══════════════════════════╪═╚════════════════════╝══╝
          │                                          │                                            │                              │                          │
          │                                          │                                            │                              │────┐                     │
          │                                          │                                            │                              │    │ Calculate new CID   │
          │                                          │                                            │                              │<───┘                     │
          │                                          │                                            │                              │                          │
          │                                          │                                            │                              │────┐                     │
          │                                          │                                            │                              │    │ Update directoryCID│
          │                                          │                                            │                              │<───┘                     │
          │                                          │                                            │                              │                          │
          │                                          │                        Return CID of stored file/directory node           │                          │
          │                                          │<──────────────────────────────────────────────────────────────────────────│                          │
          │                                          │                                            │                              │                          │
          │  storeFile response (CID of new node)    │                                            │                              │                          │
          │<─────────────────────────────────────────│                                            │                              │                          │
          │                                          │                                            │                              │                          │
          │────┐                                  ╔══╧══════════════════════════════╗             │                              │                          │
          │    │ Store returned CID for sharing   ║Client can use CID for sharing  ░║             │                              │                          │
          │<───┘                                  ╚══╤══════════════════════════════╝             │                              │                          │
     ┌────┴───┐                             ┌────────┴───────┐                              ┌─────┴─────┐                      ┌──┴─┐                ┌──────┴──────┐
     │Client A│                             │WebSocketHandler│                              │PeerManager│                      │Peer│                │HAMTDirectory│
     └────────┘                             └────────────────┘                              └───────────┘                      └────┘                └─────────────┘
```

**Key Points:**

1. **Dual API**:
   - `storeFile(path: string, content: string | Uint8Array)`: Creates file nodes
     - `path`: Full pathname in HAMTDirectory tree (e.g., "docs/readme.txt")
     - `content`: File content as string (UTF-8 encoded) or Uint8Array (binary data)
   - `createDirectory(path: string)`: Creates directory nodes
     - `path`: Full pathname in HAMTDirectory tree (e.g., "docs")
   - Both methods use the underlying storefile protocol with appropriate directory flag
   - Both methods trigger file update notifications if configured (see point 4 below)

2. **Implicit Peer Context**: The storeFile() operation has no peerID parameter. The client can only modify its own peer's directory.

3. **Ownership Enforcement**: WebSocketHandler's enforceFileOwnership() extracts the peerID from the connection context and ensures the request operates only on that peer.

4. **File Availability Notifications** (not shown in diagram): After successfully storing the file and updating the directory CID, the Peer calls `publishFileUpdateNotification()` which:
   - Checks if `fileUpdateNotifyTopic` is configured (from P2PConfig)
   - Checks if peer is subscribed to that topic
   - If both true, publishes message: `{"type":"p2p-webapp-file-update","peer":"<peerID>"}`
   - This allows applications to notify other peers of file changes for automatic UI refresh
   - Privacy-friendly: only publishes when peer is subscribed (opt-in behavior)

4. **Handler → PeerManager → Peer**: WebSocketHandler calls PeerManager.GetPeer(peerID) to get the Peer, then calls peer.StoreFile() directly on the Peer.

5. **IPFS Node Creation**: Peer creates a file or directory node in IPFS and stores it via ipfs-lite, which returns the new node with CID.

6. **Path-based Update**: Uses the path to find the correct subdirectory in the Peer's HAMTDirectory and adds the new node there.

7. **CID Management**: After modifying the HAMTDirectory, Peer updates its own directoryCID.

8. **Pinning**: The peer pins its directory for persistence across sessions.

9. **Security Model**: The implicit peerID design prevents clients from modifying other peers' directories. Only listFiles() has an explicit peerID parameter because it's a read-only query operation.

**Related:**
- crc-P2PWebAppClient.md: Client-side file storage
- crc-PeerManager.md: Peer lifecycle management, provides GetPeer()
- crc-Peer.md: Peer operations including storeFile()
- crc-WebSocketHandler.md: Ownership enforcement
- seq-list-files.md: Related file list retrieval sequence
