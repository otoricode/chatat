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

func topicAuthReq(method, url string, body []byte, userID uuid.UUID) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, url, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	return r.WithContext(middleware.WithUserID(r.Context(), userID))
}

func withTopicIDParam(r *http.Request, id uuid.UUID) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id.String())
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestTopicHandler_Create(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		topic := &model.Topic{ID: uuid.New(), Name: "General", CreatedAt: time.Now()}
		h := handler.NewTopicHandler(&mockTopicService{topic: topic}, &mockTopicMessageService{})
		parentID := uuid.New()
		body, _ := json.Marshal(map[string]any{
			"name":     "General",
			"icon":     "chat",
			"parentId": parentID.String(),
		})
		w := httptest.NewRecorder()
		h.Create(w, topicAuthReq(http.MethodPost, "/topics", body, userID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		h.Create(w, topicAuthReq(http.MethodPost, "/topics", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid parentId", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		body, _ := json.Marshal(map[string]any{"name": "Test", "parentId": "bad"})
		w := httptest.NewRecorder()
		h.Create(w, topicAuthReq(http.MethodPost, "/topics", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/topics", nil)
		h.Create(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: apperror.BadRequest("name required")},
			&mockTopicMessageService{},
		)
		parentID := uuid.New()
		body, _ := json.Marshal(map[string]any{"name": "", "parentId": parentID.String()})
		w := httptest.NewRecorder()
		h.Create(w, topicAuthReq(http.MethodPost, "/topics", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: errors.New("db")},
			&mockTopicMessageService{},
		)
		parentID := uuid.New()
		body, _ := json.Marshal(map[string]any{"name": "Test", "parentId": parentID.String()})
		w := httptest.NewRecorder()
		h.Create(w, topicAuthReq(http.MethodPost, "/topics", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid member id", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		parentID := uuid.New()
		body, _ := json.Marshal(map[string]any{"name": "Test", "parentId": parentID.String(), "memberIds": []string{"bad"}})
		w := httptest.NewRecorder()
		h.Create(w, topicAuthReq(http.MethodPost, "/topics", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("with valid members", func(t *testing.T) {
		topic := &model.Topic{ID: uuid.New(), Name: "General", CreatedAt: time.Now()}
		h := handler.NewTopicHandler(&mockTopicService{topic: topic}, &mockTopicMessageService{})
		parentID := uuid.New()
		body, _ := json.Marshal(map[string]any{"name": "General", "parentId": parentID.String(), "memberIds": []string{uuid.New().String()}})
		w := httptest.NewRecorder()
		h.Create(w, topicAuthReq(http.MethodPost, "/topics", body, userID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestTopicHandler_GetByID(t *testing.T) {
	userID := uuid.New()
	topicID := uuid.New()

	t.Run("success", func(t *testing.T) {
		dtl := &service.TopicDetail{Topic: model.Topic{ID: topicID, Name: "General"}}
		h := handler.NewTopicHandler(&mockTopicService{topicDtl: dtl}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/topics/"+topicID.String(), nil, userID)
		h.GetByID(w, withTopicIDParam(r, topicID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: apperror.NotFound("topic", topicID.String())},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/topics/"+topicID.String(), nil, userID)
		h.GetByID(w, withTopicIDParam(r, topicID))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/topics/"+topicID.String(), nil)
		h.GetByID(w, withTopicIDParam(r, topicID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: errors.New("db")},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/topics/"+topicID.String(), nil, userID)
		h.GetByID(w, withTopicIDParam(r, topicID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid topic id", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/topics/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.GetByID(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTopicHandler_ListByChat(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		items := []*service.TopicListItem{{Topic: model.Topic{ID: uuid.New(), Name: "T1"}}}
		h := handler.NewTopicHandler(&mockTopicService{topicList: items}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/chats/"+chatID.String()+"/topics", nil, userID)
		h.ListByChat(w, withTopicIDParam(r, chatID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String()+"/topics", nil)
		h.ListByChat(w, withTopicIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: apperror.Internal(nil)},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/chats/"+chatID.String()+"/topics", nil, userID)
		h.ListByChat(w, withTopicIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: errors.New("db")},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/chats/"+chatID.String()+"/topics", nil, userID)
		h.ListByChat(w, withTopicIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/chats/invalid/topics", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.ListByChat(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTopicHandler_ListByUser(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		items := []*service.TopicListItem{{Topic: model.Topic{ID: uuid.New(), Name: "T1"}}}
		h := handler.NewTopicHandler(&mockTopicService{topicList: items}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		h.ListByUser(w, topicAuthReq(http.MethodGet, "/topics", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/topics", nil)
		h.ListByUser(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: apperror.Internal(nil)},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		h.ListByUser(w, topicAuthReq(http.MethodGet, "/topics", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: errors.New("db")},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		h.ListByUser(w, topicAuthReq(http.MethodGet, "/topics", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestTopicHandler_Update(t *testing.T) {
	userID := uuid.New()
	tid := uuid.New()

	t.Run("success", func(t *testing.T) {
		topic := &model.Topic{ID: tid, Name: "Updated"}
		h := handler.NewTopicHandler(&mockTopicService{topic: topic}, &mockTopicMessageService{})
		body, _ := json.Marshal(map[string]any{"name": "Updated"})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPut, "/topics/"+tid.String(), body, userID)
		h.Update(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPut, "/topics/"+tid.String(), []byte("bad"), userID)
		h.Update(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/topics/"+tid.String(), nil)
		h.Update(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: apperror.Forbidden("not allowed")},
			&mockTopicMessageService{},
		)
		body, _ := json.Marshal(map[string]any{"name": "Blocked"})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPut, "/topics/"+tid.String(), body, userID)
		h.Update(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: errors.New("db")},
			&mockTopicMessageService{},
		)
		body, _ := json.Marshal(map[string]any{"name": "Test"})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPut, "/topics/"+tid.String(), body, userID)
		h.Update(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid topic id", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		body, _ := json.Marshal(map[string]any{"name": "Test"})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPut, "/topics/invalid", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Update(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTopicHandler_Delete(t *testing.T) {
	userID := uuid.New()
	tid := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodDelete, "/topics/"+tid.String(), nil, userID)
		h.Delete(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: apperror.Forbidden("not owner")},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodDelete, "/topics/"+tid.String(), nil, userID)
		h.Delete(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/topics/"+tid.String(), nil)
		h.Delete(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: apperror.Internal(nil)},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodDelete, "/topics/"+tid.String(), nil, userID)
		h.Delete(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: errors.New("db")},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodDelete, "/topics/"+tid.String(), nil, userID)
		h.Delete(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid topic id", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodDelete, "/topics/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Delete(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTopicHandler_AddMember(t *testing.T) {
	userID := uuid.New()
	tid := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		memberID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": memberID.String()})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/"+tid.String()+"/members", body, userID)
		h.AddMember(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid userId", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		body, _ := json.Marshal(map[string]string{"userId": "bad"})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/"+tid.String()+"/members", body, userID)
		h.AddMember(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/topics/"+tid.String()+"/members", nil)
		h.AddMember(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/"+tid.String()+"/members", []byte("bad"), userID)
		h.AddMember(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: apperror.Forbidden("not allowed")},
			&mockTopicMessageService{},
		)
		memberID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": memberID.String()})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/"+tid.String()+"/members", body, userID)
		h.AddMember(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: errors.New("db")},
			&mockTopicMessageService{},
		)
		memberID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": memberID.String()})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/"+tid.String()+"/members", body, userID)
		h.AddMember(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid topic id", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		memberID := uuid.New()
		body, _ := json.Marshal(map[string]string{"userId": memberID.String()})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/invalid/members", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.AddMember(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTopicHandler_RemoveMember(t *testing.T) {
	userID := uuid.New()
	tid := uuid.New()
	memberID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(
			http.MethodDelete,
			"/topics/"+tid.String()+"/members/"+memberID.String(),
			nil, userID,
		)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", tid.String())
		rctx.URLParams.Add("userId", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveMember(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(
			http.MethodDelete,
			"/topics/"+tid.String()+"/members/"+memberID.String(),
			nil,
		)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", tid.String())
		rctx.URLParams.Add("userId", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveMember(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid userId", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(
			http.MethodDelete,
			"/topics/"+tid.String()+"/members/bad",
			nil, userID,
		)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", tid.String())
		rctx.URLParams.Add("userId", "bad")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveMember(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: apperror.Forbidden("not allowed")},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(
			http.MethodDelete,
			"/topics/"+tid.String()+"/members/"+memberID.String(),
			nil, userID,
		)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", tid.String())
		rctx.URLParams.Add("userId", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveMember(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: errors.New("db")},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(
			http.MethodDelete,
			"/topics/"+tid.String()+"/members/"+memberID.String(),
			nil, userID,
		)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", tid.String())
		rctx.URLParams.Add("userId", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveMember(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid topic id", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(
			http.MethodDelete,
			"/topics/invalid/members/"+memberID.String(),
			nil, userID,
		)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		rctx.URLParams.Add("userId", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveMember(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTopicHandler_SendMessage(t *testing.T) {
	userID := uuid.New()
	tid := uuid.New()

	t.Run("success", func(t *testing.T) {
		msg := &model.TopicMessage{ID: uuid.New(), Content: "Hello"}
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{message: msg})
		body, _ := json.Marshal(map[string]string{"content": "Hello"})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/"+tid.String()+"/messages", body, userID)
		h.SendMessage(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/"+tid.String()+"/messages", []byte("bad"), userID)
		h.SendMessage(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("with reply", func(t *testing.T) {
		msg := &model.TopicMessage{ID: uuid.New(), Content: "Reply"}
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{message: msg})
		replyID := uuid.New().String()
		body, _ := json.Marshal(map[string]any{"content": "Reply", "replyToId": replyID})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/"+tid.String()+"/messages", body, userID)
		h.SendMessage(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid replyToId", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		body, _ := json.Marshal(map[string]any{"content": "Reply", "replyToId": "bad-uuid"})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/"+tid.String()+"/messages", body, userID)
		h.SendMessage(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("with custom type", func(t *testing.T) {
		msg := &model.TopicMessage{ID: uuid.New(), Content: "image.png"}
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{message: msg})
		body, _ := json.Marshal(map[string]any{"content": "image.png", "type": "image"})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/"+tid.String()+"/messages", body, userID)
		h.SendMessage(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/topics/"+tid.String()+"/messages", nil)
		h.SendMessage(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{},
			&mockTopicMessageService{err: apperror.Forbidden("not member")},
		)
		body, _ := json.Marshal(map[string]string{"content": "Hello"})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/"+tid.String()+"/messages", body, userID)
		h.SendMessage(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{},
			&mockTopicMessageService{err: errors.New("db")},
		)
		body, _ := json.Marshal(map[string]string{"content": "Hello"})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/"+tid.String()+"/messages", body, userID)
		h.SendMessage(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid topic id", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		body, _ := json.Marshal(map[string]string{"content": "Hello"})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodPost, "/topics/invalid/messages", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.SendMessage(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTopicHandler_ListMessages(t *testing.T) {
	userID := uuid.New()
	tid := uuid.New()

	t.Run("success", func(t *testing.T) {
		dtl := &service.TopicDetail{Topic: model.Topic{ID: tid}}
		page := &service.TopicMessagePage{
			Messages: []*model.TopicMessage{{ID: uuid.New()}},
			HasMore:  false,
		}
		h := handler.NewTopicHandler(
			&mockTopicService{topicDtl: dtl},
			&mockTopicMessageService{messagePage: page},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/topics/"+tid.String()+"/messages", nil, userID)
		h.ListMessages(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("topic not found", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: apperror.NotFound("topic", tid.String())},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/topics/"+tid.String()+"/messages", nil, userID)
		h.ListMessages(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/topics/"+tid.String()+"/messages", nil)
		h.ListMessages(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("message service error", func(t *testing.T) {
		dtl := &service.TopicDetail{Topic: model.Topic{ID: tid}}
		h := handler.NewTopicHandler(
			&mockTopicService{topicDtl: dtl},
			&mockTopicMessageService{err: apperror.Internal(nil)},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/topics/"+tid.String()+"/messages", nil, userID)
		h.ListMessages(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error topic", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{err: errors.New("db")},
			&mockTopicMessageService{},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/topics/"+tid.String()+"/messages", nil, userID)
		h.ListMessages(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error messages", func(t *testing.T) {
		dtl := &service.TopicDetail{Topic: model.Topic{ID: tid}}
		h := handler.NewTopicHandler(
			&mockTopicService{topicDtl: dtl},
			&mockTopicMessageService{err: errors.New("db")},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/topics/"+tid.String()+"/messages", nil, userID)
		h.ListMessages(w, withTopicIDParam(r, tid))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid topic id", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(http.MethodGet, "/topics/invalid/messages", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.ListMessages(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTopicHandler_DeleteMessage(t *testing.T) {
	userID := uuid.New()
	tid := uuid.New()
	msgID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(
			http.MethodDelete,
			"/topics/"+tid.String()+"/messages/"+msgID.String(),
			nil, userID,
		)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", tid.String())
		rctx.URLParams.Add("messageId", msgID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.DeleteMessage(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid messageId", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(
			http.MethodDelete,
			"/topics/"+tid.String()+"/messages/bad",
			nil, userID,
		)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", tid.String())
		rctx.URLParams.Add("messageId", "bad")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.DeleteMessage(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("with forAll", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := topicAuthReq(
			http.MethodDelete,
			"/topics/"+tid.String()+"/messages/"+msgID.String()+"?forAll=true",
			nil, userID,
		)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", tid.String())
		rctx.URLParams.Add("messageId", msgID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.DeleteMessage(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewTopicHandler(&mockTopicService{}, &mockTopicMessageService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(
			http.MethodDelete,
			"/topics/"+tid.String()+"/messages/"+msgID.String(),
			nil,
		)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", tid.String())
		rctx.URLParams.Add("messageId", msgID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.DeleteMessage(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{},
			&mockTopicMessageService{err: apperror.NotFound("message", msgID.String())},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(
			http.MethodDelete,
			"/topics/"+tid.String()+"/messages/"+msgID.String(),
			nil, userID,
		)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", tid.String())
		rctx.URLParams.Add("messageId", msgID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.DeleteMessage(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewTopicHandler(
			&mockTopicService{},
			&mockTopicMessageService{err: errors.New("db")},
		)
		w := httptest.NewRecorder()
		r := topicAuthReq(
			http.MethodDelete,
			"/topics/"+tid.String()+"/messages/"+msgID.String(),
			nil, userID,
		)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", tid.String())
		rctx.URLParams.Add("messageId", msgID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.DeleteMessage(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
