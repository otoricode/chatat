package model

import (
	"time"

	"github.com/google/uuid"
)

// Entity represents a real-world entity that can be referenced in documents.
type Entity struct {
	ID            uuid.UUID         `json:"id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Fields        map[string]string `json:"fields"`
	OwnerID       uuid.UUID         `json:"ownerId"`
	ContactUserID *uuid.UUID        `json:"contactUserId,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
}

// CreateEntityInput holds data needed to create a new entity.
type CreateEntityInput struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Fields        map[string]string `json:"fields"`
	OwnerID       uuid.UUID         `json:"ownerId"`
	ContactUserID *uuid.UUID        `json:"contactUserId"`
}

// UpdateEntityInput holds data for updating an entity.
type UpdateEntityInput struct {
	Name   *string            `json:"name,omitempty"`
	Type   *string            `json:"type,omitempty"`
	Fields *map[string]string `json:"fields,omitempty"`
}

// EntityListItem extends Entity with document count.
type EntityListItem struct {
	Entity
	DocumentCount int `json:"documentCount"`
}
