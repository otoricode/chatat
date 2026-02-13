package service

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"

	"github.com/otoritech/chatat/internal/model"
)

// FCMPushSender sends push notifications via Firebase Cloud Messaging.
type FCMPushSender struct {
	client *messaging.Client
}

// NewFCMPushSender creates a new FCM push sender.
// credentialsFile is the path to the Firebase service account JSON file.
func NewFCMPushSender(ctx context.Context, credentialsFile string) (PushSender, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("get Firebase messaging client: %w", err)
	}

	return &FCMPushSender{client: client}, nil
}

func (s *FCMPushSender) Send(ctx context.Context, token string, notif model.Notification) error {
	msg := buildFCMMessage(token, notif)
	_, err := s.client.Send(ctx, msg)
	if err != nil {
		if messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err) {
			log.Warn().Str("token", token[:min(len(token), 20)]+"...").Msg("invalid FCM token, should remove")
			return fmt.Errorf("invalid token: %w", err)
		}
		return fmt.Errorf("send FCM message: %w", err)
	}
	return nil
}

func (s *FCMPushSender) SendMulti(ctx context.Context, tokens []string, notif model.Notification) error {
	if len(tokens) == 0 {
		return nil
	}

	// FCM SendEachForMulticast supports up to 500 tokens
	const batchSize = 500
	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		batch := tokens[i:end]

		msg := buildFCMMulticastMessage(batch, notif)
		resp, err := s.client.SendEachForMulticast(ctx, msg)
		if err != nil {
			log.Warn().Err(err).Int("batch", i/batchSize).Msg("FCM batch send error")
			continue
		}

		if resp.FailureCount > 0 {
			for idx, sendResp := range resp.Responses {
				if sendResp.Error != nil {
					if messaging.IsUnregistered(sendResp.Error) || messaging.IsInvalidArgument(sendResp.Error) {
						log.Warn().
							Str("token", batch[idx][:min(len(batch[idx]), 20)]+"...").
							Msg("invalid FCM token in batch, should remove")
					}
				}
			}
		}
	}

	return nil
}

func buildFCMMessage(token string, notif model.Notification) *messaging.Message {
	badge := notif.Badge
	sound := notif.Sound
	if sound == "" {
		sound = "default"
	}

	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: notif.Title,
			Body:  notif.Body,
		},
		Data: notif.Data,
		Android: &messaging.AndroidConfig{
			Priority: mapPriority(notif.Priority),
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Badge: &badge,
					Sound: sound,
				},
			},
		},
	}
	return msg
}

func buildFCMMulticastMessage(tokens []string, notif model.Notification) *messaging.MulticastMessage {
	badge := notif.Badge
	sound := notif.Sound
	if sound == "" {
		sound = "default"
	}

	return &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: notif.Title,
			Body:  notif.Body,
		},
		Data: notif.Data,
		Android: &messaging.AndroidConfig{
			Priority: mapPriority(notif.Priority),
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Badge: &badge,
					Sound: sound,
				},
			},
		},
	}
}

func mapPriority(p string) string {
	if p == "high" {
		return "high"
	}
	return "normal"
}
