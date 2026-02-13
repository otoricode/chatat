package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

const maxUploadSize = 100 * 1024 * 1024 // 100 MB

// MediaHandler handles media upload and retrieval endpoints.
type MediaHandler struct {
	mediaService service.MediaService
}

// NewMediaHandler creates a new MediaHandler.
func NewMediaHandler(mediaService service.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

// Upload handles POST /api/v1/media/upload
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	uploaderID, err := uuid.Parse(userID)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user id tidak valid"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		response.Error(w, apperror.BadRequest("gagal parsing form: ukuran file melebihi batas"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, apperror.BadRequest("file wajib dikirim"))
		return
	}
	defer file.Close()

	contextType := r.FormValue("contextType")
	contextID := r.FormValue("contextId")

	result, err := h.mediaService.Upload(r.Context(), service.MediaUploadInput{
		UploaderID:  uploaderID,
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
		Data:        file,
		ContextType: contextType,
		ContextID:   contextID,
	})
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(fmt.Errorf("gagal mengupload media")))
		return
	}

	response.Created(w, result)
}

// GetByID handles GET /api/v1/media/{id}
func (h *MediaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	mediaID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("media ID tidak valid"))
		return
	}

	result, err := h.mediaService.GetByID(r.Context(), mediaID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(fmt.Errorf("gagal mengambil media")))
		return
	}

	response.OK(w, result)
}

// Download handles GET /api/v1/media/{id}/download — redirects to signed URL.
func (h *MediaHandler) Download(w http.ResponseWriter, r *http.Request) {
	mediaID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("media ID tidak valid"))
		return
	}

	url, err := h.mediaService.GetDownloadURL(r.Context(), mediaID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(fmt.Errorf("gagal mendapatkan URL download")))
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Delete handles DELETE /api/v1/media/{id}
func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	uploaderID, err := uuid.Parse(userID)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user id tidak valid"))
		return
	}

	mediaID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("media ID tidak valid"))
		return
	}

	if err := h.mediaService.Delete(r.Context(), mediaID, uploaderID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.Error(w, appErr)
			return
		}
		response.Error(w, apperror.Internal(fmt.Errorf("gagal menghapus media")))
		return
	}

	response.OK(w, map[string]string{"message": "media berhasil dihapus"})
}
