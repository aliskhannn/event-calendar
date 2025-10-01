package model

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	EventDate   time.Time  `json:"event_date"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ReminderAt  *time.Time `json:"reminder_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
