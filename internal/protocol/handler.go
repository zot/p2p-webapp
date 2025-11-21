// CRC: crc-WebSocketHandler.md, Spec: main.md
package protocol

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/zot/p2p-webapp/internal/peer"
)

// Handler routes and processes protocol messages
// CRC: crc-WebSocketHandler.md
type Handler struct {
	peerManager PeerManager
	nextReqID   int
	mu          sync.Mutex
	pending     map[int]chan *Message
	onSendAck   func(peerID string, ack int) // Callback to send ack messages to WebSocket clients
}

// PeerManager interface for peer operations (Dependency Inversion)
type PeerManager interface {
	CreatePeer(requestedPeerKey string, rootDirectory string) (peerID string, peerKey string, err error)
	RemovePeer(peerID string) error
	GetPeer(peerID string) (peer.PeerOperations, error)
	// Callback setters
	SetPeerFilesCallback(cb func(receiverPeerID, targetPeerID, dirCID string, entries map[string]any))
	SetGotFileCallback(cb func(receiverPeerID string, cid string, success bool, content any))
}

// NewHandler creates a new protocol handler
// CRC: crc-WebSocketHandler.md
func NewHandler(pm PeerManager) *Handler {
	return &Handler{
		peerManager: pm,
		pending:     make(map[int]chan *Message),
	}
}

// SetAckCallback sets the callback for sending ack messages
func (h *Handler) SetAckCallback(callback func(peerID string, ack int)) {
	h.onSendAck = callback
}

// HandleClientMessage processes messages from the client
// CRC: crc-WebSocketHandler.md
func (h *Handler) HandleClientMessage(msg *Message, peerID string) (*Message, error) {
	switch msg.Method {
	case "peer":
		return h.handlePeer(msg)
	case "start":
		return h.handleStart(msg, peerID)
	case "stop":
		return h.handleStop(msg, peerID)
	case "send":
		return h.handleSend(msg, peerID)
	case "subscribe":
		return h.handleSubscribe(msg, peerID)
	case "publish":
		return h.handlePublish(msg, peerID)
	case "unsubscribe":
		return h.handleUnsubscribe(msg, peerID)
	case "listpeers":
		return h.handleListPeers(msg, peerID)
	case "listfiles":
		return h.handleListFiles(msg, peerID)
	case "getfile":
		return h.handleGetFile(msg, peerID)
	case "storefile":
		return h.handleStoreFile(msg, peerID)
	case "removefile":
		return h.handleRemoveFile(msg, peerID)
	default:
		return h.errorResponse(msg.RequestID, 400, fmt.Sprintf("unknown method: %s", msg.Method))
	}
}

// Client request handlers

func (h *Handler) handlePeer(msg *Message) (*Message, error) {
	var req PeerRequest
	if msg.Params != nil {
		if err := json.Unmarshal(msg.Params, &req); err != nil {
			return h.errorResponse(msg.RequestID, 400, "invalid params")
		}
	}

	peerID, peerKey, err := h.peerManager.CreatePeer(req.PeerKey, req.RootDirectory)
	if err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	return h.peerResponse(msg.RequestID, peerID, peerKey)
}

func (h *Handler) handleStart(msg *Message, peerID string) (*Message, error) {
	var req StartRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return h.errorResponse(msg.RequestID, 400, "invalid params")
	}

	peer, err := h.peerManager.GetPeer(peerID)
	if err != nil {
		return h.errorResponse(msg.RequestID, 404, err.Error())
	}

	if err := peer.Start(req.Protocol); err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	return h.emptyResponse(msg.RequestID)
}

func (h *Handler) handleStop(msg *Message, peerID string) (*Message, error) {
	var req StopRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return h.errorResponse(msg.RequestID, 400, "invalid params")
	}

	peer, err := h.peerManager.GetPeer(peerID)
	if err != nil {
		return h.errorResponse(msg.RequestID, 404, err.Error())
	}

	if err := peer.Stop(req.Protocol); err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	return h.emptyResponse(msg.RequestID)
}

