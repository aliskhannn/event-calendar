//go:build integration
// +build integration

package event

import (
	"context"
	"errors"
	"github.com/joho/godotenv"
	"log"
	"os"
	"testing"
	"time"

	"github.com/aliskhannn/calendar-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testRepo *Repository
var testUserID uuid.UUID

func TestMain(m *testing.M) {
	_ = godotenv.Load(".env.test")
	// Connecting to a test database.
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		log.Fatal("TEST_DATABASE_URL is not set")
	}
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	testRepo = New(db)
	testUserID = uuid.New()

	// Очистка таблицы перед тестами
	db.Exec(context.Background(), "DELETE FROM events")

	os.Exit(m.Run())
}

func createTestEvent(t *testing.T, title string) model.Event {
	e := model.Event{
		UserID:    testUserID,
		Title:     title,
		EventDate: time.Now().Truncate(time.Second),
	}
	id, err := testRepo.CreateEvent(context.Background(), e)
	if err != nil {
		t.Fatalf("failed to create test event: %v", err)
	}
	e.ID = id
	return e
}

func TestRepository_CreateUpdateDelete(t *testing.T) {
	// Create
	testEvent := createTestEvent(t, "My Test Event")

	// Update
	testEvent.Title = "Updated Title"
	err := testRepo.UpdateEvent(context.Background(), testEvent)
	if err != nil {
		t.Fatalf("failed to update event: %v", err)
	}

	// Delete
	err = testRepo.DeleteEvent(context.Background(), testEvent.ID, testEvent.UserID)
	if err != nil {
		t.Fatalf("failed to delete event: %v", err)
	}

	// Delete again -> should return ErrEventNotFound
	err = testRepo.DeleteEvent(context.Background(), testEvent.ID, testEvent.UserID)
	if !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got: %v", err)
	}
}

func TestRepository_GetEvents(t *testing.T) {
	// Создаем два события для тестового пользователя
	event1 := createTestEvent(t, "Event 1")
	createTestEvent(t, "Event 2") // второе событие для недели и месяца

	// GetEventsForDay
	dayEvents, err := testRepo.GetEventsForDay(context.Background(), testUserID, event1.EventDate)
	if err != nil {
		t.Fatalf("GetEventsForDay failed: %v", err)
	}
	if len(dayEvents) == 0 {
		t.Fatalf("expected at least one event for the day")
	}

	// GetEventsForWeek
	weekEvents, err := testRepo.GetEventsForWeek(context.Background(), testUserID, time.Now())
	if err != nil {
		t.Fatalf("GetEventsForWeek failed: %v", err)
	}
	if len(weekEvents) < 2 {
		t.Fatalf("expected at least 2 events for the week, got %d", len(weekEvents))
	}

	// GetEventsForMonth
	monthEvents, err := testRepo.GetEventsForMonth(context.Background(), testUserID, time.Now())
	if err != nil {
		t.Fatalf("GetEventsForMonth failed: %v", err)
	}
	if len(monthEvents) < 2 {
		t.Fatalf("expected at least 2 events for the month, got %d", len(monthEvents))
	}
}
