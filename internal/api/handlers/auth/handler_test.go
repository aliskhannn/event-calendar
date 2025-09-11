package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mocksusersvc "github.com/aliskhannn/calendar-service/internal/mocks/api/handlers/user"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aliskhannn/calendar-service/internal/service/user"
)

func setupUserHandler(t *testing.T) (*gomock.Controller, *mocksusersvc.MockuserService, *Handler) {
	ctrl := gomock.NewController(t)
	mockService := mocksusersvc.NewMockuserService(ctrl)
	logger, _ := zap.NewDevelopment()
	validate := validator.New()
	handler := New(mockService, logger, validate)
	return ctrl, mockService, handler
}

func TestHandler_Register_Success(t *testing.T) {
	ctrl, mockService, h := setupUserHandler(t)
	defer ctrl.Finish()

	reqBody := RegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockID := uuid.New()
	mockService.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(mockID, nil)

	h.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestHandler_Register_UserAlreadyExists(t *testing.T) {
	ctrl, mockService, h := setupUserHandler(t)
	defer ctrl.Finish()

	reqBody := RegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(uuid.Nil, user.ErrUserAlreadyExists)

	h.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestHandler_Login_Success(t *testing.T) {
	ctrl, mockService, h := setupUserHandler(t)
	defer ctrl.Finish()

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.EXPECT().
		GetByEmail(gomock.Any(), reqBody.Email, reqBody.Password).
		Return("token123", nil)

	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	ctrl, mockService, h := setupUserHandler(t)
	defer ctrl.Finish()

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpass",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.EXPECT().
		GetByEmail(gomock.Any(), reqBody.Email, reqBody.Password).
		Return("", user.ErrInvalidCredentials)

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestHandler_Login_UserNotFound(t *testing.T) {
	ctrl, mockService, h := setupUserHandler(t)
	defer ctrl.Finish()

	reqBody := LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.EXPECT().
		GetByEmail(gomock.Any(), reqBody.Email, reqBody.Password).
		Return("", errors.New("not found"))

	h.Login(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestHandler_Register_InvalidBody(t *testing.T) {
	ctrl, _, h := setupUserHandler(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte("{invalid json")))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandler_Login_InvalidBody(t *testing.T) {
	ctrl, _, h := setupUserHandler(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("{invalid json")))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
