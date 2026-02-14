package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/middleware"
)

const testSecret = "test-jwt-secret-key-for-testing"

func generateToken(t *testing.T, userID uuid.UUID, secret string, exp time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": jwt.NewNumericDate(exp),
		"iat": jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return tokenString
}

func TestAuth_ValidToken(t *testing.T) {
	userID := uuid.New()
	token := generateToken(t, userID, testSecret, time.Now().Add(time.Hour))

	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, ok := middleware.GetUserID(r.Context())
		assert.True(t, ok)
		assert.Equal(t, userID, gotID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuth_MissingHeader(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_InvalidFormat(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidScheme token-value")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Should fail because scheme is not "bearer" (case-insensitive "bearer" is accepted)
	// Actually our code checks EqualFold for "bearer" so "InvalidScheme" would fail
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	token := generateToken(t, userID, testSecret, time.Now().Add(-time.Hour))

	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_WrongSecret(t *testing.T) {
	userID := uuid.New()
	token := generateToken(t, userID, "wrong-secret", time.Now().Add(time.Hour))

	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetUserID_NotSet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	_, ok := middleware.GetUserID(req.Context())
	assert.False(t, ok)
}

func TestWithUserID(t *testing.T) {
	uid := uuid.New()
	ctx := middleware.WithUserID(context.Background(), uid)
	id, ok := middleware.GetUserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, uid, id)
}

func TestAuth_InvalidSubject(t *testing.T) {
	// Token with non-UUID subject
	claims := jwt.MapClaims{
		"sub": "not-a-uuid",
		"exp": jwt.NewNumericDate(time.Now().Add(time.Hour)),
		"iat": jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_MissingSubject(t *testing.T) {
	// Token without "sub" claim
	claims := jwt.MapClaims{
		"exp": jwt.NewNumericDate(time.Now().Add(time.Hour)),
		"iat": jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_NoSpaceInHeader(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "BearerTokenWithoutSpace")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
