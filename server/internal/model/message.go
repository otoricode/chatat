package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageType represents the type of a message.
type MessageType string

const (
	MessageTypeText         MessageType = "text"
	MessageTypeImage        MessageType = "image"
	MessageTypeFile         MessageType = "file"
	MessageTypeDocumentCard MessageType = "document_card"
	MessageTypeSystem       MessageType = "system"
)

// Message represents a chat message.
type Message struct {
	ID            uuid.UUID       `json:"id"`
	ChatID        uuid.UUID       `json:"chatId"`
	SenderID      uuid.UUID       `json:"senderId"`
	Content       string          `json:"content"`
	ReplyToID     *uuid.UUID      `json:"replyToId,omitempty"`
	Type          MessageType     `json:"type"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
	IsDeleted     bool            `json:"isDeleted"`
	DeletedForAll bool            `json:"deletedForAll"`
	CreatedAt     time.Time       `json:"createdAt"`
}

// CreateMessageInput holds data needed to create a new message.
type CreateMessageInput struct {
	ChatID    uuid.UUID       `json:"chatId"`
	SenderID  uuid.UUID       `json:"senderId"`
	Content   string          `json:"content"`
	ReplyToID *uuid.UUID      `json:"replyToId"`
	Type      MessageType     `json:"type"`
	Metadata  json.RawMessage `json:"metadata"`
}
