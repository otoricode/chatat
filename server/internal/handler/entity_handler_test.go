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
	"github.com/otoritech/chatat/pkg/apperror"
)

func entityAuthReq(method, url string, body []byte, userID uuid.UUID) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, url, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	return r.WithContext(middleware.WithUserID(r.Context(), userID))
}

func withEntityIDParam(r *http.Request, id uuid.UUID) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id.String())
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestEntityHandler_Create(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ent := &model.Entity{ID: uuid.New(), Name: "Client A", Type: "client", OwnerID: userID, CreatedAt: time.Now()}
		h := handler.NewEntityHandler(&mockEntityService{entity: ent})
		body, _ := json.Marshal(map[string]any{"name": "Client A", "type": "client"})
		w := httptest.NewRecorder()
		h.Create(w, entityAuthReq(http.MethodPost, "/entities", body, userID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		h.Create(w, entityAuthReq(http.MethodPost, "/entities", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/entities", nil)
		h.Create(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: apperror.BadRequest("name required")})
		body, _ := json.Marshal(map[string]any{"name": "", "type": "client"})
		w := httptest.NewRecorder()
		h.Create(w, entityAuthReq(http.MethodPost, "/entities", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestEntityHandler_List(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		items := []*model.EntityListItem{{Entity: model.Entity{ID: uuid.New(), Name: "E1"}}}
		h := handler.NewEntityHandler(&mockEntityService{listItems: items, total: 1})
		w := httptest.NewRecorder()
		h.List(w, entityAuthReq(http.MethodGet, "/entities?type=client", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/entities", nil)
		h.List(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: apperror.Internal(nil)})
		w := httptest.NewRecorder()
		h.List(w, entityAuthReq(http.MethodGet, "/entities?type=client", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: errors.New("db")})
		w := httptest.NewRecorder()
		h.List(w, entityAuthReq(http.MethodGet, "/entities?type=client", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("nil items", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		h.List(w, entityAuthReq(http.MethodGet, "/entities", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestEntityHandler_GetByID(t *testing.T) {
	userID := uuid.New()
	entityID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ent := &model.Entity{ID: entityID, Name: "E1", OwnerID: userID}
		h := handler.NewEntityHandler(&mockEntityService{entity: ent})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/entities/"+entityID.String(), nil, userID)
		h.GetByID(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: apperror.NotFound("entity", entityID.String())})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/entities/"+entityID.String(), nil, userID)
		h.GetByID(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/entities/"+entityID.String(), nil)
		h.GetByID(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid entity id", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/entities/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.GetByID(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/entities/"+entityID.String(), nil, userID)
		h.GetByID(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestEntityHandler_Update(t *testing.T) {
	userID := uuid.New()
	entityID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ent := &model.Entity{ID: entityID, Name: "Updated"}
		h := handler.NewEntityHandler(&mockEntityService{entity: ent})
		body, _ := json.Marshal(map[string]any{"name": "Updated"})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodPut, "/entities/"+entityID.String(), body, userID)
		h.Update(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entities/"+entityID.String(), nil)
		h.Update(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid entity id", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		body, _ := json.Marshal(map[string]any{"name": "Updated"})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodPut, "/entities/invalid", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Update(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodPut, "/entities/"+entityID.String(), []byte("bad"), userID)
		h.Update(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: apperror.Forbidden("not owner")})
		body, _ := json.Marshal(map[string]any{"name": "Updated"})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodPut, "/entities/"+entityID.String(), body, userID)
		h.Update(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: errors.New("db")})
		body, _ := json.Marshal(map[string]any{"name": "Updated"})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodPut, "/entities/"+entityID.String(), body, userID)
		h.Update(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestEntityHandler_Delete(t *testing.T) {
	userID := uuid.New()
	entityID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodDelete, "/entities/"+entityID.String(), nil, userID)
		h.Delete(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: apperror.Forbidden("not owner")})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodDelete, "/entities/"+entityID.String(), nil, userID)
		h.Delete(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/entities/"+entityID.String(), nil)
		h.Delete(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid entity id", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodDelete, "/entities/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Delete(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodDelete, "/entities/"+entityID.String(), nil, userID)
		h.Delete(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestEntityHandler_Search(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ents := []*model.Entity{{ID: uuid.New(), Name: "Found"}}
		h := handler.NewEntityHandler(&mockEntityService{entities: ents})
		w := httptest.NewRecorder()
		h.Search(w, entityAuthReq(http.MethodGet, "/entities/search?q=found", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("empty query returns all", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{entities: []*model.Entity{}})
		w := httptest.NewRecorder()
		h.Search(w, entityAuthReq(http.MethodGet, "/entities/search", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/entities/search?q=test", nil)
		h.Search(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: apperror.Internal(nil)})
		w := httptest.NewRecorder()
		h.Search(w, entityAuthReq(http.MethodGet, "/entities/search?q=test", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: errors.New("db")})
		w := httptest.NewRecorder()
		h.Search(w, entityAuthReq(http.MethodGet, "/entities/search?q=test", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("nil entities", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		h.Search(w, entityAuthReq(http.MethodGet, "/entities/search?q=test", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestEntityHandler_ListTypes(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{types: []string{"client", "vendor"}})
		w := httptest.NewRecorder()
		h.ListTypes(w, entityAuthReq(http.MethodGet, "/entities/types", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/entities/types", nil)
		h.ListTypes(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: apperror.Internal(nil)})
		w := httptest.NewRecorder()
		h.ListTypes(w, entityAuthReq(http.MethodGet, "/entities/types", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: errors.New("db")})
		w := httptest.NewRecorder()
		h.ListTypes(w, entityAuthReq(http.MethodGet, "/entities/types", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("nil types", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		h.ListTypes(w, entityAuthReq(http.MethodGet, "/entities/types", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestEntityHandler_LinkToDocument(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	entityID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		body, _ := json.Marshal(map[string]string{"entityId": entityID.String()})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodPost, "/documents/"+docID.String()+"/entities", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.LinkToDocument(w, r)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		body, _ := json.Marshal(map[string]string{"entityId": entityID.String()})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/documents/"+docID.String()+"/entities", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.LinkToDocument(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: apperror.Internal(nil)})
		body, _ := json.Marshal(map[string]string{"entityId": entityID.String()})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodPost, "/documents/"+docID.String()+"/entities", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.LinkToDocument(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: errors.New("db")})
		body, _ := json.Marshal(map[string]string{"entityId": entityID.String()})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodPost, "/documents/"+docID.String()+"/entities", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.LinkToDocument(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodPost, "/documents/"+docID.String()+"/entities", []byte("bad"), userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.LinkToDocument(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		body, _ := json.Marshal(map[string]string{"entityId": entityID.String()})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodPost, "/documents/invalid/entities", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.LinkToDocument(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid entity id in body", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		body, _ := json.Marshal(map[string]string{"entityId": "not-a-uuid"})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodPost, "/documents/"+docID.String()+"/entities", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.LinkToDocument(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestEntityHandler_UnlinkFromDocument(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	entityID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/entities/"+entityID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("entityId", entityID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UnlinkFromDocument(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/documents/"+docID.String()+"/entities/"+entityID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("entityId", entityID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UnlinkFromDocument(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: apperror.Internal(nil)})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/entities/"+entityID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("entityId", entityID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UnlinkFromDocument(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/entities/"+entityID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("entityId", entityID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UnlinkFromDocument(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodDelete, "/documents/invalid/entities/"+entityID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		rctx.URLParams.Add("entityId", entityID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UnlinkFromDocument(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid entity id", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodDelete, "/documents/"+docID.String()+"/entities/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		rctx.URLParams.Add("entityId", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UnlinkFromDocument(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestEntityHandler_GetDocumentEntities(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ents := []*model.Entity{{ID: uuid.New(), Name: "E1"}}
		h := handler.NewEntityHandler(&mockEntityService{entities: ents})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/documents/"+docID.String()+"/entities", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.GetDocumentEntities(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: apperror.Internal(nil)})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/documents/"+docID.String()+"/entities", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.GetDocumentEntities(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/documents/"+docID.String()+"/entities", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.GetDocumentEntities(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/documents/invalid/entities", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.GetDocumentEntities(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("nil entities", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/documents/"+docID.String()+"/entities", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", docID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.GetDocumentEntities(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestEntityHandler_ListDocuments(t *testing.T) {
	userID := uuid.New()
	entityID := uuid.New()

	t.Run("success", func(t *testing.T) {
		docs := []*model.Document{{ID: uuid.New(), Title: "D1"}}
		h := handler.NewEntityHandler(&mockEntityService{docs: docs})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/entities/"+entityID.String()+"/documents", nil, userID)
		h.ListDocuments(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: apperror.Internal(nil)})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/entities/"+entityID.String()+"/documents", nil, userID)
		h.ListDocuments(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/entities/"+entityID.String()+"/documents", nil, userID)
		h.ListDocuments(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid entity id", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/entities/invalid/documents", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.ListDocuments(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("nil docs", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := entityAuthReq(http.MethodGet, "/entities/"+entityID.String()+"/documents", nil, userID)
		h.ListDocuments(w, withEntityIDParam(r, entityID))
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestEntityHandler_CreateFromContact(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ent := &model.Entity{ID: uuid.New(), Name: "Contact"}
		h := handler.NewEntityHandler(&mockEntityService{entity: ent})
		body, _ := json.Marshal(map[string]string{"contactUserId": uuid.New().String()})
		w := httptest.NewRecorder()
		h.CreateFromContact(w, entityAuthReq(http.MethodPost, "/entities/from-contact", body, userID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		h.CreateFromContact(w, entityAuthReq(http.MethodPost, "/entities/from-contact", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		body, _ := json.Marshal(map[string]string{"contactUserId": "bad-uuid"})
		w := httptest.NewRecorder()
		h.CreateFromContact(w, entityAuthReq(http.MethodPost, "/entities/from-contact", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/entities/from-contact", nil)
		h.CreateFromContact(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: apperror.Internal(nil)})
		body, _ := json.Marshal(map[string]string{"contactUserId": uuid.New().String()})
		w := httptest.NewRecorder()
		h.CreateFromContact(w, entityAuthReq(http.MethodPost, "/entities/from-contact", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewEntityHandler(&mockEntityService{err: errors.New("db")})
		body, _ := json.Marshal(map[string]string{"contactUserId": uuid.New().String()})
		w := httptest.NewRecorder()
		h.CreateFromContact(w, entityAuthReq(http.MethodPost, "/entities/from-contact", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
