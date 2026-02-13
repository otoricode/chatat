package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/ws"
	"github.com/otoritech/chatat/pkg/apperror"
)

// SendMessageInput holds the input for sending a message.
type SendMessageInput struct {
	ChatID    uuid.UUID         `json:"chatId"`
	SenderID  uuid.UUID         `json:"-"`
	Content   string            `json:"content"`
	ReplyToID *uuid.UUID        `json:"replyToId"`
	Type      model.MessageType `json:"type"`
	Metadata  json.RawMessage   `json:"metadata"`
}

// MessagePage represents a paginated list of messages.
type MessagePage struct {
	Messages []*model.Message `json:"messages"`
	Cursor   string           `json:"cursor"`
	HasMore  bool             `json:"hasMore"`
}

// WSMessageEvent is the WebSocket event payload for new messages.
type WSMessageEvent struct {
	Type    string         `json:"type"`
	Payload *model.Message `json:"payload"`
}

// MessageService defines operations for message management.
type MessageService interface {
	SendMessage(ctx context.Context, input SendMessageInput) (*model.Message, error)
	GetMessages(ctx context.Context, chatID uuid.UUID, cursor string, limit int) (*MessagePage, error)
	ForwardMessage(ctx context.Context, messageID, senderID, targetChatID uuid.UUID) (*model.Message, error)
	DeleteMessage(ctx context.Context, messageID, userID uuid.UUID, forAll bool) error
	SearchMessages(ctx context.Context, chatID uuid.UUID, query string) ([]*model.Message, error)
	MarkChatAsRead(ctx context.Context, chatID, userID uuid.UUID) error
}

type messageService struct {
	messageRepo     repository.MessageRepository
	messageStatRepo repository.MessageStatusRepository
	chatRepo        repository.ChatRepository
	userRepo        repository.UserRepository
	hub             *ws.Hub
	notifSvc        NotificationService
}

// NewMessageService creates a new MessageService.
func NewMessageService(
	messageRepo repository.MessageRepository,
	messageStatRepo repository.MessageStatusRepository,
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
	hub *ws.Hub,
	notifSvc NotificationService,
) MessageService {
	return &messageService{
		messageRepo:     messageRepo,
		messageStatRepo: messageStatRepo,
		chatRepo:        chatRepo,
		userRepo:        userRepo,
		hub:             hub,
		notifSvc:        notifSvc,
	}
}

