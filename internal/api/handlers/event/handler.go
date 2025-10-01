package event

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aliskhannn/calendar-service/internal/model"
)

//go:generate mockgen -source=handler.go -destination=../../../mocks/api/handlers/event/mock_event_service.go -package=mocks

// eventService defines the interface for event-related operations.
// It provides methods for creating, updating, deleting, and retrieving events for a user.
type eventService interface {
	// CreateEvent creates a new event for the specified user and returns the event ID.
	CreateEvent(ctx context.Context, userID uuid.UUID, title, description string, date time.Time, reminderAt *time.Time) (uuid.UUID, error)

	// UpdateEvent updates an existing event for the specified user and event ID.
	UpdateEvent(ctx context.Context, eventID, userID uuid.UUID, title, description string, date time.Time, reminderAt *time.Time) error

	// DeleteEvent deletes an event for the specified user and event ID.
	DeleteEvent(ctx context.Context, eventID, userID uuid.UUID) error

	// GetEventsForDay retrieves all events for a specific user on a given day.
	GetEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)

	// GetEventsForWeek retrieves all events for a specific user within a week starting from the given date.
	GetEventsForWeek(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)

	// GetEventsForMonth retrieves all events for a specific user within a month starting from the given date.
	GetEventsForMonth(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)
}

// Handler manages HTTP requests for event-related operations.
// It encapsulates the event service, reminder channel, logger, and validator for handling requests.
type Handler struct {
	service    eventService          // service handles business logic for event operations
	reminderCh chan<- model.Reminder // reminderCh sends reminders for events
	logger     *zap.Logger           // logger logs application events and errors
	validator  *validator.Validate   // validator validates incoming request data
}

// New creates a new Handler instance with the provided dependencies.
// It initializes the Handler with an event service, reminder channel, logger, and validator.
//
// Parameters:
//   - s: The event service for handling event-related operations.
//   - reminderCh: The channel for sending event reminders.
//   - l: The logger for logging application events and errors.
//   - v: The validator for validating request data.
//
// Returns:
//   - A pointer to the initialized Handler.
func New(
	s eventService,
	reminderCh chan<- model.Reminder,
	l *zap.Logger,
	v *validator.Validate,
) *Handler {
	return &Handler{
		service:    s,
		reminderCh: reminderCh,
		logger:     l,
		validator:  v,
	}
}
