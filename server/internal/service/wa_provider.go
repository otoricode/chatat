package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// WhatsAppProvider defines the interface for WhatsApp messaging.
type WhatsAppProvider interface {
	GetBusinessNumber() string
	SendMessage(ctx context.Context, phone string, message string) error
}

// -- Log (dev) provider --

// LogWhatsAppProvider is a development-only WA provider that logs messages.
type LogWhatsAppProvider struct {
	businessNumber string
}

// NewLogWhatsAppProvider creates a new log-based WA provider for development.
func NewLogWhatsAppProvider(businessNumber string) *LogWhatsAppProvider {
	return &LogWhatsAppProvider{businessNumber: businessNumber}
}

// GetBusinessNumber returns the WhatsApp business phone number.
func (p *LogWhatsAppProvider) GetBusinessNumber() string {
	log.Debug().Str("number", p.businessNumber).Msg("[DEV] WA business number")
	return p.businessNumber
}

// SendMessage logs the message instead of sending it (development mode).
func (p *LogWhatsAppProvider) SendMessage(_ context.Context, phone string, message string) error {
	log.Info().
		Str("to", phone).
		Str("message", message).
		Msg("[DEV] WA message (not sent)")
	return nil
}

// -- GOWA provider (real) --

// gowaResponse is the standard response from GOWA API.
type gowaResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GOWAProvider is a WhatsApp provider using go-whatsapp-web-multidevice REST API.
type GOWAProvider struct {
	baseURL        string
	businessNumber string
	httpClient     *http.Client
}

// NewGOWAProvider creates a new GOWA-based WhatsApp provider.
func NewGOWAProvider(baseURL string, businessNumber string) *GOWAProvider {
	return &GOWAProvider{
		baseURL:        baseURL,
		businessNumber: businessNumber,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetBusinessNumber returns the WhatsApp business phone number.
func (p *GOWAProvider) GetBusinessNumber() string {
	return p.businessNumber
}

// SendMessage sends a text message via GOWA REST API.
func (p *GOWAProvider) SendMessage(ctx context.Context, phoneNumber string, message string) error {
	// GOWA expects phone in format: 6289685028129@s.whatsapp.net
	// We receive E.164 format like +6289685028129, strip the + and add suffix
	jid := phoneNumber
	if len(jid) > 0 && jid[0] == '+' {
		jid = jid[1:]
	}
	jid = jid + "@s.whatsapp.net"

	payload := map[string]string{
		"phone":   jid,
		"message": message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/send/message", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error().
			Int("status", resp.StatusCode).
			Str("body", string(respBody)).
			Msg("GOWA send message failed")
		return fmt.Errorf("GOWA API error: status %d", resp.StatusCode)
	}

	var gowaResp gowaResponse
	if err := json.Unmarshal(respBody, &gowaResp); err == nil {
		log.Debug().
			Str("code", gowaResp.Code).
			Str("to", phoneNumber).
			Msg("WA message sent via GOWA")
	}

	return nil
}
