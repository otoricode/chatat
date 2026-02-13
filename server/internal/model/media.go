package model

import (
	"time"

	"github.com/google/uuid"
)

// MediaType represents the type of uploaded media.
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeFile  MediaType = "file"
)

// Media represents an uploaded file stored in S3-compatible storage.
type Media struct {
	ID           uuid.UUID  `json:"id"`
	UploaderID   uuid.UUID  `json:"uploaderId"`
	Type         MediaType  `json:"type"`
	Filename     string     `json:"filename"`
	ContentType  string     `json:"contentType"`
	Size         int        `json:"size"`
	Width        *int       `json:"width,omitempty"`
	Height       *int       `json:"height,omitempty"`
	StorageKey   string     `json:"storageKey"`
	ThumbnailKey *string    `json:"thumbnailKey,omitempty"`
	ContextType  *string    `json:"contextType,omitempty"`
	ContextID    *uuid.UUID `json:"contextId,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// MediaResponse is the API response for media info.
type MediaResponse struct {
	ID           uuid.UUID `json:"id"`
	Type         MediaType `json:"type"`
	Filename     string    `json:"filename"`
	ContentType  string    `json:"contentType"`
	Size         int       `json:"size"`
	Width        *int      `json:"width,omitempty"`
	Height       *int      `json:"height,omitempty"`
	URL          string    `json:"url"`
	ThumbnailURL string    `json:"thumbnailUrl,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}
