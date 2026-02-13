package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/ws"
	"github.com/otoritech/chatat/pkg/apperror"
)

// CreateTopicInput holds data needed to create a new topic.
type CreateTopicInput struct {
	Name        string      `json:"name"`
	Icon        string      `json:"icon"`
	Description string      `json:"description"`
	ParentID    uuid.UUID   `json:"parentId"`
	MemberIDs   []uuid.UUID `json:"memberIds"`
}

// UpdateTopicInput holds optional fields for updating a topic.
type UpdateTopicInput struct {
	Name        *string `json:"name"`
	Icon        *string `json:"icon"`
	Description *string `json:"description"`
}

// TopicListItem represents a topic in a list with metadata.
type TopicListItem struct {
	Topic       model.Topic         `json:"topic"`
	LastMessage *model.TopicMessage `json:"lastMessage"`
	MemberCount int                 `json:"memberCount"`
}

// TopicDetail represents detailed topic information.
type TopicDetail struct {
	Topic   model.Topic   `json:"topic"`
	Members []*MemberInfo `json:"members"`
	Parent  *model.Chat   `json:"parent"`
}

// TopicService defines operations for topic management.
type TopicService interface {
	CreateTopic(ctx context.Context, userID uuid.UUID, input CreateTopicInput) (*model.Topic, error)
	GetTopic(ctx context.Context, topicID, userID uuid.UUID) (*TopicDetail, error)
	ListByChat(ctx context.Context, chatID, userID uuid.UUID) ([]*TopicListItem, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*TopicListItem, error)
	UpdateTopic(ctx context.Context, topicID, userID uuid.UUID, input UpdateTopicInput) (*model.Topic, error)
	AddMember(ctx context.Context, topicID, userID, addedBy uuid.UUID) error
	RemoveMember(ctx context.Context, topicID, userID, removedBy uuid.UUID) error
	DeleteTopic(ctx context.Context, topicID, userID uuid.UUID) error
}

type topicService struct {
	topicRepo    repository.TopicRepository
	topicMsgRepo repository.TopicMessageRepository
	chatRepo     repository.ChatRepository
	userRepo     repository.UserRepository
	hub          *ws.Hub
}

// NewTopicService creates a new TopicService.
func NewTopicService(
	topicRepo repository.TopicRepository,
	topicMsgRepo repository.TopicMessageRepository,
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
	hub *ws.Hub,
) TopicService {
	return &topicService{
		topicRepo:    topicRepo,
		topicMsgRepo: topicMsgRepo,
		chatRepo:     chatRepo,
		userRepo:     userRepo,
		hub:          hub,
	}
}

func (s *topicService) CreateTopic(ctx context.Context, userID uuid.UUID, input CreateTopicInput) (*model.Topic, error) {
	// Validate input
	if input.Name == "" {
		return nil, apperror.Validation("name", "topic name is required")
	}
	if input.Icon == "" {
		return nil, apperror.Validation("icon", "topic icon is required")
	}

	// Validate parent chat exists
	parentChat, err := s.chatRepo.FindByID(ctx, input.ParentID)
	if err != nil {
		if apperror.IsNotFound(err) {
			return nil, apperror.BadRequest("parent chat not found")
		}
		return nil, fmt.Errorf("find parent chat: %w", err)
	}

	// Validate creator is member of parent chat
	parentMembers, err := s.chatRepo.GetMembers(ctx, input.ParentID)
	if err != nil {
		return nil, fmt.Errorf("get parent members: %w", err)
	}

	creatorIsMember := false
	parentMemberSet := make(map[uuid.UUID]bool)
	for _, m := range parentMembers {
		parentMemberSet[m.UserID] = true
		if m.UserID == userID {
			creatorIsMember = true
		}
	}
	if !creatorIsMember {
		return nil, apperror.Forbidden("you are not a member of the parent chat")
	}

	// Determine members
	var memberIDs []uuid.UUID
	if parentChat.Type == model.ChatTypePersonal {
		// For personal chat, auto-include both participants
		for _, m := range parentMembers {
			memberIDs = append(memberIDs, m.UserID)
		}
	} else {
		// For group, validate all memberIDs are in parent chat
		if len(input.MemberIDs) == 0 {
			// If no members specified, include all parent members
			for _, m := range parentMembers {
				memberIDs = append(memberIDs, m.UserID)
			}
		} else {
			// Ensure creator is always included
			creatorIncluded := false
			for _, id := range input.MemberIDs {
				if id == userID {
					creatorIncluded = true
				}
				if !parentMemberSet[id] {
					return nil, apperror.BadRequest("member " + id.String() + " is not in parent chat")
				}
			}
			memberIDs = input.MemberIDs
			if !creatorIncluded {
				memberIDs = append([]uuid.UUID{userID}, memberIDs...)
			}
		}
	}

	// Create topic
	topic, err := s.topicRepo.Create(ctx, model.CreateTopicInput{
		Name:        input.Name,
		Icon:        input.Icon,
		Description: input.Description,
		ParentType:  parentChat.Type,
		ParentID:    input.ParentID,
		CreatedBy:   userID,
	})
	if err != nil {
		return nil, fmt.Errorf("create topic: %w", err)
	}

	// Add members
	for _, memberID := range memberIDs {
		role := model.MemberRoleMember
		if memberID == userID {
			role = model.MemberRoleAdmin
		}
		if err := s.topicRepo.AddMember(ctx, topic.ID, memberID, role); err != nil {
			// Log but continue
			continue
		}
	}

	// Get creator name for system message
	creator, _ := s.userRepo.FindByID(ctx, userID)
	creatorName := "Seseorang"
	if creator != nil {
		creatorName = creator.Name
	}

	// Send system message to parent chat about topic creation
	sysContent := creatorName + " membuat topik " + topic.Name
	s.sendParentSystemMessage(ctx, input.ParentID, userID, sysContent)

	// Create WS room for topic
	if s.hub != nil {
		roomID := "topic:" + topic.ID.String()
		_ = roomID // Room will be created when users connect
	}

	return topic, nil
}

