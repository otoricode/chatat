package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/otoritech/chatat/pkg/apperror"
)

// SessionService manages one-device-per-user sessions.
type SessionService interface {
	Register(ctx context.Context, userID uuid.UUID, deviceID string, refreshToken string) error
	Validate(ctx context.Context, userID uuid.UUID, deviceID string) error
	Invalidate(ctx context.Context, userID uuid.UUID) error
}

type deviceInfo struct {
	DeviceID     string `json:"deviceId"`
	RefreshToken string `json:"refreshToken"`
}

type sessionService struct {
	redis        *redis.Client
	tokenService TokenService
	sessionTTL   time.Duration
}

// NewSessionService creates a new session service.
func NewSessionService(redisClient *redis.Client, tokenService TokenService, ttl time.Duration) SessionService {
	if ttl == 0 {
		ttl = 30 * 24 * time.Hour // 30 days
	}
	return &sessionService{
		redis:        redisClient,
		tokenService: tokenService,
		sessionTTL:   ttl,
	}
}

func (s *sessionService) Register(ctx context.Context, userID uuid.UUID, deviceID string, refreshToken string) error {
	key := fmt.Sprintf("device:%s", userID.String())

	// Check for existing device
	raw, err := s.redis.Get(ctx, key).Bytes()
	if err == nil {
		var existing deviceInfo
		if err := json.Unmarshal(raw, &existing); err == nil {
			if existing.DeviceID != deviceID {
				// Different device — revoke old session
				log.Info().
					Str("user_id", userID.String()).
					Str("old_device", existing.DeviceID).
					Str("new_device", deviceID).
					Msg("revoking previous device session")

				_ = s.tokenService.Revoke(ctx, "", existing.RefreshToken)
			}
		}
	} else if err != redis.Nil {
		return apperror.Internal(err)
	}

	// Store new device
	data := deviceInfo{
		DeviceID:     deviceID,
		RefreshToken: refreshToken,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return apperror.Internal(err)
	}

	if err := s.redis.Set(ctx, key, jsonData, s.sessionTTL).Err(); err != nil {
		return apperror.Internal(err)
	}

	return nil
}

func (s *sessionService) Validate(ctx context.Context, userID uuid.UUID, deviceID string) error {
	key := fmt.Sprintf("device:%s", userID.String())
	raw, err := s.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return apperror.Unauthorized("no active session")
	}
	if err != nil {
		return apperror.Internal(err)
	}

	var info deviceInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return apperror.Internal(err)
	}

	if info.DeviceID != deviceID {
		return apperror.Unauthorized("session active on another device")
	}

	return nil
}

func (s *sessionService) Invalidate(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf("device:%s", userID.String())
	return s.redis.Del(ctx, key).Err()
}
