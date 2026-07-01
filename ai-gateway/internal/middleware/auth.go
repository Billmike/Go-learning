package middleware

import (
	"errors"
	"net/http"

	"github.com/kayodeayelegun/ai-gateway/internal/auth"
	"github.com/kayodeayelegun/ai-gateway/pkg/response"
)

// Auth returns middleware that validates a static Bearer token on every request.
// Missing or invalid tokens receive 401 with a JSON error envelope.
func Auth(expectedToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := auth.ValidateBearer(r, expectedToken); err != nil {
				msg := "unauthorized"
				switch {
				case errors.Is(err, auth.ErrMissingToken):
					msg = "missing authorization token"
				case errors.Is(err, auth.ErrInvalidToken):
					msg = "invalid authorization token"
				}
				response.Error(w, http.StatusUnauthorized, msg)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
