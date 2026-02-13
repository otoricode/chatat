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

// CreateGroupInput holds data for creating a group chat.
type CreateGroupInput struct {
	Name        string      `json:"name"`
	Icon        string      `json:"icon"`
	Description string      `json:"description"`
	MemberIDs   []uuid.UUID `json:"memberIds"`
}

// UpdateGroupInput holds optional fields for updating a group.
type UpdateGroupInput struct {
	Name        *string `json:"name"`
	Icon        *string `json:"icon"`
	Description *string `json:"description"`
}

// GroupInfo represents detailed group information.
type GroupInfo struct {
	Chat    model.Chat    `json:"chat"`
	Members []*MemberInfo `json:"members"`
}

// MemberInfo represents a group member with profile and status.
type MemberInfo struct {
	User     model.User `json:"user"`
	Role     string     `json:"role"`
	IsOnline bool       `json:"isOnline"`
	JoinedAt time.Time  `json:"joinedAt"`
}

// GroupService defines operations for group management.
type GroupService interface {
	CreateGroup(ctx context.Context, creatorID uuid.UUID, input CreateGroupInput) (*model.Chat, error)
	UpdateGroup(ctx context.Context, chatID, userID uuid.UUID, input UpdateGroupInput) (*model.Chat, error)
	AddMember(ctx context.Context, chatID, userID, addedBy uuid.UUID) error
	RemoveMember(ctx context.Context, chatID, userID, removedBy uuid.UUID) error
	PromoteToAdmin(ctx context.Context, chatID, userID, promotedBy uuid.UUID) error
	LeaveGroup(ctx context.Context, chatID, userID uuid.UUID) error
	DeleteGroup(ctx context.Context, chatID, userID uuid.UUID) error
	GetGroupInfo(ctx context.Context, chatID, userID uuid.UUID) (*GroupInfo, error)
}

type groupService struct {
	chatRepo        repository.ChatRepository
	messageRepo     repository.MessageRepository
	messageStatRepo repository.MessageStatusRepository
	userRepo        repository.UserRepository
	hub             *ws.Hub
}

// NewGroupService creates a new GroupService.
func NewGroupService(
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
	messageStatRepo repository.MessageStatusRepository,
	userRepo repository.UserRepository,
	hub *ws.Hub,
) GroupService {
	return &groupService{
		chatRepo:        chatRepo,
		messageRepo:     messageRepo,
		messageStatRepo: messageStatRepo,
		userRepo:        userRepo,
		hub:             hub,
	}
}

func (s *groupService) CreateGroup(ctx context.Context, creatorID uuid.UUID, input CreateGroupInput) (*model.Chat, error) {
	// Validate name
	if input.Name == "" {
		return nil, apperror.Validation("name", "group name is required")
	}
	if len(input.Name) > 100 {
		return nil, apperror.Validation("name", "group name must be at most 100 characters")
	}

	// Validate icon
	if input.Icon == "" {
		return nil, apperror.Validation("icon", "group icon is required")
	}

	// Validate member count: at least 2 other members
	if len(input.MemberIDs) < 2 {
		return nil, apperror.Validation("memberIds", "group must have at least 2 other members")
	}

	// Remove duplicates and make sure creator is not in the member list
	uniqueMembers := make(map[uuid.UUID]bool)
	var memberIDs []uuid.UUID
	for _, id := range input.MemberIDs {
		if id == creatorID {
			continue
		}
		if !uniqueMembers[id] {
			uniqueMembers[id] = true
			memberIDs = append(memberIDs, id)
		}
	}

	if len(memberIDs) < 2 {
		return nil, apperror.Validation("memberIds", "group must have at least 2 other members")
	}

	// Verify all member users exist
	for _, memberID := range memberIDs {
		if _, err := s.userRepo.FindByID(ctx, memberID); err != nil {
			if apperror.IsNotFound(err) {
				return nil, apperror.NotFound("user", memberID.String())
			}
			return nil, fmt.Errorf("verify member: %w", err)
		}
	}

	// Create the group chat
	chat, err := s.chatRepo.Create(ctx, model.CreateChatInput{
		Type:        model.ChatTypeGroup,
		Name:        input.Name,
		Icon:        input.Icon,
		Description: input.Description,
		CreatedBy:   creatorID,
	})
	if err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}

	// Add creator as admin
	if err := s.chatRepo.AddMember(ctx, chat.ID, creatorID, model.MemberRoleAdmin); err != nil {
		return nil, fmt.Errorf("add creator: %w", err)
	}

	// Add all members
	for _, memberID := range memberIDs {
		if err := s.chatRepo.AddMember(ctx, chat.ID, memberID, model.MemberRoleMember); err != nil {
			return nil, fmt.Errorf("add member: %w", err)
		}
	}

	// Get creator's name for system message
	creator, _ := s.userRepo.FindByID(ctx, creatorID)
	creatorName := "Seseorang"
	if creator != nil && creator.Name != "" {
		creatorName = creator.Name
	}

	// Send system message
	sysMsg := creatorName + " membuat grup"
	s.sendSystemMessage(ctx, chat.ID, creatorID, sysMsg)

	// Broadcast group creation
	s.broadcastGroupEvent(chat.ID, "group_created", map[string]interface{}{
		"chatId": chat.ID.String(),
		"name":   chat.Name,
		"icon":   chat.Icon,
	})

	return chat, nil
}

