package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/handler"
	"github.com/otoritech/chatat/internal/service"
)

func TestAuthHandler_SendOTP_InvalidPhone(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	sms := &mockSMSProvider{}
	wa := &mockWAProvider{businessNumber: "+628001234567"}
	otpSvc := service.NewOTPService(client, sms, service.DefaultOTPConfig())
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	tokenSvc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret"))
	sessionSvc := service.NewSessionService(client, tokenSvc, 0)

	h := handler.NewAuthHandler(otpSvc, reverseOTPSvc, tokenSvc, sessionSvc, nil)

	body, _ := json.Marshal(map[string]string{"phone": "invalid"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/otp/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SendOTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_SendOTP_ValidPhone(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	sms := &mockSMSProvider{}
	wa := &mockWAProvider{businessNumber: "+628001234567"}
	otpSvc := service.NewOTPService(client, sms, service.DefaultOTPConfig())
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	tokenSvc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret"))
	sessionSvc := service.NewSessionService(client, tokenSvc, 0)

	h := handler.NewAuthHandler(otpSvc, reverseOTPSvc, tokenSvc, sessionSvc, nil)

	body, _ := json.Marshal(map[string]string{"phone": "+6281234567890"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/otp/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SendOTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
}

func TestAuthHandler_SendOTP_InvalidBody(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	sms := &mockSMSProvider{}
	wa := &mockWAProvider{businessNumber: "+628001234567"}
	otpSvc := service.NewOTPService(client, sms, service.DefaultOTPConfig())
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	tokenSvc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret"))
	sessionSvc := service.NewSessionService(client, tokenSvc, 0)

	h := handler.NewAuthHandler(otpSvc, reverseOTPSvc, tokenSvc, sessionSvc, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/otp/send", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SendOTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
