# File Operations Refactor Checklist

## Overview
Refactoring file operations to match specs/main.md (lines 195-324, 343-346)
- Replace simple map storage with HAMTDirectory
- Add peer-to-peer file list protocol on reserved "p2p-webapp" protocol
- Implement structured entries with type, CID, MIME type
- Change to async server messages (peerFiles, gotFile)

---

## Phase 1: Core Data Structures

### internal/peer/manager.go
- [ ] Add imports for HAMTDirectory (boxo/ipld/unixfs/io)
- [ ] Update Peer struct:
  - [ ] Add `directory *uio.HAMTDirectory` field
  - [ ] Add `directoryCID cid.Cid` field
  - [ ] Remove old file tracking if incompatible
- [ ] Update Manager struct:
  - [ ] Change fileListHandlers to store peerID -> callback mapping
  - [ ] Add onPeerFiles callback: `func(receiverPeerID, targetPeerID, dirCID string, entries map[string]FileEntry)`
  - [ ] Add onGotFile callback: `func(receiverPeerID string, cid string, success bool, content any)`

### Define FileEntry struct
- [ ] Create FileEntry struct with:
  - [ ] Type string ("file" or "directory")
  - [ ] CID string
  - [ ] MimeType string (for files only)

---

## Phase 2: Peer Creation & Restoration

### internal/peer/manager.go - CreatePeer()
- [ ] Add rootDirectory parameter (optional CID string)
- [ ] If rootDirectory provided:
  - [ ] Parse CID
  - [ ] Load existing HAMTDirectory from IPFS
  - [ ] Assign to peer.directory
  - [ ] Set peer.directoryCID
- [ ] If no rootDirectory:
  - [ ] Create new empty HAMTDirectory
  - [ ] Get initial node and CID
  - [ ] Pin the directory CID
  - [ ] Assign to peer
- [ ] Update Manager.peerFiles initialization (or remove if replaced)

---

## Phase 3: libp2p Protocol Handlers

### internal/peer/manager.go - Protocol Messages
- [ ] Define P2PWebAppProtocol constant: `const P2PWebAppProtocol = "/p2p-webapp/1.0.0"`
- [ ] Create message types:
  - [ ] GetFileListMessage struct (empty or with request fields)
  - [ ] FileListMessage struct with CID and directory data

### Register Protocol Handler
- [ ] In NewPeer or similar:
  - [ ] Register handler: `host.SetStreamHandler(P2PWebAppProtocol, handleP2PWebAppProtocol)`
- [ ] Implement handleP2PWebAppProtocol(stream):
  - [ ] Read message type from stream
  - [ ] Route to handleGetFileList or handleFileList

### handleGetFileList (receiver)
- [ ] Read GetFileListMessage from stream
- [ ] Get peer's directory and CID
- [ ] Build entries map by walking HAMTDirectory tree:
  - [ ] Use ForEachLink or EnumLinksAsync
  - [ ] For each link, determine type (file/directory)
  - [ ] Get MIME type for files
  - [ ] Build pathname (handle nested structure)
- [ ] Send FileListMessage back to requester
- [ ] Close stream

### handleFileList (requester)
- [ ] Read FileListMessage from stream
- [ ] Extract sender peerID from stream
- [ ] Look up fileListHandlers for that peerID
- [ ] Call onPeerFiles callback with (receiverPeerID, senderPeerID, CID, entries)
- [ ] Remove handler (one-time use)
- [ ] Close stream

---

## Phase 4: File Operations Implementation

### ListFiles(peerID)
- [ ] If peerID is local peer:
  - [ ] Build entries from own directory
  - [ ] Spawn goroutine to call onPeerFiles immediately
  - [ ] Return nil
- [ ] If peerID is remote:
  - [ ] Check if handler already exists (deduplicate)
  - [ ] If yes, return nil (already pending)
  - [ ] Register fileListHandler
  - [ ] Open stream to remote peer with P2PWebAppProtocol
  - [ ] Send GetFileListMessage
  - [ ] Handle errors: remove handler, return error
  - [ ] On success, spawn goroutine and return nil

