package model

import (
	"time"

	"github.com/google/uuid"
)

// Reminder represents a notification for an event.
// It includes the user and event IDs, the message (event title), and the time to send the reminder.
type Reminder struct {
	UserID   uuid.UUID // identifier of the user to receive the reminder
	EventID  uuid.UUID // identifier of the associated event
	Message  string    // message content, typically the event title
	RemindAt time.Time // time when the reminder should be sent
}
