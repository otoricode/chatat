package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// BackupPlatform represents the cloud storage platform.
type BackupPlatform string

const (
	// BackupPlatformGoogleDrive is Google Drive backup.
	BackupPlatformGoogleDrive BackupPlatform = "google_drive"
	// BackupPlatformICloud is iCloud backup.
	BackupPlatformICloud BackupPlatform = "icloud"
)

// BackupStatus represents the status of a backup operation.
type BackupStatus string

const (
	// BackupStatusInProgress means backup is currently running.
	BackupStatusInProgress BackupStatus = "in_progress"
	// BackupStatusCompleted means backup finished successfully.
	BackupStatusCompleted BackupStatus = "completed"
	// BackupStatusFailed means backup failed.
	BackupStatusFailed BackupStatus = "failed"
)

// BackupRecord represents a single backup operation record.
type BackupRecord struct {
	ID        uuid.UUID      `json:"id"`
	UserID    uuid.UUID      `json:"userId"`
	SizeBytes int64          `json:"sizeBytes"`
	Platform  BackupPlatform `json:"platform"`
	Status    BackupStatus   `json:"status"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

// LogBackupInput holds data for logging a backup record.
type LogBackupInput struct {
	SizeBytes int64          `json:"sizeBytes"`
	Platform  BackupPlatform `json:"platform"`
	Status    BackupStatus   `json:"status,omitempty"`
}

// BackupBundle is the export data container.
type BackupBundle struct {
	Version   int        `json:"version"`
	UserID    string     `json:"userId"`
	CreatedAt time.Time  `json:"createdAt"`
	Data      BackupData `json:"data"`
}

// BackupData holds all exported user data.
type BackupData struct {
	Profile   *UserExport      `json:"profile"`
	Chats     []ChatExport     `json:"chats"`
	Messages  []MessageExport  `json:"messages"`
	Contacts  []ContactExport  `json:"contacts"`
	Documents []DocumentExport `json:"documents"`
}

// UserExport is the user profile in a backup.
type UserExport struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Status string `json:"status"`
}

// ChatExport represents a chat in a backup bundle.
type ChatExport struct {
	ServerID    string `json:"serverId"`
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

// MessageExport represents a message in a backup bundle.
type MessageExport struct {
	ServerID  string `json:"serverId"`
	ChatID    string `json:"chatId"`
	SenderID  string `json:"senderId"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
}

// ContactExport represents a contact in a backup bundle.
type ContactExport struct {
	UserID    string `json:"userId"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Nickname  string `json:"nickname,omitempty"`
	AddedAt   string `json:"addedAt"`
}

// DocumentExport represents a document in a backup bundle.
type DocumentExport struct {
	ServerID string `json:"serverId"`
	Title    string `json:"title"`
	Icon     string `json:"icon"`
	ChatID   string `json:"chatId,omitempty"`
	TopicID  string `json:"topicId,omitempty"`
}
