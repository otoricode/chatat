// Package model defines domain models and input types for the application.
package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user in the system.
type User struct {
	ID        uuid.UUID `json:"id"`
	Phone     string    `json:"phone"`
	PhoneHash string    `json:"-"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	Status    string    `json:"status"`
	Language  string    `json:"language"`
	LastSeen  time.Time `json:"lastSeen"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateUserInput holds the data needed to create a new user.
type CreateUserInput struct {
	Phone  string `json:"phone"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// UpdateUserInput holds optional fields for updating a user.
type UpdateUserInput struct {
	Name     *string `json:"name"`
	Avatar   *string `json:"avatar"`
	Status   *string `json:"status"`
	Language *string `json:"language"`
}
