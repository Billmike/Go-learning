package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/kayodeayelegun/ai-gateway/internal/requestid"
)

// Logging returns middleware that emits one structured JSON log line per request
// after the downstream handler completes. Phase 1 uses r.RemoteAddr for client_ip.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			logger.Info("request completed",
				"request_id", requestid.FromContext(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"latency_ms", time.Since(start).Milliseconds(),
				"client_ip", clientIP(r),
			)
		})
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
