package service

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/otoritech/chatat/internal/model"
)

// LogPushSender logs push notifications instead of sending them.
// Used in development/testing environments.
type LogPushSender struct{}

// NewLogPushSender creates a new log-based push sender.
func NewLogPushSender() PushSender {
	return &LogPushSender{}
}

func (s *LogPushSender) Send(_ context.Context, token string, notif model.Notification) error {
	log.Info().
		Str("token", token[:min(len(token), 20)]+"...").
		Str("type", string(notif.Type)).
		Str("title", notif.Title).
		Str("body", notif.Body).
		Msg("[PUSH] notification sent (log mode)")
	return nil
}

func (s *LogPushSender) SendMulti(_ context.Context, tokens []string, notif model.Notification) error {
	log.Info().
		Int("recipients", len(tokens)).
		Str("type", string(notif.Type)).
		Str("title", notif.Title).
		Str("body", notif.Body).
		Msg("[PUSH] notification sent to multiple (log mode)")
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
