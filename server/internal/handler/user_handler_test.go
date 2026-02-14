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
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/handler"
	"github.com/otoritech/chatat/internal/middleware"
	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
)

// --- Mock UserService ---

type mockUserService struct {
	user      *model.User
	err       error
	updateErr error
}

func (m *mockUserService) GetProfile(_ context.Context, _ uuid.UUID) (*model.User, error) {
	return m.user, m.err
}

func (m *mockUserService) UpdateProfile(_ context.Context, _ uuid.UUID, input model.UpdateUserInput) (*model.User, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	if m.err != nil {
		return nil, m.err
	}
	u := *m.user
	if input.Name != nil {
		u.Name = *input.Name
	}
	if input.Avatar != nil {
		u.Avatar = *input.Avatar
	}
	if input.Status != nil {
		u.Status = *input.Status
	}
	return &u, nil
}

func (m *mockUserService) SetupProfile(_ context.Context, _ uuid.UUID, name, avatar string) (*model.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	u := *m.user
	u.Name = name
	u.Avatar = avatar
	return &u, nil
}

func (m *mockUserService) UpdateLastSeen(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockUserService) DeleteAccount(_ context.Context, _ uuid.UUID) error {
	return m.err
}

// --- Mock ContactService ---

type mockContactService struct {
	matches []service.ContactMatch
	list    []service.ContactInfo
	user    *model.User
	profile *service.ContactInfo
	err     error
}

func (m *mockContactService) SyncContacts(_ context.Context, _ uuid.UUID, _ []string) ([]service.ContactMatch, error) {
	return m.matches, m.err
}

func (m *mockContactService) GetContacts(_ context.Context, _ uuid.UUID) ([]service.ContactInfo, error) {
	return m.list, m.err
}

func (m *mockContactService) SearchByPhone(_ context.Context, _ string) (*model.User, error) {
	return m.user, m.err
}

func (m *mockContactService) GetContactProfile(_ context.Context, _ uuid.UUID) (*service.ContactInfo, error) {
	return m.profile, m.err
}

// --- Helpers ---

func authReq(method, url string, body []byte, userID uuid.UUID) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	return req.WithContext(middleware.WithUserID(req.Context(), userID))
}

// --- UserHandler Tests ---