### GetFile(cidStr)
- [ ] Parse CID
- [ ] Get node from IPFS: `ipfsPeer.GetNode(ctx, cid)`
- [ ] Determine type:
  - [ ] Check unixfs type (File, Directory, HAMTShard)
- [ ] For files:
  - [ ] Read content
  - [ ] Detect MIME type
  - [ ] Spawn goroutine to call onGotFile with {type: "file", mimeType, content}
- [ ] For directories:
  - [ ] Build entries map
  - [ ] Spawn goroutine to call onGotFile with {type: "directory", entries}
- [ ] Return nil (async)

### StoreFile(peerID, path, content, directory)
- [ ] Get peer's directory
- [ ] If directory is true:
  - [ ] Create new HAMTDirectory node
  - [ ] Error if content is not null
- [ ] If directory is false:
  - [ ] Error if content is null
  - [ ] Create UnixFS file node with content
- [ ] Parse path to find parent directory and name
- [ ] Navigate to parent in HAMTDirectory (create if needed)
- [ ] AddChild with name and new node
- [ ] Update root directory CID
- [ ] Pin new CID
- [ ] Unpin old CID
- [ ] Return nil

### RemoveFile(peerID, path)
- [ ] Get peer's directory
- [ ] Parse path to find parent and name
- [ ] Navigate to parent directory
- [ ] RemoveChild by name
- [ ] Update root directory CID
- [ ] Pin new CID
- [ ] Unpin old CID
- [ ] Return nil

---

## Phase 5: Protocol Handler Integration

### internal/protocol/handler.go - Message Types
- [ ] Update messages.go:
  - [ ] Add PeerFilesRequest struct (peerID, CID, entries map[string]FileEntry)
  - [ ] Add GotFileRequest struct (cid string, success bool, content any)
  - [ ] Update StoreFileRequest: add `directory bool` field
  - [ ] Update ListFilesRequest: just `peerid string`
  - [ ] Remove old file operation response types

### Update PeerManager Interface
- [ ] Add SetPeerFilesCallback(func(...))
- [ ] Add SetGotFileCallback(func(...))
- [ ] Update CreatePeer signature: add rootDirectory parameter
- [ ] Update StoreFile signature: `(peerID, path, content string, directory bool)`
- [ ] Update ListFiles signature: `(peerID string)` - returns nil, uses callback
- [ ] Update GetFile signature: `(cid string)` - returns nil, uses callback

### Handler Methods
- [ ] handleListFiles:
  - [ ] Parse ListFilesRequest (just peerid)
  - [ ] Call peerManager.ListFiles(peerid)
  - [ ] Return emptyResponse (actual response via peerFiles server message)
- [ ] handleGetFile:
  - [ ] Parse GetFileRequest (just cid)
  - [ ] Call peerManager.GetFile(cid)
  - [ ] Return emptyResponse (actual response via gotFile server message)
- [ ] handleStoreFile:
  - [ ] Parse StoreFileRequest (path, content, directory)
  - [ ] Decode base64 content if present
  - [ ] Call peerManager.StoreFile(peerID, path, content, directory)
  - [ ] Return emptyResponse
- [ ] handleRemoveFile:
  - [ ] No changes needed

### Server Message Creators
- [ ] CreatePeerFilesMessage(peerID, cid string, entries):
  - [ ] Build PeerFilesRequest
  - [ ] Return Message with method "peerFiles"
- [ ] CreateGotFileMessage(cid string, success bool, content):
  - [ ] Build GotFileRequest
  - [ ] Return Message with method "gotFile"

---

## Phase 6: WebSocket Integration

### internal/server/websocket.go
- [ ] In NewWSConnection or connection setup:
  - [ ] Register peerFiles callback: manager.SetPeerFilesCallback(...)
  - [ ] Register gotFile callback: manager.SetGotFileCallback(...)
- [ ] Implement peerFiles callback:
  - [ ] Create message via handler.CreatePeerFilesMessage
  - [ ] Send to client via WebSocket
- [ ] Implement gotFile callback:
  - [ ] Create message via handler.CreateGotFileMessage
  - [ ] Send to client via WebSocket

---

## Phase 7: Client Library Updates

### pkg/client/src/types.ts
- [ ] Add FileEntry interface:
  - [ ] type: "file" | "directory"
  - [ ] cid: string
  - [ ] mimeType?: string
