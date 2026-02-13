package model

import (
	"time"

	"github.com/google/uuid"
)

// DeviceToken represents a user's device push notification token.
type DeviceToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Token     string    `json:"token"`
	Platform  string    `json:"platform"` // ios, android
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NotificationType defines the type of push notification.
type NotificationType string

const (
	NotifTypeMessage          NotificationType = "message"
	NotifTypeGroupMessage     NotificationType = "group_message"
	NotifTypeTopicMessage     NotificationType = "topic_message"
	NotifTypeSignatureRequest NotificationType = "signature_request"
	NotifTypeDocumentLocked   NotificationType = "document_locked"
	NotifTypeGroupInvite      NotificationType = "group_invite"
)

// Notification represents a push notification payload.
type Notification struct {
	Type     NotificationType  `json:"type"`
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Data     map[string]string `json:"data"`
	Badge    int               `json:"badge,omitempty"`
	Sound    string            `json:"sound,omitempty"`
	Priority string            `json:"priority,omitempty"` // high, normal
}
