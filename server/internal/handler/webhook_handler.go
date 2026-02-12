package handler

import (
	"net/http"

	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/phone"
	"github.com/otoritech/chatat/pkg/response"
)

// WebhookHandler handles incoming webhooks.
type WebhookHandler struct {
	reverseOTP service.ReverseOTPService
}

// NewWebhookHandler creates a new webhook handler.
func NewWebhookHandler(reverseOTP service.ReverseOTPService) *WebhookHandler {
	return &WebhookHandler{reverseOTP: reverseOTP}
}

type whatsappWebhookRequest struct {
	From    string `json:"from"`
	Message string `json:"message"`
}

// HandleWhatsApp handles incoming WhatsApp webhook messages.
func (h *WebhookHandler) HandleWhatsApp(w http.ResponseWriter, r *http.Request) {
	var req whatsappWebhookRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	// Normalize sender phone
	normalized, err := phone.Normalize(req.From, "")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid phone number"))
		return
	}

	if err := h.reverseOTP.HandleIncomingMessage(r.Context(), normalized, req.Message); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, map[string]string{"status": "received"})
}
