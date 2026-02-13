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

// SendTopicMessageInput holds input for sending a topic message.
type SendTopicMessageInput struct {
	TopicID   uuid.UUID         `json:"topicId"`
	SenderID  uuid.UUID         `json:"-"`
	Content   string            `json:"content"`
	ReplyToID *uuid.UUID        `json:"replyToId"`
	Type      model.MessageType `json:"type"`
}

// TopicMessagePage represents a paginated list of topic messages.
type TopicMessagePage struct {
	Messages []*model.TopicMessage `json:"messages"`
	Cursor   string                `json:"cursor"`
	HasMore  bool                  `json:"hasMore"`
}

// TopicMessageService defines operations for topic message management.
type TopicMessageService interface {
	SendMessage(ctx context.Context, input SendTopicMessageInput) (*model.TopicMessage, error)
	GetMessages(ctx context.Context, topicID uuid.UUID, cursor string, limit int) (*TopicMessagePage, error)
	DeleteMessage(ctx context.Context, messageID, userID uuid.UUID, forAll bool) error
}

type topicMessageService struct {
	topicMsgRepo repository.TopicMessageRepository
	topicRepo    repository.TopicRepository
	hub          *ws.Hub
}

// NewTopicMessageService creates a new TopicMessageService.
func NewTopicMessageService(
	topicMsgRepo repository.TopicMessageRepository,
	topicRepo repository.TopicRepository,
	hub *ws.Hub,
) TopicMessageService {
	return &topicMessageService{
		topicMsgRepo: topicMsgRepo,
		topicRepo:    topicRepo,
		hub:          hub,
	}
}

func (s *topicMessageService) SendMessage(ctx context.Context, input SendTopicMessageInput) (*model.TopicMessage, error) {
	// Validate content
	if input.Type == "" {
		input.Type = model.MessageTypeText
	}
	if input.Type == model.MessageTypeText && input.Content == "" {
		return nil, apperror.Validation("content", "message content cannot be empty")
	}

	// Verify sender is topic member
	members, err := s.topicRepo.GetMembers(ctx, input.TopicID)
	if err != nil {
		return nil, fmt.Errorf("get topic members: %w", err)
	}
	isMember := false
	for _, m := range members {
		if m.UserID == input.SenderID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, apperror.Forbidden("you are not a member of this topic")
	}

	// Validate replyToID
	if input.ReplyToID != nil {
		replyMsg, err := s.topicMsgRepo.FindByID(ctx, *input.ReplyToID)
		if err != nil {
			if apperror.IsNotFound(err) {
				return nil, apperror.BadRequest("reply message not found")
			}
			return nil, fmt.Errorf("find reply message: %w", err)
		}
		if replyMsg.TopicID != input.TopicID {
			return nil, apperror.BadRequest("reply message must be in the same topic")
		}
	}

	// Create message
	msg, err := s.topicMsgRepo.Create(ctx, model.CreateTopicMessageInput{
		TopicID:   input.TopicID,
		SenderID:  input.SenderID,
		Content:   input.Content,
		ReplyToID: input.ReplyToID,
		Type:      input.Type,
	})
	if err != nil {
		return nil, fmt.Errorf("create topic message: %w", err)
	}

	// Broadcast via WebSocket to topic room
	roomID := "topic:" + input.TopicID.String()
	event := map[string]interface{}{
		"type":    "new_topic_message",
		"payload": msg,
	}
	data, err := json.Marshal(event)
	if err == nil && s.hub != nil {
		s.hub.SendToRoom(roomID, data, uuid.Nil)
	}

	return msg, nil
}

func (s *topicMessageService) GetMessages(ctx context.Context, topicID uuid.UUID, cursor string, limit int) (*TopicMessagePage, error) {
	if limit <= 0 {
		limit = 20
	}

	var cursorTime *time.Time
	if cursor != "" {
		t, err := time.Parse(time.RFC3339Nano, cursor)
		if err != nil {
			return nil, apperror.BadRequest("invalid cursor format")
		}
		cursorTime = &t
	}

	// Fetch one extra to check hasMore
	msgs, err := s.topicMsgRepo.ListByTopic(ctx, topicID, cursorTime, limit+1)
	if err != nil {
		return nil, fmt.Errorf("list topic messages: %w", err)
	}

	hasMore := len(msgs) > limit
	if hasMore {
		msgs = msgs[:limit]
	}

	var nextCursor string
	if len(msgs) > 0 {
		nextCursor = msgs[len(msgs)-1].CreatedAt.Format(time.RFC3339Nano)
	}

	return &TopicMessagePage{
		Messages: msgs,
		Cursor:   nextCursor,
		HasMore:  hasMore,
	}, nil
}

func (s *topicMessageService) DeleteMessage(ctx context.Context, messageID, userID uuid.UUID, forAll bool) error {
	msg, err := s.topicMsgRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}

	// Only sender can delete their own message
	if msg.SenderID != userID {
		return apperror.Forbidden("you can only delete your own messages")
	}

	if err := s.topicMsgRepo.MarkAsDeleted(ctx, messageID, forAll); err != nil {
		return fmt.Errorf("delete topic message: %w", err)
	}

	// Broadcast deletion via WebSocket
	if forAll && s.hub != nil {
		roomID := "topic:" + msg.TopicID.String()
		event := map[string]interface{}{
			"type": "topic_message_deleted",
			"payload": map[string]string{
				"messageId": messageID.String(),
				"topicId":   msg.TopicID.String(),
			},
		}
		data, _ := json.Marshal(event)
		s.hub.SendToRoom(roomID, data, uuid.Nil)
	}

	return nil
}
