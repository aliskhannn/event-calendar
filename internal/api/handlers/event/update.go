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

// UpdateRequest represents the expected JSON structure for updating an event.
// It includes fields for the event title, description, event date, and optional reminder time,
// with validation rules applied to ensure data integrity.
type UpdateRequest struct {
	Title       string     `json:"title" validate:"required,min=3,max=255"` // Title of the event, required, 3-255 characters
	Description string     `json:"description" validate:"max=1000"`         // optional description, max 1000 characters
	EventDate   time.Time  `json:"event_date" validate:"required"`          // date and time of the event, required
	ReminderAt  *time.Time `json:"reminder_at"`                             // optional reminder time for the event
}

// Update handles HTTP requests to update an existing event by its ID.
// It extracts and validates the user ID from the request context, the event ID from the URL,
// and the event data from the request body. It then calls the service to update the event.
// If successful, it returns a success response; otherwise, it returns an appropriate error response.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	// Extract and validate user ID from request context.
	userIDVal := r.Context().Value(middlewares.UserIDKey)
	userID, ok := userIDVal.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		h.logger.Warn("missing or invalid user id in context")
		response.Fail(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	// Extract event ID from URL parameter.
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.logger.Warn("missing event id in path")
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("missing event id"))
		return
	}

	// Parse event ID into UUID.
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("invalid event id", zap.Error(err))
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("invalid event id"))
		return
	}

	// Decode and validate request body.
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	// Validate request data using the validator.
	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("validation failed", zap.Error(err))
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("validation error: %s", err.Error()))
		return
	}

	// Update the event using the service.
	if err := h.service.UpdateEvent(r.Context(), eventID, userID, req.Title, req.Description, req.EventDate, req.ReminderAt); err != nil {
		// Handle case where event is not found.
		if errors.Is(err, eventrepo.ErrEventNotFound) {
			h.logger.Info("event not found", zap.String("eventID", eventID.String()))
			response.Fail(w, http.StatusNotFound, fmt.Errorf("event not found"))
			return
		}

		// Log and handle unexpected errors.
		h.logger.Error("unexpected error updating event", zap.Error(err))
		response.Fail(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// Return success response.
	response.OK(w, "event updated")
}
