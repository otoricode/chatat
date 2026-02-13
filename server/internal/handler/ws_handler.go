package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/ws"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// WSHandler handles WebSocket connections.
type WSHandler struct {
	hub             *ws.Hub
	jwtSecret       string
	chatRepo        repository.ChatRepository
	topicRepo       repository.TopicRepository
	messageStatRepo repository.MessageStatusRepository
	redis           *redis.Client
}

// NewWSHandler creates a new WebSocket handler.
func NewWSHandler(hub *ws.Hub, jwtSecret string, chatRepo repository.ChatRepository, topicRepo repository.TopicRepository, messageStatRepo repository.MessageStatusRepository, redisClient *redis.Client) *WSHandler {
	return &WSHandler{
		hub:             hub,
		jwtSecret:       jwtSecret,
		chatRepo:        chatRepo,
		topicRepo:       topicRepo,
		messageStatRepo: messageStatRepo,
		redis:           redisClient,
	}
}

// HandleConnection upgrades an HTTP connection to WebSocket.
func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		response.Error(w, apperror.Unauthorized("missing token query parameter"))
		return
	}

	userID, err := h.validateToken(tokenString)
	if err != nil {
		response.Error(w, apperror.Unauthorized("invalid or expired token"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket upgrade failed")
		return
	}

	client := ws.NewClient(h.hub, conn, userID)
	client.MessageHandler = h.handleClientMessage
	h.hub.RegisterClient(client)

	// Join user's personal notification room
	h.hub.JoinRoom(client, "user:"+userID.String())

	// Join all chat rooms the user belongs to
	go h.joinUserChatRooms(client)

	// Join all topic rooms the user belongs to
	go h.joinUserTopicRooms(client)

	go client.WritePump()
	go client.ReadPump()
}

// joinUserChatRooms loads user's chats and joins their WS rooms.
func (h *WSHandler) joinUserChatRooms(client *ws.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chats, err := h.chatRepo.ListByUser(ctx, client.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", client.UserID.String()).Msg("failed to load user chats for WS rooms")
		return
	}

	for _, c := range chats {
		roomID := "chat:" + c.Chat.ID.String()
		h.hub.JoinRoom(client, roomID)
	}

	log.Debug().
		Str("user_id", client.UserID.String()).
		Int("rooms", len(chats)).
		Msg("user joined chat rooms")
}

// joinUserTopicRooms loads user's topics and joins their WS rooms.
func (h *WSHandler) joinUserTopicRooms(client *ws.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topics, err := h.topicRepo.ListByUser(ctx, client.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", client.UserID.String()).Msg("failed to load user topics for WS rooms")
		return
	}

	for _, t := range topics {
		roomID := "topic:" + t.ID.String()
		h.hub.JoinRoom(client, roomID)
	}

	log.Debug().
		Str("user_id", client.UserID.String()).
		Int("rooms", len(topics)).
		Msg("user joined topic rooms")
}

// handleClientMessage routes incoming WS messages from clients.
func (h *WSHandler) handleClientMessage(client *ws.Client, msg ws.WSMessage) {
	switch msg.Type {
	case ws.WSTypeTyping:
		h.handleTyping(client, msg.Payload)
	case ws.WSTypeMessageAck:
		h.handleMessageAck(client, msg.Payload)
	case ws.WSTypeReadReceipt:
		h.handleReadReceipt(client, msg.Payload)
	default:
		log.Debug().
			Str("user_id", client.UserID.String()).
			Str("type", msg.Type).
			Msg("unhandled ws message type")
	}
}

// --- Typing ---

type typingPayload struct {
	ChatID   string `json:"chatId"`
	IsTyping bool   `json:"isTyping"`
}

type typingBroadcast struct {
	ChatID   string `json:"chatId"`
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
	IsTyping bool   `json:"isTyping"`
}

func (h *WSHandler) handleTyping(client *ws.Client, payload json.RawMessage) {
	var p typingPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}

	// Store in Redis with 3s TTL
	if h.redis != nil {
		key := "typing:" + p.ChatID + ":" + client.UserID.String()
		if p.IsTyping {
			h.redis.Set(context.Background(), key, "1", 3*time.Second)
		} else {
			h.redis.Del(context.Background(), key)
		}
	}

	// Broadcast to chat room (exclude sender)
	broadcast := typingBroadcast{
		ChatID:   p.ChatID,
		UserID:   client.UserID.String(),
		UserName: "", // Will be resolved on client side from memberMap
		IsTyping: p.IsTyping,
	}

	bPayload, err := json.Marshal(broadcast)
	if err != nil {
		return
	}

	wsMsg := ws.WSMessage{
		Type:    ws.WSTypeTyping,
		Payload: bPayload,
	}

	data, err := json.Marshal(wsMsg)
	if err != nil {
		return
	}

	roomID := "chat:" + p.ChatID
	h.hub.SendToRoom(roomID, data, client.UserID)
}

