package model

import (
	"time"

	"github.com/google/uuid"
)

// MessageSearchRow represents a message search result from the database.
type MessageSearchRow struct {
	ID         uuid.UUID   `json:"id"`
	ChatID     uuid.UUID   `json:"chatId"`
	SenderID   uuid.UUID   `json:"senderId"`
	Content    string      `json:"content"`
	Type       MessageType `json:"type"`
	CreatedAt  time.Time   `json:"createdAt"`
	ChatName   string      `json:"chatName"`
	SenderName string      `json:"senderName"`
	Highlight  string      `json:"highlight"`
}

// DocumentSearchRow represents a document search result from the database.
type DocumentSearchRow struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Icon      string    `json:"icon"`
	OwnerID   uuid.UUID `json:"ownerId"`
	Locked    bool      `json:"locked"`
	UpdatedAt time.Time `json:"updatedAt"`
	Highlight string    `json:"highlight"`
}
