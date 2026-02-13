package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

// RedisRateLimitConfig configures a Redis-backed sliding window rate limiter.
type RedisRateLimitConfig struct {
	Requests  int           // max requests per window
	Window    time.Duration // sliding window duration
	KeyPrefix string        // redis key prefix
	ByIP      bool          // if true, use IP instead of user ID
}

// RedisRateLimit returns middleware that rate-limits using Redis counters.
// Keys are scoped per user (authenticated) or per IP (public).
func RedisRateLimit(rdb *redis.Client, cfg RedisRateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			identifier := extractIdentifier(r, cfg.ByIP)
			key := "rl:" + cfg.KeyPrefix + ":" + identifier

			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				// On Redis failure, allow the request (fail open)
				next.ServeHTTP(w, r)
				return
			}

			// Set expiry on first request in window
			if count == 1 {
				rdb.Expire(ctx, key, cfg.Window)
			}

			remaining := cfg.Requests - int(count)
			if remaining < 0 {
				remaining = 0
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(cfg.Requests))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(cfg.Window).Unix(), 10))

			if int(count) > cfg.Requests {
				ttl, _ := rdb.TTL(ctx, key).Result()
				w.Header().Set("Retry-After", strconv.Itoa(int(ttl.Seconds())))
				response.Error(w, apperror.RateLimited())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractIdentifier gets the rate-limit key from the request.
// Uses user ID from context if authenticated, otherwise falls back to IP.
func extractIdentifier(r *http.Request, forceIP bool) string {
	if !forceIP {
		if userID, ok := GetUserID(r.Context()); ok {
			return userID.String()
		}
	}
	// Fall back to IP address
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}
