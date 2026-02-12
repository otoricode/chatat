package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/otoritech/chatat/pkg/apperror"
)

// TokenService handles JWT token generation, validation, and revocation.
type TokenService interface {
	Generate(ctx context.Context, userID uuid.UUID) (*TokenPair, error)
	Validate(tokenString string) (*Claims, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)
	Revoke(ctx context.Context, accessToken string, refreshToken string) error
}

// TokenPair holds access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

// Claims represents JWT claims.
type Claims struct {
	UserID uuid.UUID `json:"userId"`
	jwt.RegisteredClaims
}

// TokenConfig holds token configuration.
type TokenConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// DefaultTokenConfig returns sensible defaults.
func DefaultTokenConfig(secret string) TokenConfig {
	return TokenConfig{
		Secret:          secret,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 30 * 24 * time.Hour, // 30 days
	}
}

type tokenService struct {
	redis  *redis.Client
	config TokenConfig
}

// NewTokenService creates a new token service.
func NewTokenService(redisClient *redis.Client, config TokenConfig) TokenService {
	return &tokenService{
		redis:  redisClient,
		config: config,
	}
}

func (s *tokenService) Generate(ctx context.Context, userID uuid.UUID) (*TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(s.config.AccessTokenTTL)
	refreshExp := now.Add(s.config.RefreshTokenTTL)

	// Generate access token
	accessClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExp),
			ID:        uuid.New().String(),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.config.Secret))
	if err != nil {
		return nil, apperror.Internal(err)
	}

	// Generate refresh token
	refreshTokenID := uuid.New().String()
	refreshClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExp),
			ID:        refreshTokenID,
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.config.Secret))
	if err != nil {
		return nil, apperror.Internal(err)
	}

	// Store refresh token in Redis
	refreshKey := fmt.Sprintf("refresh:%s", refreshTokenID)
	if err := s.redis.Set(ctx, refreshKey, userID.String(), s.config.RefreshTokenTTL).Err(); err != nil {
		return nil, apperror.Internal(err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp.Unix(),
	}, nil
}

func (s *tokenService) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.config.Secret), nil
	})
	if err != nil {
		return nil, apperror.Unauthorized("invalid or expired token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apperror.Unauthorized("invalid token claims")
	}

	return claims, nil
}

func (s *tokenService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.Validate(refreshToken)
	if err != nil {
		return nil, apperror.Unauthorized("invalid refresh token")
	}

	// Check if refresh token exists in Redis
	refreshKey := fmt.Sprintf("refresh:%s", claims.ID)
	storedUserID, err := s.redis.Get(ctx, refreshKey).Result()
	if err == redis.Nil {
		return nil, apperror.Unauthorized("refresh token revoked or expired")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}

	if storedUserID != claims.UserID.String() {
		return nil, apperror.Unauthorized("refresh token mismatch")
	}

	// Delete old refresh token
	_ = s.redis.Del(ctx, refreshKey).Err()

	// Generate new token pair
	return s.Generate(ctx, claims.UserID)
}

func (s *tokenService) Revoke(ctx context.Context, accessToken string, refreshToken string) error {
	// Blacklist access token
	if accessToken != "" {
		accessClaims, err := s.Validate(accessToken)
		if err == nil && accessClaims.ExpiresAt != nil {
			ttl := time.Until(accessClaims.ExpiresAt.Time)
			if ttl > 0 {
				blacklistKey := fmt.Sprintf("blacklist:%s", accessClaims.ID)
				_ = s.redis.Set(ctx, blacklistKey, "1", ttl).Err()
			}
		}
	}

	// Delete refresh token
	if refreshToken != "" {
		refreshClaims, err := s.Validate(refreshToken)
		if err == nil {
			refreshKey := fmt.Sprintf("refresh:%s", refreshClaims.ID)
			_ = s.redis.Del(ctx, refreshKey).Err()
		}
	}

	return nil
}
