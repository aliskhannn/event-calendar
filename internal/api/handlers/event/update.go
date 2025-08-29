package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aliskhannn/calendar-service/internal/api/response"
	"github.com/aliskhannn/calendar-service/internal/model"
	eventrepo "github.com/aliskhannn/calendar-service/internal/repository/event"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type UpdateRequest struct {
	Title       string    `json:"title" validate:"required,min=3,max=255"`
	Description string    `json:"description" validate:"max=1000"`
	EventDate   time.Time `json:"event_date" validate:"required"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.logger.Warn("missing event id in path")
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("missing event id"))
		return
	}

	eventID, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("invalid event id", zap.Error(err))
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("invalid event id"))
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("validation failed", zap.Error(err))
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("validation error: %s", err.Error()))
		return
	}

	event := model.Event{
		ID:          eventID,
		Title:       req.Title,
		Description: req.Description,
		EventDate:   req.EventDate,
		UpdatedAt:   time.Now(),
	}

	if err := h.service.UpdateEvent(r.Context(), event); err != nil {
		if errors.Is(err, eventrepo.ErrEventNotFound) {
			h.logger.Info("event not found", zap.String("eventID", eventID.String()))
			response.Fail(w, http.StatusNotFound, fmt.Errorf("event not found"))
			return
		}

		h.logger.Error("unexpected error updating event", zap.Error(err))
		response.Fail(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	response.OK(w, "event updated")
}
