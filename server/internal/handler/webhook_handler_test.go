package handler_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/internal/handler"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
)

func makeGOWAPayload(from string, body string) []byte {
	payload := map[string]interface{}{
		"event":     "message",
		"device_id": "628001234567@s.whatsapp.net",
		"payload": map[string]interface{}{
			"id":         "3EB0C127D7BACC83D6A1",
			"chat_id":    from,
			"from":       from,
			"from_name":  "Test User",
			"body":       body,
			"is_from_me": false,
			"timestamp":  "2025-01-15T10:30:00Z",
		},
	}
	b, _ := json.Marshal(payload)
	return b
}

func signPayload(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestWebhookHandler_HandleWhatsApp_InvalidBody(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	wa := &mockWAProvider{businessNumber: "+628001234567"}
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	h := handler.NewWebhookHandler(reverseOTPSvc, "") // empty secret = no verification

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookHandler_HandleWhatsApp_InvalidJID(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	wa := &mockWAProvider{businessNumber: "+628001234567"}
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	h := handler.NewWebhookHandler(reverseOTPSvc, "")

	// From field with no valid phone
	body := makeGOWAPayload("@s.whatsapp.net", "ABC123")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	// Should return 200 with "ignored" since we gracefully skip invalid phones
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWebhookHandler_HandleWhatsApp_Success(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	wa := &mockWAProvider{businessNumber: "+628001234567"}
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	h := handler.NewWebhookHandler(reverseOTPSvc, "")

	body := makeGOWAPayload("6281234567890@s.whatsapp.net", "ABC123")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWebhookHandler_HandleWhatsApp_SkipsOwnMessages(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	wa := &mockWAProvider{businessNumber: "+628001234567"}
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	h := handler.NewWebhookHandler(reverseOTPSvc, "")

	payload := map[string]interface{}{
		"event":     "message",
		"device_id": "628001234567@s.whatsapp.net",
		"payload": map[string]interface{}{
			"from":       "628001234567@s.whatsapp.net",
			"body":       "Hello",
			"is_from_me": true,
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWebhookHandler_HandleWhatsApp_SignatureVerification(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	wa := &mockWAProvider{businessNumber: "+628001234567"}
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	secret := "test-secret"
	h := handler.NewWebhookHandler(reverseOTPSvc, secret)

	body := makeGOWAPayload("6281234567890@s.whatsapp.net", "ABC123")

	// Test with valid signature
	t.Run("valid signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature-256", signPayload(body, secret))
		w := httptest.NewRecorder()

		h.HandleWhatsApp(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test with invalid signature
	t.Run("invalid signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
		w := httptest.NewRecorder()

		h.HandleWhatsApp(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test with missing signature
	t.Run("missing signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.HandleWhatsApp(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test with signature missing sha256= prefix
	t.Run("no sha256 prefix", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature-256", "justahexstring")
		w := httptest.NewRecorder()

		h.HandleWhatsApp(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestWebhookHandler_HandleWhatsApp_NonMessageEvent(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	wa := &mockWAProvider{businessNumber: "+628001234567"}
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	h := handler.NewWebhookHandler(reverseOTPSvc, "")

	payload := map[string]interface{}{
		"event":     "status_update",
		"device_id": "628001234567@s.whatsapp.net",
		"payload":   map[string]interface{}{},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWebhookHandler_HandleWhatsApp_EmptyBody(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	wa := &mockWAProvider{businessNumber: "+628001234567"}
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	h := handler.NewWebhookHandler(reverseOTPSvc, "")

	payload := map[string]interface{}{
		"event":     "message",
		"device_id": "628001234567@s.whatsapp.net",
		"payload": map[string]interface{}{
			"from":       "6281234567890@s.whatsapp.net",
			"body":       "",
			"is_from_me": false,
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWebhookHandler_HandleWhatsApp_ServiceError(t *testing.T) {
	h := handler.NewWebhookHandler(&mockReverseOTPService{err: apperror.Internal(nil)}, "")

	body := makeGOWAPayload("6281234567890@s.whatsapp.net", "ABC123")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWebhookHandler_HandleWhatsApp_GenericError(t *testing.T) {
	h := handler.NewWebhookHandler(&mockReverseOTPService{err: errors.New("db")}, "")

	body := makeGOWAPayload("6281234567890@s.whatsapp.net", "ABC123")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWebhookHandler_HandleWhatsApp_EmptyJID(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	wa := &mockWAProvider{businessNumber: "+628001234567"}
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	h := handler.NewWebhookHandler(reverseOTPSvc, "")

	payload := map[string]interface{}{
		"event":     "message",
		"device_id": "628001234567@s.whatsapp.net",
		"payload": map[string]interface{}{
			"from":       "",
			"body":       "ABC123",
			"is_from_me": false,
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWebhookHandler_HandleWhatsApp_PhoneNormalizeError(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	wa := &mockWAProvider{businessNumber: "+628001234567"}
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	h := handler.NewWebhookHandler(reverseOTPSvc, "")

	// Send a phone that extracts but fails normalization
	body := makeGOWAPayload("abc@s.whatsapp.net", "ABC123")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