func (s *groupService) UpdateGroup(ctx context.Context, chatID, userID uuid.UUID, input UpdateGroupInput) (*model.Chat, error) {
	// Verify user is admin
	if err := s.requireAdmin(ctx, chatID, userID); err != nil {
		return nil, err
	}

	// Get current chat
	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("find chat: %w", err)
	}

	// Ensure it's a group
	if chat.Type != model.ChatTypeGroup {
		return nil, apperror.BadRequest("only group chats can be updated")
	}

	// Build update input
	updateInput := model.UpdateChatInput{
		Name:        input.Name,
		Icon:        input.Icon,
		Description: input.Description,
	}

	updated, err := s.chatRepo.Update(ctx, chatID, updateInput)
	if err != nil {
		return nil, fmt.Errorf("update group: %w", err)
	}

	// Get updater name for system message
	user, _ := s.userRepo.FindByID(ctx, userID)
	userName := "Seseorang"
	if user != nil && user.Name != "" {
		userName = user.Name
	}

	// System message for the update
	sysMsg := userName + " mengubah info grup"
	s.sendSystemMessage(ctx, chatID, userID, sysMsg)

	// Broadcast
	s.broadcastGroupEvent(chatID, "group_updated", map[string]interface{}{
		"chatId": chatID.String(),
		"name":   updated.Name,
		"icon":   updated.Icon,
	})

	return updated, nil
}

func (s *groupService) AddMember(ctx context.Context, chatID, userID, addedBy uuid.UUID) error {
	// Verify adder is admin
	if err := s.requireAdmin(ctx, chatID, addedBy); err != nil {
		return err
	}

	// Ensure it's a group
	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("find chat: %w", err)
	}
	if chat.Type != model.ChatTypeGroup {
		return apperror.BadRequest("can only add members to group chats")
	}

	// Check if user exists
	newUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if apperror.IsNotFound(err) {
			return apperror.NotFound("user", userID.String())
		}
		return fmt.Errorf("find user: %w", err)
	}

	// Check if already a member
	members, err := s.chatRepo.GetMembers(ctx, chatID)
	if err != nil {
		return fmt.Errorf("get members: %w", err)
	}
	for _, m := range members {
		if m.UserID == userID {
			return apperror.Conflict("user is already a member of this group")
		}
	}

	// Add the member
	if err := s.chatRepo.AddMember(ctx, chatID, userID, model.MemberRoleMember); err != nil {
		return fmt.Errorf("add member: %w", err)
	}

	// System message
	adder, _ := s.userRepo.FindByID(ctx, addedBy)
	adderName := "Seseorang"
	if adder != nil && adder.Name != "" {
		adderName = adder.Name
	}
	newUserName := "seseorang"
	if newUser.Name != "" {
		newUserName = newUser.Name
	}

	sysMsg := adderName + " menambahkan " + newUserName
	s.sendSystemMessage(ctx, chatID, addedBy, sysMsg)

	// Broadcast
	s.broadcastGroupEvent(chatID, "member_added", map[string]interface{}{
		"chatId": chatID.String(),
		"userId": userID.String(),
		"name":   newUserName,
	})

	return nil
}

