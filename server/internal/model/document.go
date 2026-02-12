package model

import (
	"time"

	"github.com/google/uuid"
)

// CollaboratorRole represents a document collaborator's role.
type CollaboratorRole string

const (
	CollaboratorRoleEditor CollaboratorRole = "editor"
	CollaboratorRoleViewer CollaboratorRole = "viewer"
)

// LockedByType represents how a document was locked.
type LockedByType string

const (
	LockedByManual     LockedByType = "manual"
	LockedBySignatures LockedByType = "signatures"
)

// Document represents a collaborative document.
type Document struct {
	ID           uuid.UUID     `json:"id"`
	Title        string        `json:"title"`
	Icon         string        `json:"icon"`
	Cover        *string       `json:"cover,omitempty"`
	OwnerID      uuid.UUID     `json:"ownerId"`
	ChatID       *uuid.UUID    `json:"chatId,omitempty"`
	TopicID      *uuid.UUID    `json:"topicId,omitempty"`
	IsStandalone bool          `json:"isStandalone"`
	RequireSigs  bool          `json:"requireSigs"`
	Locked       bool          `json:"locked"`
	LockedAt     *time.Time    `json:"lockedAt,omitempty"`
	LockedBy     *LockedByType `json:"lockedBy,omitempty"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

// DocumentCollaborator represents a user collaborating on a document.
type DocumentCollaborator struct {
	DocumentID uuid.UUID        `json:"documentId"`
	UserID     uuid.UUID        `json:"userId"`
	Role       CollaboratorRole `json:"role"`
	AddedAt    time.Time        `json:"addedAt"`
}

// DocumentSigner represents a user who can sign a document.
type DocumentSigner struct {
	DocumentID uuid.UUID  `json:"documentId"`
	UserID     uuid.UUID  `json:"userId"`
	SignedAt   *time.Time `json:"signedAt,omitempty"`
	SignerName string     `json:"signerName,omitempty"`
}

// CreateDocumentInput holds data needed to create a new document.
type CreateDocumentInput struct {
	Title        string     `json:"title"`
	Icon         string     `json:"icon"`
	OwnerID      uuid.UUID  `json:"ownerId"`
	ChatID       *uuid.UUID `json:"chatId"`
	TopicID      *uuid.UUID `json:"topicId"`
	IsStandalone bool       `json:"isStandalone"`
}

// UpdateDocumentInput holds optional fields for updating a document.
type UpdateDocumentInput struct {
	Title       *string `json:"title"`
	Icon        *string `json:"icon"`
	Cover       *string `json:"cover"`
	RequireSigs *bool   `json:"requireSigs"`
}

// DocumentHistory records an action performed on a document.
type DocumentHistory struct {
	ID         uuid.UUID `json:"id"`
	DocumentID uuid.UUID `json:"documentId"`
	UserID     uuid.UUID `json:"userId"`
	Action     string    `json:"action"`
	Details    string    `json:"details,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}
