// Package model defines domain models and input types for the application.
package model

import (
	"time"

	"github.com/google/uuid"
)

// PrivacySettings holds user privacy preferences.
type PrivacySettings struct {
	LastSeenVisibility     string `json:"lastSeenVisibility"`     // everyone, contacts, nobody
	OnlineVisibility       string `json:"onlineVisibility"`       // everyone, contacts, nobody
	ReadReceipts           bool   `json:"readReceipts"`           // show/hide blue checks
	ProfilePhotoVisibility string `json:"profilePhotoVisibility"` // everyone, contacts
}

// DefaultPrivacySettings returns privacy settings with default values.
func DefaultPrivacySettings() PrivacySettings {
	return PrivacySettings{
		LastSeenVisibility:     "everyone",
		OnlineVisibility:       "everyone",
		ReadReceipts:           true,
		ProfilePhotoVisibility: "everyone",
	}
}

// User represents a registered user in the system.
type User struct {
	ID              uuid.UUID       `json:"id"`
	Phone           string          `json:"phone"`
	PhoneHash       string          `json:"-"`
	Name            string          `json:"name"`
	Avatar          string          `json:"avatar"`
	Status          string          `json:"status"`
	Language        string          `json:"language"`
	PrivacySettings PrivacySettings `json:"privacySettings"`
	LastSeen        time.Time       `json:"lastSeen"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// CreateUserInput holds the data needed to create a new user.
type CreateUserInput struct {
	Phone  string `json:"phone"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// UpdateUserInput holds optional fields for updating a user.
type UpdateUserInput struct {
	Name            *string          `json:"name"`
	Avatar          *string          `json:"avatar"`
	Status          *string          `json:"status"`
	Language        *string          `json:"language"`
	PrivacySettings *PrivacySettings `json:"privacySettings"`
}
