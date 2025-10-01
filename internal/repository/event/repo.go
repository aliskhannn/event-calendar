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

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

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

func (r *Repository) ArchiveOldEvents(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert into archive
	_, err = tx.Exec(ctx, `
        INSERT INTO archived_events (id, user_id, event_date, title, description, created_at, updated_at)
        SELECT id, user_id, event_date, title, description, created_at, updated_at
        FROM events
        WHERE event_date < CURRENT_DATE
    `)
	if err != nil {
		return fmt.Errorf("failed to insert old events: %w", err)
	}

	// Delete old events
	_, err = tx.Exec(ctx, `DELETE FROM events WHERE event_date < CURRENT_DATE`)
	if err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

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
