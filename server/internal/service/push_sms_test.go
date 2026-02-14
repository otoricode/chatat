package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/internal/model"
)

func TestLogPushSender_Send(t *testing.T) {
	sender := NewLogPushSender()
	err := sender.Send(context.Background(), "device-token-1234567890", model.Notification{
		Type:  "message",
		Title: "Test",
		Body:  "Hello",
	})
	assert.NoError(t, err)
}

func TestLogPushSender_SendMulti(t *testing.T) {
	sender := NewLogPushSender()
	err := sender.SendMulti(context.Background(), []string{"tok1", "tok2"}, model.Notification{
		Type:  "message",
		Title: "Test",
		Body:  "Hello",
	})
	assert.NoError(t, err)
}

func TestLogSMSProvider_Send(t *testing.T) {
	p := NewLogSMSProvider()
	err := p.Send("+6281234567890", "Your OTP is 123456")
	assert.NoError(t, err)
}

func TestLogWhatsAppProvider_GetBusinessNumber(t *testing.T) {
	p := NewLogWhatsAppProvider("+6281000000000")
	assert.Equal(t, "+6281000000000", p.GetBusinessNumber())
}

func TestLogWhatsAppProvider_SendMessage(t *testing.T) {
	p := NewLogWhatsAppProvider("+6281000000000")
	err := p.SendMessage(context.Background(), "+6289999999999", "Hello from test")
	assert.NoError(t, err)
}

func TestGOWAProvider_GetBusinessNumber(t *testing.T) {
	p := NewGOWAProvider("http://localhost:3000", "+6281000000000")
	assert.Equal(t, "+6281000000000", p.GetBusinessNumber())
}

func TestGOWAProvider_SendMessage_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/send/message", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload map[string]string
		_ = json.NewDecoder(r.Body).Decode(&payload)
		assert.Equal(t, "6289999999999@s.whatsapp.net", payload["phone"])
		assert.Equal(t, "Hello WA", payload["message"])

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":"200","message":"ok"}`))
	}))
	defer ts.Close()

	p := NewGOWAProvider(ts.URL, "+6281000000000")
	err := p.SendMessage(context.Background(), "+6289999999999", "Hello WA")
	assert.NoError(t, err)
}

func TestGOWAProvider_SendMessage_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":"500","message":"internal error"}`))
	}))
	defer ts.Close()

	p := NewGOWAProvider(ts.URL, "+6281000000000")
	err := p.SendMessage(context.Background(), "+62899", "fail")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestGOWAProvider_SendMessage_Unreachable(t *testing.T) {
	p := NewGOWAProvider("http://127.0.0.1:19999", "+6281000000000")
	err := p.SendMessage(context.Background(), "+62899", "fail")
	assert.Error(t, err)
}
