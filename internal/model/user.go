package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the calendar service.
// It contains the user's unique ID, email, name, password (excluded from JSON),
// and timestamps for creation and updates.
type User struct {
	ID        uuid.UUID `json:"id"`         // unique identifier for the user
	Email     string    `json:"email"`      // user's email address
	Name      string    `json:"name"`       // user's name
	Password  string    `json:"-"`          // user's password (not serialized to JSON)
	CreatedAt time.Time `json:"created_at"` // timestamp when the user was created
	UpdatedAt time.Time `json:"updated_at"` // timestamp when the user was last updated
}
