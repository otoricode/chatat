package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/internal/handler"
	"github.com/otoritech/chatat/internal/middleware"
	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
)

func searchAuthReq(method, url string, userID uuid.UUID) *http.Request {
	r := httptest.NewRequest(method, url, nil)
	return r.WithContext(middleware.WithUserID(r.Context(), userID))
}

func TestSearchHandler_SearchAll(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		results := &service.SearchResults{}
		h := handler.NewSearchHandler(&mockSearchService{searchResults: results})
		w := httptest.NewRecorder()
		h.SearchAll(w, searchAuthReq(http.MethodGet, "/search?q=hello&limit=5", userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("query too short", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		h.SearchAll(w, searchAuthReq(http.MethodGet, "/search?q=h", userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/search?q=test", nil)
		h.SearchAll(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{err: assert.AnError})
		w := httptest.NewRecorder()
		h.SearchAll(w, searchAuthReq(http.MethodGet, "/search?q=hello", userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSearchHandler_SearchMessages(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := []*model.MessageSearchRow{{Content: "hello"}}
		h := handler.NewSearchHandler(&mockSearchService{msgRows: rows})
		w := httptest.NewRecorder()
		h.SearchMessages(w, searchAuthReq(http.MethodGet, "/search/messages?q=hello", userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("query too short", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		h.SearchMessages(w, searchAuthReq(http.MethodGet, "/search/messages?q=h", userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/search/messages?q=hello", nil)
		h.SearchMessages(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{err: assert.AnError})
		w := httptest.NewRecorder()
		h.SearchMessages(w, searchAuthReq(http.MethodGet, "/search/messages?q=hello", userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSearchHandler_SearchDocuments(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := []*model.DocumentSearchRow{{Title: "Doc1"}}
		h := handler.NewSearchHandler(&mockSearchService{docRows: rows})
		w := httptest.NewRecorder()
		h.SearchDocuments(w, searchAuthReq(http.MethodGet, "/search/documents?q=doc", userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/search/documents?q=doc", nil)
		h.SearchDocuments(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("query too short", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		h.SearchDocuments(w, searchAuthReq(http.MethodGet, "/search/documents?q=d", userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{err: assert.AnError})
		w := httptest.NewRecorder()
		h.SearchDocuments(w, searchAuthReq(http.MethodGet, "/search/documents?q=doc", userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSearchHandler_SearchContacts(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		users := []*model.User{{ID: uuid.New(), Name: "Alice"}}
		h := handler.NewSearchHandler(&mockSearchService{contacts: users})
		w := httptest.NewRecorder()
		h.SearchContacts(w, searchAuthReq(http.MethodGet, "/search/contacts?q=alice", userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/search/contacts?q=alice", nil)
		h.SearchContacts(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("query too short", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		h.SearchContacts(w, searchAuthReq(http.MethodGet, "/search/contacts?q=a", userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{err: assert.AnError})
		w := httptest.NewRecorder()
		h.SearchContacts(w, searchAuthReq(http.MethodGet, "/search/contacts?q=alice", userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSearchHandler_SearchEntities(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ents := []*model.Entity{{ID: uuid.New(), Name: "Entity1"}}
		h := handler.NewSearchHandler(&mockSearchService{entities: ents})
		w := httptest.NewRecorder()
		h.SearchEntities(w, searchAuthReq(http.MethodGet, "/search/entities?q=ent", userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/search/entities?q=ent", nil)
		h.SearchEntities(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("query too short", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		h.SearchEntities(w, searchAuthReq(http.MethodGet, "/search/entities?q=e", userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{err: assert.AnError})
		w := httptest.NewRecorder()
		h.SearchEntities(w, searchAuthReq(http.MethodGet, "/search/entities?q=ent", userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSearchHandler_SearchInChat(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := []*model.MessageSearchRow{{Content: "hi"}}
		h := handler.NewSearchHandler(&mockSearchService{msgRows: rows})
		w := httptest.NewRecorder()
		r := searchAuthReq(http.MethodGet, "/search/chats/"+chatID.String()+"?q=hello", userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.SearchInChat(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid chatId", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		r := searchAuthReq(http.MethodGet, "/search/chats/invalid?q=hello", userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.SearchInChat(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("query too short", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		r := searchAuthReq(http.MethodGet, "/search/chats/"+chatID.String()+"?q=h", userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.SearchInChat(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{err: apperror.NotFound("chat", chatID.String())})
		w := httptest.NewRecorder()
		r := searchAuthReq(http.MethodGet, "/search/chats/"+chatID.String()+"?q=hello", userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.SearchInChat(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewSearchHandler(&mockSearchService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/search/chats/"+chatID.String()+"?q=hello", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.SearchInChat(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("success with offset and limit", func(t *testing.T) {
		rows := []*model.MessageSearchRow{{Content: "hi"}}
		h := handler.NewSearchHandler(&mockSearchService{msgRows: rows})
		w := httptest.NewRecorder()
		r := searchAuthReq(http.MethodGet, "/search/chats/"+chatID.String()+"?q=hello&offset=5&limit=50", userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.SearchInChat(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("success with invalid offset and limit", func(t *testing.T) {
		rows := []*model.MessageSearchRow{{Content: "hi"}}
		h := handler.NewSearchHandler(&mockSearchService{msgRows: rows})
		w := httptest.NewRecorder()
		r := searchAuthReq(http.MethodGet, "/search/chats/"+chatID.String()+"?q=hello&offset=abc&limit=999", userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.SearchInChat(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
