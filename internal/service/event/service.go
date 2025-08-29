package event

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/aliskhannn/calendar-service/internal/model"
)

type eventRepo interface {
	CreateEvent(ctx context.Context, event model.Event) (uuid.UUID, error)
	UpdateEvent(ctx context.Context, event model.Event) error
	DeleteEvent(ctx context.Context, eventID, userID uuid.UUID) error
	GetEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)
	GetEventsForWeek(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)
	GetEventsForMonth(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)
}

type Service struct {
	eventRepo eventRepo
}

func New(r eventRepo) *Service {
	return &Service{
		eventRepo: r,
	}
}

func (s *Service) CreateEvent(ctx context.Context, event model.Event) (uuid.UUID, error) {
	id, err := s.eventRepo.CreateEvent(ctx, event)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create event: %w", err)
	}

	return id, nil
}

func (s *Service) UpdateEvent(ctx context.Context, event model.Event) error {
	err := s.eventRepo.UpdateEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}

	return nil
}

func (s *Service) DeleteEvent(ctx context.Context, eventID, userID uuid.UUID) error {
	err := s.eventRepo.DeleteEvent(ctx, eventID, userID)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}

	return nil
}

func (s *Service) GetEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error) {
	event, err := s.eventRepo.GetEventsForDay(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("get events for day: %w", err)
	}

	return event, nil
}

func (s *Service) GetEventsForWeek(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error) {
	event, err := s.eventRepo.GetEventsForWeek(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("get events for week: %w", err)
	}

	return event, nil
}

func (s *Service) GetEventsForMonth(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error) {
	event, err := s.eventRepo.GetEventsForMonth(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("get events for month: %w", err)
	}

	return event, nil
}