// --- Message Ack (Delivered) ---

type messageAckPayload struct {
	MessageID string `json:"messageId"`
	ChatID    string `json:"chatId"`
	Status    string `json:"status"`
}

func (h *WSHandler) handleMessageAck(client *ws.Client, payload json.RawMessage) {
	var p messageAckPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}

	msgID, err := uuid.Parse(p.MessageID)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Update message status to delivered
	if err := h.messageStatRepo.UpdateStatus(ctx, msgID, client.UserID, "delivered"); err != nil {
		log.Warn().Err(err).Str("message_id", p.MessageID).Msg("failed to update delivery status")
		return
	}

	// Broadcast status change to chat room
	h.broadcastMessageStatus(p.ChatID, p.MessageID, client.UserID.String(), "delivered")
}

// --- Read Receipt ---

type readReceiptPayload struct {
	ChatID           string `json:"chatId"`
	LastReadMessageID string `json:"lastReadMessageId"`
}

func (h *WSHandler) handleReadReceipt(client *ws.Client, payload json.RawMessage) {
	var p readReceiptPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}

	chatID, err := uuid.Parse(p.ChatID)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Mark all messages in chat as read for this user
	if err := h.messageStatRepo.MarkChatAsRead(ctx, chatID, client.UserID); err != nil {
		log.Warn().Err(err).
			Str("chat_id", p.ChatID).
			Str("user_id", client.UserID.String()).
			Msg("failed to mark messages as read")
		return
	}

	// Broadcast read receipt to chat room
	lastReadID := p.LastReadMessageID
	if lastReadID == "" {
		lastReadID = p.ChatID // fallback
	}
	h.broadcastMessageStatus(p.ChatID, lastReadID, client.UserID.String(), "read")
}

// --- Helpers ---

type messageStatusBroadcast struct {
	ChatID    string `json:"chatId"`
	MessageID string `json:"messageId"`
	UserID    string `json:"userId"`
	Status    string `json:"status"`
}

func (h *WSHandler) broadcastMessageStatus(chatID, messageID, userID, status string) {
	broadcast := messageStatusBroadcast{
		ChatID:    chatID,
		MessageID: messageID,
		UserID:    userID,
		Status:    status,
	}

	bPayload, err := json.Marshal(broadcast)
	if err != nil {
		return
	}

	wsMsg := ws.WSMessage{
		Type:    ws.WSTypeMessageStatus,
		Payload: bPayload,
	}

	data, err := json.Marshal(wsMsg)
	if err != nil {
		return
	}

	roomID := "chat:" + chatID
	h.hub.SendToRoom(roomID, data, uuid.Nil)
}

func (h *WSHandler) validateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, apperror.Unauthorized("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, apperror.Unauthorized("invalid claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, apperror.Unauthorized("missing subject")
	}

	return uuid.Parse(sub)
}
