package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/ws"
	"github.com/otoritech/chatat/pkg/apperror"
)

// ChatListItem represents a chat in the user's chat list with metadata.
type ChatListItem struct {
	Chat        model.Chat     `json:"chat"`
	LastMessage *model.Message `json:"lastMessage"`
	UnreadCount int            `json:"unreadCount"`
	OtherUser   *model.User    `json:"otherUser,omitempty"` // for personal chats
	IsOnline    bool           `json:"isOnline"`
}

// ChatDetail represents detailed chat information with members.
type ChatDetail struct {
	Chat    model.Chat    `json:"chat"`
	Members []*model.User `json:"members"`
}

// ChatService defines operations for chat management.
type ChatService interface {
	CreatePersonalChat(ctx context.Context, userID, contactID uuid.UUID) (*model.Chat, error)
	GetOrCreatePersonalChat(ctx context.Context, userID, contactID uuid.UUID) (*model.Chat, error)
	ListChats(ctx context.Context, userID uuid.UUID) ([]*ChatListItem, error)
	GetChat(ctx context.Context, chatID, userID uuid.UUID) (*ChatDetail, error)
	PinChat(ctx context.Context, chatID, userID uuid.UUID) error
	UnpinChat(ctx context.Context, chatID, userID uuid.UUID) error
	IsMember(ctx context.Context, chatID, userID uuid.UUID) (bool, error)
}

type chatService struct {
	chatRepo        repository.ChatRepository
	messageRepo     repository.MessageRepository
	messageStatRepo repository.MessageStatusRepository
	userRepo        repository.UserRepository
	hub             *ws.Hub
}

// NewChatService creates a new ChatService.
func NewChatService(
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
	messageStatRepo repository.MessageStatusRepository,
	userRepo repository.UserRepository,
	hub *ws.Hub,
) ChatService {
	return &chatService{
		chatRepo:        chatRepo,
		messageRepo:     messageRepo,
		messageStatRepo: messageStatRepo,
		userRepo:        userRepo,
		hub:             hub,
	}
}

func (s *chatService) CreatePersonalChat(ctx context.Context, userID, contactID uuid.UUID) (*model.Chat, error) {
	if userID == contactID {
		return nil, apperror.BadRequest("cannot create chat with yourself")
	}

	// Check if contact user exists
	_, err := s.userRepo.FindByID(ctx, contactID)
	if err != nil {
		if apperror.IsNotFound(err) {
			return nil, apperror.NotFound("user", contactID.String())
		}
		return nil, fmt.Errorf("find contact user: %w", err)
	}

	// Create chat
	chat, err := s.chatRepo.Create(ctx, model.CreateChatInput{
		Type:      model.ChatTypePersonal,
		CreatedBy: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("create personal chat: %w", err)
	}

	// Add both users as members
	if err := s.chatRepo.AddMember(ctx, chat.ID, userID, model.MemberRoleAdmin); err != nil {
		return nil, fmt.Errorf("add creator member: %w", err)
	}
	if err := s.chatRepo.AddMember(ctx, chat.ID, contactID, model.MemberRoleMember); err != nil {
		return nil, fmt.Errorf("add contact member: %w", err)
	}

	return chat, nil
}

func (s *chatService) GetOrCreatePersonalChat(ctx context.Context, userID, contactID uuid.UUID) (*model.Chat, error) {
	if userID == contactID {
		return nil, apperror.BadRequest("cannot create chat with yourself")
	}

	// Try to find existing personal chat
	chat, err := s.chatRepo.FindPersonalChat(ctx, userID, contactID)
	if err == nil {
		return chat, nil
	}

	// If not found, create new
	if apperror.IsNotFound(err) {
		return s.CreatePersonalChat(ctx, userID, contactID)
	}

	return nil, fmt.Errorf("find personal chat: %w", err)
}

func (s *chatService) ListChats(ctx context.Context, userID uuid.UUID) ([]*ChatListItem, error) {
	chatsWithMsg, err := s.chatRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list chats: %w", err)
	}

	items := make([]*ChatListItem, 0, len(chatsWithMsg))
	for _, cwm := range chatsWithMsg {
		item := &ChatListItem{
			Chat: cwm.Chat,
		}

		// Get unread count
		unread, err := s.messageStatRepo.GetUnreadCount(ctx, cwm.Chat.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("get unread count: %w", err)
		}
		item.UnreadCount = unread

		// Get last message
		messages, err := s.messageRepo.ListByChat(ctx, cwm.Chat.ID, nil, 1)
		if err != nil {
			return nil, fmt.Errorf("get last message: %w", err)
		}
		if len(messages) > 0 {
			item.LastMessage = messages[0]
		}

		// For personal chats, include other user info
		if cwm.Chat.Type == model.ChatTypePersonal {
			members, err := s.chatRepo.GetMembers(ctx, cwm.Chat.ID)
			if err != nil {
				return nil, fmt.Errorf("get chat members: %w", err)
			}
			for _, m := range members {
				if m.UserID != userID {
					otherUser, err := s.userRepo.FindByID(ctx, m.UserID)
					if err != nil {
						if apperror.IsNotFound(err) {
							continue
						}
						return nil, fmt.Errorf("get other user: %w", err)
					}
					item.OtherUser = otherUser
					item.IsOnline = s.hub.IsOnline(otherUser.ID)
					break
				}
			}
		}

		items = append(items, item)
	}

	// Sort: pinned first, then by last message time DESC
	sortChatList(items)

	return items, nil
}

