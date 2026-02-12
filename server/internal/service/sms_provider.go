package service

import (
	"github.com/rs/zerolog/log"
)

// SMSProvider defines the interface for sending SMS messages.
type SMSProvider interface {
	Send(phone string, message string) error
}

// LogSMSProvider is a development-only SMS provider that logs OTP codes.
type LogSMSProvider struct{}

// NewLogSMSProvider creates a new log-based SMS provider for development.
func NewLogSMSProvider() *LogSMSProvider {
	return &LogSMSProvider{}
}

// Send logs the SMS message instead of sending it.
func (p *LogSMSProvider) Send(phone string, message string) error {
	log.Info().
		Str("phone", phone).
		Str("message", message).
		Msg("[DEV] SMS would be sent")
	return nil
}