func TestUserHandler_GetMe(t *testing.T) {
	userID := uuid.New()
	user := &model.User{ID: userID, Phone: "+6281234567890", Name: "Test", Avatar: "\U0001F60A"}

	t.Run("success", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		w := httptest.NewRecorder()
		h.GetMe(w, authReq(http.MethodGet, "/api/v1/users/me", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{err: apperror.NotFound("user", userID.String())})
		w := httptest.NewRecorder()
		h.GetMe(w, authReq(http.MethodGet, "/api/v1/users/me", nil, userID))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		w := httptest.NewRecorder()
		h.GetMe(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{err: errors.New("db")})
		w := httptest.NewRecorder()
		h.GetMe(w, authReq(http.MethodGet, "/api/v1/users/me", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_UpdateMe(t *testing.T) {
	userID := uuid.New()
	user := &model.User{ID: userID, Phone: "+6281234567890", Name: "Old", Avatar: "\U0001F60A"}

	t.Run("success", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		body, _ := json.Marshal(map[string]any{"name": "New"})
		w := httptest.NewRecorder()
		h.UpdateMe(w, authReq(http.MethodPut, "/api/v1/users/me", body, userID))
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		data := resp["data"].(map[string]any)
		assert.Equal(t, "New", data["name"])
	})

	t.Run("no fields", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		body, _ := json.Marshal(map[string]any{})
		w := httptest.NewRecorder()
		h.UpdateMe(w, authReq(http.MethodPut, "/api/v1/users/me", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/api/v1/users/me", nil)
		h.UpdateMe(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		w := httptest.NewRecorder()
		h.UpdateMe(w, authReq(http.MethodPut, "/api/v1/users/me", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{err: apperror.Internal(nil)})
		body, _ := json.Marshal(map[string]any{"name": "New"})
		w := httptest.NewRecorder()
		h.UpdateMe(w, authReq(http.MethodPut, "/api/v1/users/me", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{err: errors.New("db")})
		body, _ := json.Marshal(map[string]any{"name": "New"})
		w := httptest.NewRecorder()
		h.UpdateMe(w, authReq(http.MethodPut, "/api/v1/users/me", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_SetupProfile(t *testing.T) {
	userID := uuid.New()
	user := &model.User{ID: userID, Phone: "+6281234567890", Name: "", Avatar: ""}

	t.Run("success", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		body, _ := json.Marshal(map[string]any{"name": "Andi", "avatar": "\U0001F60A"})
		w := httptest.NewRecorder()
		h.SetupProfile(w, authReq(http.MethodPost, "/api/v1/users/me/setup", body, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/setup", nil)
		h.SetupProfile(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		w := httptest.NewRecorder()
		h.SetupProfile(w, authReq(http.MethodPost, "/api/v1/users/me/setup", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{err: apperror.Internal(nil)})
		body, _ := json.Marshal(map[string]any{"name": "Andi", "avatar": "\U0001F60A"})
		w := httptest.NewRecorder()
		h.SetupProfile(w, authReq(http.MethodPost, "/api/v1/users/me/setup", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{err: errors.New("db")})
		body, _ := json.Marshal(map[string]any{"name": "Andi", "avatar": "\U0001F60A"})
		w := httptest.NewRecorder()
		h.SetupProfile(w, authReq(http.MethodPost, "/api/v1/users/me/setup", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_DeleteAccount(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{})
		w := httptest.NewRecorder()
		h.DeleteAccount(w, authReq(http.MethodDelete, "/api/v1/users/me", nil, userID))
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/api/v1/users/me", nil)
		h.DeleteAccount(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{err: apperror.Internal(nil)})
		w := httptest.NewRecorder()
		h.DeleteAccount(w, authReq(http.MethodDelete, "/api/v1/users/me", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{err: errors.New("db")})
		w := httptest.NewRecorder()
		h.DeleteAccount(w, authReq(http.MethodDelete, "/api/v1/users/me", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// --- ContactHandler Tests ---

func TestContactHandler_Sync(t *testing.T) {
	userID := uuid.New()
	matches := []service.ContactMatch{
		{PhoneHash: "abc123", UserID: uuid.New(), Name: "Alice"},
	}

	t.Run("success", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{matches: matches})
		body, _ := json.Marshal(map[string]any{"phoneHashes": []string{"abc123"}})
		w := httptest.NewRecorder()
		h.Sync(w, authReq(http.MethodPost, "/api/v1/contacts/sync", body, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("empty hashes", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{})
		body, _ := json.Marshal(map[string]any{"phoneHashes": []string{}})
		w := httptest.NewRecorder()
		h.Sync(w, authReq(http.MethodPost, "/api/v1/contacts/sync", body, userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{})
		w := httptest.NewRecorder()
		h.Sync(w, authReq(http.MethodPost, "/api/v1/contacts/sync", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/contacts/sync", nil)
		h.Sync(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{err: apperror.Internal(nil)})
		body, _ := json.Marshal(map[string]any{"phoneHashes": []string{"abc123"}})
		w := httptest.NewRecorder()
		h.Sync(w, authReq(http.MethodPost, "/api/v1/contacts/sync", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{err: errors.New("db")})
		body, _ := json.Marshal(map[string]any{"phoneHashes": []string{"abc123"}})
		w := httptest.NewRecorder()
		h.Sync(w, authReq(http.MethodPost, "/api/v1/contacts/sync", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestContactHandler_List(t *testing.T) {
	userID := uuid.New()
	list := []service.ContactInfo{
		{UserID: uuid.New(), Name: "Alice", IsOnline: true},
	}

	t.Run("success", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{list: list})
		w := httptest.NewRecorder()
		h.List(w, authReq(http.MethodGet, "/api/v1/contacts", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/contacts", nil)
		h.List(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{err: apperror.Internal(nil)})
		w := httptest.NewRecorder()
		h.List(w, authReq(http.MethodGet, "/api/v1/contacts", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{err: errors.New("db")})
		w := httptest.NewRecorder()
		h.List(w, authReq(http.MethodGet, "/api/v1/contacts", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestContactHandler_Search(t *testing.T) {
	user := &model.User{ID: uuid.New(), Phone: "+6281234567890", Name: "Found"}

	t.Run("success", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{user: user})
		req := authReq(http.MethodGet, "/api/v1/contacts/search?phone=%2B6281234567890", nil, uuid.New())
		w := httptest.NewRecorder()
		h.Search(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing phone", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{})
		req := authReq(http.MethodGet, "/api/v1/contacts/search", nil, uuid.New())
		w := httptest.NewRecorder()
		h.Search(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid phone", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{})
		req := authReq(http.MethodGet, "/api/v1/contacts/search?phone=abc", nil, uuid.New())
		w := httptest.NewRecorder()
		h.Search(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{err: apperror.Internal(nil)})
		req := authReq(http.MethodGet, "/api/v1/contacts/search?phone=%2B6281234567890", nil, uuid.New())
		w := httptest.NewRecorder()
		h.Search(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{err: errors.New("db")})
		req := authReq(http.MethodGet, "/api/v1/contacts/search?phone=%2B6281234567890", nil, uuid.New())
		w := httptest.NewRecorder()
		h.Search(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestContactHandler_GetProfile(t *testing.T) {
	contactID := uuid.New()
	profile := &service.ContactInfo{
		UserID:   contactID,
		Name:     "Profile",
		IsOnline: false,
		LastSeen: time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{profile: profile})
		req := authReq(http.MethodGet, "/api/v1/contacts/"+contactID.String(), nil, uuid.New())
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("userId", contactID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.GetProfile(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid userId", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{profile: profile})
		req := authReq(http.MethodGet, "/api/v1/contacts/not-a-uuid", nil, uuid.New())
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("userId", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.GetProfile(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error not found", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{err: apperror.NotFound("contact", contactID.String())})
		req := authReq(http.MethodGet, "/api/v1/contacts/"+contactID.String(), nil, uuid.New())
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("userId", contactID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.GetProfile(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewContactHandler(&mockContactService{err: errors.New("db")})
		req := authReq(http.MethodGet, "/api/v1/contacts/"+contactID.String(), nil, uuid.New())
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("userId", contactID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.GetProfile(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// --- Privacy Settings Tests ---

func TestUserHandler_GetPrivacySettings(t *testing.T) {
	userID := uuid.New()
	ps := model.DefaultPrivacySettings()
	user := &model.User{ID: userID, Name: "Test", PrivacySettings: ps}

	t.Run("success", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		w := httptest.NewRecorder()
		h.GetPrivacySettings(w, authReq(http.MethodGet, "/api/v1/users/me/privacy", nil, userID))
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "everyone")
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/privacy", nil)
		h.GetPrivacySettings(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{err: apperror.Internal(nil)})
		w := httptest.NewRecorder()
		h.GetPrivacySettings(w, authReq(http.MethodGet, "/api/v1/users/me/privacy", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{err: errors.New("db")})
		w := httptest.NewRecorder()
		h.GetPrivacySettings(w, authReq(http.MethodGet, "/api/v1/users/me/privacy", nil, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_UpdatePrivacySettings(t *testing.T) {
	userID := uuid.New()
	ps := model.DefaultPrivacySettings()
	user := &model.User{ID: userID, Name: "Test", PrivacySettings: ps}

	t.Run("success", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		body, _ := json.Marshal(map[string]any{"lastSeenVisibility": "contacts"})
		w := httptest.NewRecorder()
		h.UpdatePrivacySettings(w, authReq(http.MethodPut, "/api/v1/users/me/privacy", body, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		w := httptest.NewRecorder()
		h.UpdatePrivacySettings(w, authReq(http.MethodPut, "/api/v1/users/me/privacy", []byte("bad"), userID))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/privacy", nil)
		h.UpdatePrivacySettings(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("get profile error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{err: apperror.NotFound("user", userID.String())})
		body, _ := json.Marshal(map[string]any{"lastSeenVisibility": "contacts"})
		w := httptest.NewRecorder()
		h.UpdatePrivacySettings(w, authReq(http.MethodPut, "/api/v1/users/me/privacy", body, userID))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get profile generic error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{err: errors.New("db")})
		body, _ := json.Marshal(map[string]any{"lastSeenVisibility": "contacts"})
		w := httptest.NewRecorder()
		h.UpdatePrivacySettings(w, authReq(http.MethodPut, "/api/v1/users/me/privacy", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("success all fields", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user})
		readReceipts := true
		body, _ := json.Marshal(map[string]any{
			"lastSeenVisibility":     "nobody",
			"onlineVisibility":       "contacts",
			"readReceipts":           readReceipts,
			"profilePhotoVisibility": "everyone",
		})
		w := httptest.NewRecorder()
		h.UpdatePrivacySettings(w, authReq(http.MethodPut, "/api/v1/users/me/privacy", body, userID))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("update profile service error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user, updateErr: apperror.Internal(nil)})
		body, _ := json.Marshal(map[string]any{"lastSeenVisibility": "contacts"})
		w := httptest.NewRecorder()
		h.UpdatePrivacySettings(w, authReq(http.MethodPut, "/api/v1/users/me/privacy", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("update profile generic error", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserService{user: user, updateErr: errors.New("db")})
		body, _ := json.Marshal(map[string]any{"lastSeenVisibility": "contacts"})
		w := httptest.NewRecorder()
		h.UpdatePrivacySettings(w, authReq(http.MethodPut, "/api/v1/users/me/privacy", body, userID))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