func (s *chatService) GetChat(ctx context.Context, chatID, userID uuid.UUID) (*ChatDetail, error) {
	// Verify user is member
	isMember, err := s.IsMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, apperror.Forbidden("you are not a member of this chat")
	}

	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat: %w", err)
	}

	members, err := s.chatRepo.GetMembers(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	// Fetch user profiles for all members
	users := make([]*model.User, 0, len(members))
	for _, m := range members {
		user, err := s.userRepo.FindByID(ctx, m.UserID)
		if err != nil {
			if apperror.IsNotFound(err) {
				continue
			}
			return nil, fmt.Errorf("get member user: %w", err)
		}
		users = append(users, user)
	}

	return &ChatDetail{
		Chat:    *chat,
		Members: users,
	}, nil
}

func (s *chatService) PinChat(ctx context.Context, chatID, userID uuid.UUID) error {
	isMember, err := s.IsMember(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return apperror.Forbidden("you are not a member of this chat")
	}

	_, err = s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("find chat: %w", err)
	}

	// Update pinned_at timestamp
	now := "NOW()"
	_ = now
	// Use the repository's raw update to set pinned_at
	if err := s.chatRepo.Pin(ctx, chatID); err != nil {
		return fmt.Errorf("pin chat: %w", err)
	}
	return nil
}

func (s *chatService) UnpinChat(ctx context.Context, chatID, userID uuid.UUID) error {
	isMember, err := s.IsMember(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return apperror.Forbidden("you are not a member of this chat")
	}

	if err := s.chatRepo.Unpin(ctx, chatID); err != nil {
		return fmt.Errorf("unpin chat: %w", err)
	}
	return nil
}

func (s *chatService) IsMember(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	members, err := s.chatRepo.GetMembers(ctx, chatID)
	if err != nil {
		if apperror.IsNotFound(err) {
			return false, apperror.NotFound("chat", chatID.String())
		}
		return false, fmt.Errorf("check membership: %w", err)
	}
	for _, m := range members {
		if m.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

// sortChatList sorts chats: pinned first, then by last message time DESC.
func sortChatList(items []*ChatListItem) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0; j-- {
			if shouldSwapChat(items[j-1], items[j]) {
				items[j-1], items[j] = items[j], items[j-1]
			} else {
				break
			}
		}
	}
}

func shouldSwapChat(a, b *ChatListItem) bool {
	// Pinned always comes first
	aPinned := a.Chat.PinnedAt != nil
	bPinned := b.Chat.PinnedAt != nil

	if aPinned && !bPinned {
		return false
	}
	if !aPinned && bPinned {
		return true
	}

	// Both pinned or both unpinned: sort by last message time DESC
	aTime := a.Chat.UpdatedAt
	bTime := b.Chat.UpdatedAt

	if a.LastMessage != nil {
		aTime = a.LastMessage.CreatedAt
	}
	if b.LastMessage != nil {
		bTime = b.LastMessage.CreatedAt
	}

	return aTime.Before(bTime)
}
