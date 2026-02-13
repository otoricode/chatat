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
	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/pkg/apperror"
)

func backupAuthReq(method, url string, body []byte, userID uuid.UUID) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, url, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	return r.WithContext(middleware.WithUserID(r.Context(), userID))
}

func TestBackupHandler_Export(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		bundle := &model.BackupBundle{}
		h := handler.NewBackupHandler(&mockBackupService{bundle: bundle})
		w := httptest.NewRecorder()
		h.Export(w, backupAuthReq(http.MethodGet, "/api/v1/backup/export", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/backup/export", nil)
		h.Export(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{err: apperror.Internal(assert.AnError)})
		w := httptest.NewRecorder()
		h.Export(w, backupAuthReq(http.MethodGet, "/api/v1/backup/export", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{err: errors.New("db")})
		w := httptest.NewRecorder()
		h.Export(w, backupAuthReq(http.MethodGet, "/api/v1/backup/export", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestBackupHandler_Import(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{})
		body, _ := json.Marshal(model.BackupBundle{})
		w := httptest.NewRecorder()
		h.Import(w, backupAuthReq(http.MethodPost, "/api/v1/backup/import", body, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{})
		w := httptest.NewRecorder()
		h.Import(w, backupAuthReq(http.MethodPost, "/api/v1/backup/import", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/backup/import", nil)
		h.Import(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{err: apperror.Internal(nil)})
		body, _ := json.Marshal(model.BackupBundle{})
		w := httptest.NewRecorder()
		h.Import(w, backupAuthReq(http.MethodPost, "/api/v1/backup/import", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{err: errors.New("db")})
		body, _ := json.Marshal(model.BackupBundle{})
		w := httptest.NewRecorder()
		h.Import(w, backupAuthReq(http.MethodPost, "/api/v1/backup/import", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestBackupHandler_LogBackup(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rec := &model.BackupRecord{ID: uuid.New(), UserID: userID}
		h := handler.NewBackupHandler(&mockBackupService{record: rec})
		body, _ := json.Marshal(model.LogBackupInput{SizeBytes: 1024, Platform: "android"})
		w := httptest.NewRecorder()
		h.LogBackup(w, backupAuthReq(http.MethodPost, "/api/v1/backup/log", body, userID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{})
		w := httptest.NewRecorder()
		h.LogBackup(w, backupAuthReq(http.MethodPost, "/api/v1/backup/log", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/backup/log", nil)
		h.LogBackup(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{err: apperror.Internal(nil)})
		body, _ := json.Marshal(model.LogBackupInput{SizeBytes: 1024, Platform: "android"})
		w := httptest.NewRecorder()
		h.LogBackup(w, backupAuthReq(http.MethodPost, "/api/v1/backup/log", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{err: errors.New("db")})
		body, _ := json.Marshal(model.LogBackupInput{SizeBytes: 1024, Platform: "android"})
		w := httptest.NewRecorder()
		h.LogBackup(w, backupAuthReq(http.MethodPost, "/api/v1/backup/log", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestBackupHandler_GetHistory(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		records := []model.BackupRecord{{ID: uuid.New()}}
		h := handler.NewBackupHandler(&mockBackupService{records: records})
		w := httptest.NewRecorder()
		h.GetHistory(w, backupAuthReq(http.MethodGet, "/api/v1/backup/history", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/backup/history", nil)
		h.GetHistory(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{err: apperror.Internal(nil)})
		w := httptest.NewRecorder()
		h.GetHistory(w, backupAuthReq(http.MethodGet, "/api/v1/backup/history", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{err: errors.New("db")})
		w := httptest.NewRecorder()
		h.GetHistory(w, backupAuthReq(http.MethodGet, "/api/v1/backup/history", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestBackupHandler_GetLatest(t *testing.T) {
	userID := uuid.New()

	t.Run("success with record", func(t *testing.T) {
		rec := &model.BackupRecord{ID: uuid.New()}
		h := handler.NewBackupHandler(&mockBackupService{record: rec})
		w := httptest.NewRecorder()
		h.GetLatest(w, backupAuthReq(http.MethodGet, "/api/v1/backup/latest", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("no backup", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{})
		w := httptest.NewRecorder()
		h.GetLatest(w, backupAuthReq(http.MethodGet, "/api/v1/backup/latest", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/backup/latest", nil)
		h.GetLatest(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{err: apperror.Internal(nil)})
		w := httptest.NewRecorder()
		h.GetLatest(w, backupAuthReq(http.MethodGet, "/api/v1/backup/latest", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewBackupHandler(&mockBackupService{err: errors.New("db")})
		w := httptest.NewRecorder()
		h.GetLatest(w, backupAuthReq(http.MethodGet, "/api/v1/backup/latest", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
