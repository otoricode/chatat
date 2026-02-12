package ws

import "encoding/json"

// WSMessage is the envelope for all WebSocket messages.
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// WebSocket message types.
const (
	WSTypeMessage      = "message"
	WSTypeTyping       = "typing"
	WSTypeOnlineStatus = "online_status"
	WSTypeReadReceipt  = "read_receipt"
	WSTypeDocUpdate    = "doc_update"
	WSTypeDocLock      = "doc_lock"
	WSTypeNotification = "notification"
)