func (s *groupService) RemoveMember(ctx context.Context, chatID, userID, removedBy uuid.UUID) error {
	// Verify remover is admin
	if err := s.requireAdmin(ctx, chatID, removedBy); err != nil {
		return err
	}

	// Cannot remove yourself (use LeaveGroup instead)
	if userID == removedBy {
		return apperror.BadRequest("use leave endpoint to leave the group")
	}

	// Cannot remove the creator
	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("find chat: %w", err)
	}
	if chat.Type != model.ChatTypeGroup {
		return apperror.BadRequest("can only remove members from group chats")
	}
	if userID == chat.CreatedBy {
		return apperror.Forbidden("cannot remove the group creator")
	}

	// Get the user being removed
	removedUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if apperror.IsNotFound(err) {
			return apperror.NotFound("user", userID.String())
		}
		return fmt.Errorf("find user: %w", err)
	}

	// Remove the member
	if err := s.chatRepo.RemoveMember(ctx, chatID, userID); err != nil {
		return fmt.Errorf("remove member: %w", err)
	}

	// System message
	remover, _ := s.userRepo.FindByID(ctx, removedBy)
	removerName := "Seseorang"
	if remover != nil && remover.Name != "" {
		removerName = remover.Name
	}
	removedName := "seseorang"
	if removedUser.Name != "" {
		removedName = removedUser.Name
	}

	sysMsg := removerName + " mengeluarkan " + removedName
	s.sendSystemMessage(ctx, chatID, removedBy, sysMsg)

	// Broadcast
	s.broadcastGroupEvent(chatID, "member_removed", map[string]interface{}{
		"chatId": chatID.String(),
		"userId": userID.String(),
	})

	return nil
}

func (s *groupService) PromoteToAdmin(ctx context.Context, chatID, userID, promotedBy uuid.UUID) error {
	// Verify promoter is admin
	if err := s.requireAdmin(ctx, chatID, promotedBy); err != nil {
		return err
	}

	// Ensure the target user is a member
	members, err := s.chatRepo.GetMembers(ctx, chatID)
	if err != nil {
		return fmt.Errorf("get members: %w", err)
	}

	var targetMember *model.ChatMember
	for _, m := range members {
		if m.UserID == userID {
			targetMember = m
			break
		}
	}
	if targetMember == nil {
		return apperror.NotFound("member", userID.String())
	}

	// Already admin
	if targetMember.Role == model.MemberRoleAdmin {
		return apperror.Conflict("user is already an admin")
	}

	// Remove and re-add as admin (since we don't have UpdateMemberRole)
	if err := s.chatRepo.RemoveMember(ctx, chatID, userID); err != nil {
		return fmt.Errorf("remove member for promotion: %w", err)
	}
	if err := s.chatRepo.AddMember(ctx, chatID, userID, model.MemberRoleAdmin); err != nil {
		return fmt.Errorf("add member as admin: %w", err)
	}

	// System message
	promoter, _ := s.userRepo.FindByID(ctx, promotedBy)
	promoterName := "Seseorang"
	if promoter != nil && promoter.Name != "" {
		promoterName = promoter.Name
	}
	target, _ := s.userRepo.FindByID(ctx, userID)
	targetName := "seseorang"
	if target != nil && target.Name != "" {
		targetName = target.Name
	}

	sysMsg := promoterName + " menjadikan " + targetName + " sebagai admin"
	s.sendSystemMessage(ctx, chatID, promotedBy, sysMsg)

	// Broadcast
	s.broadcastGroupEvent(chatID, "member_promoted", map[string]interface{}{
		"chatId": chatID.String(),
		"userId": userID.String(),
	})

	return nil
}

