package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the current state of a user account.
// @Description UserStatus can be active, inactive, or deleted.
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusDeleted  UserStatus = "deleted"
)

// User represents a registered user in the system.
// @Description User stores personal information and account status.
// @Description Passwords are stored as hashes for security.
type User struct {
	ID           int        `json:"-"`
	UUID         uuid.UUID  `json:"userId"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	PasswordHash string     `json:"-"`
	Status       UserStatus `json:"status"`
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}
