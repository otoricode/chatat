package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/middleware"
)

func TestRedisRateLimit(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	cfg := middleware.RedisRateLimitConfig{
		Requests:  3,
		Window:    time.Minute,
		KeyPrefix: "test",
		ByIP:      true,
	}

	handler := middleware.RedisRateLimit(rdb, cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("allows requests within limit", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = "10.0.0.1:1234"
			handler.ServeHTTP(w, r)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("blocks after limit exceeded", func(t *testing.T) {
		// The 4th request should be blocked
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "10.0.0.1:1234"
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("sets rate limit headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "10.0.0.2:1234"
		handler.ServeHTTP(w, r)
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
	})

	t.Run("uses userID when ByIP false", func(t *testing.T) {
		cfg2 := middleware.RedisRateLimitConfig{
			Requests:  2,
			Window:    time.Minute,
			KeyPrefix: "user",
			ByIP:      false,
		}
		h := middleware.RedisRateLimit(rdb, cfg2)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		userID := uuid.New()
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := middleware.WithUserID(r.Context(), userID)
			h.ServeHTTP(w, r.WithContext(ctx))
			assert.Equal(t, http.StatusOK, w.Code)
		}
		// 3rd should be blocked
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := middleware.WithUserID(r.Context(), userID)
		h.ServeHTTP(w, r.WithContext(ctx))
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("falls back to X-Real-IP header", func(t *testing.T) {
		cfg3 := middleware.RedisRateLimitConfig{
			Requests:  5,
			Window:    time.Minute,
			KeyPrefix: "xrip",
			ByIP:      true,
		}
		h := middleware.RedisRateLimit(rdb, cfg3)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("X-Real-IP", "1.2.3.4")
		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("falls back to X-Forwarded-For", func(t *testing.T) {
		cfg4 := middleware.RedisRateLimitConfig{
			Requests:  5,
			Window:    time.Minute,
			KeyPrefix: "xff",
			ByIP:      true,
		}
		h := middleware.RedisRateLimit(rdb, cfg4)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("X-Forwarded-For", "5.6.7.8")
		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
