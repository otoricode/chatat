package handler_test

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/internal/handler"
	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/pkg/apperror"
)

func mediaCtx(r *http.Request, userID uuid.UUID) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), "userID", userID.String()))
}

func withMediaIDParam(r *http.Request, id uuid.UUID) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id.String())
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestMediaHandler_GetByID(t *testing.T) {
	mediaID := uuid.New()

	t.Run("success", func(t *testing.T) {
		resp := &model.MediaResponse{ID: mediaID, Filename: "test.png"}
		h := handler.NewMediaHandler(&mockMediaService{mediaResp: resp})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/media/"+mediaID.String(), nil)
		h.GetByID(w, withMediaIDParam(r, mediaID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/media/bad-id", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "bad-id")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.GetByID(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{err: apperror.NotFound("media", mediaID.String())})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/media/"+mediaID.String(), nil)
		h.GetByID(w, withMediaIDParam(r, mediaID))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/media/"+mediaID.String(), nil)
		h.GetByID(w, withMediaIDParam(r, mediaID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestMediaHandler_Download(t *testing.T) {
	mediaID := uuid.New()

	t.Run("success redirect", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{downloadURL: "https://example.com/file.png"})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/media/"+mediaID.String()+"/download", nil)
		h.Download(w, withMediaIDParam(r, mediaID))
		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, "https://example.com/file.png", w.Header().Get("Location"))
	})

	t.Run("invalid id", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/media/bad/download", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "bad")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Download(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{err: apperror.NotFound("media", mediaID.String())})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/media/"+mediaID.String()+"/download", nil)
		h.Download(w, withMediaIDParam(r, mediaID))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/media/"+mediaID.String()+"/download", nil)
		h.Download(w, withMediaIDParam(r, mediaID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestMediaHandler_Delete(t *testing.T) {
	userID := uuid.New()
	mediaID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/media/"+mediaID.String(), nil)
		r = mediaCtx(r, userID)
		h.Delete(w, withMediaIDParam(r, mediaID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/media/"+mediaID.String(), nil)
		h.Delete(w, withMediaIDParam(r, mediaID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid media id", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/media/bad", nil)
		r = mediaCtx(r, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "bad")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Delete(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{err: apperror.Forbidden("not uploader")})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/media/"+mediaID.String(), nil)
		r = mediaCtx(r, userID)
		h.Delete(w, withMediaIDParam(r, mediaID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/media/"+mediaID.String(), nil)
		r = mediaCtx(r, userID)
		h.Delete(w, withMediaIDParam(r, mediaID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid userID format", func(t *testing.T) {
		h := handler.NewMediaHandler(&mockMediaService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/media/"+mediaID.String(), nil)
		r = r.WithContext(context.WithValue(r.Context(), "userID", "not-a-uuid"))
		h.Delete(w, withMediaIDParam(r, mediaID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestMediaHandler_Upload_Unauthorized(t *testing.T) {
	h := handler.NewMediaHandler(&mockMediaService{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/media/upload", nil)
	h.Upload(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMediaHandler_Upload_NoFile(t *testing.T) {
	userID := uuid.New()
	h := handler.NewMediaHandler(&mockMediaService{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/media/upload", nil)
	r = mediaCtx(r, userID)
	r.Header.Set("Content-Type", "multipart/form-data; boundary=----boundary")
	h.Upload(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaHandler_Upload_MissingFileField(t *testing.T) {
	userID := uuid.New()
	h := handler.NewMediaHandler(&mockMediaService{})

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("contextType", "chat")
	writer.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/media/upload", &buf)
	r = mediaCtx(r, userID)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	h.Upload(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaHandler_Upload_InvalidUserID(t *testing.T) {
	h := handler.NewMediaHandler(&mockMediaService{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/media/upload", nil)
	r = r.WithContext(context.WithValue(r.Context(), "userID", "not-a-uuid"))
	h.Upload(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMediaHandler_Upload_Success(t *testing.T) {
	userID := uuid.New()
	resp := &model.MediaResponse{ID: uuid.New(), Filename: "test.png"}
	h := handler.NewMediaHandler(&mockMediaService{mediaResp: resp})

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "test.png")
	part.Write([]byte("fake image content"))
	writer.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/media/upload", &buf)
	r = mediaCtx(r, userID)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	h.Upload(w, r)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestMediaHandler_Upload_ServiceError(t *testing.T) {
	userID := uuid.New()
	h := handler.NewMediaHandler(&mockMediaService{err: apperror.Internal(nil)})

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "test.png")
	part.Write([]byte("fake image content"))
	writer.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/media/upload", &buf)
	r = mediaCtx(r, userID)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	h.Upload(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMediaHandler_Upload_GenericError(t *testing.T) {
	userID := uuid.New()
	h := handler.NewMediaHandler(&mockMediaService{err: errors.New("db")})

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "test.png")
	part.Write([]byte("fake image content"))
	writer.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/media/upload", &buf)
	r = mediaCtx(r, userID)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	h.Upload(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
