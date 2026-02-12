package service

import (
	"github.com/rs/zerolog/log"
)

// WhatsAppProvider defines the interface for WhatsApp Business API.
type WhatsAppProvider interface {
	GetBusinessNumber() string
}

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
