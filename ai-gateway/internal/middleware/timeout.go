package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/kayodeayelegun/ai-gateway/pkg/response"
)

// Timeout returns middleware that enforces a deadline on downstream handlers.
// The timeout is attached to the request context so handlers and outbound
// calls can respect cancellation. If the deadline is exceeded before the
// handler finishes, the client receives 504 with {"error":"gateway timeout"}.
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()

			tw := &timeoutWriter{ResponseWriter: w}
			done := make(chan struct{})
			var panicVal any

			go func() {
				defer func() {
					if v := recover(); v != nil {
						panicVal = v
					}
					close(done)
				}()
				next.ServeHTTP(tw, r.WithContext(ctx))
			}()

			select {
			case <-done:
				if panicVal != nil {
					panic(panicVal)
				}
				return
			case <-ctx.Done():
			}

			tw.mu.Lock()
			tw.timedOut = true
			alreadyWritten := tw.written
			tw.mu.Unlock()

			<-done
			if panicVal != nil {
				panic(panicVal)
			}

			if !alreadyWritten && ctx.Err() == context.DeadlineExceeded {
				response.Error(w, http.StatusGatewayTimeout, "gateway timeout")
			}
		})
	}
}

// timeoutWriter drops writes after a timeout so a slow handler cannot race
// with the gateway's timeout response.
type timeoutWriter struct {
	http.ResponseWriter
	mu       sync.Mutex
	timedOut bool
	written  bool
}

func (tw *timeoutWriter) WriteHeader(status int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return
	}
	tw.written = true
	tw.ResponseWriter.WriteHeader(status)
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return 0, http.ErrHandlerTimeout
	}
	tw.written = true
	return tw.ResponseWriter.Write(b)
}