func (h *Handler) handleSend(msg *Message, peerID string) (*Message, error) {
	var req SendRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return h.errorResponse(msg.RequestID, 400, "invalid params")
	}

	peer, err := h.peerManager.GetPeer(peerID)
	if err != nil {
		return h.errorResponse(msg.RequestID, 404, err.Error())
	}

	if err := peer.SendToPeer(req.Peer, req.Protocol, req.Data); err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	// If ack was requested (>= 0), send ack message back to client
	if req.Ack >= 0 && h.onSendAck != nil {
		h.onSendAck(peerID, req.Ack)
	}

	return h.emptyResponse(msg.RequestID)
}

func (h *Handler) handleSubscribe(msg *Message, peerID string) (*Message, error) {
	var req SubscribeRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return h.errorResponse(msg.RequestID, 400, "invalid params")
	}

	peer, err := h.peerManager.GetPeer(peerID)
	if err != nil {
		return h.errorResponse(msg.RequestID, 404, err.Error())
	}

	if err := peer.Subscribe(req.Topic); err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	// Automatically start monitoring peer join/leave events
	if err := peer.Monitor(req.Topic); err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	return h.emptyResponse(msg.RequestID)
}

func (h *Handler) handlePublish(msg *Message, peerID string) (*Message, error) {
	var req PublishRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return h.errorResponse(msg.RequestID, 400, "invalid params")
	}

	peer, err := h.peerManager.GetPeer(peerID)
	if err != nil {
		return h.errorResponse(msg.RequestID, 404, err.Error())
	}

	if err := peer.Publish(req.Topic, req.Data); err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	return h.emptyResponse(msg.RequestID)
}

func (h *Handler) handleUnsubscribe(msg *Message, peerID string) (*Message, error) {
	var req UnsubscribeRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return h.errorResponse(msg.RequestID, 400, "invalid params")
	}

	peer, err := h.peerManager.GetPeer(peerID)
	if err != nil {
		return h.errorResponse(msg.RequestID, 404, err.Error())
	}

	if err := peer.Unsubscribe(req.Topic); err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	// Automatically stop monitoring peer join/leave events
	if err := peer.StopMonitor(req.Topic); err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	return h.emptyResponse(msg.RequestID)
}

func (h *Handler) handleListPeers(msg *Message, peerID string) (*Message, error) {
	var req ListPeersRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return h.errorResponse(msg.RequestID, 400, "invalid params")
	}

	peer, err := h.peerManager.GetPeer(peerID)
	if err != nil {
		return h.errorResponse(msg.RequestID, 404, err.Error())
	}

	peers, err := peer.ListPeers(req.Topic)
	if err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	resp := ListPeersResponse{Peers: peers}
	result, _ := json.Marshal(resp)
	return &Message{
		RequestID:  msg.RequestID,
		IsResponse: true,
		Result:     result,
	}, nil
}

func (h *Handler) handleListFiles(msg *Message, peerID string) (*Message, error) {
	var req ListFilesRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return h.errorResponse(msg.RequestID, 400, "invalid params")
	}

	// Get the requesting peer
	peer, err := h.peerManager.GetPeer(peerID)
	if err != nil {
		return h.errorResponse(msg.RequestID, 404, err.Error())
	}

	// Async operation - actual result comes via peerFiles server message
	if err := peer.ListFiles(req.PeerID); err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	return h.emptyResponse(msg.RequestID)
}

func (h *Handler) handleGetFile(msg *Message, peerID string) (*Message, error) {
	var req GetFileRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return h.errorResponse(msg.RequestID, 400, "invalid params")
	}

	// Get the requesting peer
	peer, err := h.peerManager.GetPeer(peerID)
	if err != nil {
		return h.errorResponse(msg.RequestID, 404, err.Error())
	}

	// Async operation - actual result comes via gotFile server message
	if err := peer.GetFile(req.CID); err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	return h.emptyResponse(msg.RequestID)
}

func (h *Handler) handleStoreFile(msg *Message, peerID string) (*Message, error) {
	var req StoreFileRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return h.errorResponse(msg.RequestID, 400, "invalid params")
	}

	peer, err := h.peerManager.GetPeer(peerID)
	if err != nil {
		return h.errorResponse(msg.RequestID, 404, err.Error())
	}

	var content []byte
	if !req.Directory {
		// Decode base64 content for files
		var err error
		content, err = base64.StdEncoding.DecodeString(req.Content)
		if err != nil {
			return h.errorResponse(msg.RequestID, 400, "invalid base64 content")
		}
	}

	fileCID, rootCID, err := peer.StoreFile(req.Path, content, req.Directory)
	if err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	// Return file CID and root CID in response
	response := map[string]string{
		"fileCid": fileCID,
		"rootCid": rootCID,
	}
	result, _ := json.Marshal(response)
	return &Message{
		RequestID:  msg.RequestID,
		IsResponse: true,
		Result:     result,
	}, nil
}

