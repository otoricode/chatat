package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
)

// PushSender abstracts the push notification delivery (FCM, Expo, etc.).
type PushSender interface {
	Send(ctx context.Context, token string, notif model.Notification) error
	SendMulti(ctx context.Context, tokens []string, notif model.Notification) error
}

// NotificationService defines operations for push notification management.
type NotificationService interface {
	RegisterDevice(ctx context.Context, userID uuid.UUID, token, platform string) error
	UnregisterDevice(ctx context.Context, userID uuid.UUID, token string) error
	SendToUser(ctx context.Context, userID uuid.UUID, notif model.Notification) error
	SendToUsers(ctx context.Context, userIDs []uuid.UUID, notif model.Notification) error
	SendToChat(ctx context.Context, chatID uuid.UUID, excludeUserID uuid.UUID, notif model.Notification) error
}

type notificationService struct {
	deviceRepo repository.DeviceTokenRepository
	chatRepo   repository.ChatRepository
	sender     PushSender
}

// NewNotificationService creates a new notification service.
func NewNotificationService(
	deviceRepo repository.DeviceTokenRepository,
	chatRepo repository.ChatRepository,
	sender PushSender,
) NotificationService {
	return &notificationService{
		deviceRepo: deviceRepo,
		chatRepo:   chatRepo,
		sender:     sender,
	}
}

func (s *notificationService) RegisterDevice(ctx context.Context, userID uuid.UUID, token, platform string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("device token kosong")
	}
	platform = strings.ToLower(strings.TrimSpace(platform))
	if platform != "ios" && platform != "android" {
		return fmt.Errorf("platform tidak valid: %s", platform)
	}
	_, err := s.deviceRepo.Upsert(ctx, userID, token, platform)
	return err
}

func (s *notificationService) UnregisterDevice(ctx context.Context, userID uuid.UUID, token string) error {
	return s.deviceRepo.Delete(ctx, userID, token)
}

func (s *notificationService) SendToUser(ctx context.Context, userID uuid.UUID, notif model.Notification) error {
	tokens, err := s.deviceRepo.FindByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("find device tokens: %w", err)
	}
	if len(tokens) == 0 {
		return nil // user has no registered devices
	}

	var tokenStrings []string
	for _, t := range tokens {
		tokenStrings = append(tokenStrings, t.Token)
	}

	if err := s.sender.SendMulti(ctx, tokenStrings, notif); err != nil {
		log.Warn().Err(err).Str("userId", userID.String()).Msg("failed to send push notification")
		return err
	}
	return nil
}

func (s *notificationService) SendToUsers(ctx context.Context, userIDs []uuid.UUID, notif model.Notification) error {
	if len(userIDs) == 0 {
		return nil
	}
	tokens, err := s.deviceRepo.FindByUsers(ctx, userIDs)
	if err != nil {
		return fmt.Errorf("find device tokens for users: %w", err)
	}
	if len(tokens) == 0 {
		return nil
	}

	var tokenStrings []string
	for _, t := range tokens {
		tokenStrings = append(tokenStrings, t.Token)
	}

	if err := s.sender.SendMulti(ctx, tokenStrings, notif); err != nil {
		log.Warn().Err(err).Msg("failed to send push notifications to users")
		return err
	}
	return nil
}

func (s *notificationService) SendToChat(ctx context.Context, chatID uuid.UUID, excludeUserID uuid.UUID, notif model.Notification) error {
	// Get chat members
	members, err := s.chatRepo.GetMembers(ctx, chatID)
	if err != nil {
		return fmt.Errorf("get chat members: %w", err)
	}

	var userIDs []uuid.UUID
	for _, m := range members {
		if m.UserID != excludeUserID {
			userIDs = append(userIDs, m.UserID)
		}
	}

	if len(userIDs) == 0 {
		return nil
	}

	return s.SendToUsers(ctx, userIDs, notif)
}

// -- Notification builder helpers --

// BuildMessageNotif creates a notification for a new chat message.
func BuildMessageNotif(senderName, content string, chatID uuid.UUID, chatType string) model.Notification {
	preview := truncate(content, 50)
	body := fmt.Sprintf("%s: %s", senderName, preview)

	notifType := model.NotifTypeMessage
	if chatType == "group" {
		notifType = model.NotifTypeGroupMessage
	}

	return model.Notification{
		Type:  notifType,
		Title: senderName,
		Body:  body,
		Data: map[string]string{
			"type":   string(notifType),
			"chatId": chatID.String(),
		},
		Sound:    "default",
		Priority: "high",
	}
}

// BuildGroupMessageNotif creates a notification for a group chat message.
func BuildGroupMessageNotif(groupName, senderName, content string, chatID uuid.UUID) model.Notification {
	preview := truncate(content, 50)
	title := groupName
	body := fmt.Sprintf("%s: %s", senderName, preview)

	return model.Notification{
		Type:  model.NotifTypeGroupMessage,
		Title: title,
		Body:  body,
		Data: map[string]string{
			"type":   string(model.NotifTypeGroupMessage),
			"chatId": chatID.String(),
		},
		Sound:    "default",
		Priority: "high",
	}
}

// BuildTopicMessageNotif creates a notification for a topic message.
func BuildTopicMessageNotif(topicName, senderName, content string, topicID uuid.UUID) model.Notification {
	preview := truncate(content, 50)
	title := topicName
	body := fmt.Sprintf("%s: %s", senderName, preview)

	return model.Notification{
		Type:  model.NotifTypeTopicMessage,
		Title: title,
		Body:  body,
		Data: map[string]string{
			"type":    string(model.NotifTypeTopicMessage),
			"topicId": topicID.String(),
		},
		Sound:    "default",
		Priority: "high",
	}
}

// BuildSignatureRequestNotif creates a notification for a signature request.
func BuildSignatureRequestNotif(requesterName, docTitle string, docID uuid.UUID) model.Notification {
	body := fmt.Sprintf("%s meminta tanda tangan Anda untuk '%s'", requesterName, truncate(docTitle, 40))

	return model.Notification{
		Type:  model.NotifTypeSignatureRequest,
		Title: "Permintaan Tanda Tangan",
		Body:  body,
		Data: map[string]string{
			"type":       string(model.NotifTypeSignatureRequest),
			"documentId": docID.String(),
		},
		Sound:    "default",
		Priority: "high",
	}
}

// BuildDocLockedNotif creates a notification for a locked document.
func BuildDocLockedNotif(ownerName, docTitle string, docID uuid.UUID) model.Notification {
	body := fmt.Sprintf("Dokumen '%s' telah dikunci oleh %s", truncate(docTitle, 40), ownerName)

	return model.Notification{
		Type:  model.NotifTypeDocumentLocked,
		Title: "Dokumen Dikunci",
		Body:  body,
		Data: map[string]string{
			"type":       string(model.NotifTypeDocumentLocked),
			"documentId": docID.String(),
		},
		Sound:    "default",
		Priority: "normal",
	}
}

// BuildGroupInviteNotif creates a notification for a group invite.
func BuildGroupInviteNotif(inviterName, groupName string, chatID uuid.UUID) model.Notification {
	body := fmt.Sprintf("%s mengundang Anda ke grup '%s'", inviterName, groupName)

	return model.Notification{
		Type:  model.NotifTypeGroupInvite,
		Title: "Undangan Grup",
		Body:  body,
		Data: map[string]string{
			"type":   string(model.NotifTypeGroupInvite),
			"chatId": chatID.String(),
		},
		Sound:    "default",
		Priority: "high",
	}
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
