package event

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/aliskhannn/calendar-service/internal/model"
)

//go:generate mockgen -source=service.go -destination=../../mocks/service/event/mock_event.go -package=mocks

// eventRepo defines the interface for event-related database operations.
// It provides methods for creating, updating, deleting, archiving, and retrieving events.
type eventRepo interface {
	// CreateEvent inserts a new event into the database and returns its ID.
	CreateEvent(ctx context.Context, event model.Event) (uuid.UUID, error)

	// UpdateEvent updates an existing event in the database.
	UpdateEvent(ctx context.Context, event model.Event) error

	// DeleteEvent removes an event from the database for the specified event and user IDs.
	DeleteEvent(ctx context.Context, eventID, userID uuid.UUID) error

	// ArchiveOldEvents moves old events to an archive table and deletes them from the events table.
	ArchiveOldEvents(ctx context.Context) error

	// GetEventsForDay retrieves all events for a user on a specific day.
	GetEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)

	// GetEventsForWeek retrieves all events for a user within a week from the given date.
	GetEventsForWeek(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)

	// GetEventsForMonth retrieves all events for a user within a month from the given date.
	GetEventsForMonth(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)
}

// Service manages business logic for event-related operations.
// It interacts with the event repository to perform CRUD operations and archiving.
type Service struct {
	eventRepo eventRepo // Repository for event database operations
}

// New creates a new Service instance with the provided event repository.
//
// Parameters:
//   - r: The event repository for database operations.
//
// Returns:
//   - A pointer to the initialized Service.
func New(r eventRepo) *Service {
	return &Service{
		eventRepo: r,
	}
}

// CreateEvent creates a new event for the specified user and returns its ID.
// It constructs an event model and delegates to the repository for database insertion.
//
// Parameters:
//   - ctx: The context for the operation.
//   - userID: The UUID of the user creating the event.
//   - title: The title of the event.
//   - description: The description of the event.
//   - date: The date and time of the event.
//   - reminderAt: The optional reminder time for the event.
//
// Returns:
//   - The UUID of the created event.
//   - An error if the creation fails.
func (s *Service) CreateEvent(ctx context.Context, userID uuid.UUID, title, description string, date time.Time, reminderAt *time.Time) (uuid.UUID, error) {
	event := model.Event{
		UserID:      userID,
		Title:       title,
		Description: description,
		EventDate:   date,
		ReminderAt:  reminderAt,
	}

	id, err := s.eventRepo.CreateEvent(ctx, event)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create event: %w", err)
	}

	return id, nil
}

// UpdateEvent updates an existing event for the specified user and event ID.
// It constructs an event model with updated fields and delegates to the repository.
//
// Parameters:
//   - ctx: The context for the operation.
//   - eventID: The UUID of the event to update.
//   - userID: The UUID of the user who owns the event.
//   - title: The updated title of the event.
//   - description: The updated description of the event.
//   - date: The updated date and time of the event.
//   - reminderAt: The updated optional reminder time for the event.
//
// Returns:
//   - An error if the update fails.
func (s *Service) UpdateEvent(ctx context.Context, eventID, userID uuid.UUID, title, description string, date time.Time, reminderAt *time.Time) error {
	event := model.Event{
		ID:          eventID,
		UserID:      userID,
		EventDate:   date,
		Title:       title,
		Description: description,
		ReminderAt:  reminderAt,
		UpdatedAt:   time.Now(),
	}

	err := s.eventRepo.UpdateEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}

	return nil
}

// DeleteEvent deletes an event for the specified user and event ID.
// It delegates to the repository to perform the deletion.
//
// Parameters:
//   - ctx: The context for the operation.
//   - eventID: The UUID of the event to delete.
//   - userID: The UUID of the user who owns the event.
//
// Returns:
//   - An error if the deletion fails.
func (s *Service) DeleteEvent(ctx context.Context, eventID, userID uuid.UUID) error {
	err := s.eventRepo.DeleteEvent(ctx, eventID, userID)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}

	return nil
}

// ArchiveOldEvents archives events older than the current date.
// It delegates to the repository to move old events to an archive table and delete them from the events table.
//
// Parameters:
//   - ctx: The context for the operation.
//
// Returns:
//   - An error if the archiving fails.
func (s *Service) ArchiveOldEvents(ctx context.Context) error {
	err := s.eventRepo.ArchiveOldEvents(ctx)
	if err != nil {
		return fmt.Errorf("archive old events: %w", err)
	}

	return nil
}

// GetEventsForDay retrieves all events for a specific user on a given day.
// It delegates to the repository to fetch the events.
//
// Parameters:
//   - ctx: The context for the operation.
//   - userID: The UUID of the user whose events are retrieved.
//   - date: The date for which to retrieve events.
//
// Returns:
//   - A slice of events for the specified day.
//   - An error if the retrieval fails.
func (s *Service) GetEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error) {
	events, err := s.eventRepo.GetEventsForDay(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("get events for day: %w", err)
	}

	return events, nil
}

// GetEventsForWeek retrieves all events for a specific user within a week from the given date.
// It delegates to the repository to fetch the events.
//
// Parameters:
//   - ctx: The context for the operation.
//   - userID: The UUID of the user whose events are retrieved.
//   - date: The reference date for the week.
//
// Returns:
//   - A slice of events for the specified week.
//   - An error if the retrieval fails.
func (s *Service) GetEventsForWeek(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error) {
	events, err := s.eventRepo.GetEventsForWeek(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("get events for week: %w", err)
	}

	return events, nil
}

// GetEventsForMonth retrieves all events for a specific user within a month from the given date.
// It delegates to the repository to fetch the events.
//
// Parameters:
//   - ctx: The context for the operation.
//   - userID: The UUID of the user whose events are retrieved.
//   - date: The reference date for the month.
//
// Returns:
//   - A slice of events for the specified month.
//   - An error if the retrieval fails.
func (s *Service) GetEventsForMonth(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error) {
	events, err := s.eventRepo.GetEventsForMonth(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("get events for month: %w", err)
	}

	return events, nil
}
