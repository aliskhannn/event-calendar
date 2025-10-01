package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aliskhannn/calendar-service/internal/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aliskhannn/calendar-service/internal/api/response"
	eventrepo "github.com/aliskhannn/calendar-service/internal/repository/event"
)

type UpdateRequest struct {
	Title       string    `json:"title" validate:"required,min=3,max=255"`
	Description string    `json:"description" validate:"max=1000"`
	EventDate   time.Time `json:"event_date" validate:"required"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middlewares.UserIDKey)
	userID, ok := userIDVal.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		h.logger.Warn("missing or invalid user id in context")
		response.Fail(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

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

	if err := h.service.UpdateEvent(r.Context(), eventID, userID, req.Title, req.Description, req.EventDate); err != nil {
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
