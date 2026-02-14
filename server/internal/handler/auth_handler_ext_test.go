package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/handler"
	"github.com/otoritech/chatat/internal/middleware"
	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
)

func newAuthHandler(otp *mockOTPService, rev *mockReverseOTPService, tok *mockTokenService, sess *mockSessionService, repo *mockUserRepo) *handler.AuthHandler {
	return handler.NewAuthHandler(otp, rev, tok, sess, repo)
}

func authReqNoUser(method, url string, body []byte) *http.Request {
	if body != nil {
		r := httptest.NewRequest(method, url, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		return r
	}
	return httptest.NewRequest(method, url, nil)
}

func authReqWithUserID(method, url string, body []byte, userID uuid.UUID) *http.Request {
	r := authReqNoUser(method, url, body)
	return r.WithContext(middleware.WithUserID(r.Context(), userID))
}

func defaultTokenPair() *service.TokenPair {
	return &service.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
	}
}

func defaultUser() *model.User {
	return &model.User{
		ID:    uuid.New(),
		Phone: "+6281234567890",
		Name:  "Test User",
	}
}

// --- SendOTP Tests ---

func TestAuthHandler_SendOTP(t *testing.T) {
	t.Run("invalid body", func(t *testing.T) {
		h := newAuthHandler(&mockOTPService{}, &mockReverseOTPService{}, &mockTokenService{}, &mockSessionService{}, &mockUserRepo{})
		w := httptest.NewRecorder()
		h.SendOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/send", []byte("bad")))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid phone", func(t *testing.T) {
		h := newAuthHandler(&mockOTPService{}, &mockReverseOTPService{}, &mockTokenService{}, &mockSessionService{}, &mockUserRepo{})
		body, _ := json.Marshal(map[string]string{"phone": "invalid"})
		w := httptest.NewRecorder()
		h.SendOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/send", body))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error AppError", func(t *testing.T) {
		h := newAuthHandler(&mockOTPService{err: apperror.RateLimited()}, &mockReverseOTPService{}, &mockTokenService{}, &mockSessionService{}, &mockUserRepo{})
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890"})
		w := httptest.NewRecorder()
		h.SendOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/send", body))
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("service error generic", func(t *testing.T) {
		h := newAuthHandler(&mockOTPService{err: errors.New("redis down")}, &mockReverseOTPService{}, &mockTokenService{}, &mockSessionService{}, &mockUserRepo{})
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890"})
		w := httptest.NewRecorder()
		h.SendOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/send", body))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		h := newAuthHandler(&mockOTPService{code: "123456"}, &mockReverseOTPService{}, &mockTokenService{}, &mockSessionService{}, &mockUserRepo{})
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890"})
		w := httptest.NewRecorder()
		h.SendOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/send", body))
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAuthHandler_VerifyOTP(t *testing.T) {
	userID := uuid.New()
	user := &model.User{ID: userID, Phone: "+6281234567890", Name: "User"}
	tokens := &service.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresAt: time.Now().Add(time.Hour).Unix()}

	t.Run("success", func(t *testing.T) {
		h := newAuthHandler(
			&mockOTPService{},
			nil,
			&mockTokenService{tokenPair: tokens},
			&mockSessionService{},
			&mockUserRepo{user: user},
		)
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890", "code": "123456", "deviceId": "dev1"})
		w := httptest.NewRecorder()
		h.VerifyOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/verify", body))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newAuthHandler(&mockOTPService{}, nil, nil, nil, nil)
		w := httptest.NewRecorder()
		h.VerifyOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/verify", []byte("bad")))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid phone", func(t *testing.T) {
		h := newAuthHandler(&mockOTPService{}, nil, nil, nil, nil)
		body, _ := json.Marshal(map[string]string{"phone": "abc", "code": "123456"})
		w := httptest.NewRecorder()
		h.VerifyOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/verify", body))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("wrong code", func(t *testing.T) {
		h := newAuthHandler(
			&mockOTPService{err: apperror.Unauthorized("invalid code")},
			nil, nil, nil, nil,
		)
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890", "code": "000000"})
		w := httptest.NewRecorder()
		h.VerifyOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/verify", body))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("otp verify generic error", func(t *testing.T) {
		h := newAuthHandler(
			&mockOTPService{err: errors.New("redis down")},
			nil, nil, nil, nil,
		)
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890", "code": "000000"})
		w := httptest.NewRecorder()
		h.VerifyOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/verify", body))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("new user created", func(t *testing.T) {
		findErr := errors.New("not found")
		h := newAuthHandler(
			&mockOTPService{}, nil,
			&mockTokenService{tokenPair: tokens},
			&mockSessionService{},
			&mockUserRepo{user: user, findPhoneErr: &findErr},
		)
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890", "code": "123456", "deviceId": "d1"})
		w := httptest.NewRecorder()
		h.VerifyOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/verify", body))
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		data := resp["data"].(map[string]any)
		assert.Equal(t, true, data["isNewUser"])
	})

	t.Run("create user fails", func(t *testing.T) {
		findErr := errors.New("not found")
		h := newAuthHandler(
			&mockOTPService{}, nil,
			&mockTokenService{tokenPair: tokens},
			&mockSessionService{},
			&mockUserRepo{findPhoneErr: &findErr, err: errors.New("db error")},
		)
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890", "code": "123456", "deviceId": "d1"})
		w := httptest.NewRecorder()
		h.VerifyOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/verify", body))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("token generate fails", func(t *testing.T) {
		h := newAuthHandler(
			&mockOTPService{}, nil,
			&mockTokenService{err: errors.New("token error")},
			&mockSessionService{},
			&mockUserRepo{user: user},
		)
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890", "code": "123456", "deviceId": "d1"})
		w := httptest.NewRecorder()
		h.VerifyOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/verify", body))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("success without device id", func(t *testing.T) {
		h := newAuthHandler(
			&mockOTPService{}, nil,
			&mockTokenService{tokenPair: tokens},
			&mockSessionService{},
			&mockUserRepo{user: user},
		)
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890", "code": "123456"})
		w := httptest.NewRecorder()
		h.VerifyOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/otp/verify", body))
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAuthHandler_InitReverseOTP(t *testing.T) {
	session := &service.ReverseOTPSession{
		SessionID:      "sess-1",
		TargetWANumber: "+628999",
		UniqueCode:     "ABCD",
		ExpiresAt:      time.Now().Add(5 * time.Minute),
	}

	t.Run("success", func(t *testing.T) {
		h := newAuthHandler(nil, &mockReverseOTPService{session: session}, nil, nil, nil)
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890"})
		w := httptest.NewRecorder()
		h.InitReverseOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/reverse-otp/init", body))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newAuthHandler(nil, &mockReverseOTPService{}, nil, nil, nil)
		w := httptest.NewRecorder()
		h.InitReverseOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/reverse-otp/init", []byte("bad")))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid phone", func(t *testing.T) {
		h := newAuthHandler(nil, &mockReverseOTPService{}, nil, nil, nil)
		body, _ := json.Marshal(map[string]string{"phone": "bad"})
		w := httptest.NewRecorder()
		h.InitReverseOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/reverse-otp/init", body))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		h := newAuthHandler(nil, &mockReverseOTPService{err: apperror.Internal(nil)}, nil, nil, nil)
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890"})
		w := httptest.NewRecorder()
		h.InitReverseOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/reverse-otp/init", body))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("service generic error", func(t *testing.T) {
		h := newAuthHandler(nil, &mockReverseOTPService{err: errors.New("redis")}, nil, nil, nil)
		body, _ := json.Marshal(map[string]string{"phone": "+6281234567890"})
		w := httptest.NewRecorder()
		h.InitReverseOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/reverse-otp/init", body))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAuthHandler_CheckReverseOTP(t *testing.T) {
	userID := uuid.New()
	user := &model.User{ID: userID, Phone: "+6281234567890", Name: "User"}
	tokens := &service.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresAt: time.Now().Add(time.Hour).Unix()}

	t.Run("verified", func(t *testing.T) {
		h := newAuthHandler(
			nil,
			&mockReverseOTPService{result: &service.VerificationResult{Status: "verified", Phone: "+6281234567890"}},
			&mockTokenService{tokenPair: tokens},
			&mockSessionService{},
			&mockUserRepo{user: user},
		)
		body, _ := json.Marshal(map[string]string{"sessionId": "s1", "deviceId": "d1"})
		w := httptest.NewRecorder()
		h.CheckReverseOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/reverse-otp/check", body))
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "accessToken")
	})

	t.Run("pending", func(t *testing.T) {
		h := newAuthHandler(
			nil,
			&mockReverseOTPService{result: &service.VerificationResult{Status: "pending"}},
			nil, nil, nil,
		)
		body, _ := json.Marshal(map[string]string{"sessionId": "s1"})
		w := httptest.NewRecorder()
		h.CheckReverseOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/reverse-otp/check", body))
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "pending")
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newAuthHandler(nil, nil, nil, nil, nil)
		w := httptest.NewRecorder()
		h.CheckReverseOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/reverse-otp/check", []byte("bad")))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("check fails AppError", func(t *testing.T) {
		h := newAuthHandler(nil, &mockReverseOTPService{err: apperror.NotFound("session", "s1")}, nil, nil, nil)
		body, _ := json.Marshal(map[string]string{"sessionId": "s1"})
		w := httptest.NewRecorder()
		h.CheckReverseOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/reverse-otp/check", body))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("check fails generic", func(t *testing.T) {
		h := newAuthHandler(nil, &mockReverseOTPService{err: errors.New("redis")}, nil, nil, nil)
		body, _ := json.Marshal(map[string]string{"sessionId": "s1"})
		w := httptest.NewRecorder()
		h.CheckReverseOTP(w, authReqNoUser(http.MethodPost, "/api/v1/auth/reverse-otp/check", body))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	tokens := &service.TokenPair{AccessToken: "new-at", RefreshToken: "new-rt", ExpiresAt: time.Now().Add(time.Hour).Unix()}

	t.Run("success", func(t *testing.T) {
		h := newAuthHandler(nil, nil, &mockTokenService{tokenPair: tokens}, nil, nil)
		body, _ := json.Marshal(map[string]string{"refreshToken": "old-rt"})
		w := httptest.NewRecorder()
		h.RefreshToken(w, authReqNoUser(http.MethodPost, "/api/v1/auth/refresh", body))
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "new-at")
	})

	t.Run("invalid body", func(t *testing.T) {
		h := newAuthHandler(nil, nil, &mockTokenService{}, nil, nil)
		w := httptest.NewRecorder()
		h.RefreshToken(w, authReqNoUser(http.MethodPost, "/api/v1/auth/refresh", []byte("bad")))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("expired token", func(t *testing.T) {
		h := newAuthHandler(nil, nil, &mockTokenService{err: apperror.Unauthorized("expired")}, nil, nil)
		body, _ := json.Marshal(map[string]string{"refreshToken": "bad"})
		w := httptest.NewRecorder()
		h.RefreshToken(w, authReqNoUser(http.MethodPost, "/api/v1/auth/refresh", body))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("generic error", func(t *testing.T) {
		h := newAuthHandler(nil, nil, &mockTokenService{err: errors.New("redis")}, nil, nil)
		body, _ := json.Marshal(map[string]string{"refreshToken": "bad"})
		w := httptest.NewRecorder()
		h.RefreshToken(w, authReqNoUser(http.MethodPost, "/api/v1/auth/refresh", body))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("success with auth header", func(t *testing.T) {
		userID := uuid.New()
		h := newAuthHandler(nil, nil, &mockTokenService{}, &mockSessionService{}, nil)
		body, _ := json.Marshal(map[string]string{"refreshToken": "rt"})
		w := httptest.NewRecorder()
		r := authReqWithUserID(http.MethodPost, "/api/v1/auth/logout", body, userID)
		r.Header.Set("Authorization", "Bearer some-token")
		h.Logout(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("success without auth", func(t *testing.T) {
		h := newAuthHandler(nil, nil, &mockTokenService{}, &mockSessionService{}, nil)
		w := httptest.NewRecorder()
		h.Logout(w, authReqNoUser(http.MethodPost, "/api/v1/auth/logout", nil))
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
