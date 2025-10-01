package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	usersvc "github.com/aliskhannn/calendar-service/internal/service/user"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aliskhannn/calendar-service/internal/api/response"
	userrepo "github.com/aliskhannn/calendar-service/internal/repository/user"
)

//go:generate mockgen -source=handler.go -destination=../../../mocks/api/handlers/user/mock_user_service.go -package=mocks

// userService defines the interface for user-related operations.
// It is used internally by the Handler to perform registration and login logic.
type userService interface {
	// Create registers a new user with the given email, name, and password.
	// Returns the newly created user's UUID or an error.
	Create(ctx context.Context, email, name, password string) (uuid.UUID, error)

	// GetByEmail validates the user's credentials and returns a JWT token if successful.
	GetByEmail(ctx context.Context, email, password string) (string, error)
}

// Handler handles HTTP requests for user registration and login.
type Handler struct {
	service   userService
	logger    *zap.Logger
	validator *validator.Validate
}

// New creates a new Handler instance with the given user service, logger, and validator.
func New(s userService, l *zap.Logger, v *validator.Validate) *Handler {
	return &Handler{
		service:   s,
		logger:    l,
		validator: v,
	}
}

// RegisterRequest represents the JSON payload for registering a new user.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest represents the JSON payload for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// Register handles user registration requests.
// It validates the request body, creates a new user, and responds with the user ID if successful.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode register request body", zap.Error(err))
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	id, err := h.service.Create(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		if errors.Is(err, usersvc.ErrUserAlreadyExists) {
			h.logger.Warn("user already exists", zap.Error(err))
			response.Fail(w, http.StatusConflict, err)
			return
		}
		if errors.Is(err, userrepo.ErrUserNotFound) {
			h.logger.Warn("user with provided email not found", zap.Error(err))
			response.Fail(w, http.StatusServiceUnavailable, err)
			return
		}

		h.logger.Error("failed to register user", zap.String("email", req.Email), zap.Error(err))
		response.Fail(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	h.logger.Info("user registered successfully", zap.String("user_id", id.String()), zap.String("email", req.Email))
	response.Created(w, id)
}

// Login handles user login requests.
// It validates the credentials, generates a JWT token, and returns it if authentication succeeds.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode login request body", zap.Error(err))
		response.Fail(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	token, err := h.service.GetByEmail(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usersvc.ErrInvalidCredentials) {
			response.Fail(w, http.StatusUnauthorized, err)
		}
		if errors.Is(err, userrepo.ErrUserNotFound) {
			h.logger.Info("user not found", zap.String("email", req.Email))
			response.Fail(w, http.StatusServiceUnavailable, err)
			return
		}

		h.logger.Warn("failed login", zap.String("email", req.Email), zap.Error(err))
		response.Fail(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	h.logger.Info("user logged in successfully", zap.String("email", req.Email))
	response.OK(w, map[string]string{"token": token})
}
