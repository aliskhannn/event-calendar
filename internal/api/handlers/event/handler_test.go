//go:build unit
// +build unit

package event

import (
	"bytes"
	"context"
	"encoding/json"
	mockseventsvc "github.com/aliskhannn/calendar-service/internal/mocks/api/handlers/event"
	"github.com/aliskhannn/calendar-service/internal/repository/event"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aliskhannn/calendar-service/internal/middlewares"
	"github.com/aliskhannn/calendar-service/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func setupHandler(t *testing.T) (*gomock.Controller, *mockseventsvc.MockeventService, *Handler) {
	ctrl := gomock.NewController(t)
	mockService := mockseventsvc.NewMockeventService(ctrl)
	logger, _ := zap.NewDevelopment()
	validate := validator.New()
	handler := New(mockService, logger, validate)
	return ctrl, mockService, handler
}

func TestHandler_Create_Success(t *testing.T) {
	ctrl, mockService, h := setupHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	reqBody := CreateRequest{
		Title:     "Test Event",
		EventDate: time.Now(),
		UserID:    userID,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, userID))
	w := httptest.NewRecorder()

	mockService.EXPECT().
		CreateEvent(gomock.Any(), gomock.Any()).
		Return(uuid.New(), nil)

	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestHandler_Create_InvalidBody(t *testing.T) {
	ctrl, _, h := setupHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte("{invalid json")))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, userID))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	ctrl, mockService, h := setupHandler(t)
	defer ctrl.Finish()

	eventID := uuid.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/events/"+eventID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, userID))

	// chi RouteContext для URLParam
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", eventID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	w := httptest.NewRecorder()

	mockService.EXPECT().
		DeleteEvent(gomock.Any(), eventID, userID).
		Return(nil)

	h.Delete(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_GetDay_Success(t *testing.T) {
	ctrl, mockService, h := setupHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	date := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/events/day?date="+date.Format("2006-01-02"), nil)
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, userID))
	w := httptest.NewRecorder()

	mockEvents := []model.Event{{Title: "Event 1", EventDate: date}}
	mockService.EXPECT().
		GetEventsForDay(gomock.Any(), userID, gomock.Any()).
		Return(mockEvents, nil)

	h.GetDay(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	ctrl, mockService, h := setupHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	eventID := uuid.New()
	reqBody := UpdateRequest{
		Title:     "Updated",
		EventDate: time.Now(),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/events/"+eventID.String(), bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, userID))

	// chi RouteContext для URLParam
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", eventID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	w := httptest.NewRecorder()

	mockService.EXPECT().
		UpdateEvent(gomock.Any(), gomock.Any()).
		Return(nil)

	h.Update(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	ctrl, mockService, h := setupHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	eventID := uuid.New()
	reqBody := UpdateRequest{
		Title:     "Updated",
		EventDate: time.Now(),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/events/"+eventID.String(), bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, userID))

	// chi RouteContext для URLParam
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", eventID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	w := httptest.NewRecorder()

	mockService.EXPECT().
		UpdateEvent(gomock.Any(), gomock.Any()).
		Return(event.ErrEventNotFound)

	h.Update(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
