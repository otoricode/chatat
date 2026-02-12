package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/otoritech/chatat/pkg/apperror"
)

// OTPService handles OTP generation, storage, and verification.
type OTPService interface {
	Generate(ctx context.Context, phone string) (string, error)
	Verify(ctx context.Context, phone string, code string) error
}

// OTPConfig holds OTP configuration.
type OTPConfig struct {
	Length         int
	TTL            time.Duration
	MaxAttempts    int
	CooldownPeriod time.Duration
	MaxPerDay      int
}

// DefaultOTPConfig returns sensible defaults.
func DefaultOTPConfig() OTPConfig {
	return OTPConfig{
		Length:         6,
		TTL:            5 * time.Minute,
		MaxAttempts:    3,
		CooldownPeriod: 60 * time.Second,
		MaxPerDay:      5,
	}
}

type otpData struct {
	Code     string `json:"code"`
	Attempts int    `json:"attempts"`
}

type otpService struct {
	redis  *redis.Client
	sms    SMSProvider
	config OTPConfig
}

// NewOTPService creates a new OTP service.
func NewOTPService(redisClient *redis.Client, smsProvider SMSProvider, config OTPConfig) OTPService {
	return &otpService{
		redis:  redisClient,
		sms:    smsProvider,
		config: config,
	}
}

func (s *otpService) Generate(ctx context.Context, phone string) (string, error) {
	// Check cooldown (1 per 60 seconds)
	cooldownKey := fmt.Sprintf("otp_cooldown:%s", phone)
	exists, err := s.redis.Exists(ctx, cooldownKey).Result()
	if err != nil {
		return "", apperror.Internal(err)
	}
	if exists > 0 {
		return "", apperror.BadRequest("please wait before requesting another OTP")
	}

	// Check daily limit
	dailyKey := fmt.Sprintf("otp_daily:%s", phone)
	dailyCount, err := s.redis.Get(ctx, dailyKey).Int()
	if err != nil && err != redis.Nil {
		return "", apperror.Internal(err)
	}
	if dailyCount >= s.config.MaxPerDay {
		return "", apperror.RateLimited()
	}

	// Generate OTP code
	code, err := generateNumericCode(s.config.Length)
	if err != nil {
		return "", apperror.Internal(err)
	}

	// Store OTP in Redis
	otpKey := fmt.Sprintf("otp:%s", phone)
	data := otpData{Code: code, Attempts: 0}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", apperror.Internal(err)
	}
	if err := s.redis.Set(ctx, otpKey, jsonData, s.config.TTL).Err(); err != nil {
		return "", apperror.Internal(err)
	}

	// Set cooldown
	if err := s.redis.Set(ctx, cooldownKey, "1", s.config.CooldownPeriod).Err(); err != nil {
		return "", apperror.Internal(err)
	}

	// Increment daily counter
	pipe := s.redis.Pipeline()
	pipe.Incr(ctx, dailyKey)
	// Set TTL to end of day if first request
	if dailyCount == 0 {
		pipe.Expire(ctx, dailyKey, 24*time.Hour)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return "", apperror.Internal(err)
	}

	// Send SMS
	message := fmt.Sprintf("Your Chatat verification code is: %s", code)
	if err := s.sms.Send(phone, message); err != nil {
		return "", apperror.Internal(err)
	}

	return code, nil
}

func (s *otpService) Verify(ctx context.Context, phone string, code string) error {
	otpKey := fmt.Sprintf("otp:%s", phone)

	raw, err := s.redis.Get(ctx, otpKey).Bytes()
	if err == redis.Nil {
		return apperror.InvalidOTP()
	}
	if err != nil {
		return apperror.Internal(err)
	}

	var data otpData
	if err := json.Unmarshal(raw, &data); err != nil {
		return apperror.Internal(err)
	}

	// Check max attempts
	if data.Attempts >= s.config.MaxAttempts {
		_ = s.redis.Del(ctx, otpKey).Err()
		return apperror.InvalidOTP()
	}

	// Constant-time comparison
	if subtle.ConstantTimeCompare([]byte(data.Code), []byte(code)) != 1 {
		// Increment attempts
		data.Attempts++
		jsonData, _ := json.Marshal(data)
		ttl, _ := s.redis.TTL(ctx, otpKey).Result()
		if ttl > 0 {
			_ = s.redis.Set(ctx, otpKey, jsonData, ttl).Err()
		}
		return apperror.InvalidOTP()
	}

	// OTP verified, delete it
	_ = s.redis.Del(ctx, otpKey).Err()
	return nil
}

func generateNumericCode(length int) (string, error) {
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code[i] = byte('0') + byte(n.Int64())
	}
	return string(code), nil
}