func (s *messageService) SendMessage(ctx context.Context, input SendMessageInput) (*model.Message, error) {
	// Validate content
	if input.Type == "" {
		input.Type = model.MessageTypeText
	}
	if input.Type == model.MessageTypeText && input.Content == "" {
		return nil, apperror.Validation("content", "message content cannot be empty")
	}

	// Verify sender is member of the chat
	members, err := s.chatRepo.GetMembers(ctx, input.ChatID)
	if err != nil {
		return nil, fmt.Errorf("get chat members: %w", err)
	}
	isMember := false
	for _, m := range members {
		if m.UserID == input.SenderID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, apperror.Forbidden("you are not a member of this chat")
	}

	// Validate replyToID if provided
	if input.ReplyToID != nil {
		replyMsg, err := s.messageRepo.FindByID(ctx, *input.ReplyToID)
		if err != nil {
			if apperror.IsNotFound(err) {
				return nil, apperror.BadRequest("reply message not found")
			}
			return nil, fmt.Errorf("find reply message: %w", err)
		}
		// Reply must be in the same chat
		if replyMsg.ChatID != input.ChatID {
			return nil, apperror.BadRequest("reply message must be in the same chat")
		}
	}

	// Create message
	msg, err := s.messageRepo.Create(ctx, model.CreateMessageInput{
		ChatID:    input.ChatID,
		SenderID:  input.SenderID,
		Content:   input.Content,
		ReplyToID: input.ReplyToID,
		Type:      input.Type,
		Metadata:  input.Metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	// Create message_status entries for other members (status: sent)
	for _, m := range members {
		if m.UserID == input.SenderID {
			continue
		}
		if err := s.messageStatRepo.Create(ctx, msg.ID, m.UserID, model.DeliveryStatusSent); err != nil {
			// Log but don't fail the send
			continue
		}
	}

	// Broadcast via WebSocket to chat room
	roomID := "chat:" + input.ChatID.String()
	event := WSMessageEvent{
		Type:    "new_message",
		Payload: msg,
	}
	data, err := json.Marshal(event)
	if err == nil {
		s.hub.SendToRoom(roomID, data, uuid.Nil)
	}

	// Send push notification (fire-and-forget)
	if s.notifSvc != nil {
		go func() {
			senderName := "Seseorang"
			if sender, err := s.userRepo.FindByID(context.Background(), input.SenderID); err == nil && sender.Name != "" {
				senderName = sender.Name
			}

			chat, err := s.chatRepo.FindByID(context.Background(), input.ChatID)
			if err != nil {
				return
			}

			var notif model.Notification
			if chat.Type == model.ChatTypeGroup {
				notif = BuildGroupMessageNotif(chat.Name, senderName, input.Content, input.ChatID)
			} else {
				notif = BuildMessageNotif(senderName, input.Content, input.ChatID, string(chat.Type))
			}

			_ = s.notifSvc.SendToChat(context.Background(), input.ChatID, input.SenderID, notif)
		}()
	}

	return msg, nil
}

func (s *messageService) GetMessages(ctx context.Context, chatID uuid.UUID, cursor string, limit int) (*MessagePage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var cursorTime *time.Time
	if cursor != "" {
		t, err := time.Parse(time.RFC3339Nano, cursor)
		if err != nil {
			return nil, apperror.BadRequest("invalid cursor format")
		}
		cursorTime = &t
	}

	// Fetch limit+1 to check if there are more
	messages, err := s.messageRepo.ListByChat(ctx, chatID, cursorTime, limit+1)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	var nextCursor string
	if hasMore && len(messages) > 0 {
		nextCursor = messages[len(messages)-1].CreatedAt.Format(time.RFC3339Nano)
	}

	return &MessagePage{
		Messages: messages,
		Cursor:   nextCursor,
		HasMore:  hasMore,
	}, nil
}

func (s *messageService) ForwardMessage(ctx context.Context, messageID, senderID, targetChatID uuid.UUID) (*model.Message, error) {
	// Find original message
	originalMsg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("find original message: %w", err)
	}

	// Verify sender is member of target chat
	members, err := s.chatRepo.GetMembers(ctx, targetChatID)
	if err != nil {
		return nil, fmt.Errorf("get target chat members: %w", err)
	}
	isMember := false
	for _, m := range members {
		if m.UserID == senderID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, apperror.Forbidden("you are not a member of the target chat")
	}

	// Build forwarded metadata
	meta := map[string]interface{}{
		"forwarded":         true,
		"originalChatId":    originalMsg.ChatID.String(),
		"originalMessageId": originalMsg.ID.String(),
	}
	metaJSON, _ := json.Marshal(meta)

	// Create forwarded message in target chat
	msg, err := s.messageRepo.Create(ctx, model.CreateMessageInput{
		ChatID:   targetChatID,
		SenderID: senderID,
		Content:  originalMsg.Content,
		Type:     originalMsg.Type,
		Metadata: metaJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("create forwarded message: %w", err)
	}

	// Create status entries for other members
	for _, m := range members {
		if m.UserID == senderID {
			continue
		}
		_ = s.messageStatRepo.Create(ctx, msg.ID, m.UserID, model.DeliveryStatusSent)
	}

	// Broadcast to target chat
	roomID := "chat:" + targetChatID.String()
	event := WSMessageEvent{
		Type:    "new_message",
		Payload: msg,
	}
	data, err := json.Marshal(event)
	if err == nil {
		s.hub.SendToRoom(roomID, data, uuid.Nil)
	}

	return msg, nil
}

func (s *messageService) DeleteMessage(ctx context.Context, messageID, userID uuid.UUID, forAll bool) error {
	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("find message: %w", err)
	}

	if forAll {
		// Only sender can delete for all
		if msg.SenderID != userID {
			return apperror.Forbidden("only the sender can delete a message for everyone")
		}

		// Only within 1 hour
		if time.Since(msg.CreatedAt) > time.Hour {
			return apperror.BadRequest("can only delete for everyone within 1 hour of sending")
		}
	}

	if err := s.messageRepo.MarkAsDeleted(ctx, messageID, forAll); err != nil {
		return fmt.Errorf("mark as deleted: %w", err)
	}

	// Broadcast deletion via WebSocket
	if forAll {
		roomID := "chat:" + msg.ChatID.String()
		event := map[string]interface{}{
			"type": "message_deleted",
			"payload": map[string]interface{}{
				"messageId":     messageID.String(),
				"chatId":        msg.ChatID.String(),
				"deletedForAll": true,
			},
		}
		data, _ := json.Marshal(event)
		s.hub.SendToRoom(roomID, data, uuid.Nil)
	}

	return nil
}

func (s *messageService) SearchMessages(ctx context.Context, chatID uuid.UUID, query string) ([]*model.Message, error) {
	if query == "" {
		return nil, apperror.BadRequest("search query cannot be empty")
	}

	messages, err := s.messageRepo.Search(ctx, chatID, query)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}

	return messages, nil
}

func (s *messageService) MarkChatAsRead(ctx context.Context, chatID, userID uuid.UUID) error {
	if err := s.messageStatRepo.MarkChatAsRead(ctx, chatID, userID); err != nil {
		return fmt.Errorf("mark chat as read: %w", err)
	}
	return nil
}
