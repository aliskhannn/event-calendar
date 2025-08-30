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

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
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

	userIDVal := r.Context().Value(middlewares.UserIDKey)
	userID, ok := userIDVal.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		h.logger.Warn("missing or invalid user id in context")
		response.Fail(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	if err := h.service.DeleteEvent(r.Context(), eventID, userID); err != nil {
		if errors.Is(err, eventrepo.ErrEventNotFound) {
			h.logger.Info("event not found", zap.String("eventID", eventID.String()))
			response.Fail(w, http.StatusNotFound, fmt.Errorf("event not found"))
			return
		}

		h.logger.Error("failed to delete event",
			zap.String("event_id", eventID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		response.Fail(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	response.OK(w, "event deleted")
}
