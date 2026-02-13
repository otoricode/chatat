package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/internal/handler"
	"github.com/otoritech/chatat/internal/middleware"
	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
)

func docAuthReq(method, url string, body []byte, userID uuid.UUID) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, url, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	return r.WithContext(middleware.WithUserID(r.Context(), userID))
}

func withDocIDParam(r *http.Request, id uuid.UUID) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id.String())
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func newDocHandler(docSvc *mockDocumentService, blockSvc *mockBlockService, tmplSvc *mockTemplateService) *handler.DocumentHandler {
	return handler.NewDocumentHandler(docSvc, blockSvc, tmplSvc)
}

// --- Document CRUD ---

func TestDocumentHandler_Create(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		doc := &service.DocumentFull{Document: model.Document{ID: uuid.New(), Title: "Doc1", CreatedAt: time.Now()}}
		h := newDocHandler(&mockDocumentService{docFull: doc}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"title": "Doc1"})
		w := httptest.NewRecorder()
		h.Create(w, docAuthReq(http.MethodPost, "/documents", body, userID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		h.Create(w, docAuthReq(http.MethodPost, "/documents", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents", nil)
		h.Create(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("with chatId", func(t *testing.T) {
		doc := &service.DocumentFull{Document: model.Document{ID: uuid.New()}}
		h := newDocHandler(&mockDocumentService{docFull: doc}, &mockBlockService{}, &mockTemplateService{})
		chatID := uuid.New().String()
		body, _ := json.Marshal(map[string]any{"title": "Doc", "chatId": chatID})
		w := httptest.NewRecorder()
		h.Create(w, docAuthReq(http.MethodPost, "/documents", body, userID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("with topicId", func(t *testing.T) {
		doc := &service.DocumentFull{Document: model.Document{ID: uuid.New()}}
		h := newDocHandler(&mockDocumentService{docFull: doc}, &mockBlockService{}, &mockTemplateService{})
		topicID := uuid.New().String()
		body, _ := json.Marshal(map[string]any{"title": "Doc", "topicId": topicID})
		w := httptest.NewRecorder()
		h.Create(w, docAuthReq(http.MethodPost, "/documents", body, userID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid chatId", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"title": "Doc", "chatId": "bad"})
		w := httptest.NewRecorder()
		h.Create(w, docAuthReq(http.MethodPost, "/documents", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid topicId", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"title": "Doc", "topicId": "bad"})
		w := httptest.NewRecorder()
		h.Create(w, docAuthReq(http.MethodPost, "/documents", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"title": "Doc1"})
		w := httptest.NewRecorder()
		h.Create(w, docAuthReq(http.MethodPost, "/documents", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.BadRequest("title required")}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"title": "Doc1"})
		w := httptest.NewRecorder()
		h.Create(w, docAuthReq(http.MethodPost, "/documents", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_GetByID(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		doc := &service.DocumentFull{Document: model.Document{ID: docID}}
		h := newDocHandler(&mockDocumentService{docFull: doc}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodGet, "/documents/"+docID.String(), nil, userID)
		h.GetByID(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.NotFound("document", docID.String())}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodGet, "/documents/"+docID.String(), nil, userID)
		h.GetByID(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/documents/"+docID.String(), nil)
		h.GetByID(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodGet, "/documents/"+docID.String(), nil, userID)
		h.GetByID(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodGet, "/documents/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.GetByID(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_List(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		items := []*service.DocumentListItem{{ID: uuid.New(), Title: "D1"}}
		h := newDocHandler(&mockDocumentService{docList: items}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		h.List(w, docAuthReq(http.MethodGet, "/documents", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/documents", nil)
		h.List(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Internal(nil)}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		h.List(w, docAuthReq(http.MethodGet, "/documents", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		h.List(w, docAuthReq(http.MethodGet, "/documents", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestDocumentHandler_ListByChat(t *testing.T) {
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		items := []*service.DocumentListItem{{ID: uuid.New()}}
		h := newDocHandler(&mockDocumentService{docList: items}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String()+"/documents", nil)
		h.ListByChat(w, withDocIDParam(r, chatID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Internal(nil)}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String()+"/documents", nil)
		h.ListByChat(w, withDocIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String()+"/documents", nil)
		h.ListByChat(w, withDocIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/chats/invalid/documents", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.ListByChat(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_ListByTopic(t *testing.T) {
	tid := uuid.New()

	t.Run("success", func(t *testing.T) {
		items := []*service.DocumentListItem{{ID: uuid.New()}}
		h := newDocHandler(&mockDocumentService{docList: items}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/topics/"+tid.String()+"/documents", nil)
		h.ListByTopic(w, withDocIDParam(r, tid))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Internal(nil)}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/topics/"+tid.String()+"/documents", nil)
		h.ListByTopic(w, withDocIDParam(r, tid))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/topics/"+tid.String()+"/documents", nil)
		h.ListByTopic(w, withDocIDParam(r, tid))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid topic id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/topics/invalid/documents", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.ListByTopic(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_Update(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		doc := &model.Document{ID: docID, Title: "Updated"}
		h := newDocHandler(&mockDocumentService{doc: doc}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"title": "Updated"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String(), body, userID)
		h.Update(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String(), []byte("bad"), userID)
		h.Update(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/documents/"+docID.String(), nil)
		h.Update(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Forbidden("not owner")}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"title": "Updated"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String(), body, userID)
		h.Update(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"title": "Updated"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String(), body, userID)
		h.Update(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"title": "Updated"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/invalid", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Update(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_Delete(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String(), nil, userID)
		h.Delete(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Forbidden("not owner")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String(), nil, userID)
		h.Delete(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/documents/"+docID.String(), nil)
		h.Delete(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String(), nil, userID)
		h.Delete(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Delete(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_Duplicate(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		doc := &service.DocumentFull{Document: model.Document{ID: uuid.New()}}
		h := newDocHandler(&mockDocumentService{docFull: doc}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/duplicate", nil, userID)
		h.Duplicate(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/duplicate", nil)
		h.Duplicate(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.NotFound("document", docID.String())}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/duplicate", nil, userID)
		h.Duplicate(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/duplicate", nil, userID)
		h.Duplicate(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/invalid/duplicate", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Duplicate(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- Lock / Unlock / Sign ---

func TestDocumentHandler_Lock(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"mode": "manual"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/lock", body, userID)
		h.Lock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid mode", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"mode": "invalid"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/lock", body, userID)
		h.Lock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/lock", nil)
		h.Lock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/lock", []byte("bad"), userID)
		h.Lock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Forbidden("not owner")}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"mode": "manual"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/lock", body, userID)
		h.Lock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"mode": "manual"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/lock", body, userID)
		h.Lock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"mode": "manual"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/invalid/lock", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Lock(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_Unlock(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/unlock", nil, userID)
		h.Unlock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/unlock", nil)
		h.Unlock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Forbidden("not owner")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/unlock", nil, userID)
		h.Unlock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/unlock", nil, userID)
		h.Unlock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/invalid/unlock", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Unlock(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_Sign(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		doc := &model.Document{ID: docID}
		h := newDocHandler(&mockDocumentService{doc: doc}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"name": "John Doe"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/sign", body, userID)
		h.Sign(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/sign", []byte("bad"), userID)
		h.Sign(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/sign", nil)
		h.Sign(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.BadRequest("not locked")}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"name": "John Doe"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/sign", body, userID)
		h.Sign(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"name": "John Doe"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/sign", body, userID)
		h.Sign(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"name": "John Doe"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/invalid/sign", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Sign(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_AddSigner(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		signerID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": signerID.String()})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/signers", body, userID)
		h.AddSigner(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid userId", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"userId": "bad"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/signers", body, userID)
		h.AddSigner(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/signers", nil)
		h.AddSigner(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/signers", []byte("bad"), userID)
		h.AddSigner(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Forbidden("not owner")}, &mockBlockService{}, &mockTemplateService{})
		signerID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": signerID.String()})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/signers", body, userID)
		h.AddSigner(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		signerID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": signerID.String()})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/signers", body, userID)
		h.AddSigner(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		signerID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": signerID.String()})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/invalid/signers", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.AddSigner(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_RemoveSigner(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	signerID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/signers/"+signerID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", signerID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveSigner(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/documents/"+docID.String()+"/signers/"+signerID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", signerID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveSigner(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.NotFound("signer", signerID.String())}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/signers/"+signerID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", signerID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveSigner(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/signers/"+signerID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", signerID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveSigner(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/invalid/signers/"+signerID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		rctx.URLParams.Add("userID", signerID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveSigner(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid signer id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/signers/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveSigner(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_ListSigners(t *testing.T) {
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		signers := []*model.DocumentSigner{{DocumentID: uuid.New(), UserID: uuid.New()}}
		h := newDocHandler(&mockDocumentService{signers: signers}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/documents/"+docID.String()+"/signers", nil)
		h.ListSigners(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("nil signers returns empty array", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/documents/"+docID.String()+"/signers", nil)
		h.ListSigners(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Internal(nil)}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/documents/"+docID.String()+"/signers", nil)
		h.ListSigners(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/documents/"+docID.String()+"/signers", nil)
		h.ListSigners(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/documents/invalid/signers", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.ListSigners(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- Block endpoints ---

func TestDocumentHandler_AddBlock(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		blk := &model.Block{ID: uuid.New(), Type: "paragraph"}
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{block: blk}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"type": "paragraph", "content": "hello", "position": 0})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/blocks", body, userID)
		h.AddBlock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/blocks", []byte("bad"), userID)
		h.AddBlock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/blocks", nil)
		h.AddBlock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{err: apperror.Forbidden("locked")}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"type": "paragraph", "content": "hello", "position": 0})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/blocks", body, userID)
		h.AddBlock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{err: errors.New("db")}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"type": "paragraph", "content": "hello", "position": 0})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/blocks", body, userID)
		h.AddBlock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"type": "paragraph", "content": "hello", "position": 0})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/invalid/blocks", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.AddBlock(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("with parentBlockId", func(t *testing.T) {
		blk := &model.Block{ID: uuid.New(), Type: "paragraph"}
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{block: blk}, &mockTemplateService{})
		parentID := uuid.New().String()
		body, _ := json.Marshal(map[string]any{"type": "paragraph", "content": "hello", "position": 0, "parentBlockId": parentID})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/blocks", body, userID)
		h.AddBlock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid parentBlockId", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"type": "paragraph", "content": "hello", "position": 0, "parentBlockId": "bad-id"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/blocks", body, userID)
		h.AddBlock(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_UpdateBlock(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		blk := &model.Block{ID: blockID}
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{block: blk}, &mockTemplateService{})
		content := "updated"
		body, _ := json.Marshal(map[string]any{"content": content})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/blocks/"+blockID.String(), body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("blockId", blockID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateBlock(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/documents/"+docID.String()+"/blocks/"+blockID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("blockId", blockID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateBlock(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/blocks/"+blockID.String(), []byte("bad"), userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("blockId", blockID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateBlock(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{err: apperror.NotFound("block", blockID.String())}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"content": "x"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/blocks/"+blockID.String(), body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("blockId", blockID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateBlock(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{err: errors.New("db")}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"content": "x"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/blocks/"+blockID.String(), body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("blockId", blockID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateBlock(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid block id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"content": "x"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/blocks/invalid", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("blockId", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateBlock(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_DeleteBlock(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/blocks/"+blockID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("blockId", blockID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.DeleteBlock(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/documents/"+docID.String()+"/blocks/"+blockID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("blockId", blockID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.DeleteBlock(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{err: apperror.NotFound("block", blockID.String())}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/blocks/"+blockID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("blockId", blockID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.DeleteBlock(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{err: errors.New("db")}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/blocks/"+blockID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("blockId", blockID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.DeleteBlock(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid block id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/blocks/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("blockId", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.DeleteBlock(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_ReorderBlocks(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		ids := []string{uuid.New().String(), uuid.New().String()}
		body, _ := json.Marshal(map[string]any{"blockIds": ids})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/blocks/reorder", body, userID)
		h.ReorderBlocks(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid block id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"blockIds": []string{"bad-id"}})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/blocks/reorder", body, userID)
		h.ReorderBlocks(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/documents/"+docID.String()+"/blocks/reorder", nil)
		h.ReorderBlocks(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/blocks/reorder", []byte("bad"), userID)
		h.ReorderBlocks(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{err: apperror.Internal(nil)}, &mockTemplateService{})
		ids := []string{uuid.New().String()}
		body, _ := json.Marshal(map[string]any{"blockIds": ids})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/blocks/reorder", body, userID)
		h.ReorderBlocks(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{err: errors.New("db")}, &mockTemplateService{})
		ids := []string{uuid.New().String()}
		body, _ := json.Marshal(map[string]any{"blockIds": ids})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/blocks/reorder", body, userID)
		h.ReorderBlocks(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		ids := []string{uuid.New().String()}
		body, _ := json.Marshal(map[string]any{"blockIds": ids})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/invalid/blocks/reorder", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.ReorderBlocks(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_BatchBlocks(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"operations": []map[string]any{}})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/blocks/batch", body, userID)
		h.BatchBlocks(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/blocks/batch", []byte("bad"), userID)
		h.BatchBlocks(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/blocks/batch", nil)
		h.BatchBlocks(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{err: apperror.BadRequest("locked")}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"operations": []map[string]any{}})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/blocks/batch", body, userID)
		h.BatchBlocks(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{err: errors.New("db")}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"operations": []map[string]any{}})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/blocks/batch", body, userID)
		h.BatchBlocks(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]any{"operations": []map[string]any{}})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/invalid/blocks/batch", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.BatchBlocks(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- Collaborator endpoints ---

func TestDocumentHandler_AddCollaborator(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		collabID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": collabID.String(), "role": "editor"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/collaborators", body, userID)
		h.AddCollaborator(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid role", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		collabID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": collabID.String(), "role": "admin"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/collaborators", body, userID)
		h.AddCollaborator(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid userId", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"userId": "bad", "role": "editor"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/collaborators", body, userID)
		h.AddCollaborator(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/collaborators", nil)
		h.AddCollaborator(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Forbidden("not owner")}, &mockBlockService{}, &mockTemplateService{})
		collabID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": collabID.String(), "role": "editor"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/collaborators", body, userID)
		h.AddCollaborator(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		collabID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": collabID.String(), "role": "editor"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/collaborators", body, userID)
		h.AddCollaborator(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		collabID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": collabID.String(), "role": "editor"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/invalid/collaborators", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.AddCollaborator(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPost, "/documents/"+docID.String()+"/collaborators", []byte("bad"), userID)
		h.AddCollaborator(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_RemoveCollaborator(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	collabID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/collaborators/"+collabID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", collabID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveCollaborator(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/documents/"+docID.String()+"/collaborators/"+collabID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", collabID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveCollaborator(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Forbidden("not owner")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/collaborators/"+collabID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", collabID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveCollaborator(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/collaborators/"+collabID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", collabID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveCollaborator(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/invalid/collaborators/"+collabID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		rctx.URLParams.Add("userID", collabID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveCollaborator(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid user id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/collaborators/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveCollaborator(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_UpdateCollaboratorRole(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	collabID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"role": "viewer"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/collaborators/"+collabID.String(), body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", collabID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateCollaboratorRole(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid role", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"role": "owner"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/collaborators/"+collabID.String(), body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", collabID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateCollaboratorRole(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/documents/"+docID.String()+"/collaborators/"+collabID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", collabID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateCollaboratorRole(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Forbidden("not owner")}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"role": "viewer"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/collaborators/"+collabID.String(), body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", collabID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateCollaboratorRole(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"role": "viewer"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/collaborators/"+collabID.String(), body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", collabID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateCollaboratorRole(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"role": "viewer"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/invalid/collaborators/"+collabID.String(), body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		rctx.URLParams.Add("userID", collabID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateCollaboratorRole(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid user id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"role": "viewer"})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/collaborators/invalid", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateCollaboratorRole(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := docAuthReq(http.MethodPut, "/documents/"+docID.String()+"/collaborators/"+collabID.String(), []byte("bad"), userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("userID", collabID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UpdateCollaboratorRole(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- Tag endpoints ---

func TestDocumentHandler_AddTag(t *testing.T) {
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"tag": "important"})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/tags", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		h.AddTag(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/tags", bytes.NewReader([]byte("bad")))
		r.Header.Set("Content-Type", "application/json")
		h.AddTag(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.BadRequest("empty tag")}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"tag": ""})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/tags", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		h.AddTag(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"tag": "test"})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/tags", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		h.AddTag(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		body, _ := json.Marshal(map[string]string{"tag": "test"})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/invalid/tags", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.AddTag(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_RemoveTag(t *testing.T) {
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/documents/"+docID.String()+"/tags/important", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("tag", "important")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveTag(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("empty tag", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/documents/"+docID.String()+"/tags/", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("tag", "")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveTag(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Internal(nil)}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/documents/"+docID.String()+"/tags/urgent", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("tag", "urgent")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveTag(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/documents/"+docID.String()+"/tags/urgent", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("tag", "urgent")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveTag(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/documents/invalid/tags/urgent", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		rctx.URLParams.Add("tag", "urgent")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveTag(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- History & Template endpoints ---

func TestDocumentHandler_GetHistory(t *testing.T) {
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		hist := []*model.DocumentHistory{{ID: uuid.New()}}
		h := newDocHandler(&mockDocumentService{history: hist}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/documents/"+docID.String()+"/history", nil)
		h.GetHistory(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: apperror.Internal(nil)}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/documents/"+docID.String()+"/history", nil)
		h.GetHistory(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("nil history returns empty array", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/documents/"+docID.String()+"/history", nil)
		h.GetHistory(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{err: errors.New("db")}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/documents/"+docID.String()+"/history", nil)
		h.GetHistory(w, withDocIDParam(r, docID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/documents/invalid/history", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.GetHistory(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_ListTemplates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		templates := []*service.DocumentTemplate{{ID: "basic", Name: "Basic"}}
		h := newDocHandler(&mockDocumentService{}, &mockBlockService{}, &mockTemplateService{templates: templates})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/templates", nil)
		h.ListTemplates(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
