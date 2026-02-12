package model

import (
	"time"

	"github.com/google/uuid"
)

// DeliveryStatus represents the delivery status of a message.
type DeliveryStatus string

const (
	DeliveryStatusSent      DeliveryStatus = "sent"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusRead      DeliveryStatus = "read"
)

// MessageStatus tracks delivery/read status of a message for a user.
type MessageStatus struct {
	MessageID uuid.UUID      `json:"messageId"`
	UserID    uuid.UUID      `json:"userId"`
	Status    DeliveryStatus `json:"status"`
	UpdatedAt time.Time      `json:"updatedAt"`
}
