# User Manual

<!-- Spec: main.md -->

## Table of Contents

- [Introduction](#introduction)
- [Getting Started](#getting-started)
- [Using the Demo Application](#using-the-demo-application)
- [Understanding P2P Concepts](#understanding-p2p-concepts)
- [Troubleshooting](#troubleshooting)

## Introduction

<!-- Spec: main.md -->

**What is p2p-webapp?**

p2p-webapp is a local server that enables web applications to communicate peer-to-peer without requiring central servers. It combines web hosting, peer-to-peer networking, and decentralized file storage into a single executable.

**Who is it for?**

- **End users**: Run decentralized web applications on your own computer
- **Developers**: Build collaborative applications without server infrastructure

**Key Features**:
- Peer-to-peer messaging between users
- Group chat via topic subscriptions
- IPFS-based file storage and sharing
- Automatic peer discovery on local networks and globally
- Works offline after initial peer discovery
- No central server required

## Getting Started

### Installation

1. Download the `p2p-webapp` executable for your platform
2. Place it in a convenient location
3. Run it from the command line

### First Run

**With bundled demo**:
```bash
# Simply run the executable
./p2p-webapp

# Server starts and opens browser automatically
# Demo application loads at http://localhost:10000
```

**With your own application**:
```bash
# Extract the demo to get started
mkdir my-app
cd my-app
/path/to/p2p-webapp extract

# Run from the directory
./p2p-webapp --dir .
```

### Configuration (Optional)

<!-- Spec: main.md (FR7: Configuration System) -->

You can customize server behavior with a `p2p-webapp.toml` file:

**Location**: Place in the same directory as `index.html`

**Example** (`p2p-webapp.toml`):
```toml
[server]
port = 10000  # Starting port number

[behavior]
autoOpenBrowser = true  # Open browser automatically
verbosity = 0           # Logging level (0-3)

[p2p]
# Enable automatic file update notifications
# When you or peers upload/remove files, everyone gets notified
fileUpdateNotifyTopic = "chatroom"  # Use your app's chat topic
```

See `p2p-webapp.example.toml` for all available options.

## Using the Demo Application

### Overview

<!-- UI: demo/index.html -->

The bundled demo is a peer-to-peer chatroom with file sharing. It demonstrates all key features of p2p-webapp.

**Main Interface**:
```
┌─────────────────────────────────────────────────┐
│ P2P Chatroom Demo          [Browse Files]       │
│ Connected                                        │
├─────────────────────────┬───────────────────────┤
│                         │  Peers:               │
│  Chat Messages          │  • Chat room          │
│  (Room or DM mode)      │  • Alice              │
│                         │  • Bob                │
│                         │  • Carol              │
│                         │                       │
│                         │                       │
└─────────────────────────┴───────────────────────┘
```

### Chat Features

<!-- Spec: main.md (FR3: PubSub Messaging) -->

#### Room Chat

**Purpose**: Group conversation with all connected peers

**How to use**:
1. Select "Chat room" in the peer list (selected by default)
2. Type your message in the input box
3. Press Enter or click Send
4. Messages appear in the chat area for all participants

**Behind the scenes**:
- Uses topic-based publish/subscribe (pubsub)
- Broadcasts to all peers subscribed to the "chatroom" topic
- Messages delivered via GossipSub protocol

#### Direct Messages

**Purpose**: Private one-on-one conversation with a specific peer

**How to use**:
1. Click a peer name in the peer list (e.g., "Alice")
2. Type your message in the input box
3. Press Enter or click Send
4. Messages only visible between you and that peer

**Behind the scenes**:
- Uses peer-to-peer protocol messaging
- Direct connection between two peers
- Server manages stream lifecycle automatically

**Tip**: The demo stores all messages (both room and DM) so you can switch between conversations without losing messages.

### File Sharing Features

<!-- Spec: main.md (FR4: IPFS File Storage, FR5: File Listing, FR6: File Retrieval) -->

#### Accessing the File Browser

**How to access**:
1. Click "Browse Files" button in the top-right corner
2. File browser modal opens

**Context-aware button text**:
- "Browse My Files" - when chatroom is selected
- "Browse Peer's Files" - when a peer is selected

#### Viewing Files

**Your own files**:
1. In file browser, select your peer ID from dropdown
2. See hierarchical tree of your files and folders
3. Folders shown with 📁 icon, files with 📄 icon
4. Click folder to expand/collapse

**Peer's files**:
1. In file browser, select a peer from dropdown
2. See their shared files (if any)
3. Browse their directory structure
4. Download their files

**File information displayed**:
- File/folder name
- CID (Content Identifier) - unique hash of the content
- MIME type (for files)

#### Uploading Files

**How to upload**:
1. Browse your own files
2. (Optional) Select a folder to upload into
3. Click "Upload Files" button OR drag-and-drop files
4. Files are added to IPFS and directory updated

**What happens**:
- File content stored in IPFS
- CID generated for the file
- File added to your directory at the selected path
- Directory CID updated
- **Other peers get notified** (if file notifications enabled)

#### Creating Directories

**How to create**:
1. Browse your own files
2. (Optional) Select parent folder
3. Click "Create Directory" button
4. Enter directory name
5. Directory created and automatically selected

**Tip**: Use directories to organize files into categories or projects.

#### Downloading Files

**How to download**:
1. Browse files (yours or peer's)
2. Click on a file name
3. File retrieved from IPFS by CID
4. Browser downloads the file

**Behind the scenes**:
- File retrieved using content-addressed storage (IPFS)
- Can retrieve from original peer or any peer that has the content
- Content verification via CID ensures integrity

#### Removing Files

**How to remove**:
1. Browse your own files
2. Select file or folder to remove
3. Click "Remove" button
4. File/folder removed from directory

**What happens**:
- Entry removed from your directory
- Directory CID updated
- **Other peers get notified** (if file notifications enabled)
- Content may still exist in IPFS if other peers have it

### Automatic File List Updates

<!-- Spec: main.md (FR15: File Update Notifications) -->

**What is it?**

When file notifications are enabled, peers automatically receive updates when other peers add or remove files. This keeps file lists synchronized without manual refreshing.

**How it works**:

1. **Configuration**: Application sets `fileUpdateNotifyTopic` in `p2p-webapp.toml`
   ```toml
   [p2p]
   fileUpdateNotifyTopic = "chatroom"
   ```

2. **Automatic notifications**: When you upload/remove files:
   - Server publishes notification to the configured topic
   - All subscribed peers receive the notification
   - If viewing your files, their list refreshes automatically

3. **Visual feedback**: File list updates seamlessly without user action

**Privacy design**:
- Only publishes if you're subscribed to the notification topic
- No notifications sent if you're not participating in the group
- Opt-in mechanism prevents unintended broadcasts

**Example scenario**:

1. Alice, Bob, and Carol are in the chatroom
2. Alice browses Bob's files in the file browser
3. Bob uploads a new file
4. Alice's file list automatically refreshes to show the new file
5. No manual refresh needed!

## Understanding P2P Concepts

### Peer Identity

<!-- Spec: main.md (FR1: Peer Management) -->

**Peer ID**: Unique identifier derived from your peer key

**Peer Key**: Private key that proves your identity

**Session persistence**:
- First visit: p2p-webapp generates new peer key
- Demo saves peer key in browser's localStorage
- Next visit: Same peer ID restored from saved key
- Different browser/device: New peer ID (unless you copy the key)

**Tip**: To maintain the same identity across devices, export your peer key and import it on other devices.

### Peer Discovery

<!-- Spec: main.md (FR11: Peer Discovery, FR12: NAT Traversal) -->

**How peers find each other**:

1. **Local Discovery (mDNS)**:
   - Discovers peers on same network (LAN)
   - Fast (sub-second)
   - Zero configuration

2. **Global Discovery (DHT)**:
   - Discovers peers across the internet
   - Uses distributed hash table
   - Bootstraps via public IPFS nodes

3. **Topic Subscription**:
   - When you subscribe to a topic (e.g., "chatroom")
   - DHT advertises your subscription
   - Other subscribers find you automatically

**NAT Traversal**:

Most home networks use NAT (Network Address Translation), which can block direct connections:

- **Relay**: Connect via intermediate peer
- **Hole Punching**: Attempt direct connection through NAT
- **Automatic**: p2p-webapp handles this transparently

**You don't need to configure anything** - discovery happens automatically!

### Content-Addressed Storage (IPFS)

<!-- Spec: main.md (FR4: IPFS File Storage) -->

**What is a CID?**

CID (Content Identifier) is a hash of file content. It's like a fingerprint:
- Same content = same CID
- Different content = different CID
- Verifiable integrity

**How it works**:

1. Upload file → IPFS generates CID
2. File stored in IPFS network
3. Anyone with the CID can retrieve the file
4. Content verified automatically via CID

**Benefits**:
- **Deduplication**: Same file uploaded twice = one copy
- **Integrity**: Content can't be modified without changing CID
- **Distribution**: Any peer can share content they have

**Directory Structure**:

Your files organized in HAMTDirectory:
- Hierarchical like traditional filesystem
- Each directory has its own CID
- Root CID represents your entire file tree
- Efficient for large directories

### Message Types

<!-- Spec: main.md (FR2: Protocol-Based Messaging, FR3: PubSub Messaging) -->

**Protocol Messages** (Direct):
- Sent to specific peer
- Private one-on-one communication
- Server manages connection lifecycle
- Delivery confirmation via promises

**PubSub Messages** (Broadcast):
- Published to topic
- Received by all topic subscribers
- Group communication
- Best-effort delivery (no confirmation)

**File Notifications** (Special):
- Published to configured topic
- Automatic after file operations
- Contains peer ID and notification type
- Applications can filter and handle

## Troubleshooting

### Connection Issues

**Problem**: "Connecting to network..." status doesn't change

**Possible causes**:
- Firewall blocking connections
- Network doesn't allow P2P traffic
- DHT bootstrap nodes unreachable

**Solutions**:
1. Check firewall settings
2. Try a different network
3. Ensure internet connectivity for DHT bootstrap
4. Check verbose logs: `./p2p-webapp -vv`

---

**Problem**: No peers discovered

**Possible causes**:
- Not subscribed to any topics
- Peers on different networks
- Network blocks mDNS

**Solutions**:
1. Ensure all peers subscribe to same topic (e.g., "chatroom")
2. Wait 10-30 seconds for DHT discovery
3. Try local network first (same WiFi) for faster mDNS discovery
4. Check peer list in UI

### File Operation Issues

**Problem**: File upload fails

**Possible causes**:
- File too large
- IPFS storage error
- Invalid file path

**Solutions**:
1. Try smaller files first
2. Check console for error messages
3. Use simple filenames without special characters
4. Ensure parent directories exist

---

**Problem**: Can't see peer's files

**Possible causes**:
- Peer offline
- Peer hasn't shared any files
- Network connectivity issue

**Solutions**:
1. Verify peer is in peer list (online)
2. Ask peer to share files
3. Try refreshing after a moment
4. Check if you can send messages to the peer

---

**Problem**: File list doesn't auto-refresh

**Possible causes**:
- File notifications not configured
- Peer not subscribed to notification topic
- Configuration mismatch

**Solutions**:
1. Check `p2p-webapp.toml` has `fileUpdateNotifyTopic` set
2. Ensure topic matches subscription (e.g., "chatroom")
3. Manually refresh by re-opening file browser
4. Verify peer is subscribed to the topic

### Performance Issues

**Problem**: Slow file downloads

**Possible causes**:
- Content not available from nearby peers
- Relayed connection (via intermediary)
- Large file size

**Solutions**:
1. Wait for DHT discovery to find more peers
2. Try local network peers first
3. Use direct connections when possible
4. Consider smaller files for better performance

---

**Problem**: High memory usage

**Possible causes**:
- Many large files in IPFS
- Many concurrent connections
- Long-running session

**Solutions**:
1. Restart p2p-webapp periodically
2. Limit number of files stored
3. Close unused connections

### Browser Issues

**Problem**: Duplicate peer ID error

**Cause**: Same peer key used in multiple tabs

**Solution**:
1. Close other tabs with the application
2. Clear localStorage and reload
3. Use different browser/profile for separate identities

---

**Problem**: Messages not appearing

**Possible causes**:
- Not subscribed to correct topic
- Protocol not started
- WebSocket disconnected

**Solutions**:
1. Check connection status indicator
2. Reload page to reconnect
3. Verify topic subscription
4. Check browser console for errors

### Command Line Issues

**Problem**: Port already in use

**Cause**: Another instance running or port occupied

**Solution**:
```bash
# Check running instances
./p2p-webapp ps

# Kill all instances
./p2p-webapp killall

# Or use different port
./p2p-webapp -p 10001
```

---

**Problem**: Bundle not found

**Cause**: Binary not bundled with demo

**Solution**:
```bash
# Check if bundled
./p2p-webapp ls

# If not, build with demo
make build

# Or extract existing bundle
./p2p-webapp extract
```

### Getting Help

**Verbose Logging**:

For detailed troubleshooting information:

```bash
# Level 1: Basic info
./p2p-webapp -v

# Level 2: WebSocket details
./p2p-webapp -vv

# Level 3: Everything
./p2p-webapp -vvv
```

**Check logs for**:
- Peer creation events
- Connection attempts
- Message send/receive
- Error messages

**Where to report issues**:
- Check documentation: `docs/` directory
- Review developer guide: `docs/developer-guide.md`
- Open issue on project repository

---

*Last updated: 2025-11-20 - Added file notification documentation*
