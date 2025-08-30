package event

import (
	"context"
	"errors"
	"fmt"
	"github.com/aliskhannn/calendar-service/internal/api/response"
	"github.com/aliskhannn/calendar-service/internal/middlewares"
	"github.com/aliskhannn/calendar-service/internal/model"
	eventrepo "github.com/aliskhannn/calendar-service/internal/repository/event"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func (h *Handler) GetDay(w http.ResponseWriter, r *http.Request) {
	h.getEvents(w, r, h.service.GetEventsForDay)
}

func (h *Handler) GetWeek(w http.ResponseWriter, r *http.Request) {
	h.getEvents(w, r, h.service.GetEventsForWeek)
}

func (h *Handler) GetMonth(w http.ResponseWriter, r *http.Request) {
	h.getEvents(w, r, h.service.GetEventsForMonth)
}

func (h *Handler) getEvents(w http.ResponseWriter, r *http.Request, fetch func(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)) {
	userIDVal := r.Context().Value(middlewares.UserIDKey)
	userID, ok := userIDVal.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		h.logger.Warn("missing or invalid user id in context")
		response.Fail(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	dateStr := chi.URLParam(r, "date")
	if dateStr == "" {
		h.logger.Warn("missing date in path")
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("missing date"))
		return
	}

	eventDate, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		h.logger.Warn("invalid date", zap.Error(err))
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("invalid date"))
		return
	}

	events, err := fetch(r.Context(), userID, eventDate)
	if err != nil {
		if errors.Is(err, eventrepo.ErrEventNotFound) {
			h.logger.Info("events not found", zap.String("userID", userID.String()), zap.Time("date", eventDate))
			response.Fail(w, http.StatusNotFound, fmt.Errorf("events not found"))
			return
		}
		h.logger.Error("failed to fetch events", zap.Error(err))
		response.Fail(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	response.OK(w, events)
}
