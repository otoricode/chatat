package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/internal/handler"
	"github.com/otoritech/chatat/internal/middleware"
)

func notifAuthReq(method, url string, body []byte, userID uuid.UUID) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, url, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	return r.WithContext(middleware.WithUserID(r.Context(), userID))
}

func TestNotificationHandler_RegisterDevice(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewNotificationHandler(&mockNotificationService{})
		body, _ := json.Marshal(map[string]string{"token": "abc123", "platform": "ios"})
		w := httptest.NewRecorder()
		h.RegisterDevice(w, notifAuthReq(http.MethodPost, "/notifications/devices", body, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing token", func(t *testing.T) {
		h := handler.NewNotificationHandler(&mockNotificationService{})
		body, _ := json.Marshal(map[string]string{"token": "", "platform": "ios"})
		w := httptest.NewRecorder()
		h.RegisterDevice(w, notifAuthReq(http.MethodPost, "/notifications/devices", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid platform", func(t *testing.T) {
		h := handler.NewNotificationHandler(&mockNotificationService{})
		body, _ := json.Marshal(map[string]string{"token": "abc", "platform": "web"})
		w := httptest.NewRecorder()
		h.RegisterDevice(w, notifAuthReq(http.MethodPost, "/notifications/devices", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewNotificationHandler(&mockNotificationService{})
		w := httptest.NewRecorder()
		h.RegisterDevice(w, notifAuthReq(http.MethodPost, "/notifications/devices", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewNotificationHandler(&mockNotificationService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/notifications/devices", nil)
		h.RegisterDevice(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewNotificationHandler(&mockNotificationService{err: assert.AnError})
		body, _ := json.Marshal(map[string]string{"token": "abc", "platform": "android"})
		w := httptest.NewRecorder()
		h.RegisterDevice(w, notifAuthReq(http.MethodPost, "/notifications/devices", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotificationHandler_UnregisterDevice(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewNotificationHandler(&mockNotificationService{})
		body, _ := json.Marshal(map[string]string{"token": "abc123"})
		w := httptest.NewRecorder()
		h.UnregisterDevice(w, notifAuthReq(http.MethodDelete, "/notifications/devices", body, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewNotificationHandler(&mockNotificationService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/notifications/devices", nil)
		h.UnregisterDevice(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewNotificationHandler(&mockNotificationService{})
		w := httptest.NewRecorder()
		h.UnregisterDevice(w, notifAuthReq(http.MethodDelete, "/notifications/devices", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty token", func(t *testing.T) {
		h := handler.NewNotificationHandler(&mockNotificationService{})
		body, _ := json.Marshal(map[string]string{"token": ""})
		w := httptest.NewRecorder()
		h.UnregisterDevice(w, notifAuthReq(http.MethodDelete, "/notifications/devices", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewNotificationHandler(&mockNotificationService{err: errors.New("fail")})
		body, _ := json.Marshal(map[string]string{"token": "abc123"})
		w := httptest.NewRecorder()
		h.UnregisterDevice(w, notifAuthReq(http.MethodDelete, "/notifications/devices", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
