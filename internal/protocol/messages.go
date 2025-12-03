package protocol

import "encoding/json"

// Message envelope for all WebSocket communications
type Message struct {
	RequestID  int             `json:"requestid"`
	Method     string          `json:"method"`
	Params     json.RawMessage `json:"params,omitempty"`
	Result     json.RawMessage `json:"result,omitempty"`
	Error      *ErrorResponse  `json:"error,omitempty"`
	IsResponse bool            `json:"isresponse"`
}

// Shared response structs (minimize duplication)

// EmptyResponse is used for operations that return "null or error"
type EmptyResponse struct{}

// StringResponse is used for operations that return a single string (peerid, connectionID)
type StringResponse struct {
	Value string `json:"value"`
}

// PeerResponse is used for the Peer command, returning {peerid, peerkey, version}
type PeerResponse struct {
	PeerID  string `json:"peerid"`
	PeerKey string `json:"peerkey"`
	Version string `json:"version"`
}

// ErrorResponse provides standardized error structure
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Client Request Messages

// PeerRequest creates or restores a peer
type PeerRequest struct {
	PeerKey       string `json:"peerkey,omitempty"`
	RootDirectory string `json:"rootDirectory,omitempty"` // Optional CID of peer's root directory
}

// StartRequest starts listening for a protocol
type StartRequest struct {
	Protocol string `json:"protocol"`
}

// StopRequest stops listening for a protocol
type StopRequest struct {
	Protocol string `json:"protocol"`
}

// SendRequest sends data to a peer on a protocol
type SendRequest struct {
	Peer     string `json:"peer"`
	Protocol string `json:"protocol"`
	Data     any    `json:"data"`
	Ack      int    `json:"ack"` // If >= 0, server sends ack message when delivered; -1 = no ack
}

// SubscribeRequest subscribes to a topic
type SubscribeRequest struct {
	Topic string `json:"topic"`
}

// PublishRequest publishes data to a topic
type PublishRequest struct {
	Topic string `json:"topic"`
	Data  any    `json:"data"`
}

// UnsubscribeRequest unsubscribes from a topic
type UnsubscribeRequest struct {
	Topic string `json:"topic"`
}

// ListPeersRequest lists peers subscribed to a topic
type ListPeersRequest struct {
	Topic string `json:"topic"`
}

// ListPeersResponse returns list of peer IDs
type ListPeersResponse struct {
	Peers []string `json:"peers"`
}

// AddPeersRequest protects and tags peer connections
// CRC: crc-Peer.md
// Sequence: seq-add-peers.md
type AddPeersRequest struct {
	PeerIDs []string `json:"peerIds"`
}

// RemovePeersRequest unprotects and untags peer connections
// CRC: crc-Peer.md
// Sequence: seq-remove-peers.md
type RemovePeersRequest struct {
	PeerIDs []string `json:"peerIds"`
}

// Server Request Messages (sent from server to client)

// PeerDataRequest delivers data from a peer on a protocol
type PeerDataRequest struct {
	Peer     string `json:"peer"`
	Protocol string `json:"protocol"`
	Data     any    `json:"data"`
}

// TopicDataRequest delivers data from a topic
type TopicDataRequest struct {
	Topic  string `json:"topic"`
	PeerID string `json:"peerid"`
	Data   any    `json:"data"`
}

// PeerChangeRequest notifies client that a peer joined or left a subscribed topic
type PeerChangeRequest struct {
	Topic  string `json:"topic"`
	PeerID string `json:"peerid"`
	Joined bool   `json:"joined"` // true = joined, false = left
}

// AckRequest notifies client that a message was successfully delivered
type AckRequest struct {
	Ack int `json:"ack"`
}

// PeerFilesRequest notifies client of a peer's file list (server-to-client)
type PeerFilesRequest struct {
	PeerID  string                   `json:"peerid"`  // Target peer whose files were listed
	CID     string                   `json:"cid"`     // Root directory CID
	Entries map[string]FileEntryInfo `json:"entries"` // Full pathname tree
}

// FileEntryInfo contains metadata about a file or directory
type FileEntryInfo struct {
	Type     string `json:"type"`               // "file" or "directory"
	CID      string `json:"cid"`                // Content identifier
	MimeType string `json:"mimeType,omitempty"` // MIME type for files
}

// GotFileRequest notifies client of file retrieval result (server-to-client)
type GotFileRequest struct {
	CID     string `json:"cid"`     // Requested CID
	Success bool   `json:"success"` // Whether retrieval was successful
	Content any    `json:"content"` // File content or error info
}

// File Operation Messages

// ListFilesRequest requests a peer's file list (async, result via peerFiles server message)
type ListFilesRequest struct {
	PeerID string `json:"peerid"` // Peer whose files to list
}

// GetFileRequest requests file content by CID (async, result via gotFile server message)
type GetFileRequest struct {
	CID            string `json:"cid"`
	FallbackPeerID string `json:"fallbackPeerID,omitempty"` // Optional peer to request from if not found locally
}

// StoreFileRequest stores file or directory content
type StoreFileRequest struct {
	Path      string `json:"path"`
	Content   string `json:"content,omitempty"` // base64 encoded file content (null for directories)
	Directory bool   `json:"directory"`          // true = create directory, false = create file
}

// RemoveFileRequest removes a file or directory
type RemoveFileRequest struct {
	Path string `json:"path"`
}
