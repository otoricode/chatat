// Package cache provides a Redis-backed caching layer for frequently accessed data.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/otoritech/chatat/internal/model"
)

// Default TTLs for cached entities.
const (
	UserTTL       = 5 * time.Minute
	ChatListTTL   = 1 * time.Minute
	OnlineUserTTL = 30 * time.Second
)

// Service provides cache get/set/invalidate operations backed by Redis.
type Service struct {
	redis *redis.Client
}

// NewService creates a new cache service.
func NewService(redisClient *redis.Client) *Service {
	return &Service{redis: redisClient}
}

// --- User Profile Cache ---

func userKey(userID uuid.UUID) string {
	return "cache:user:" + userID.String()
}

// GetUser retrieves a cached user profile.
// Returns nil, nil if not found in cache.
func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	data, err := s.redis.Get(ctx, userKey(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("cache: failed to get user")
		return nil, nil // fail-open
	}

	var user model.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, nil
	}
	return &user, nil
}

// SetUser stores a user profile in cache.
func (s *Service) SetUser(ctx context.Context, user *model.User) {
	data, err := json.Marshal(user)
	if err != nil {
		return
	}
	if err := s.redis.Set(ctx, userKey(user.ID), data, UserTTL).Err(); err != nil {
		log.Warn().Err(err).Str("user_id", user.ID.String()).Msg("cache: failed to set user")
	}
}

// InvalidateUser removes a user from cache.
func (s *Service) InvalidateUser(ctx context.Context, userID uuid.UUID) {
	if err := s.redis.Del(ctx, userKey(userID)).Err(); err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("cache: failed to invalidate user")
	}
}

// --- Online Status Cache ---

func onlineKey(userID uuid.UUID) string {
	return "cache:online:" + userID.String()
}

// SetOnline marks a user as online with a TTL.
func (s *Service) SetOnline(ctx context.Context, userID uuid.UUID) {
	if err := s.redis.Set(ctx, onlineKey(userID), "1", OnlineUserTTL).Err(); err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("cache: failed to set online")
	}
}

// IsOnline checks if a user is marked as online in cache.
func (s *Service) IsOnline(ctx context.Context, userID uuid.UUID) bool {
	val, err := s.redis.Exists(ctx, onlineKey(userID)).Result()
	if err != nil {
		return false
	}
	return val > 0
}

// SetOffline removes the online flag for a user.
func (s *Service) SetOffline(ctx context.Context, userID uuid.UUID) {
	if err := s.redis.Del(ctx, onlineKey(userID)).Err(); err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("cache: failed to set offline")
	}
}

// --- Chat List Cache ---

func chatListKey(userID uuid.UUID) string {
	return "cache:chatlist:" + userID.String()
}

// GetChatList retrieves cached chat list JSON for a user.
// Returns nil if not found.
func (s *Service) GetChatList(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	data, err := s.redis.Get(ctx, chatListKey(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("cache: failed to get chat list")
		return nil, nil
	}
	return data, nil
}

// SetChatList stores chat list JSON in cache.
func (s *Service) SetChatList(ctx context.Context, userID uuid.UUID, data []byte) {
	if err := s.redis.Set(ctx, chatListKey(userID), data, ChatListTTL).Err(); err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("cache: failed to set chat list")
	}
}

// InvalidateChatList removes a user's chat list from cache.
func (s *Service) InvalidateChatList(ctx context.Context, userID uuid.UUID) {
	if err := s.redis.Del(ctx, chatListKey(userID)).Err(); err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("cache: failed to invalidate chat list")
	}
}

// --- Generic helpers ---

// Get retrieves a cached value by key. Returns nil if not found.
func (s *Service) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := s.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache get %s: %w", key, err)
	}
	return data, nil
}

// Set stores a value in cache with the given TTL.
func (s *Service) Set(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	return s.redis.Set(ctx, key, data, ttl).Err()
}

// Del removes a key from cache.
func (s *Service) Del(ctx context.Context, key string) error {
	return s.redis.Del(ctx, key).Err()
}