func (s *groupService) LeaveGroup(ctx context.Context, chatID, userID uuid.UUID) error {
	// Verify user is a member
	members, err := s.chatRepo.GetMembers(ctx, chatID)
	if err != nil {
		return fmt.Errorf("get members: %w", err)
	}

	isMember := false
	for _, m := range members {
		if m.UserID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return apperror.Forbidden("you are not a member of this group")
	}

	// Creator cannot leave
	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("find chat: %w", err)
	}
	if chat.Type != model.ChatTypeGroup {
		return apperror.BadRequest("can only leave group chats")
	}
	if userID == chat.CreatedBy {
		return apperror.Forbidden("group creator cannot leave the group, delete it instead")
	}

	// Remove the member
	if err := s.chatRepo.RemoveMember(ctx, chatID, userID); err != nil {
		return fmt.Errorf("leave group: %w", err)
	}

	// System message
	user, _ := s.userRepo.FindByID(ctx, userID)
	userName := "Seseorang"
	if user != nil && user.Name != "" {
		userName = user.Name
	}

	sysMsg := userName + " keluar dari grup"
	s.sendSystemMessage(ctx, chatID, userID, sysMsg)

	// Broadcast
	s.broadcastGroupEvent(chatID, "member_left", map[string]interface{}{
		"chatId": chatID.String(),
		"userId": userID.String(),
	})

	return nil
}

func (s *groupService) DeleteGroup(ctx context.Context, chatID, userID uuid.UUID) error {
	// Only creator can delete
	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("find chat: %w", err)
	}
	if chat.Type != model.ChatTypeGroup {
		return apperror.BadRequest("can only delete group chats")
	}
	if chat.CreatedBy != userID {
		return apperror.Forbidden("only the group creator can delete the group")
	}

	// Broadcast before deletion so members get notified
	s.broadcastGroupEvent(chatID, "group_deleted", map[string]interface{}{
		"chatId": chatID.String(),
		"name":   chat.Name,
	})

	// Delete cascades via FK
	if err := s.chatRepo.Delete(ctx, chatID); err != nil {
		return fmt.Errorf("delete group: %w", err)
	}

	return nil
}

func (s *groupService) GetGroupInfo(ctx context.Context, chatID, userID uuid.UUID) (*GroupInfo, error) {
	// Verify membership
	members, err := s.chatRepo.GetMembers(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	isMember := false
	for _, m := range members {
		if m.UserID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, apperror.Forbidden("you are not a member of this group")
	}

	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("find chat: %w", err)
	}

	// Build member info
	memberInfos := make([]*MemberInfo, 0, len(members))
	for _, m := range members {
		user, err := s.userRepo.FindByID(ctx, m.UserID)
		if err != nil {
			if apperror.IsNotFound(err) {
				continue
			}
			return nil, fmt.Errorf("find member user: %w", err)
		}
		memberInfos = append(memberInfos, &MemberInfo{
			User:     *user,
			Role:     string(m.Role),
			IsOnline: s.hub.IsOnline(m.UserID),
			JoinedAt: m.JoinedAt,
		})
	}

	return &GroupInfo{
		Chat:    *chat,
		Members: memberInfos,
	}, nil
}

// --- Helper Methods ---

// requireAdmin checks that the user is an admin of the chat.
func (s *groupService) requireAdmin(ctx context.Context, chatID, userID uuid.UUID) error {
	members, err := s.chatRepo.GetMembers(ctx, chatID)
	if err != nil {
		return fmt.Errorf("get members: %w", err)
	}

	for _, m := range members {
		if m.UserID == userID {
			if m.Role != model.MemberRoleAdmin {
				return apperror.Forbidden("only admins can perform this action")
			}
			return nil
		}
	}

	return apperror.Forbidden("you are not a member of this group")
}

// sendSystemMessage creates a system-type message in the chat.
func (s *groupService) sendSystemMessage(ctx context.Context, chatID, senderID uuid.UUID, content string) {
	_, _ = s.messageRepo.Create(ctx, model.CreateMessageInput{
		ChatID:   chatID,
		SenderID: senderID,
		Content:  content,
		Type:     model.MessageTypeSystem,
	})
}

// broadcastGroupEvent broadcasts a group event via WebSocket.
func (s *groupService) broadcastGroupEvent(chatID uuid.UUID, eventType string, payload map[string]interface{}) {
	roomID := "chat:" + chatID.String()
	event := map[string]interface{}{
		"type":    eventType,
		"payload": payload,
	}
	data, err := json.Marshal(event)
	if err == nil {
		s.hub.SendToRoom(roomID, data, uuid.Nil)
	}
}
