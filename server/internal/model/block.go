package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// BlockType represents the type of a document block.
type BlockType string

const (
	BlockTypeParagraph    BlockType = "paragraph"
	BlockTypeHeading1     BlockType = "heading1"
	BlockTypeHeading2     BlockType = "heading2"
	BlockTypeHeading3     BlockType = "heading3"
	BlockTypeBulletList   BlockType = "bullet-list"
	BlockTypeNumberedList BlockType = "numbered-list"
	BlockTypeChecklist    BlockType = "checklist"
	BlockTypeTable        BlockType = "table"
	BlockTypeCallout      BlockType = "callout"
	BlockTypeCode         BlockType = "code"
	BlockTypeToggle       BlockType = "toggle"
	BlockTypeDivider      BlockType = "divider"
	BlockTypeQuote        BlockType = "quote"
)

// Block represents a content block within a document.
type Block struct {
	ID            uuid.UUID       `json:"id"`
	DocumentID    uuid.UUID       `json:"documentId"`
	Type          BlockType       `json:"type"`
	Content       string          `json:"content,omitempty"`
	Checked       *bool           `json:"checked,omitempty"`
	Rows          json.RawMessage `json:"rows,omitempty"`
	Columns       json.RawMessage `json:"columns,omitempty"`
	Language      string          `json:"language,omitempty"`
	Emoji         string          `json:"emoji,omitempty"`
	Color         string          `json:"color,omitempty"`
	SortOrder     int             `json:"sortOrder"`
	ParentBlockID *uuid.UUID      `json:"parentBlockId,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

// CreateBlockInput holds data needed to create a new block.
type CreateBlockInput struct {
	DocumentID    uuid.UUID       `json:"documentId"`
	Type          BlockType       `json:"type"`
	Content       string          `json:"content"`
	Checked       *bool           `json:"checked"`
	Rows          json.RawMessage `json:"rows"`
	Columns       json.RawMessage `json:"columns"`
	Language      string          `json:"language"`
	Emoji         string          `json:"emoji"`
	Color         string          `json:"color"`
	SortOrder     int             `json:"sortOrder"`
	ParentBlockID *uuid.UUID      `json:"parentBlockId"`
}

// UpdateBlockInput holds optional fields for updating a block.
type UpdateBlockInput struct {
	Content   *string         `json:"content"`
	Checked   *bool           `json:"checked"`
	Rows      json.RawMessage `json:"rows"`
	Columns   json.RawMessage `json:"columns"`
	Language  *string         `json:"language"`
	Emoji     *string         `json:"emoji"`
	Color     *string         `json:"color"`
	SortOrder *int            `json:"sortOrder"`
}