func (h *Handler) handleRemoveFile(msg *Message, peerID string) (*Message, error) {
	var req RemoveFileRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return h.errorResponse(msg.RequestID, 400, "invalid params")
	}

	peer, err := h.peerManager.GetPeer(peerID)
	if err != nil {
		return h.errorResponse(msg.RequestID, 404, err.Error())
	}

	if err := peer.RemoveFile(req.Path); err != nil {
		return h.errorResponse(msg.RequestID, 500, err.Error())
	}

	return h.emptyResponse(msg.RequestID)
}

// Server message senders (to be called by peer manager)

func (h *Handler) NextRequestID() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	id := h.nextReqID
	h.nextReqID++
	return id
}

func (h *Handler) CreatePeerDataMessage(peer, protocol string, data any) *Message {
	req := PeerDataRequest{
		Peer:     peer,
		Protocol: protocol,
		Data:     data,
	}
	params, _ := json.Marshal(req)
	return &Message{
		RequestID: h.NextRequestID(),
		Method:    "peerData",
		Params:    params,
	}
}

func (h *Handler) CreateTopicDataMessage(topic, peerID string, data any) *Message {
	req := TopicDataRequest{
		Topic:  topic,
		PeerID: peerID,
		Data:   data,
	}
	params, _ := json.Marshal(req)
	return &Message{
		RequestID: h.NextRequestID(),
		Method:    "topicData",
		Params:    params,
	}
}

func (h *Handler) CreatePeerChangeMessage(topic, peerID string, joined bool) *Message {
	req := PeerChangeRequest{
		Topic:  topic,
		PeerID: peerID,
		Joined: joined,
	}
	params, _ := json.Marshal(req)
	return &Message{
		RequestID: h.NextRequestID(),
		Method:    "peerChange",
		Params:    params,
	}
}

func (h *Handler) CreateAckMessage(ack int) *Message {
	req := AckRequest{
		Ack: ack,
	}
	params, _ := json.Marshal(req)
	return &Message{
		RequestID: h.NextRequestID(),
		Method:    "ack",
		Params:    params,
	}
}

func (h *Handler) CreatePeerFilesMessage(peerID, cid string, entries map[string]FileEntryInfo) *Message {
	req := PeerFilesRequest{
		PeerID:  peerID,
		CID:     cid,
		Entries: entries,
	}
	params, _ := json.Marshal(req)
	return &Message{
		RequestID: h.NextRequestID(),
		Method:    "peerFiles",
		Params:    params,
	}
}

func (h *Handler) CreateGotFileMessage(cid string, success bool, content any) *Message {
	req := GotFileRequest{
		CID:     cid,
		Success: success,
		Content: content,
	}
	params, _ := json.Marshal(req)
	return &Message{
		RequestID: h.NextRequestID(),
		Method:    "gotFile",
		Params:    params,
	}
}

// Response helpers

func (h *Handler) emptyResponse(requestID int) (*Message, error) {
	return &Message{
		RequestID:  requestID,
		IsResponse: true,
		Result:     json.RawMessage("null"),
	}, nil
}

func (h *Handler) stringResponse(requestID int, value string) (*Message, error) {
	resp := StringResponse{Value: value}
	result, _ := json.Marshal(resp)
	return &Message{
		RequestID:  requestID,
		IsResponse: true,
		Result:     result,
	}, nil
}

func (h *Handler) peerResponse(requestID int, peerID, peerKey string) (*Message, error) {
	resp := PeerResponse{PeerID: peerID, PeerKey: peerKey}
	result, _ := json.Marshal(resp)
	return &Message{
		RequestID:  requestID,
		IsResponse: true,
		Result:     result,
	}, nil
}

func (h *Handler) errorResponse(requestID, code int, message string) (*Message, error) {
	return &Message{
		RequestID:  requestID,
		IsResponse: true,
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
		},
	}, nil
}
