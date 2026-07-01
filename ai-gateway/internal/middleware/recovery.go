package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/kayodeayelegun/ai-gateway/internal/requestid"
	"github.com/kayodeayelegun/ai-gateway/pkg/response"
)

// Recovery returns middleware that catches panics in downstream handlers,
// logs the panic value and stack trace, and returns a 500 JSON error response.
// Panics are recovered only at this HTTP boundary; handlers should still prefer
// returning errors over panicking.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					logger.Error("panic recovered",
						"request_id", requestid.FromContext(r.Context()),
						"method", r.Method,
						"path", r.URL.Path,
						"panic", v,
						"stack", string(debug.Stack()),
					)
					response.Error(w, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
