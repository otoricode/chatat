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

	"github.com/otoritech/chatat/internal/handler"
	"github.com/otoritech/chatat/internal/service"
)

func TestWebhookHandler_HandleWhatsApp_InvalidBody(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	wa := &mockWAProvider{businessNumber: "+628001234567"}
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	h := handler.NewWebhookHandler(reverseOTPSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookHandler_HandleWhatsApp_InvalidPhone(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	wa := &mockWAProvider{businessNumber: "+628001234567"}
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	h := handler.NewWebhookHandler(reverseOTPSvc)

	body, _ := json.Marshal(map[string]string{"from": "invalid", "message": "ABC123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookHandler_HandleWhatsApp_Success(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer func() { _ = client.Close() }()

	wa := &mockWAProvider{businessNumber: "+628001234567"}
	reverseOTPSvc := service.NewReverseOTPService(client, wa, 0)
	h := handler.NewWebhookHandler(reverseOTPSvc)

	body, _ := json.Marshal(map[string]string{"from": "+6281234567890", "message": "ABC123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/whatsapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWhatsApp(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