- [ ] Update ListFilesRequest: `peerid: string`
- [ ] Remove ListFilesResponse (uses server message)
- [ ] Add PeerFilesRequest:
  - [ ] peerid: string
  - [ ] cid: string
  - [ ] entries: { [path: string]: FileEntry }
- [ ] Update GetFileRequest: `cid: string`
- [ ] Remove GetFileResponse (uses server message)
- [ ] Add GotFileRequest:
  - [ ] cid: string
  - [ ] success: boolean
  - [ ] content: any
- [ ] Update StoreFileRequest: add `directory: boolean`

### pkg/client/src/client.ts - File List Handlers
- [ ] Add field: `fileListHandlers: Map<string, (cid: string, entries: {[path: string]: FileEntry}) => void>`
- [ ] Update listFiles(peerid):
  - [ ] Create promise
  - [ ] Check if handler exists for peerid
  - [ ] If no handler, add to map and send request
  - [ ] If handler exists, just add promise to queue
  - [ ] Return promise
- [ ] Add routePeerFiles handler in handleServerRequest:
  - [ ] Extract peerid, cid, entries
  - [ ] Look up handler
  - [ ] Resolve all pending promises
  - [ ] Remove handler

### pkg/client/src/client.ts - Get File Handlers
- [ ] Add field: `getFileHandlers: Map<string, (success: boolean, content: any) => void>`
- [ ] Update getFile(cid):
  - [ ] Create promise
  - [ ] Add handler to map
  - [ ] Send request
  - [ ] Return promise
- [ ] Add routeGotFile handler in handleServerRequest:
  - [ ] Extract cid, success, content
  - [ ] Look up handler
  - [ ] Resolve/reject promise
  - [ ] Remove handler

### pkg/client/src/client.ts - Store/Remove
- [ ] Update storeFile signature: `(path: string, content: Uint8Array | null, directory: boolean)`
  - [ ] Handle null content for directories
  - [ ] Encode content as base64 if present
  - [ ] Include directory flag
- [ ] Update removeFile: no changes needed

---

## Phase 8: Main.go Updates

### cmd/p2p-webapp/main.go
- [ ] Update CreatePeer calls in protocol handler:
  - [ ] Parse rootDirectory from PeerRequest if present
  - [ ] Pass to CreatePeer

### internal/protocol/handler.go - handlePeer
- [ ] Update PeerRequest struct to include rootDirectory (optional)
- [ ] Extract rootDirectory from request
- [ ] Pass to peerManager.CreatePeer(requestedPeerKey, rootDirectory)

---

## Phase 9: Testing & Verification

### Build
- [ ] Run `make build`
- [ ] Fix any compilation errors

### Manual Testing
- [ ] Test peer creation with no rootDirectory
- [ ] Test peer creation with rootDirectory CID
- [ ] Test storeFile creating a file
- [ ] Test storeFile creating a directory
- [ ] Test listFiles on local peer
- [ ] Test listFiles on remote peer
- [ ] Test getFile for a file
- [ ] Test getFile for a directory
- [ ] Test removeFile

### Update Tests
- [ ] Create test file for file operations
- [ ] Test HAMTDirectory creation
- [ ] Test file storage and retrieval
- [ ] Test directory creation
- [ ] Test peer-to-peer file list protocol
- [ ] Test getFile with different types

---

## Phase 10: Documentation

### Update traceability.md
- [ ] Mark file operations as redesigned
- [ ] Update implementation status
- [ ] Reference new sequence diagrams

### Code Comments
- [ ] Add traceability comments to new functions
- [ ] Reference seq-list-files.md and seq-store-file.md
- [ ] Document HAMTDirectory usage

---

## Progress Tracking
- Phase 1: ⬜ Not Started
- Phase 2: ⬜ Not Started
- Phase 3: ⬜ Not Started
- Phase 4: ⬜ Not Started
- Phase 5: ⬜ Not Started
- Phase 6: ⬜ Not Started
- Phase 7: ⬜ Not Started
- Phase 8: ⬜ Not Started
- Phase 9: ⬜ Not Started
- Phase 10: ⬜ Not Started
