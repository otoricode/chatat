package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/phone"
	"github.com/otoritech/chatat/pkg/response"
)

// WebhookHandler handles incoming webhooks.
type WebhookHandler struct {
	reverseOTP    service.ReverseOTPService
	webhookSecret string
}

// NewWebhookHandler creates a new webhook handler.
func NewWebhookHandler(reverseOTP service.ReverseOTPService, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		reverseOTP:    reverseOTP,
		webhookSecret: webhookSecret,
	}
}

// gowaWebhookPayload is the payload from GOWA webhook.
type gowaWebhookPayload struct {
	Event    string             `json:"event"`
	DeviceID string             `json:"device_id"`
	Payload  gowaMessagePayload `json:"payload"`
}

// gowaMessagePayload is the message payload inside GOWA webhook.
type gowaMessagePayload struct {
	ID        string `json:"id"`
	ChatID    string `json:"chat_id"`
	From      string `json:"from"`
	FromName  string `json:"from_name"`
	Body      string `json:"body"`
	IsFromMe  bool   `json:"is_from_me"`
	Timestamp string `json:"timestamp"`
}

// HandleWhatsApp handles incoming WhatsApp webhook messages from GOWA.
func (h *WebhookHandler) HandleWhatsApp(w http.ResponseWriter, r *http.Request) {
	// Read raw body for signature verification
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, apperror.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Verify HMAC signature if secret is configured
	if h.webhookSecret != "" {
		signature := r.Header.Get("X-Hub-Signature-256")
		if !h.verifySignature(rawBody, signature) {
			log.Warn().Msg("webhook signature verification failed")
			response.Error(w, apperror.Unauthorized("invalid webhook signature"))
			return
		}
	}

	// Parse GOWA webhook payload
	var payload gowaWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		response.Error(w, apperror.BadRequest("invalid webhook payload"))
		return
	}

	// Only process text message events
	if payload.Event != "message" {
		response.OK(w, map[string]string{"status": "ignored"})
		return
	}

	// Skip messages sent by us
	if payload.Payload.IsFromMe {
		response.OK(w, map[string]string{"status": "ignored"})
		return
	}

	// Skip if no text body (media messages etc.)
	if payload.Payload.Body == "" {
		response.OK(w, map[string]string{"status": "ignored"})
		return
	}

	// Extract phone from GOWA "from" field (format: 628xxx@s.whatsapp.net)
	senderPhone := extractPhoneFromJID(payload.Payload.From)
	if senderPhone == "" {
		log.Warn().Str("from", payload.Payload.From).Msg("could not extract phone from JID")
		response.OK(w, map[string]string{"status": "ignored"})
		return
	}

	// Normalize the extracted phone number
	normalized, err := phone.Normalize(senderPhone, "")
	if err != nil {
		log.Warn().Str("phone", senderPhone).Err(err).Msg("could not normalize sender phone")
		response.OK(w, map[string]string{"status": "ignored"})
		return
	}

	// Handle the incoming message for reverse OTP verification
	if err := h.reverseOTP.HandleIncomingMessage(r.Context(), normalized, payload.Payload.Body); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, map[string]string{"status": "received"})
}

// verifySignature verifies the HMAC SHA256 signature from GOWA webhook.
func (h *WebhookHandler) verifySignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	// Signature format: sha256={hex}
	sig := strings.TrimPrefix(signature, "sha256=")
	if sig == signature {
		return false // no prefix found
	}

	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedMAC), []byte(sig))
}

// extractPhoneFromJID extracts phone number from WhatsApp JID.
// Example: "628123456789@s.whatsapp.net" -> "+628123456789"
func extractPhoneFromJID(jid string) string {
	if jid == "" {
		return ""
	}
	parts := strings.Split(jid, "@")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	return "+" + parts[0]
}
