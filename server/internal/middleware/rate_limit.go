package middleware

import (
	"net/http"

	"golang.org/x/time/rate"

	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

// RateLimit returns middleware that limits requests using a token bucket.
func RateLimit(requestsPerSecond float64, burst int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				response.Error(w, apperror.RateLimited())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
