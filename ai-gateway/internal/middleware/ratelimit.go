package middleware

import (
	"net/http"

	"github.com/kayodeayelegun/ai-gateway/internal/auth"
	"github.com/kayodeayelegun/ai-gateway/internal/ratelimit"
	"github.com/kayodeayelegun/ai-gateway/pkg/response"
)

// RateLimit returns middleware that enforces per-token request limits.
// It expects Auth to run first so each request carries a bearer token.
func RateLimit(limiter ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := auth.ExtractToken(r)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !limiter.Allow(key) {
				response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
