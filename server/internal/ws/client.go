package ws

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096

	// WebSocket rate limiting: max messages per window
	wsRateLimitMessages = 30
	wsRateLimitWindow   = 60 * time.Second
)

// MessageHandler is a callback invoked when the client receives a typed message.
type MessageHandler func(client *Client, msg WSMessage)

// Client represents a single WebSocket connection.
type Client struct {
	UserID         uuid.UUID
	Conn           *websocket.Conn
	Send           chan []byte
	Hub            *Hub
	MessageHandler MessageHandler

	// Rate limiting state (per-client, not shared)
	msgCount       int
	msgWindowStart time.Time
}

// NewClient creates a new client.
func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID) *Client {
	return &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub.
// Should be called in a goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.UnregisterClient(c)
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Error().Err(err).Str("user_id", c.UserID.String()).Msg("failed to set read deadline")
		return
	}
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Warn().Err(err).Str("user_id", c.UserID.String()).Msg("websocket unexpected close")
			}
			break
		}

		// Rate limiting: sliding window counter
		now := time.Now()
		if now.Sub(c.msgWindowStart) > wsRateLimitWindow {
			c.msgCount = 0
			c.msgWindowStart = now
		}
		c.msgCount++
		if c.msgCount > wsRateLimitMessages {
			log.Warn().Str("user_id", c.UserID.String()).Msg("websocket rate limit exceeded, disconnecting")
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Warn().Err(err).Str("user_id", c.UserID.String()).Msg("invalid ws message format")
			continue
		}

		if c.MessageHandler != nil {
			c.MessageHandler(c, wsMsg)
		} else {
			log.Debug().
				Str("user_id", c.UserID.String()).
				Str("type", wsMsg.Type).
				Msg("ws message received (no handler)")
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
// Should be called in a goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Error().Err(err).Str("user_id", c.UserID.String()).Msg("failed to set write deadline")
				return
			}
			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Debug().Err(err).Msg("failed to write close message")
				}
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Warn().Err(err).Str("user_id", c.UserID.String()).Msg("failed to write ws message")
				return
			}

		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Error().Err(err).Str("user_id", c.UserID.String()).Msg("failed to set write deadline for ping")
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
