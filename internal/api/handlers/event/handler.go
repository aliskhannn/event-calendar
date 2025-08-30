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
type eventService interface {
	CreateEvent(ctx context.Context, event model.Event) (uuid.UUID, error)
	UpdateEvent(ctx context.Context, event model.Event) error
	DeleteEvent(ctx context.Context, eventID, userID uuid.UUID) error
	GetEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)
	GetEventsForWeek(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)
	GetEventsForMonth(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)
}

type Handler struct {
	service   eventService
	logger    *zap.Logger
	validator *validator.Validate
}

func New(s eventService, l *zap.Logger, v *validator.Validate) *Handler {
	return &Handler{
		service:   s,
		logger:    l,
		validator: v,
	}
}
