package handler

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

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
	hub       *ws.Hub
	jwtSecret string
}

// NewWSHandler creates a new WebSocket handler.
func NewWSHandler(hub *ws.Hub, jwtSecret string) *WSHandler {
	return &WSHandler{hub: hub, jwtSecret: jwtSecret}
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
	h.hub.RegisterClient(client)

	// Join user's personal notification room
	h.hub.JoinRoom(client, "user:"+userID.String())

	go client.WritePump()
	go client.ReadPump()
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
