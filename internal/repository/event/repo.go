package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"

	"github.com/aliskhannn/calendar-service/internal/model"
)

var (
	ErrEventNotFound = errors.New("event not found")
)

// Repository manages interactions with the events table in the PostgreSQL database.
// It provides methods for creating, updating, deleting, archiving, and retrieving events.
type Repository struct {
	db *pgxpool.Pool // Database connection pool
}

// New creates a new Repository instance with the provided database connection pool.
//
// Parameters:
//   - db: The PostgreSQL connection pool for database operations.
//
// Returns:
//   - A pointer to the initialized Repository.
func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

// CreateEvent inserts a new event into the events table and returns its ID.
// It stores the user ID, event date, title, description, and optional reminder time.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - event: The event data to be inserted.
//
// Returns:
//   - The UUID of the created event.
//   - An error if the insertion fails.
func (r *Repository) CreateEvent(ctx context.Context, event model.Event) (uuid.UUID, error) {
	query := `
		INSERT INTO events (
		    user_id, event_date, title, description, reminder_at
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
    `

	err := r.db.QueryRow(
		ctx, query, event.UserID, event.EventDate, event.Title, event.Description, event.ReminderAt,
	).Scan(&event.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create event: %w", err)
	}

	return event.ID, nil
}

// UpdateEvent updates an existing event in the events table.
// It updates the event date, title, description, reminder time, and updated_at timestamp
// for the specified event ID and user ID.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - event: The event data containing updated fields.
//
// Returns:
//   - An error if the update fails or if the event is not found.
func (r *Repository) UpdateEvent(ctx context.Context, event model.Event) error {
	query := `
		UPDATE events
		SET
		    event_date = $1,
			title = $2,
			description = $3,
			reminder_at = $4,
			updated_at = now()
		WHERE id = $5 AND user_id = $6;
	`

	cmdTag, err := r.db.Exec(ctx, query, event.EventDate, event.Title, event.Description, event.ReminderAt, event.ID, event.UserID)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrEventNotFound
	}

	return nil
}

// DeleteEvent deletes an event from the events table.
// It removes the event with the specified ID and user ID.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - eventID: The UUID of the event to delete.
//   - userID: The UUID of the user who owns the event.
//
// Returns:
//   - An error if the deletion fails or if the event is not found.
func (r *Repository) DeleteEvent(ctx context.Context, eventID, userID uuid.UUID) error {
	query := `
   		DELETE FROM events
   		WHERE id = $1 AND user_id = $2;
    `

	cmdTag, err := r.db.Exec(ctx, query, eventID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrEventNotFound
	}

	return nil
}

// ArchiveOldEvents moves events older than the current date to the archived_events table
// and deletes them from the events table. It uses a transaction to ensure atomicity.
//
// Parameters:
//   - ctx: The context for the database operation.
//
// Returns:
//   - An error if the archiving or deletion fails, or if the transaction cannot be committed.
func (r *Repository) ArchiveOldEvents(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert old events into archived_events table.
	_, err = tx.Exec(ctx, `
        INSERT INTO archived_events (id, user_id, event_date, title, description, created_at, updated_at)
        SELECT id, user_id, event_date, title, description, created_at, updated_at
        FROM events
        WHERE event_date < CURRENT_DATE
    `)
	if err != nil {
		return fmt.Errorf("failed to insert old events: %w", err)
	}

	// Delete old events from events table.
	_, err = tx.Exec(ctx, `DELETE FROM events WHERE event_date < CURRENT_DATE`)
	if err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}

	// Commit the transaction.
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetEventsForDay retrieves all events for a specific user on a given day.
// Events are ordered by their event_date.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - userID: The UUID of the user whose events are retrieved.
//   - date: The date for which to retrieve events.
//
// Returns:
//   - A slice of events for the specified day.
//   - An error if the query fails or if no events are found.
func (r *Repository) GetEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error) {
	query := `
		SELECT id, user_id, event_date, title, description, reminder_at, created_at, updated_at
		FROM events
		WHERE user_id = $1 AND event_date = $2
		ORDER BY event_date
    `

	rows, err := r.db.Query(ctx, query, userID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for day: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.UserID, &e.EventDate, &e.Title, &e.Description, &e.ReminderAt, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	if len(events) == 0 {
		return nil, ErrEventNotFound
	}

	return events, nil
}

// GetEventsForWeek retrieves all events for a specific user within a week starting from the given date.
// The week is defined as 7 days before and 1 day after the specified date. Events are ordered by event_date.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - userID: The UUID of the user whose events are retrieved.
//   - date: The reference date for the week.
//
// Returns:
//   - A slice of events for the specified week.
//   - An error if the query fails or if no events are found.
func (r *Repository) GetEventsForWeek(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error) {
	start := date.AddDate(0, 0, -7)
	end := date.AddDate(0, 0, 1)

	query := `
		SELECT id, user_id, event_date, title, description, reminder_at, created_at, updated_at
		FROM events
		WHERE user_id = $1 AND event_date >= $2 AND event_date < $3
		ORDER BY event_date
    `

	rows, err := r.db.Query(ctx, query, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for week: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.UserID, &e.EventDate, &e.Title, &e.Description, &e.ReminderAt, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	if len(events) == 0 {
		return nil, ErrEventNotFound
	}

	return events, nil
}

// GetEventsForMonth retrieves all events for a specific user within a month starting from the first day of the given date's month.
// The month ends before the first day of the next month. Events are ordered by event_date.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - userID: The UUID of the user whose events are retrieved.
//   - date: The reference date for the month.
//
// Returns:
//   - A slice of events for the specified month.
//   - An error if the query fails or if no events are found.
func (r *Repository) GetEventsForMonth(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error) {
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	end := date.AddDate(0, 1, 0)

	query := `
		SELECT id, user_id, event_date, title, description, reminder_at, created_at, updated_at
		FROM events
		WHERE user_id = $1 AND event_date >= $2 AND event_date < $3
		ORDER BY event_date
    `

	rows, err := r.db.Query(ctx, query, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for month: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.UserID, &e.EventDate, &e.Title, &e.Description, &e.ReminderAt, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	if len(events) == 0 {
		return nil, ErrEventNotFound
	}

	return events, nil
}
