package model

import (
	"time"

	"github.com/google/uuid"
)

// Topic represents a discussion topic within a chat.
type Topic struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	Description string    `json:"description,omitempty"`
	ParentType  ChatType  `json:"parentType"`
	ParentID    uuid.UUID `json:"parentId"`
	CreatedBy   uuid.UUID `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TopicMember represents a user's membership in a topic.
type TopicMember struct {
	TopicID  uuid.UUID  `json:"topicId"`
	UserID   uuid.UUID  `json:"userId"`
	Role     MemberRole `json:"role"`
	JoinedAt time.Time  `json:"joinedAt"`
}

// CreateTopicInput holds data needed to create a new topic.
type CreateTopicInput struct {
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	ParentType  ChatType  `json:"parentType"`
	ParentID    uuid.UUID `json:"parentId"`
	CreatedBy   uuid.UUID `json:"createdBy"`
}

// UpdateTopicInput holds optional fields for updating a topic.
type UpdateTopicInput struct {
	Name        *string `json:"name"`
	Icon        *string `json:"icon"`
	Description *string `json:"description"`
}

// TopicMessage represents a message within a topic.
type TopicMessage struct {
	ID            uuid.UUID   `json:"id"`
	TopicID       uuid.UUID   `json:"topicId"`
	SenderID      uuid.UUID   `json:"senderId"`
	Content       string      `json:"content"`
	ReplyToID     *uuid.UUID  `json:"replyToId,omitempty"`
	Type          MessageType `json:"type"`
	IsDeleted     bool        `json:"isDeleted"`
	DeletedForAll bool        `json:"deletedForAll"`
	CreatedAt     time.Time   `json:"createdAt"`
}

// CreateTopicMessageInput holds data needed to create a topic message.
type CreateTopicMessageInput struct {
	TopicID   uuid.UUID   `json:"topicId"`
	SenderID  uuid.UUID   `json:"senderId"`
	Content   string      `json:"content"`
	ReplyToID *uuid.UUID  `json:"replyToId"`
	Type      MessageType `json:"type"`
}
