package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/ws"
)

const (
	onlineKeyPrefix = "online:"
	onlineTTL       = 5 * time.Minute
)

// onlineStatusPayload is the payload for online_status WS events.
type onlineStatusPayload struct {
	UserID   uuid.UUID `json:"userId"`
	IsOnline bool      `json:"isOnline"`
	LastSeen time.Time `json:"lastSeen"`
}

// StatusNotifier broadcasts online/offline status to contacts via WebSocket.
type StatusNotifier struct {
	hub         *ws.Hub
	contactRepo repository.ContactRepository
	userRepo    repository.UserRepository
	redis       *redis.Client
}

// NewStatusNotifier creates a StatusNotifier and wires Hub event callbacks.
func NewStatusNotifier(
	hub *ws.Hub,
	contactRepo repository.ContactRepository,
	userRepo repository.UserRepository,
	redisClient *redis.Client,
) *StatusNotifier {
	sn := &StatusNotifier{
		hub:         hub,
		contactRepo: contactRepo,
		userRepo:    userRepo,
		redis:       redisClient,
	}

	hub.SetEventCallbacks(sn.handleConnect, sn.handleDisconnect)
	return sn
}

func (sn *StatusNotifier) handleConnect(userID uuid.UUID) {
	ctx := context.Background()

	// Track online status in Redis
	if sn.redis != nil {
		key := onlineKeyPrefix + userID.String()
		if err := sn.redis.Set(ctx, key, "1", onlineTTL).Err(); err != nil {
			log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to set online status in redis")
		}
	}

	sn.broadcastStatus(ctx, userID, true)
}

func (sn *StatusNotifier) handleDisconnect(userID uuid.UUID) {
	ctx := context.Background()

	// Remove online status from Redis
	if sn.redis != nil {
		key := onlineKeyPrefix + userID.String()
		if err := sn.redis.Del(ctx, key).Err(); err != nil {
			log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to remove online status from redis")
		}
	}

	// Update last_seen
	if err := sn.userRepo.UpdateLastSeen(ctx, userID); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to update last_seen on disconnect")
	}

	sn.broadcastStatus(ctx, userID, false)
}

func (sn *StatusNotifier) broadcastStatus(ctx context.Context, userID uuid.UUID, isOnline bool) {
	// Find users who have this user as a contact
	observers, err := sn.contactRepo.FindContactsOf(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to find contact observers")
		return
	}

	if len(observers) == 0 {
		return
	}

	payload := onlineStatusPayload{
		UserID:   userID,
		IsOnline: isOnline,
		LastSeen: time.Now(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal online_status payload")
		return
	}

	msg := ws.WSMessage{
		Type:    ws.WSTypeOnlineStatus,
		Payload: payloadBytes,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal online_status message")
		return
	}

	for _, observerID := range observers {
		sn.hub.SendToUser(observerID, msgBytes)
	}

	log.Debug().
		Str("user_id", userID.String()).
		Bool("is_online", isOnline).
		Int("observers", len(observers)).
		Msg("broadcasted online status")
}

// RefreshOnlineStatus refreshes the Redis TTL for a connected user.
// Call this on WebSocket pong or authenticated API requests.
func (sn *StatusNotifier) RefreshOnlineStatus(ctx context.Context, userID uuid.UUID) error {
	if sn.redis == nil {
		return nil
	}
	key := onlineKeyPrefix + userID.String()
	if err := sn.redis.Expire(ctx, key, onlineTTL).Err(); err != nil {
		return fmt.Errorf("refresh online status: %w", err)
	}
	return nil
}
