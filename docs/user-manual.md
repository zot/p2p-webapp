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

See `docs/examples/p2p-webapp.toml` for all available options.

## Using the Demo Application

### Overview

<!-- UI: demo/index.html -->

The bundled demo is a peer-to-peer chatroom with file sharing. It demonstrates all key features of p2p-webapp.

**Main Interface**:
```
┌───────────────────────────────────────────────────────────┐
│ P2P Chatroom Demo    [Contacts] [Browse Files]            │
│ Connected                                                  │
├─────────────────────────────┬─────────────────────────────┤
│                             │  Peers:                      │
│  Chat Messages              │  • Chat room                 │
│  (Room or DM mode)          │  • Alice                     │
│                             │  • Bob                       │
│                             │  • Carol                     │
│                             │                              │
│                             │                              │
└─────────────────────────────┴─────────────────────────────┘
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

### Managing Contact List

<!-- Spec: main.md (Demo Chatroom Application - Contact List) -->

**What is it?**

The contact list feature allows you to protect connections to specific peers, ensuring they remain active even when the connection manager needs to prune connections. Protected peers receive higher priority, making them less likely to be disconnected.

**How to access**:

1. Click the "Contacts" button in the top-right corner of the header (next to "Browse Files")
2. Contact management modal opens

**Adding a peer to your contacts**:

1. In the contacts modal, enter a peer ID in the input box
2. Click "Add Contact" or press Enter
3. Peer ID appears in the contact list below

**What happens when you add a contact**:
- The peer's connection is protected from being closed by the connection manager
- The peer is tagged with a priority value (higher priority = more important)
- If the peer is known but not connected, a connection attempt is made

**Removing a peer from contacts**:

1. In the contacts modal, find the peer ID in the list
2. Click the "×" button next to the peer ID
3. Peer ID is removed from the contact list

**What happens when you remove a contact**:
- The peer's connection protection is removed
- The priority tag is removed
- The connection itself is NOT closed, but may be pruned later if needed

**Accepting or canceling changes**:

- **Accept**: Applies all changes (added and removed contacts) to the connection manager
  - Calls `addPeers()` for newly added peer IDs
  - Calls `removePeers()` for removed peer IDs
  - Contact list is saved for the session
- **Cancel**: Discards all changes made in the modal
  - Contact list reverts to previous state
  - No changes applied to connection manager

**Session-only persistence**:

- Contact list is stored in memory for the current browser session
- Not saved to localStorage or other persistent storage
- Closing the browser tab clears the contact list
- Each new session starts with an empty contact list

**When to use contacts**:

- Keep connections to relay nodes active
- Maintain connections to important peers for your application
- Ensure critical peers stay connected during testing
- Prioritize connections to frequently-used peers

**Input validation**:

- Duplicate peer IDs are prevented automatically
- Invalid peer ID formats are ignored silently
- Empty input is not added to the list

**Tip**: Add peers you frequently communicate with to your contacts to ensure stable connections throughout your session.

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
<!-- CRC: crc-Peer.md -->
<!-- Sequence: seq-pubsub-communication.md -->

**How peers find each other automatically**:

p2p-webapp uses two discovery systems working together:

#### 1. Local Discovery (mDNS)

**What it does**: Finds peers on your same network (WiFi/LAN)

**How it works**:
- Broadcasts "I'm here!" on your local network
- Other peers on same network instantly respond
- Very fast (under 1 second)
- Works without internet connection

**Best for**:
- Development on the same computer
- Collaboration in same office/home
- Testing locally

#### 2. Global Discovery (DHT + Topic Advertisement)

<!-- Sequence: seq-dht-bootstrap.md -->

**What it does**: Finds peers anywhere on the internet based on shared interests

**Important**: DHT operations queue automatically during bootstrap (first 5-30 seconds), so you won't see "no peers in table" errors!

**How it works**:

**When you first start the app**:
- DHT bootstrap starts in the background (5-30 seconds)
- Connects to public IPFS nodes to build routing table
- Operations queue automatically until bootstrap completes
- You can use the app immediately - queuing is transparent!

**When you join a chatroom** (or subscribe to any topic):

1. **You advertise your interest**:
   - p2p-webapp tells the DHT network "I'm in the 'chatroom' topic"
   - If DHT is still bootstrapping, this operation queues automatically
   - When bootstrap completes (5-30 seconds), advertisement starts
   - Re-advertises periodically to stay discoverable
   - Like posting "Looking for chatroom friends" on a global bulletin board

2. **You discover others**:
   - p2p-webapp asks the DHT "Who else is in 'chatroom'?"
   - If DHT is still bootstrapping, this operation queues automatically
   - When bootstrap completes, discovery starts
   - DHT returns peer addresses from around the world
   - Automatic connection attempts to discovered peers

3. **Others discover you**:
   - When new peers join "chatroom", they find your advertisement
   - They automatically try to connect to you

**Best for**:
- Connecting with users on different networks (home, office, mobile)
- Geographic collaboration (different cities/countries)
- Public applications without known server addresses

**Timeline**:
- **Local peers (mDNS)**: Found in < 1 second
- **DHT bootstrap**: 5-15 seconds (typical), max 30 seconds
- **Remote peers (DHT)**: Found 10-30 seconds after bootstrap completes
- **First join after startup**: May take 10-30 seconds for DHT operations (queued during bootstrap)
- **Subsequent joins**: Immediate (DHT already ready)
- **Be patient** - global discovery takes time, but it's automatic!

#### Why Both?

**Together they provide complete coverage**:
- **mDNS**: Fast local connections (milliseconds)
- **DHT**: Global reach (seconds to minutes)
- **Automatic**: Both work simultaneously, no setup needed

**Example scenario**:
- You and a coworker in same office → Found via mDNS instantly
- You and a friend in different cities → Found via DHT after ~15 seconds
- You work from home, coworker at office → DHT enables connection

#### NAT Traversal (Firewall/Router Workaround)

**The problem**: Most home networks have firewalls (NAT) that block direct connections

**The solution**: p2p-webapp automatically handles this:

- **Circuit Relay**: Connects via an intermediate helper peer
  - Like asking a mutual friend to pass messages
  - Slower but works everywhere

- **Hole Punching**: Attempts direct connection through firewall
  - Tries to "punch a hole" through both firewalls
  - Faster if successful

- **AutoRelay**: Automatically finds public relay servers
  - Discovers relay peers via DHT
  - No manual configuration needed

- **Port Mapping**: Tries to automatically open firewall port
  - Uses UPnP or NAT-PMP protocols
  - Works on compatible routers

**All automatic** - you don't need to do anything!

#### Discovery Process Visualized

```
You start the app
    ↓
