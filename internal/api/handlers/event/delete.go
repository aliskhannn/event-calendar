package event

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aliskhannn/calendar-service/internal/api/response"
	"github.com/aliskhannn/calendar-service/internal/middlewares"
	eventrepo "github.com/aliskhannn/calendar-service/internal/repository/event"
)

// Delete handles the HTTP request to delete an event by its ID.
// It extracts the event ID from the URL parameter and the user ID from the request context,
// validates them, and calls the service to delete the event. If successful, it returns a success response.
// In case of errors (e.g., invalid ID, unauthorized user, or event not found), it returns an appropriate error response.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
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

	// Extract and validate user ID from request context.
	userIDVal := r.Context().Value(middlewares.UserIDKey)
	userID, ok := userIDVal.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		h.logger.Warn("missing or invalid user id in context")
		response.Fail(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	// Attempt to delete the event using the service.
	if err := h.service.DeleteEvent(r.Context(), eventID, userID); err != nil {
		// Handle case where event is not found.
		if errors.Is(err, eventrepo.ErrEventNotFound) {
			h.logger.Info("event not found", zap.String("eventID", eventID.String()))
			response.Fail(w, http.StatusNotFound, fmt.Errorf("event not found"))
			return
		}

		// Log and handle unexpected errors.
		h.logger.Error("failed to delete event",
			zap.String("event_id", eventID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		response.Fail(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// Return success response.
	response.OK(w, "event deleted")
}
