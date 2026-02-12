package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/otoritech/chatat/pkg/apperror"
)

// ReverseOTPService handles reverse OTP via WhatsApp.
type ReverseOTPService interface {
	InitSession(ctx context.Context, phone string) (*ReverseOTPSession, error)
	CheckVerification(ctx context.Context, sessionID string) (*VerificationResult, error)
	HandleIncomingMessage(ctx context.Context, senderPhone string, messageBody string) error
}

// ReverseOTPSession represents an active reverse OTP session.
type ReverseOTPSession struct {
	SessionID      string    `json:"sessionId"`
	TargetWANumber string    `json:"targetWANumber"`
	UniqueCode     string    `json:"uniqueCode"`
	ExpiresAt      time.Time `json:"expiresAt"`
}

// VerificationResult represents the result of checking a reverse OTP session.
type VerificationResult struct {
	Status string `json:"status"` // "pending", "verified", "expired"
	Phone  string `json:"phone,omitempty"`
}

type reverseOTPData struct {
	Phone    string `json:"phone"`
	Code     string `json:"code"`
	Verified bool   `json:"verified"`
}

type reverseOTPService struct {
	redis      *redis.Client
	waProvider WhatsAppProvider
	ttl        time.Duration
}

// NewReverseOTPService creates a new reverse OTP service.
func NewReverseOTPService(redisClient *redis.Client, waProvider WhatsAppProvider, ttl time.Duration) ReverseOTPService {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &reverseOTPService{
		redis:      redisClient,
		waProvider: waProvider,
		ttl:        ttl,
	}
}

func (s *reverseOTPService) InitSession(ctx context.Context, phone string) (*ReverseOTPSession, error) {
	// Check cooldown
	cooldownKey := fmt.Sprintf("rotp_cooldown:%s", phone)
	exists, err := s.redis.Exists(ctx, cooldownKey).Result()
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if exists > 0 {
		return nil, apperror.BadRequest("please wait before requesting another session")
	}

	sessionID := uuid.New().String()
	code, err := generateAlphanumericCode(6)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	expiresAt := time.Now().Add(s.ttl)

	data := reverseOTPData{
		Phone:    phone,
		Code:     code,
		Verified: false,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	// Store session
	sessionKey := fmt.Sprintf("reverse_otp:%s", sessionID)
	if err := s.redis.Set(ctx, sessionKey, jsonData, s.ttl).Err(); err != nil {
		return nil, apperror.Internal(err)
	}

	// Store reverse lookup: phone -> sessionID (for incoming message matching)
	lookupKey := fmt.Sprintf("rotp_phone:%s", phone)
	if err := s.redis.Set(ctx, lookupKey, sessionID, s.ttl).Err(); err != nil {
		return nil, apperror.Internal(err)
	}

	// Set cooldown
	if err := s.redis.Set(ctx, cooldownKey, "1", 60*time.Second).Err(); err != nil {
		return nil, apperror.Internal(err)
	}

	return &ReverseOTPSession{
		SessionID:      sessionID,
		TargetWANumber: s.waProvider.GetBusinessNumber(),
		UniqueCode:     code,
		ExpiresAt:      expiresAt,
	}, nil
}

func (s *reverseOTPService) CheckVerification(ctx context.Context, sessionID string) (*VerificationResult, error) {
	sessionKey := fmt.Sprintf("reverse_otp:%s", sessionID)
	raw, err := s.redis.Get(ctx, sessionKey).Bytes()
	if err == redis.Nil {
		return &VerificationResult{Status: "expired"}, nil
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}

	var data reverseOTPData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, apperror.Internal(err)
	}

	if data.Verified {
		return &VerificationResult{Status: "verified", Phone: data.Phone}, nil
	}

	return &VerificationResult{Status: "pending"}, nil
}

func (s *reverseOTPService) HandleIncomingMessage(ctx context.Context, senderPhone string, messageBody string) error {
	// Find session by sender phone
	lookupKey := fmt.Sprintf("rotp_phone:%s", senderPhone)
	sessionID, err := s.redis.Get(ctx, lookupKey).Result()
	if err == redis.Nil {
		return nil // No active session for this phone, ignore
	}
	if err != nil {
		return apperror.Internal(err)
	}

	sessionKey := fmt.Sprintf("reverse_otp:%s", sessionID)
	raw, err := s.redis.Get(ctx, sessionKey).Bytes()
	if err != nil {
		return nil // Session expired
	}

	var data reverseOTPData
	if err := json.Unmarshal(raw, &data); err != nil {
		return apperror.Internal(err)
	}

	// Check if code matches (case-insensitive)
	if len(messageBody) >= len(data.Code) {
		// Extract first N characters as the code
		receivedCode := messageBody[:len(data.Code)]
		if equalFold(receivedCode, data.Code) {
			data.Verified = true
			jsonData, _ := json.Marshal(data)
			ttl, _ := s.redis.TTL(ctx, sessionKey).Result()
			if ttl > 0 {
				_ = s.redis.Set(ctx, sessionKey, jsonData, ttl).Err()
			}
		}
	}

	return nil
}

func generateAlphanumericCode(length int) (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		code[i] = chars[n.Int64()]
	}
	return string(code), nil
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if ca >= 'a' && ca <= 'z' {
			ca -= 32
		}
		if cb >= 'a' && cb <= 'z' {
			cb -= 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