DHT bootstrap starts (background, 5-30 seconds)
    ↓
You join "chatroom" topic
    ↓
Subscribe succeeds immediately
    ↓
[Operations queue if DHT not ready yet]
    ↓
[Instant] mDNS finds Alice (same WiFi) ───→ Connected in 500ms
    ↓
[5-15 seconds] DHT bootstrap completes
    ↓
[Queued operations execute automatically]
    ├─ Advertise "chatroom" to DHT
    └─ Query DHT for "chatroom" peers
    ↓
[10-20 seconds] DHT finds Bob (different city)
    ↓
[15-25 seconds] DHT finds Carol (via relay)
    ↓
[30+ seconds] New peer Dave joins, finds you via your DHT advertisement
    ↓
All connected and ready to chat!
```

#### What You Need to Know

**For users**:
- **Zero configuration**: Just open the app and join a topic
- **Immediate success**: Subscribe succeeds right away (queuing handled automatically)
- **Be patient**: First join may take 10-30 seconds for DHT operations (bootstrap delay)
- **Subsequent joins**: Instant (DHT already ready)
- **Same topic name**: Ensure everyone uses exact same name (case-sensitive)
- **Internet required**: DHT needs internet for bootstrap (first-time connection to DHT network)
- **Local works offline**: mDNS works without internet
- **No error messages**: "failed to find any peer in table" error is prevented by queuing

**Troubleshooting**:
- **No peers found immediately**: Normal - DHT bootstrap takes 5-30 seconds
  - Wait for bootstrap to complete
  - Queued operations will execute automatically
  - Check logs with `-vv` to see bootstrap progress
- **Only local peers**: Check internet connection for DHT bootstrap
- **Slow first discovery**: Normal - DHT bootstrap + operations take 10-30 seconds
  - Subsequent operations are immediate
- **Different networks**: DHT is designed for this - wait for bootstrap and discovery

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
- DHT bootstrap still in progress (first 5-30 seconds)
- Not subscribed to any topics
- Waiting for DHT operations to execute (queued during bootstrap)
- Different topic names (case-sensitive)
- Network blocks mDNS

**Solutions**:
1. **Wait for DHT bootstrap** (5-30 seconds after peer creation)
   - Operations queue automatically during bootstrap
   - No error messages displayed (queuing is transparent)
   - Run with `-vv` to see "DHT ready" and "Processing queued operations"
2. Ensure all peers subscribe to **exact same topic** (e.g., "chatroom") - case-sensitive!
3. **Be patient with first subscribe** - may take 10-30 seconds for DHT operations
   - Subsequent subscribes are immediate (DHT already ready)
4. Check internet connection for DHT bootstrap to public IPFS nodes
5. Try local network first (same WiFi) for faster mDNS discovery (< 1 second)
6. Check peer list in UI
7. Run with verbose logging to see bootstrap and queuing: `./p2p-webapp -vv`

---

**Problem**: Only local peers discovered, no remote peers

**Possible causes**:
- DHT bootstrap still in progress (first 5-30 seconds)
- DHT operations queued (waiting for bootstrap)
- No internet connection for DHT bootstrap
- No remote peers subscribed to topic yet

**Solutions**:
1. **Wait for DHT bootstrap** to complete (5-30 seconds typical, up to 30s max)
2. Check internet connectivity - DHT requires connection to IPFS bootstrap nodes
3. Verify other remote peers have subscribed to the same topic
4. Check verbose logs for DHT bootstrap and queuing status: `./p2p-webapp -vv`
5. Look for these log messages:
   - "Queued DHT operation" - operations queuing during bootstrap
   - "DHT ready" - bootstrap complete
   - "Processing N queued DHT operations" - queue executing
   - "DHT: advertising topic..." - advertisement started
   - "DHT: discovered peer..." - peers found via DHT

---

**Problem**: Peers discovered but can't connect

**Possible causes**:
- NAT/firewall blocking connections
- No relay peers available yet
- Network configuration issues

**Solutions**:
1. Wait for DHT to discover relay peers (automatic, takes 10-30 seconds after bootstrap)
2. Check NAT/firewall settings on router
3. Ensure UPnP enabled on router for automatic port mapping
4. Try from different network to isolate issue
5. Check verbose logs for connection errors: `./p2p-webapp -vv`

---

**Problem**: "failed to find any peer in table" error

**This error should not occur** - operations queue automatically during DHT bootstrap.

**If you see this error**:
1. This indicates a bug - operations should queue automatically
2. Report the issue with steps to reproduce
3. Include verbose logs: `./p2p-webapp -vv`
4. Check if DHT is enabled (should be by default)

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

*Last updated: 2025-11-24 - Added DHT topic discovery and enhanced troubleshooting*
