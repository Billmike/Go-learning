package requestid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// Header is the HTTP header used to carry the request ID.
const Header = "X-Request-ID"

// ctxKey is an unexported type for the context key to avoid collisions
// with other packages that also store values in context.Context.
type ctxKey struct{}

// FromContext returns the request ID stored in ctx, or an empty string if none.
func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// Middleware reads X-Request-ID from the incoming request or generates a new
// UUID when absent. It stores the ID in the request context and echoes it on
// the response so clients can correlate logs across services.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(Header)
		if id == "" {
			id = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), ctxKey{}, id)
		w.Header().Set(Header, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
