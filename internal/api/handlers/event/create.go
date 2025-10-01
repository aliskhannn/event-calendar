package event

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aliskhannn/calendar-service/internal/api/response"
	"github.com/aliskhannn/calendar-service/internal/middlewares"
	"github.com/aliskhannn/calendar-service/internal/model"
)

type CreateRequest struct {
	UserID      uuid.UUID  `json:"user_id" validate:"required"`
	Title       string     `json:"title" validate:"required,min=3,max=255"`
	Description string     `json:"description" validate:"max=1000"`
	EventDate   time.Time  `json:"event_date" validate:"required"`
	ReminderAt  *time.Time `json:"reminder_at"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middlewares.UserIDKey)
	userID, ok := userIDVal.(uuid.UUID)
	h.logger.Info("id from context", zap.String("user_id", userID.String()))
	if !ok || userID == uuid.Nil {
		h.logger.Warn("missing or invalid user id in context")
		response.Fail(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	var req CreateRequest
	req.UserID = userID

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

	if req.Title == "" || req.EventDate.IsZero() || req.UserID == uuid.Nil {
		h.logger.Warn("missing required fields",
			zap.String("title", req.Title),
			zap.Time("event_date", req.EventDate),
			zap.String("user_id", req.UserID.String()),
		)
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("missing required fields"))
		return
	}

	id, err := h.service.CreateEvent(r.Context(), req.UserID, req.Title, req.Description, req.EventDate, req.ReminderAt)
	if err != nil {
		h.logger.Error("failed to create event",
			zap.String("user_id", req.UserID.String()),
			zap.String("title", req.Title),
			zap.Error(err),
		)
		response.Fail(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// Push reminder to the worker if ReminderAt is set.
	if req.ReminderAt != nil && req.ReminderAt.After(time.Now()) {
		reminder := model.Reminder{
			UserID:   req.UserID,
			EventID:  id,
			Message:  req.Title,
			RemindAt: *req.ReminderAt,
		}

		select {
		case h.reminderCh <- reminder:
			h.logger.Info("reminder scheduled", zap.String("user_id", req.UserID.String()), zap.Time("remind_at", *req.ReminderAt))
		default:
			// channel full, log warning
			h.logger.Warn("reminder channel is full, dropping reminder", zap.String("user_id", req.UserID.String()))
		}
	}

	response.Created(w, id)
}
