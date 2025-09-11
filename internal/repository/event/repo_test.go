package event

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"

	"github.com/aliskhannn/calendar-service/internal/model"
)

func newTestRepo(t *testing.T) (*Repository, pgxmock.PgxPoolIface) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock: %v", err)
	}
	return New(mock), mock
}

func TestRepository_CreateEvent(t *testing.T) {
	repo, mock := newTestRepo(t)
	defer mock.Close()

	id := uuid.New()
	event := model.Event{
		UserID:      uuid.New(),
		Title:       "Test event",
		Description: "desc",
		EventDate:   time.Now(),
	}

	mock.ExpectQuery("INSERT INTO events").
		WithArgs(event.UserID, event.EventDate, event.Title, event.Description).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(id))

	gotID, err := repo.CreateEvent(context.Background(), event)
	assert.NoError(t, err)
	assert.Equal(t, id, gotID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateEvent(t *testing.T) {
	repo, mock := newTestRepo(t)
	defer mock.Close()

	event := model.Event{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Title:       "Updated",
		Description: "new desc",
		EventDate:   time.Now(),
	}

	mock.ExpectExec("UPDATE events").
		WithArgs(event.EventDate, event.Title, event.Description, event.ID, event.UserID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := repo.UpdateEvent(context.Background(), event)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteEvent_NotFound(t *testing.T) {
	repo, mock := newTestRepo(t)
	defer mock.Close()

	eventID := uuid.New()
	userID := uuid.New()

	mock.ExpectExec("DELETE FROM events").
		WithArgs(eventID, userID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.DeleteEvent(context.Background(), eventID, userID)
	assert.ErrorIs(t, err, ErrEventNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetEventsForDay(t *testing.T) {
	repo, mock := newTestRepo(t)
	defer mock.Close()

	userID := uuid.New()
	date := time.Now()
	id := uuid.New()

	mock.ExpectQuery("SELECT id, user_id, event_date, title, description, created_at, updated_at FROM events").
		WithArgs(userID, date).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "user_id", "event_date", "title", "description", "created_at", "updated_at"}).
				AddRow(id, userID, date, "Meeting", "Discuss", time.Now(), time.Now()),
		)

	events, err := repo.GetEventsForDay(context.Background(), userID, date)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "Meeting", events[0].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}
