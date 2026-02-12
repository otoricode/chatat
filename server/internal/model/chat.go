package model

import (
	"time"

	"github.com/google/uuid"
)

// ChatType represents the type of a chat.
type ChatType string

const (
	// ChatTypePersonal is a one-on-one chat.
	ChatTypePersonal ChatType = "personal"
	// ChatTypeGroup is a group chat.
	ChatTypeGroup ChatType = "group"
)

// MemberRole represents a member's role in a chat or topic.
type MemberRole string

const (
	// MemberRoleAdmin has administrative privileges.
	MemberRoleAdmin MemberRole = "admin"
	// MemberRoleMember is a regular member.
	MemberRoleMember MemberRole = "member"
)

// Chat represents a chat room (personal or group).
type Chat struct {
	ID          uuid.UUID  `json:"id"`
	Type        ChatType   `json:"type"`
	Name        string     `json:"name,omitempty"`
	Icon        string     `json:"icon,omitempty"`
	Description string     `json:"description,omitempty"`
	CreatedBy   uuid.UUID  `json:"createdBy"`
	PinnedAt    *time.Time `json:"pinnedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// ChatMember represents a user's membership in a chat.
type ChatMember struct {
	ChatID   uuid.UUID  `json:"chatId"`
	UserID   uuid.UUID  `json:"userId"`
	Role     MemberRole `json:"role"`
	JoinedAt time.Time  `json:"joinedAt"`
}

// CreateChatInput holds data needed to create a new chat.
type CreateChatInput struct {
	Type        ChatType  `json:"type"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	CreatedBy   uuid.UUID `json:"createdBy"`
}

// UpdateChatInput holds optional fields for updating a chat.
type UpdateChatInput struct {
	Name        *string `json:"name"`
	Icon        *string `json:"icon"`
	Description *string `json:"description"`
}

// ChatWithLastMessage combines a chat with its most recent message preview.
type ChatWithLastMessage struct {
	Chat        Chat    `json:"chat"`
	LastMessage *string `json:"lastMessage,omitempty"`
	UnreadCount int     `json:"unreadCount"`
}
