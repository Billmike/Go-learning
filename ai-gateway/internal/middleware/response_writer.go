package middleware

import "net/http"

// responseWriter wraps http.ResponseWriter to capture the status code written
// by the downstream handler. Per the HTTP spec, a handler that writes a body
// without calling WriteHeader is treated as 200 OK.
type responseWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(status int) {
	if !rw.written {
		rw.status = status
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}
