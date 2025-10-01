package event

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aliskhannn/calendar-service/internal/api/response"
	"github.com/aliskhannn/calendar-service/internal/middlewares"
	"github.com/aliskhannn/calendar-service/internal/model"
	eventrepo "github.com/aliskhannn/calendar-service/internal/repository/event"
)

// GetDay handles HTTP requests to retrieve events for a specific day.
// It delegates to the getEvents helper function, passing the service method for fetching daily events.
func (h *Handler) GetDay(w http.ResponseWriter, r *http.Request) {
	h.getEvents(w, r, h.service.GetEventsForDay)
}

// GetWeek handles HTTP requests to retrieve events for a specific week.
// It delegates to the getEvents helper function, passing the service method for fetching weekly events.
func (h *Handler) GetWeek(w http.ResponseWriter, r *http.Request) {
	h.getEvents(w, r, h.service.GetEventsForWeek)
}

// GetMonth handles HTTP requests to retrieve events for a specific month.
// It delegates to the getEvents helper function, passing the service method for fetching monthly events.
func (h *Handler) GetMonth(w http.ResponseWriter, r *http.Request) {
	h.getEvents(w, r, h.service.GetEventsForMonth)
}

// getEvents is a helper function that retrieves events for a given user and date range.
// It extracts and validates the user ID from the request context and the date from query parameters,
// then calls the provided fetch function to retrieve events. It handles errors and sends appropriate responses.
//
// Parameters:
//   - w: The HTTP response writer to send the response.
//   - r: The HTTP request containing the user context and query parameters.
//   - fetch: A function that retrieves events for a specific user and date.
func (h *Handler) getEvents(w http.ResponseWriter, r *http.Request, fetch func(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.Event, error)) {
	// Extract and validate user ID from request context.
	userIDVal := r.Context().Value(middlewares.UserIDKey)
	userID, ok := userIDVal.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		h.logger.Warn("missing or invalid user id in context")
		response.Fail(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	// Extract and validate date from query parameters.
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		h.logger.Warn("missing date in path")
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("missing date"))
		return
	}

	// Parse date string into time.Time.
	eventDate, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		h.logger.Warn("invalid date", zap.Error(err))
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("invalid date"))
		return
	}

	// Fetch events using the provided fetch function.
	events, err := fetch(r.Context(), userID, eventDate)
	if err != nil {
		// Handle case where no events are found.
		if errors.Is(err, eventrepo.ErrEventNotFound) {
			h.logger.Info("events not found", zap.String("userID", userID.String()), zap.Time("date", eventDate))
			response.Fail(w, http.StatusNotFound, fmt.Errorf("events not found"))
			return
		}
		// Log and handle unexpected errors.
		h.logger.Error("failed to fetch events", zap.Error(err))
		response.Fail(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// Return successful response with events.
	response.OK(w, events)
}
