package handler

import (
	"net/http"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

// BackupHandler handles backup endpoints.
type BackupHandler struct {
	backupService service.BackupService
}

// NewBackupHandler creates a new backup handler.
func NewBackupHandler(backupService service.BackupService) *BackupHandler {
	return &BackupHandler{backupService: backupService}
}

// Export handles GET /api/v1/backup/export
func (h *BackupHandler) Export(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	bundle, err := h.backupService.ExportUserData(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, bundle)
}

// Import handles POST /api/v1/backup/import
func (h *BackupHandler) Import(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	var bundle model.BackupBundle
	if err := DecodeJSON(r, &bundle); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	if err := h.backupService.ImportUserData(r.Context(), userID, &bundle); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, map[string]string{"message": "backup imported successfully"})
}

// LogBackup handles POST /api/v1/backup/log
func (h *BackupHandler) LogBackup(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	var input model.LogBackupInput
	if err := DecodeJSON(r, &input); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	record, err := h.backupService.LogBackup(r.Context(), userID, input)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.Created(w, record)
}

// GetHistory handles GET /api/v1/backup/history
func (h *BackupHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	records, err := h.backupService.GetBackupHistory(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, records)
}

// GetLatest handles GET /api/v1/backup/latest
func (h *BackupHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	record, err := h.backupService.GetLatestBackup(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(err))
		return
	}

	if record == nil {
		response.OK(w, map[string]any{"backup": nil})
		return
	}

	response.OK(w, record)
}
