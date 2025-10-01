package model

import (
	"time"

	"github.com/google/uuid"
)

// Event represents an event in the calendar service.
// It contains details about the event, including its unique ID, associated user,
// date, title, description, optional reminder time, and timestamps for creation and updates.
type Event struct {
	ID          uuid.UUID  `json:"id"`          // unique identifier for the event
	UserID      uuid.UUID  `json:"user_id"`     // identifier of the user who owns the event
	EventDate   time.Time  `json:"event_date"`  // date and time when the event occurs
	Title       string     `json:"title"`       // title of the event
	Description string     `json:"description"` // optional description of the event
	ReminderAt  *time.Time `json:"reminder_at"` // optional time for sending a reminder
	CreatedAt   time.Time  `json:"created_at"`  // timestamp when the event was created
	UpdatedAt   time.Time  `json:"updated_at"`  // timestamp when the event was last updated
}
