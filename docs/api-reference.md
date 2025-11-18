# p2p-webapp API Reference

**Complete API documentation for p2p-webapp WebSocket protocol and TypeScript client library**

---

## Table of Contents

1. [TypeScript Client Library](#typescript-client-library)
2. [WebSocket Protocol](#websocket-protocol)
3. [Message Types](#message-types)
4. [Error Handling](#error-handling)
5. [Examples](#examples)

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
import { connect, start, stop, send, subscribe, publish, unsubscribe, listPeers } from './client.js';
```

Or with named export:

```typescript
import * as P2P from './client.js';
```

---

### Core API

#### `connect(peerKey?: string): Promise<[string, string]>`

Connect to p2p-webapp server and initialize peer.

**Parameters**:
- `peerKey` (optional) - Existing peer key to reuse identity. If omitted, generates new key.

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

#### `send(peer: string, protocol: string, data: any, onAck?: AckCallback): Promise<void>`

Send data to peer on protocol.

**Parameters**:
- `peer` - Target peer ID
- `protocol` - Protocol identifier
- `data` - Any JSON-serializable data
- `onAck` (optional) - Callback invoked when delivery confirmed: `() => void | Promise<void>`

**Returns**: Promise resolving when request sent (NOT when delivered)

**Throws**: Error if protocol not started

**Example**:
```typescript
// Send without ack
await send(peerID, 'chat', { text: 'Hello!' });

// Send with delivery confirmation
await send(peerID, 'chat', { text: 'Important!' }, () => {
  console.log('Message delivered');
  showDeliveryCheckmark();
});
```

**Notes**:
- Protocol must be started first with `start()`
- Server manages connection lifecycle transparently
- `onAck` callback provides delivery confirmation
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

### Type Definitions

```typescript
type ProtocolDataCallback = (peer: string, data: any) => void | Promise<void>;
type TopicDataCallback = (peer: string, data: any) => void | Promise<void>;
type PeerChangeCallback = (peer: string, joined: boolean) => void | Promise<void>;
type AckCallback = () => void | Promise<void>;
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
