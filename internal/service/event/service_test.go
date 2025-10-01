package event

import (
	"context"
	"testing"
	"time"

	eventrepomocks "github.com/aliskhannn/calendar-service/internal/mocks/service/event"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/aliskhannn/calendar-service/internal/model"
)

func TestService_CreateEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := eventrepomocks.NewMockeventRepo(ctrl)
	svc := New(mockRepo)

	userID := uuid.New()
	date := time.Now()
	title := "Test Event"
	description := "Some description"
	mockID := uuid.New()

	expectedEvent := model.Event{
		UserID:      userID,
		Title:       title,
		Description: description,
		EventDate:   date,
	}

	mockRepo.EXPECT().
		CreateEvent(gomock.Any(), expectedEvent).
		Return(mockID, nil)

	id, err := svc.CreateEvent(context.Background(), userID, title, description, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != mockID {
		t.Fatalf("expected id %v, got %v", mockID, id)
	}
}

func TestService_UpdateEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := eventrepomocks.NewMockeventRepo(ctrl)
	svc := New(mockRepo)

	eventID := uuid.New()
	userID := uuid.New()
	title := "Updated Event"
	description := "Updated description"
	date := time.Now()

	mockRepo.EXPECT().
		UpdateEvent(gomock.Any(), gomock.Any()).
		Return(nil)

	err := svc.UpdateEvent(context.Background(), eventID, userID, title, description, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_DeleteEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := eventrepomocks.NewMockeventRepo(ctrl)
	svc := New(mockRepo)

	eventID := uuid.New()
	userID := uuid.New()

	mockRepo.EXPECT().
		DeleteEvent(gomock.Any(), eventID, userID).
		Return(nil)

	if err := svc.DeleteEvent(context.Background(), eventID, userID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_GetEventsForDay(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := eventrepomocks.NewMockeventRepo(ctrl)
	svc := New(mockRepo)

	mockEvents := []model.Event{
		{Title: "Event 1", EventDate: time.Now()},
	}

	mockRepo.EXPECT().
		GetEventsForDay(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockEvents, nil)

	ev, err := svc.GetEventsForDay(context.Background(), uuid.New(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ev) != len(mockEvents) {
		t.Fatalf("expected %d events, got %d", len(mockEvents), len(ev))
	}
}

func TestService_GetEventsForWeek(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := eventrepomocks.NewMockeventRepo(ctrl)
	svc := New(mockRepo)

	mockEvents := []model.Event{
		{Title: "Event Week", EventDate: time.Now()},
	}

	mockRepo.EXPECT().
		GetEventsForWeek(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockEvents, nil)

	ev, err := svc.GetEventsForWeek(context.Background(), uuid.New(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ev) != len(mockEvents) {
		t.Fatalf("expected %d events, got %d", len(mockEvents), len(ev))
	}
}

func TestService_GetEventsForMonth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := eventrepomocks.NewMockeventRepo(ctrl)
	svc := New(mockRepo)

	mockEvents := []model.Event{
		{Title: "Event Month", EventDate: time.Now()},
	}

	mockRepo.EXPECT().
		GetEventsForMonth(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockEvents, nil)

	ev, err := svc.GetEventsForMonth(context.Background(), uuid.New(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ev) != len(mockEvents) {
		t.Fatalf("expected %d events, got %d", len(mockEvents), len(ev))
	}
}
