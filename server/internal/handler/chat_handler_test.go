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

func chatAuthReq(method, url string, body []byte, userID uuid.UUID) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, url, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	return r.WithContext(middleware.WithUserID(r.Context(), userID))
}

func withChatIDParam(r *http.Request, chatID uuid.UUID) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", chatID.String())
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func withMsgIDParam(r *http.Request, chatID, msgID uuid.UUID) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", chatID.String())
	rctx.URLParams.Add("messageId", msgID.String())
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// --- List ---

func TestChatHandler_List(t *testing.T) {
	userID := uuid.New()
	items := []*service.ChatListItem{
		{Chat: model.Chat{ID: uuid.New(), Name: "Chat1"}},
	}

	t.Run("success", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{chatList: items}, nil, nil)
		w := httptest.NewRecorder()
		h.List(w, chatAuthReq(http.MethodGet, "/api/v1/chats", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/chats", nil)
		h.List(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{err: apperror.Internal(assert.AnError)}, nil, nil)
		w := httptest.NewRecorder()
		h.List(w, chatAuthReq(http.MethodGet, "/api/v1/chats", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{err: errors.New("db")}, nil, nil)
		w := httptest.NewRecorder()
		h.List(w, chatAuthReq(http.MethodGet, "/api/v1/chats", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// --- Create ---

func TestChatHandler_Create(t *testing.T) {
	userID := uuid.New()
	contactID := uuid.New()

	t.Run("personal chat success", func(t *testing.T) {
		chat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal}
		h := handler.NewChatHandler(&mockChatService{chat: chat}, nil, nil)
		body, _ := json.Marshal(map[string]string{"contactId": contactID.String()})
		w := httptest.NewRecorder()
		h.Create(w, chatAuthReq(http.MethodPost, "/api/v1/chats", body, userID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("group chat success", func(t *testing.T) {
		chat := &model.Chat{ID: uuid.New(), Type: model.ChatTypeGroup, Name: "Team"}
		h := handler.NewChatHandler(nil, nil, &mockGroupService{chat: chat})
		body, _ := json.Marshal(map[string]any{
			"type":      "group",
			"name":      "Team",
			"icon":      "💬",
			"memberIds": []string{uuid.New().String()},
		})
		w := httptest.NewRecorder()
		h.Create(w, chatAuthReq(http.MethodPost, "/api/v1/chats", body, userID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("missing contactId for personal", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		body, _ := json.Marshal(map[string]string{})
		w := httptest.NewRecorder()
		h.Create(w, chatAuthReq(http.MethodPost, "/api/v1/chats", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid contactId", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		body, _ := json.Marshal(map[string]string{"contactId": "not-uuid"})
		w := httptest.NewRecorder()
		h.Create(w, chatAuthReq(http.MethodPost, "/api/v1/chats", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid member id in group", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		body, _ := json.Marshal(map[string]any{
			"type":      "group",
			"name":      "T",
			"memberIds": []string{"invalid"},
		})
		w := httptest.NewRecorder()
		h.Create(w, chatAuthReq(http.MethodPost, "/api/v1/chats", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/chats", nil)
		h.Create(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		w := httptest.NewRecorder()
		h.Create(w, chatAuthReq(http.MethodPost, "/api/v1/chats", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("personal service error", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{err: errors.New("db")}, nil, nil)
		body, _ := json.Marshal(map[string]string{"contactId": contactID.String()})
		w := httptest.NewRecorder()
		h.Create(w, chatAuthReq(http.MethodPost, "/api/v1/chats", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("group service error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: errors.New("db")})
		body, _ := json.Marshal(map[string]any{
			"type":      "group",
			"name":      "Team",
			"memberIds": []string{uuid.New().String()},
		})
		w := httptest.NewRecorder()
		h.Create(w, chatAuthReq(http.MethodPost, "/api/v1/chats", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// --- GetByID ---

func TestChatHandler_GetByID(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		detail := &service.ChatDetail{Chat: model.Chat{ID: chatID}}
		h := handler.NewChatHandler(&mockChatService{chatDetail: detail}, nil, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String(), nil, userID)
		h.GetByID(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{err: apperror.NotFound("chat", chatID.String())}, nil, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String(), nil, userID)
		h.GetByID(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{err: errors.New("db")}, nil, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String(), nil, userID)
		h.GetByID(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/chats/"+chatID.String(), nil)
		h.GetByID(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.GetByID(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- Update ---

func TestChatHandler_Update(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		chat := &model.Chat{ID: chatID, Name: "Updated"}
		h := handler.NewChatHandler(nil, nil, &mockGroupService{chat: chat})
		body, _ := json.Marshal(map[string]string{"name": "Updated"})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/"+chatID.String(), body, userID)
		h.Update(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/api/v1/chats/"+chatID.String(), nil)
		h.Update(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/"+chatID.String(), []byte("bad"), userID)
		h.Update(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: apperror.Forbidden("not admin")})
		body, _ := json.Marshal(map[string]string{"name": "Updated"})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/"+chatID.String(), body, userID)
		h.Update(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: errors.New("db")})
		body, _ := json.Marshal(map[string]string{"name": "Updated"})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/"+chatID.String(), body, userID)
		h.Update(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		body, _ := json.Marshal(map[string]string{"name": "Updated"})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/invalid", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Update(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- Delete ---

func TestChatHandler_Delete(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String(), nil, userID)
		h.Delete(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: apperror.Forbidden("not admin")})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String(), nil, userID)
		h.Delete(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/api/v1/chats/"+chatID.String(), nil)
		h.Delete(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String(), nil, userID)
		h.Delete(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.Delete(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- PinChat / UnpinChat ---

func TestChatHandler_PinUnpin(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("pin success", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/"+chatID.String()+"/pin", nil, userID)
		h.PinChat(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("unpin success", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/pin", nil, userID)
		h.UnpinChat(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("pin unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/api/v1/chats/"+chatID.String()+"/pin", nil)
		h.PinChat(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("pin service error", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{err: apperror.Internal(nil)}, nil, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/"+chatID.String()+"/pin", nil, userID)
		h.PinChat(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("unpin unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/pin", nil)
		h.UnpinChat(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("unpin service error", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{err: apperror.Internal(nil)}, nil, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/pin", nil, userID)
		h.UnpinChat(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("pin invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/invalid/pin", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.PinChat(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unpin invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, nil, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/invalid/pin", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.UnpinChat(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- SendMessage ---

func TestChatHandler_SendMessage(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		msg := &model.Message{ID: uuid.New(), ChatID: chatID, Content: "hi", SenderID: userID, CreatedAt: time.Now()}
		h := handler.NewChatHandler(nil, &mockMessageService{message: msg}, nil)
		body, _ := json.Marshal(map[string]string{"content": "hi", "type": "text"})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/messages", body, userID)
		h.SendMessage(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/messages", []byte("bad"), userID)
		h.SendMessage(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/messages", nil)
		h.SendMessage(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{err: apperror.Forbidden("not a member")}, nil)
		body, _ := json.Marshal(map[string]string{"content": "hi", "type": "text"})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/messages", body, userID)
		h.SendMessage(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{err: errors.New("db")}, nil)
		body, _ := json.Marshal(map[string]string{"content": "hi", "type": "text"})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/messages", body, userID)
		h.SendMessage(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/invalid/messages", []byte(`{"content":"hi","type":"text"}`), userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.SendMessage(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- ListMessages ---

func TestChatHandler_ListMessages(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		page := &service.MessagePage{Messages: []*model.Message{}, HasMore: false}
		h := handler.NewChatHandler(&mockChatService{isMember: true}, &mockMessageService{messagePage: page}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages", nil, userID)
		h.ListMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages", nil)
		h.ListMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("not a member", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{isMember: false}, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages", nil, userID)
		h.ListMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{isMember: true}, &mockMessageService{err: apperror.Internal(nil)}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages", nil, userID)
		h.ListMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{isMember: true}, &mockMessageService{err: errors.New("db")}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages", nil, userID)
		h.ListMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("membership check error", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{err: errors.New("db")}, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages", nil, userID)
		h.ListMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/invalid/messages", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.ListMessages(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- DeleteMessage ---

func TestChatHandler_DeleteMessage(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()
	msgID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/messages/"+msgID.String(), nil, userID)
		h.DeleteMessage(w, withMsgIDParam(r, chatID, msgID))
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/messages/"+msgID.String(), nil)
		h.DeleteMessage(w, withMsgIDParam(r, chatID, msgID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{err: apperror.NotFound("message", msgID.String())}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/messages/"+msgID.String(), nil, userID)
		h.DeleteMessage(w, withMsgIDParam(r, chatID, msgID))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{err: errors.New("db")}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/messages/"+msgID.String(), nil, userID)
		h.DeleteMessage(w, withMsgIDParam(r, chatID, msgID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid message id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/messages/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		rctx.URLParams.Add("messageId", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.DeleteMessage(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("with forAll", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/messages/"+msgID.String()+"?forAll=true", nil, userID)
		h.DeleteMessage(w, withMsgIDParam(r, chatID, msgID))
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

// --- ForwardMessage ---

func TestChatHandler_ForwardMessage(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()
	msgID := uuid.New()
	targetID := uuid.New()

	t.Run("success", func(t *testing.T) {
		msg := &model.Message{ID: uuid.New(), Content: "fwd", CreatedAt: time.Now()}
		h := handler.NewChatHandler(nil, &mockMessageService{message: msg}, nil)
		body, _ := json.Marshal(map[string]string{"targetChatId": targetID.String()})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/messages/"+msgID.String()+"/forward", body, userID)
		h.ForwardMessage(w, withMsgIDParam(r, chatID, msgID))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/messages/"+msgID.String()+"/forward", []byte("bad"), userID)
		h.ForwardMessage(w, withMsgIDParam(r, chatID, msgID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/messages/"+msgID.String()+"/forward", nil)
		h.ForwardMessage(w, withMsgIDParam(r, chatID, msgID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{err: apperror.Internal(nil)}, nil)
		body, _ := json.Marshal(map[string]string{"targetChatId": targetID.String()})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/messages/"+msgID.String()+"/forward", body, userID)
		h.ForwardMessage(w, withMsgIDParam(r, chatID, msgID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{err: errors.New("db")}, nil)
		body, _ := json.Marshal(map[string]string{"targetChatId": targetID.String()})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/messages/"+msgID.String()+"/forward", body, userID)
		h.ForwardMessage(w, withMsgIDParam(r, chatID, msgID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid message id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/messages/invalid/forward", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		rctx.URLParams.Add("messageId", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.ForwardMessage(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid targetChatId", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		body, _ := json.Marshal(map[string]string{"targetChatId": "invalid"})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/messages/"+msgID.String()+"/forward", body, userID)
		h.ForwardMessage(w, withMsgIDParam(r, chatID, msgID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- SearchMessages ---

func TestChatHandler_SearchMessages(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		msgs := []*model.Message{{ID: uuid.New(), Content: "match", CreatedAt: time.Now()}}
		h := handler.NewChatHandler(&mockChatService{isMember: true}, &mockMessageService{messages: msgs}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages/search?q=match", nil, userID)
		h.SearchMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("no query", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{isMember: true}, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages/search", nil, userID)
		h.SearchMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages/search?q=test", nil)
		h.SearchMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("not a member", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{isMember: false}, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages/search?q=test", nil, userID)
		h.SearchMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{isMember: true}, &mockMessageService{err: apperror.Internal(nil)}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages/search?q=test", nil, userID)
		h.SearchMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{isMember: true}, &mockMessageService{err: errors.New("db")}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages/search?q=test", nil, userID)
		h.SearchMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("membership check error", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{err: errors.New("db")}, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/messages/search?q=test", nil, userID)
		h.SearchMessages(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(&mockChatService{}, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/invalid/messages/search?q=test", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.SearchMessages(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- Group operations ---

func TestChatHandler_AddMember(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		body, _ := json.Marshal(map[string]string{"userId": uuid.New().String()})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/members", body, userID)
		h.AddMember(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/members", []byte("bad"), userID)
		h.AddMember(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/members", nil)
		h.AddMember(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: apperror.Forbidden("not admin")})
		body, _ := json.Marshal(map[string]string{"userId": uuid.New().String()})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/members", body, userID)
		h.AddMember(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: errors.New("db")})
		body, _ := json.Marshal(map[string]string{"userId": uuid.New().String()})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/members", body, userID)
		h.AddMember(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		body, _ := json.Marshal(map[string]string{"userId": uuid.New().String()})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/invalid/members", body, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.AddMember(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid userId format", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		body, _ := json.Marshal(map[string]string{"userId": "not-a-uuid"})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/members", body, userID)
		h.AddMember(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestChatHandler_RemoveMember(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()
	memberID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/members/"+memberID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		rctx.URLParams.Add("memberID", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveMember(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/members/"+memberID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		rctx.URLParams.Add("memberID", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveMember(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: apperror.Forbidden("not admin")})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/members/"+memberID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		rctx.URLParams.Add("memberID", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveMember(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/members/"+memberID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		rctx.URLParams.Add("memberID", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveMember(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/invalid/members/"+memberID.String(), nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		rctx.URLParams.Add("memberID", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveMember(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid member id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodDelete, "/api/v1/chats/"+chatID.String()+"/members/invalid", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		rctx.URLParams.Add("memberID", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.RemoveMember(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestChatHandler_PromoteToAdmin(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()
	memberID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/"+chatID.String()+"/members/"+memberID.String()+"/admin", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		rctx.URLParams.Add("memberID", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.PromoteToAdmin(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/api/v1/chats/"+chatID.String()+"/members/"+memberID.String()+"/admin", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		rctx.URLParams.Add("memberID", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.PromoteToAdmin(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: apperror.Forbidden("not admin")})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/"+chatID.String()+"/members/"+memberID.String()+"/admin", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		rctx.URLParams.Add("memberID", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.PromoteToAdmin(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/"+chatID.String()+"/members/"+memberID.String()+"/admin", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		rctx.URLParams.Add("memberID", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.PromoteToAdmin(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/invalid/members/"+memberID.String()+"/admin", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		rctx.URLParams.Add("memberID", memberID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.PromoteToAdmin(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid member id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPut, "/api/v1/chats/"+chatID.String()+"/members/invalid/admin", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", chatID.String())
		rctx.URLParams.Add("memberID", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.PromoteToAdmin(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestChatHandler_LeaveGroup(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/leave", nil, userID)
		h.LeaveGroup(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/leave", nil)
		h.LeaveGroup(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: apperror.Internal(nil)})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/leave", nil, userID)
		h.LeaveGroup(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/leave", nil, userID)
		h.LeaveGroup(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/invalid/leave", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.LeaveGroup(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestChatHandler_GetGroupInfo(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		info := &service.GroupInfo{Chat: model.Chat{ID: chatID, Name: "G"}}
		h := handler.NewChatHandler(nil, nil, &mockGroupService{groupInfo: info})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/info", nil, userID)
		h.GetGroupInfo(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/info", nil)
		h.GetGroupInfo(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: apperror.NotFound("group", chatID.String())})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/info", nil, userID)
		h.GetGroupInfo(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{err: errors.New("db")})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/"+chatID.String()+"/info", nil, userID)
		h.GetGroupInfo(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, nil, &mockGroupService{})
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodGet, "/api/v1/chats/invalid/info", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.GetGroupInfo(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestChatHandler_MarkAsRead(t *testing.T) {
	userID := uuid.New()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/read", nil, userID)
		h.MarkAsRead(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/read", nil)
		h.MarkAsRead(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{err: apperror.Internal(nil)}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/read", nil, userID)
		h.MarkAsRead(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{err: errors.New("db")}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/"+chatID.String()+"/read", nil, userID)
		h.MarkAsRead(w, withChatIDParam(r, chatID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid chat id", func(t *testing.T) {
		h := handler.NewChatHandler(nil, &mockMessageService{}, nil)
		w := httptest.NewRecorder()
		r := chatAuthReq(http.MethodPost, "/api/v1/chats/invalid/read", nil, userID)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.MarkAsRead(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
