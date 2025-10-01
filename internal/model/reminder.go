package model

import (
	"time"

	"github.com/google/uuid"
)

type Reminder struct {
	UserID   uuid.UUID
	EventID  uuid.UUID
	Message  string // event title
	RemindAt time.Time
}
