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

// TODO: corrects mocks and tests
//
//go:generate mockgen -source=handler.go -destination=../../../mocks/api/handlers/user/mock_user_service.go -package=mocks
type userService interface {
	Create(ctx context.Context, email, name, password string) (uuid.UUID, error)
	GetByEmail(ctx context.Context, email, password string) (string, error)
}

type Handler struct {
	service   userService
	logger    *zap.Logger
	validator *validator.Validate
}

func New(s userService, l *zap.Logger, v *validator.Validate) *Handler {
	return &Handler{
		service:   s,
		logger:    l,
		validator: v,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

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