func (s *topicService) GetTopic(ctx context.Context, topicID, userID uuid.UUID) (*TopicDetail, error) {
	topic, err := s.topicRepo.FindByID(ctx, topicID)
	if err != nil {
		return nil, err
	}

	// Verify user is member
	if err := s.verifyMembership(ctx, topicID, userID); err != nil {
		return nil, err
	}

	// Get members
	topicMembers, err := s.topicRepo.GetMembers(ctx, topicID)
	if err != nil {
		return nil, fmt.Errorf("get topic members: %w", err)
	}

	members := make([]*MemberInfo, 0, len(topicMembers))
	for _, tm := range topicMembers {
		user, err := s.userRepo.FindByID(ctx, tm.UserID)
		if err != nil {
			continue
		}
		members = append(members, &MemberInfo{
			User:     *user,
			Role:     string(tm.Role),
			JoinedAt: tm.JoinedAt,
		})
	}

	// Get parent chat
	parent, _ := s.chatRepo.FindByID(ctx, topic.ParentID)

	return &TopicDetail{
		Topic:   *topic,
		Members: members,
		Parent:  parent,
	}, nil
}

func (s *topicService) ListByChat(ctx context.Context, chatID, userID uuid.UUID) ([]*TopicListItem, error) {
	// Verify user is member of parent chat
	parentMembers, err := s.chatRepo.GetMembers(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get parent members: %w", err)
	}
	isMember := false
	for _, m := range parentMembers {
		if m.UserID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, apperror.Forbidden("you are not a member of this chat")
	}

	topics, err := s.topicRepo.ListByParent(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("list topics by parent: %w", err)
	}

	items := make([]*TopicListItem, 0, len(topics))
	for _, t := range topics {
		item := &TopicListItem{
			Topic: *t,
		}

		// Get last message
		msgs, err := s.topicMsgRepo.ListByTopic(ctx, t.ID, nil, 1)
		if err == nil && len(msgs) > 0 {
			item.LastMessage = msgs[0]
		}

		// Get member count
		members, err := s.topicRepo.GetMembers(ctx, t.ID)
		if err == nil {
			item.MemberCount = len(members)
		}

		items = append(items, item)
	}

	return items, nil
}

func (s *topicService) ListByUser(ctx context.Context, userID uuid.UUID) ([]*TopicListItem, error) {
	topics, err := s.topicRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list topics by user: %w", err)
	}

	items := make([]*TopicListItem, 0, len(topics))
	for _, t := range topics {
		item := &TopicListItem{
			Topic: *t,
		}

		msgs, err := s.topicMsgRepo.ListByTopic(ctx, t.ID, nil, 1)
		if err == nil && len(msgs) > 0 {
			item.LastMessage = msgs[0]
		}

		members, err := s.topicRepo.GetMembers(ctx, t.ID)
		if err == nil {
			item.MemberCount = len(members)
		}

		items = append(items, item)
	}

	return items, nil
}

