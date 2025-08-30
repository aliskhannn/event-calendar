package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aliskhannn/calendar-service/internal/api/response"
	"github.com/aliskhannn/calendar-service/internal/model"
	userrepo "github.com/aliskhannn/calendar-service/internal/repository/user"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
)

type userService interface {
	Create(ctx context.Context, user model.User) (uuid.UUID, error)
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

	user := model.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	}

	id, err := h.service.Create(r.Context(), user)
	if err != nil {
		h.logger.Error("failed to register user", zap.String("email", req.Email), zap.Error(err))
		response.Fail(w, http.StatusBadRequest, err)
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
		if errors.Is(err, userrepo.ErrUserNotFound) {
			h.logger.Info("user not found", zap.String("email", req.Email))
			response.Fail(w, http.StatusNotFound, fmt.Errorf("user not found"))
			return
		}

		h.logger.Warn("failed login attempt", zap.String("email", req.Email), zap.Error(err))
		response.Fail(w, http.StatusUnauthorized, err)
		return
	}

	h.logger.Info("user logged in successfully", zap.String("email", req.Email))
	response.OK(w, map[string]string{"token": token})
}
