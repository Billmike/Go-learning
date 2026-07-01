package middleware

import (
	"net/http"
	"time"

	"github.com/kayodeayelegun/ai-gateway/internal/metrics"
)

// Metrics returns middleware that records request status and latency after each call.
func Metrics(collector metrics.Collector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			collector.RecordRequest(rw.status, time.Since(start))
		})
	}
}
