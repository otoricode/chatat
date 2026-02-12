package model

import (
	"time"

	"github.com/google/uuid"
)

// Entity represents a real-world entity that can be referenced in documents.
type Entity struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	OwnerID       uuid.UUID  `json:"ownerId"`
	ContactUserID *uuid.UUID `json:"contactUserId,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// CreateEntityInput holds data needed to create a new entity.
type CreateEntityInput struct {
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	OwnerID       uuid.UUID  `json:"ownerId"`
	ContactUserID *uuid.UUID `json:"contactUserId"`
}
