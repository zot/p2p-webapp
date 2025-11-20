# p2p-webapp API Reference

**Complete API documentation for p2p-webapp WebSocket protocol and TypeScript client library**

---

## Table of Contents

1. [TypeScript Client Library](#typescript-client-library)
2. [WebSocket Protocol](#websocket-protocol)
3. [Message Types](#message-types)
4. [File Operations](#file-operations)
5. [Error Handling](#error-handling)
6. [Examples](#examples)

---

## TypeScript Client Library

### Installation

The client library is bundled with p2p-webapp. Extract it with:

```bash
./p2p-webapp cp client.* my-project/
```

This copies:
- `client.js` - ES module implementation
- `client.d.ts` - TypeScript type definitions

### Import

```typescript
import { connect, start, stop, send, subscribe, publish, unsubscribe, listPeers, listFiles, getFile, storeFile, createDirectory, removeFile } from './client.js';
```

Or with named export:

```typescript
import * as P2P from './client.js';
```

---

### Core API

#### `connect(peerKey?: string, rootDirectory?: string): Promise<[string, string]>`

Connect to p2p-webapp server and initialize peer.

**Parameters**:
- `peerKey` (optional) - Existing peer key to reuse identity. If omitted, generates new key.
- `rootDirectory` (optional) - CID of peer's root directory to restore file state. If omitted, starts with empty directory.

**Returns**: Promise resolving to `[peerID, peerKey]`
- `peerID` - Unique identifier for this peer
- `peerKey` - Private key (save this to maintain identity across sessions)

**Throws**: Error if connection fails or peer already initialized

**Example**:
```typescript
// New peer
const [peerID, peerKey] = await connect();
localStorage.setItem('peerKey', peerKey);

// Reuse identity
const savedKey = localStorage.getItem('peerKey');
const [peerID, peerKey] = await connect(savedKey);
```

**Notes**:
- Must be called first before any other operations
- Establishes WebSocket connection and sends Peer() command
- Only call once per page load
- Duplicate peerID error means same peer key used in another tab

---

#### `start(protocol: string, onData: ProtocolDataCallback): Promise<void>`

Register listener for protocol-based messages.

**Parameters**:
- `protocol` - Protocol identifier (e.g., "chat", "game", "sync")
- `onData` - Callback receiving `(peer: string, data: any) => void | Promise<void>`

**Returns**: Promise resolving when protocol started

**Throws**: Error if protocol already started

**Example**:
```typescript
await start('chat', (peer, data) => {
  console.log(`Message from ${peer}:`, data);
  // Handle message
});
```

**Notes**:
- Must call before sending on protocol
- Only one listener per protocol
- Callback can be sync or async
- Messages processed sequentially (ordering guaranteed)
- Listener automatically removed on disconnect or `stop()`

---

#### `send(peer: string, protocol: string, data: any): Promise<void>`

Send data to peer on protocol.

**Parameters**:
- `peer` - Target peer ID
- `protocol` - Protocol identifier
- `data` - Any JSON-serializable data

**Returns**: Promise resolving when delivery confirmed (ack received)

**Throws**: Error if protocol not started

**Example**:
```typescript
// Send and wait for delivery confirmation
await send(peerID, 'chat', { text: 'Hello!' });
console.log('Message delivered');

// Or handle delivery in try/catch
try {
  await send(peerID, 'chat', { text: 'Important!' });
  showDeliveryCheckmark();
} catch (error) {
  console.error('Delivery failed:', error);
}
```

**Notes**:
- Protocol must be started first with `start()`
- Server manages connection lifecycle transparently
- Promise resolves when peer acknowledges receipt
- Ack numbers managed internally (transparent to caller)
- Server creates stream on-demand, reuses for subsequent messages

---

#### `stop(protocol: string): Promise<void>`

Stop protocol and remove message listener.

**Parameters**:
- `protocol` - Protocol identifier to stop

**Returns**: Promise resolving when protocol stopped

**Example**:
```typescript
await stop('chat');
```

**Notes**:
- Removes message listener
- Prevents sending on protocol
- Server keeps streams open (may be reused if restarted)

---

#### `subscribe(topic: string, onData: TopicDataCallback, onPeerChange?: PeerChangeCallback): Promise<void>`

Subscribe to topic for pub/sub messaging.

**Parameters**:
- `topic` - Topic identifier (e.g., "chatroom", "game-lobby")
- `onData` - Callback receiving `(peer: string, data: any) => void | Promise<void>`
- `onPeerChange` (optional) - Callback receiving `(peer: string, joined: boolean) => void | Promise<void>`

**Returns**: Promise resolving when subscribed

**Example**:
```typescript
await subscribe('chatroom',
  // Message callback
  (peer, data) => {
    console.log(`${peer}: ${data.text}`);
    addMessageToUI(peer, data.text);
  },
  // Peer change callback
  (peer, joined) => {
    if (joined) {
      console.log(`${peer} joined`);
      updatePeerList();
    } else {
      console.log(`${peer} left`);
      updatePeerList();
    }
  }
);
```

**Notes**:
- Automatic peer discovery via GossipSub + DHT
- Peer join/leave events monitored automatically
- Messages from all topic subscribers received
- Callbacks can be sync or async
- Messages processed sequentially
- **File Update Notifications**: Topics can receive special file availability notifications (see below)

**File Update Notifications**:

When `fileUpdateNotifyTopic` is configured in `p2p-webapp.toml`, the server automatically publishes file availability notifications to that topic after `storeFile()` / `removeFile()` operations (only if the peer is subscribed).

Applications can detect and handle these notifications:

```typescript
await subscribe('chatroom', (peer, data) => {
  // Check for file update notification
  if (data.type === 'p2p-webapp-file-update' && data.peer) {
    console.log(`Peer ${data.peer} updated their files`);

    // Refresh file list if viewing this peer's files
    if (currentViewPeer === data.peer) {
      await listFiles(data.peer);
    }
    return; // Don't process as regular message
  }

  // Handle regular messages
  console.log(`${peer}: ${data.text}`);
});
```

**Notification Message Format**:
```typescript
{
  type: "p2p-webapp-file-update",
  peer: "<peerID>"  // Peer whose files changed
}
```

**Configuration** (in `p2p-webapp.toml`):
```toml
[p2p]
fileUpdateNotifyTopic = "chatroom"  # Topic for notifications (empty = disabled)
```

**Privacy Design**:
- Notifications only published if topic is configured AND peer is subscribed
- Opt-in mechanism prevents unintended broadcasts
- Applications control whether to handle notifications

---

#### `publish(topic: string, data: any): Promise<void>`

Publish message to topic.

**Parameters**:
- `topic` - Topic identifier
- `data` - Any JSON-serializable data

**Returns**: Promise resolving when published (NOT when received by peers)

**Example**:
```typescript
await publish('chatroom', { text: 'Hello everyone!' });
```

**Notes**:
- Must subscribe to topic first
- Broadcast to all subscribed peers
- Uses GossipSub for efficient propagation
- No delivery confirmation (pub/sub is best-effort)

---

#### `unsubscribe(topic: string): Promise<void>`

Unsubscribe from topic.

**Parameters**:
- `topic` - Topic identifier

**Returns**: Promise resolving when unsubscribed

**Example**:
```typescript
await unsubscribe('chatroom');
```

**Notes**:
- Removes message and peer change listeners
- Stops receiving topic messages
- Peer leave event sent to other subscribers

---

#### `listPeers(topic: string): Promise<string[]>`

Get list of peers subscribed to topic.

**Parameters**:
- `topic` - Topic identifier

**Returns**: Promise resolving to array of peer IDs

**Example**:
```typescript
const peers = await listPeers('chatroom');
console.log(`${peers.length} peers in room:`, peers);
```

**Notes**:
- Returns current snapshot (may change immediately)
- Use `onPeerChange` callback for live updates
- Includes self in peer list

---

### File Operations API

Each peer maintains a HAMTDirectory (Hash Array Mapped Trie Directory) structure in IPFS for organizing files. The directory is identified by a CID (Content Identifier) and can be restored across sessions using the `rootDirectory` parameter in `connect()`.

#### `listFiles(peerID: string): Promise<{rootCID: string, entries: FileEntries}>`

Request list of files from a peer's directory.

**Parameters**:
- `peerID` - Peer ID to list files from (can be self or another peer)

**Returns**: Promise resolving with object containing:
- `rootCID` - CID of the peer's root directory
- `entries` - Object mapping pathnames to file/directory entries

**Entry Format**:
```typescript
{
  "path/to/file.txt": {
    type: "file",
    cid: "QmXg9Pp2ytZ...",
    mimeType: "text/plain"
  },
  "path/to/directory": {
    type: "directory",
    cid: "QmYwAPJzv5C..."
  }
}
```

**Example**:
```typescript
const { rootCID, entries } = await listFiles(peerID);
console.log(`Root directory CID: ${rootCID}`);
console.log(`Files from ${peerID}:`);

for (const [path, entry] of Object.entries(entries)) {
  if (entry.type === 'file') {
    console.log(`  📄 ${path} (${entry.mimeType})`);
  } else {
    console.log(`  📁 ${path}/`);
  }
}
```

**Notes**:
- For local peer, response is immediate
- For remote peer, sends request via reserved `p2p-webapp` protocol
- Multiple concurrent requests to same peer are deduplicated
- Internally uses `peerFiles` server push message to resolve promise

---

#### `getFile(cid: string): Promise<FileContent>`

Retrieve IPFS content by CID.

**Parameters**:
- `cid` - Content identifier to retrieve

**Returns**: Promise resolving with file content

**Content Format**:
```typescript
// File
{
  type: "file",
  mimeType: "text/plain",
  content: "base64-encoded-string"
}

// Directory
{
  type: "directory",
  entries: {
    "filename.txt": "QmCID1...",
    "subdir": "QmCID2..."
  }
}
```

**Example**:
```typescript
try {
  const content = await getFile(cid);

  if (content.type === 'file') {
    const decoded = atob(content.content);
    displayFile(decoded, content.mimeType);
  } else if (content.type === 'directory') {
    console.log('Directory entries:', content.entries);
  }
} catch (error) {
  console.error('Failed to retrieve:', error);
}
```

**Notes**:
- **Why base64 encoding?** Binary files (images, PDFs, executables, etc.) contain arbitrary bytes that aren't valid UTF-8. Since JSON can only safely encode UTF-8 strings, base64 encoding is required to transmit binary data without corruption. Do NOT remove this encoding - it's essential for binary file support.
- Promise rejects on retrieval failure
- Can retrieve content from any peer's files via their CIDs
- Internally uses `gotFile` server push message to resolve promise

---

#### `storeFile(path: string, content: string | Uint8Array): Promise<string>`

Store file in peer's IPFS directory.

**Parameters**:
- `path` - Unix-style path relative to root (e.g., "docs/readme.txt")
- `content` - File content as string or Uint8Array

**Returns**: Promise resolving to CID of the stored file node

**Example**:
```typescript
// Store text file with string
const textCid = await storeFile('readme.txt', 'Hello, world!');
console.log('Text file CID:', textCid);

// Store binary file with Uint8Array
const binaryContent = new Uint8Array([0x89, 0x50, 0x4E, 0x47]);
const binaryCid = await storeFile('image.png', binaryContent);
console.log('Binary file CID:', binaryCid);
```

**Notes**:
- String content is UTF-8 encoded
- Content is automatically base64-encoded before transmission
- Automatically creates parent directories if needed
- Updates peer's root directory CID after store
- **File Update Notifications**: If configured, automatically publishes notification to subscribers after successful storage

**Automatic Notifications**:

When `fileUpdateNotifyTopic` is configured in `p2p-webapp.toml` and the peer is subscribed to that topic, the server automatically publishes a file availability notification after successful `storeFile()` operations. Subscribers can detect and handle these notifications to refresh file lists.

See [`subscribe()` File Update Notifications](#file-update-notifications) for handling example.

---

#### `createDirectory(path: string): Promise<string>`

Create directory in peer's IPFS directory.

**Parameters**:
- `path` - Unix-style path relative to root (e.g., "docs")

**Returns**: Promise resolving to CID of the stored directory node

**Example**:
```typescript
const dirCid = await createDirectory('docs');
console.log('Directory CID:', dirCid);
```

**Notes**:
- Automatically creates parent directories if needed
- Updates peer's root directory CID after creation
- Returns the CID of the stored node, which can be used to share or retrieve the file directly
- The root directory CID can be obtained via `listFiles(client.peerID)`
- **File Update Notifications**: If configured, automatically publishes notification to subscribers after successful directory creation

**Automatic Notifications**:

When `fileUpdateNotifyTopic` is configured in `p2p-webapp.toml` and the peer is subscribed to that topic, the server automatically publishes a file availability notification after successful `createDirectory()` operations. Subscribers can detect and handle these notifications to refresh file lists.

See [`subscribe()` File Update Notifications](#file-update-notifications) for handling example.

---

#### `removeFile(path: string): Promise<void>`

Remove file or directory from peer's directory.

**Parameters**:
- `path` - Unix-style path to remove

**Returns**: Promise resolving when removed

**Example**:
```typescript
await removeFile('docs/readme.txt');
```

**Notes**:
- Removes entry from parent directory
- Updates peer's root directory CID
- Removing a directory removes all contents
- **File Update Notifications**: If configured, automatically publishes notification to subscribers after successful removal

**Automatic Notifications**:

When `fileUpdateNotifyTopic` is configured in `p2p-webapp.toml` and the peer is subscribed to that topic, the server automatically publishes a file availability notification after successful `removeFile()` operations. Subscribers can detect and handle these notifications to refresh file lists.

See [`subscribe()` File Update Notifications](#file-update-notifications) for handling example.

---

### Type Definitions

```typescript
type ProtocolDataCallback = (peer: string, data: any) => void | Promise<void>;
type TopicDataCallback = (peer: string, data: any) => void | Promise<void>;
type PeerChangeCallback = (peer: string, joined: boolean) => void | Promise<void>;

interface FileEntries {
  [pathname: string]: FileEntry | DirectoryEntry;
}

interface FileEntry {
  type: 'file';
  cid: string;
  mimeType: string;
}

interface DirectoryEntry {
  type: 'directory';
  cid: string;
}

type FileContent = FileContentFile | FileContentDirectory;

interface FileContentFile {
  type: 'file';
  mimeType: string;
  content: string; // base64-encoded
}

interface FileContentDirectory {
  type: 'directory';
  entries: {
    [pathname: string]: string; // CID
  };
}
```

---

## WebSocket Protocol

### Connection

**URL**: `ws://localhost:<port>/` (port auto-selected or specified with `-p`)

**Handshake**: Standard WebSocket upgrade

**Message Format**: JSON-RPC-like structure

---

### Message Structure

All messages are JSON objects with these fields:

#### Client Request

```typescript
{
  "requestID": number,     // Auto-incrementing from 0
  "command": string,       // Command name
  "args": any[]            // Command arguments (positional)
}
```

#### Server Response

```typescript
{
  "requestID": number,     // Matches request
  "result": any            // Success result
}
```

Or error:

```typescript
{
  "requestID": number,     // Matches request
  "error": string          // Error message
}
```

#### Server Push

```typescript
{
  "requestID": number,     // Server's request ID
  "command": string,       // Command name
  "args": any[]            // Command arguments
}
```

---

## Message Types

### Client Request Messages

#### Peer

**Command**: `"peer"`

**Args**: `[peerKey?]`
- `peerKey` (string, optional) - Existing peer key or omit for new key

**Response**: `[peerID, peerKey]`
- `peerID` (string) - Unique peer identifier
- `peerKey` (string) - Private key for this peer

**Error**: `"duplicate peer"` if peerID already registered

**Example**:
```json
// Request
{
  "requestID": 0,
  "command": "peer",
  "args": []
}

// Response
{
  "requestID": 0,
  "result": ["12D3KooW...", "CAA..."]
}
```

**Constraints**:
- Must be first command after WebSocket connect
- Cannot be sent more than once per connection

---

#### start

**Command**: `"start"`

**Args**: `[protocol]`
- `protocol` (string) - Protocol identifier

**Response**: `null`

**Error**: Error message if protocol already started

**Example**:
```json
{
  "requestID": 1,
  "command": "start",
  "args": ["chat"]
}
```

---

#### stop

**Command**: `"stop"`

**Args**: `[protocol]`
- `protocol` (string) - Protocol identifier

**Response**: `null`

**Example**:
```json
{
  "requestID": 2,
  "command": "stop",
  "args": ["chat"]
}
```

---

#### send

**Command**: `"send"`

**Args**: `[peer, protocol, data, ack?]`
- `peer` (string) - Target peer ID
- `protocol` (string) - Protocol identifier
- `data` (any) - JSON-serializable data
- `ack` (number, optional) - Ack number for delivery confirmation (-1 or omit for no ack)

**Response**: `null`

**Error**: Error message if protocol not started or peer unreachable

**Example**:
```json
{
  "requestID": 3,
  "command": "send",
  "args": ["12D3KooW...", "chat", {"text": "Hello"}, 0]
}
```

**Notes**:
- If `ack >= 0`, server will send `ack` command when delivered
- Client library manages ack numbers automatically

---

#### subscribe

**Command**: `"subscribe"`

**Args**: `[topic]`
- `topic` (string) - Topic identifier

**Response**: `null`

**Error**: Error message if subscription fails

**Example**:
```json
{
  "requestID": 4,
  "command": "subscribe",
  "args": ["chatroom"]
}
```

**Notes**:
- Automatically enables peer join/leave monitoring
- Will receive `peerChange` notifications

---

#### publish

**Command**: `"publish"`

**Args**: `[topic, data]`
- `topic` (string) - Topic identifier
- `data` (any) - JSON-serializable data

**Response**: `null`

**Error**: Error message if not subscribed to topic

**Example**:
```json
{
  "requestID": 5,
  "command": "publish",
  "args": ["chatroom", {"text": "Hello everyone"}]
}
```

---

#### unsubscribe

**Command**: `"unsubscribe"`

**Args**: `[topic]`
- `topic` (string) - Topic identifier

**Response**: `null`

**Example**:
```json
{
  "requestID": 6,
  "command": "unsubscribe",
  "args": ["chatroom"]
}
```

---

#### listPeers

**Command**: `"listPeers"`

**Args**: `[topic]`
- `topic` (string) - Topic identifier

**Response**: `string[]` - Array of peer IDs

**Example**:
```json
// Request
{
  "requestID": 7,
  "command": "listPeers",
  "args": ["chatroom"]
}

// Response
{
  "requestID": 7,
  "result": ["12D3KooW...", "12D3KooX...", "12D3KooY..."]
}
```

---

#### listFiles

**Command**: `"listFiles"`

**Args**: `[peerID]`
- `peerID` (string) - Peer ID to list files from

**Response**: `null`

**Example**:
```json
{
  "requestID": 8,
  "command": "listFiles",
  "args": ["12D3KooW..."]
}
```

**Notes**:
- Triggers `peerFiles` server push message with results
- For remote peers, sends request via reserved `p2p-webapp` protocol
- Multiple concurrent requests to same peer are deduplicated

---

#### getFile

**Command**: `"getFile"`

**Args**: `[cid]`
- `cid` (string) - Content identifier to retrieve

**Response**: `null`

**Example**:
```json
{
  "requestID": 9,
  "command": "getFile",
  "args": ["QmXg9Pp2ytZ..."]
}
```

**Notes**:
- Triggers `gotFile` server push message with content
- Can retrieve any content by CID from IPFS network

---

#### storeFile

**Command**: `"storeFile"`

**Args**: `[path, content, directory]`
- `path` (string) - Unix-style path relative to root
- `content` (string | null) - Base64-encoded file content, null for directories
- `directory` (boolean) - true for directory, false for file

**Response**: `{ cid: string }`
- `cid` - CID of the stored file/directory node

**Error**: Error message if validation fails

**Example**:
```json
// Request: Create directory
{
  "requestID": 10,
  "command": "storeFile",
  "args": ["docs", null, true]
}

// Response
{
  "requestID": 10,
  "result": { "cid": "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG" }
}

// Request: Store file
{
  "requestID": 11,
  "command": "storeFile",
  "args": ["docs/readme.txt", "SGVsbG8sIHdvcmxkIQ==", false]
}

// Response
{
  "requestID": 11,
  "result": { "cid": "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V" }
}
```

**Notes**:
- Content must be base64-encoded for files (required for binary data)
- Updates peer's root directory CID
- Returns the CID of the specific node stored, not the root directory CID

---

#### removeFile

**Command**: `"removeFile"`

**Args**: `[path]`
- `path` (string) - Unix-style path to remove

**Response**: `null`

**Example**:
```json
{
  "requestID": 12,
  "command": "removeFile",
  "args": ["docs/readme.txt"]
}
```

---

### Server Push Messages

#### peerData

**Command**: `"peerData"`

**Args**: `[peer, protocol, data]`
- `peer` (string) - Sender peer ID
- `protocol` (string) - Protocol identifier
- `data` (any) - Message data

**Response**: `null` (client acknowledges receipt)

**Example**:
```json
{
  "requestID": 100,
  "command": "peerData",
  "args": ["12D3KooW...", "chat", {"text": "Hello"}]
}
```

**Notes**:
- Routed to protocol listener registered with `start()`
- Listener receives `(peer, data)` parameters

---

#### topicData

**Command**: `"topicData"`

**Args**: `[topic, peer, data]`
- `topic` (string) - Topic identifier
- `peer` (string) - Sender peer ID
- `data` (any) - Message data

**Response**: `null` (client acknowledges receipt)

**Example**:
```json
{
  "requestID": 101,
  "command": "topicData",
  "args": ["chatroom", "12D3KooW...", {"text": "Hello everyone"}]
}
```

**Notes**:
- Routed to topic listener registered with `subscribe()`

---

#### peerChange

**Command**: `"peerChange"`

**Args**: `[topic, peer, joined]`
- `topic` (string) - Topic identifier
- `peer` (string) - Peer ID that changed
- `joined` (boolean) - `true` if joined, `false` if left

**Response**: `null` (client acknowledges receipt)

**Example**:
```json
{
  "requestID": 102,
  "command": "peerChange",
  "args": ["chatroom", "12D3KooW...", true]
}
```

**Notes**:
- Automatically sent for all subscribed topics
- Routed to peer change callback if provided to `subscribe()`

---

#### ack

**Command**: `"ack"`

**Args**: `[ackNumber]`
- `ackNumber` (number) - Ack number from original send request

**Response**: `null` (client acknowledges receipt)

**Example**:
```json
{
  "requestID": 103,
  "command": "ack",
  "args": [0]
}
```

**Notes**:
- Only sent if send request included non-negative ack parameter
- Client library invokes corresponding ack callback
- Ack numbers managed automatically by client library

---

#### peerFiles

**Command**: `"peerFiles"`

**Args**: `[peerID, rootCID, entries]`
- `peerID` (string) - Peer whose files are being listed
- `rootCID` (string) - CID of peer's root directory
- `entries` (object) - File/directory entries mapping pathnames to entry objects

**Response**: `null` (client acknowledges receipt)

**Entry Format**:
```json
{
  "docs/readme.txt": {
    "type": "file",
    "cid": "QmXg9Pp2ytZ...",
    "mimeType": "text/plain"
  },
  "images": {
    "type": "directory",
    "cid": "QmYwAPJzv5C..."
  }
}
```

**Example**:
```json
{
  "requestID": 104,
  "command": "peerFiles",
  "args": [
    "12D3KooW...",
    "QmRootCID...",
    {
      "readme.txt": {
        "type": "file",
        "cid": "QmFileCID...",
        "mimeType": "text/plain"
      }
    }
  ]
}
```

**Notes**:
- Sent in response to `listFiles` request
- Routed to file list callback registered with `listFiles()`
- Contains complete directory tree structure

---

#### gotFile

**Command**: `"gotFile"`

**Args**: `[cid, result]`
- `cid` (string) - CID that was requested
- `result` (object) - Result object with success flag and content

**Response**: `null` (client acknowledges receipt)

**Result Format (success)**:
```json
// File
{
  "success": true,
  "content": {
    "type": "file",
    "mimeType": "text/plain",
    "content": "SGVsbG8sIHdvcmxkIQ=="
  }
}

// Directory
{
  "success": true,
  "content": {
    "type": "directory",
    "entries": {
      "file.txt": "QmCID1...",
      "subdir": "QmCID2..."
    }
  }
}
```

**Result Format (failure)**:
```json
{
  "success": false,
  "content": {
    "error": "error message"
  }
}
```

**Example**:
```json
{
  "requestID": 105,
  "command": "gotFile",
  "args": [
    "QmXg9Pp2ytZ...",
    {
      "success": true,
      "content": {
        "type": "file",
        "mimeType": "text/plain",
        "content": "SGVsbG8sIHdvcmxkIQ=="
      }
    }
  ]
}
```

**Notes**:
- Sent in response to `getFile` request
- Routed to file content callback registered with `getFile()`
- File content is base64-encoded (required for binary data)

---

## Error Handling

### Error Response Format

```json
{
  "requestID": 1,
  "error": "error message here"
}
```

### Common Errors

**Connection Errors**:
- `"WebSocket connection failed"` - Can't reach server
- `"Connection closed"` - Server shut down

**Protocol Errors**:
- `"peer command must be first"` - Sent other command before Peer()
- `"peer already initialized"` - Sent Peer() twice
- `"duplicate peer"` - PeerID already registered (multi-tab issue)
- `"protocol not started"` - Tried to send on unstarted protocol
- `"protocol already started"` - Tried to start already-started protocol
- `"not subscribed to topic"` - Tried to publish without subscribing

**Network Errors**:
- `"peer unreachable"` - Can't connect to target peer
- `"stream failed"` - libp2p stream error
- `"timeout"` - Operation timed out

### Error Handling Best Practices

```typescript
try {
  await send(peer, 'chat', data);
} catch (error) {
  if (error.message.includes('unreachable')) {
    showNotification('Peer offline, message queued');
    // Implement retry logic
  } else if (error.message.includes('not started')) {
    await start('chat', handleMessage);
    await send(peer, 'chat', data);  // Retry
  } else {
    console.error('Send failed:', error);
  }
}
```

---

## Examples

### Example 1: Simple Chat Application

```typescript
import { connect, start, send } from './client.js';

let myPeerID;
let otherPeer;

// Initialize
async function init() {
  [myPeerID, _] = await connect();

  await start('chat', (peer, data) => {
    displayMessage(peer, data.text);
  });

  console.log('Ready!', myPeerID);
}

// Send message
async function sendMessage(text) {
  await send(otherPeer, 'chat', { text });
}

init();
```

---

### Example 2: Chat Room with Peer List

```typescript
import { connect, subscribe, publish, listPeers } from './client.js';

let myPeerID;
let roomPeers = [];

async function joinRoom() {
  [myPeerID, _] = await connect();

  await subscribe('chatroom',
    // Message callback
    (peer, data) => {
      displayMessage(peer, data.text);
    },
    // Peer change callback
    async (peer, joined) => {
      if (joined) {
        roomPeers.push(peer);
      } else {
        roomPeers = roomPeers.filter(p => p !== peer);
      }
      updatePeerList();
    }
  );

  // Get initial peer list
  roomPeers = await listPeers('chatroom');
  updatePeerList();
}

async function sendToRoom(text) {
  await publish('chatroom', { text });
}

joinRoom();
```

---

### Example 3: Direct Messaging with Delivery Confirmation

```typescript
import { connect, start, send } from './client.js';

let myPeerID;
const pendingMessages = new Map();

async function init() {
  [myPeerID, _] = await connect();

  await start('dm', (peer, data) => {
    displayDM(peer, data.text);
  });
}

async function sendDM(peer, text) {
  const messageId = Date.now();
  pendingMessages.set(messageId, { peer, text });
  showMessageAsPending(messageId);

  await send(peer, 'dm', { text }, () => {
    // Delivery confirmation
    pendingMessages.delete(messageId);
    showMessageAsDelivered(messageId);
  });
}

init();
```

---

### Example 4: Persistent Identity

```typescript
import { connect } from './client.js';

async function initWithPersistence() {
  let peerKey = localStorage.getItem('myPeerKey');

  try {
    const [peerID, newKey] = await connect(peerKey);

    if (!peerKey) {
      // First time, save key
      localStorage.setItem('myPeerKey', newKey);
    }

    console.log('Connected as:', peerID);
    return peerID;

  } catch (error) {
    if (error.message.includes('duplicate')) {
      alert('Already open in another tab!');
      throw error;
    }
  }
}

initWithPersistence();
```

---

### Example 5: Multiple Protocols

```typescript
import { connect, start, send } from './client.js';

async function init() {
  const [myPeerID, _] = await connect();

  // Chat protocol
  await start('chat', (peer, data) => {
    displayChatMessage(peer, data);
  });

  // Game state protocol
  await start('game', (peer, data) => {
    updateGameState(peer, data);
  });

  // File transfer protocol
  await start('files', (peer, data) => {
    receiveFileChunk(peer, data);
  });
}

// Different protocols for different purposes
async function sendChat(peer, text) {
  await send(peer, 'chat', { text });
}

async function sendGameMove(peer, move) {
  await send(peer, 'game', { move });
}

async function sendFile(peer, chunk) {
  await send(peer, 'files', chunk, () => {
    console.log('Chunk delivered');
  });
}

init();
```

---

## References

- **Architecture**: `docs/architecture.md`
- **Developer Guide**: `docs/developer-guide.md`
- **Design Documents**: `design/` directory
- **Specifications**: `specs/main.md`

---

*Last updated: Initial API documentation from CRC design*