func (s *topicService) UpdateTopic(ctx context.Context, topicID, userID uuid.UUID, input UpdateTopicInput) (*model.Topic, error) {
	// Verify user is admin
	if err := s.verifyAdmin(ctx, topicID, userID); err != nil {
		return nil, err
	}

	topic, err := s.topicRepo.Update(ctx, topicID, model.UpdateTopicInput{
		Name:        input.Name,
		Icon:        input.Icon,
		Description: input.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("update topic: %w", err)
	}

	return topic, nil
}

func (s *topicService) AddMember(ctx context.Context, topicID, userID, addedBy uuid.UUID) error {
	// Verify addedBy is admin
	if err := s.verifyAdmin(ctx, topicID, addedBy); err != nil {
		return err
	}

	// Get topic to find parent
	topic, err := s.topicRepo.FindByID(ctx, topicID)
	if err != nil {
		return err
	}

	// Verify user is in parent chat
	parentMembers, err := s.chatRepo.GetMembers(ctx, topic.ParentID)
	if err != nil {
		return fmt.Errorf("get parent members: %w", err)
	}
	inParent := false
	for _, m := range parentMembers {
		if m.UserID == userID {
			inParent = true
			break
		}
	}
	if !inParent {
		return apperror.BadRequest("user is not a member of the parent chat")
	}

	if err := s.topicRepo.AddMember(ctx, topicID, userID, model.MemberRoleMember); err != nil {
		return fmt.Errorf("add topic member: %w", err)
	}

	// System message
	adder, _ := s.userRepo.FindByID(ctx, addedBy)
	added, _ := s.userRepo.FindByID(ctx, userID)
	adderName := "Seseorang"
	addedName := "anggota baru"
	if adder != nil {
		adderName = adder.Name
	}
	if added != nil {
		addedName = added.Name
	}
	s.sendTopicSystemMessage(ctx, topicID, addedBy, adderName+" menambahkan "+addedName)

	return nil
}

func (s *topicService) RemoveMember(ctx context.Context, topicID, userID, removedBy uuid.UUID) error {
	// Verify removedBy is admin
	if err := s.verifyAdmin(ctx, topicID, removedBy); err != nil {
		return err
	}

	// Cannot remove yourself via this method (use leave instead?)
	if userID == removedBy {
		return apperror.BadRequest("use topic leave to remove yourself")
	}

	if err := s.topicRepo.RemoveMember(ctx, topicID, userID); err != nil {
		return fmt.Errorf("remove topic member: %w", err)
	}

	// System message
	remover, _ := s.userRepo.FindByID(ctx, removedBy)
	removed, _ := s.userRepo.FindByID(ctx, userID)
	removerName := "Admin"
	removedName := "anggota"
	if remover != nil {
		removerName = remover.Name
	}
	if removed != nil {
		removedName = removed.Name
	}
	s.sendTopicSystemMessage(ctx, topicID, removedBy, removerName+" mengeluarkan "+removedName)

	return nil
}

func (s *topicService) DeleteTopic(ctx context.Context, topicID, userID uuid.UUID) error {
	// Verify user is admin
	if err := s.verifyAdmin(ctx, topicID, userID); err != nil {
		return err
	}

	topic, err := s.topicRepo.FindByID(ctx, topicID)
	if err != nil {
		return err
	}

	// Send system message to parent chat
	deleter, _ := s.userRepo.FindByID(ctx, userID)
	deleterName := "Admin"
	if deleter != nil {
		deleterName = deleter.Name
	}
	s.sendParentSystemMessage(ctx, topic.ParentID, userID, deleterName+" menghapus topik "+topic.Name)

	if err := s.topicRepo.Delete(ctx, topicID); err != nil {
		return fmt.Errorf("delete topic: %w", err)
	}

	return nil
}

// --- Helpers ---

func (s *topicService) verifyMembership(ctx context.Context, topicID, userID uuid.UUID) error {
	members, err := s.topicRepo.GetMembers(ctx, topicID)
	if err != nil {
		return fmt.Errorf("get topic members: %w", err)
	}
	for _, m := range members {
		if m.UserID == userID {
			return nil
		}
	}
	return apperror.Forbidden("you are not a member of this topic")
}

func (s *topicService) verifyAdmin(ctx context.Context, topicID, userID uuid.UUID) error {
	members, err := s.topicRepo.GetMembers(ctx, topicID)
	if err != nil {
		return fmt.Errorf("get topic members: %w", err)
	}
	for _, m := range members {
		if m.UserID == userID && m.Role == model.MemberRoleAdmin {
			return nil
		}
	}
	return apperror.Forbidden("only topic admin can perform this action")
}

func (s *topicService) sendParentSystemMessage(ctx context.Context, chatID, senderID uuid.UUID, content string) {
	// Use a broadcast message via hub for the parent chat
	msg := map[string]interface{}{
		"chatId":   chatID.String(),
		"senderId": senderID.String(),
		"content":  content,
		"type":     "system",
	}
	data, err := json.Marshal(ws.WSMessage{
		Type:    ws.WSTypeMessage,
		Payload: mustMarshal(msg),
	})
	if err != nil {
		return
	}
	if s.hub != nil {
		roomID := "chat:" + chatID.String()
		s.hub.SendToRoom(roomID, data, uuid.Nil)
	}
}

func (s *topicService) sendTopicSystemMessage(ctx context.Context, topicID, senderID uuid.UUID, content string) {
	// Create an actual topic message of type system
	_, _ = s.topicMsgRepo.Create(ctx, model.CreateTopicMessageInput{
		TopicID:  topicID,
		SenderID: senderID,
		Content:  content,
		Type:     model.MessageTypeSystem,
	})
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
