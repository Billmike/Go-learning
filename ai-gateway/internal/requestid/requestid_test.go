package requestid_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kayodeayelegun/ai-gateway/internal/requestid"
)

func TestMiddleware_generatesIDWhenAbsent(t *testing.T) {
	t.Parallel()

	var gotCtxID string
	handler := requestid.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCtxID = requestid.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotCtxID == "" {
		t.Fatal("expected generated request ID in context")
	}
	if rec.Header().Get(requestid.Header) != gotCtxID {
		t.Fatalf("response header %q, context ID %q", rec.Header().Get(requestid.Header), gotCtxID)
	}
}

func TestMiddleware_reusesClientID(t *testing.T) {
	t.Parallel()

	const clientID = "client-supplied-id-123"

	var gotCtxID string
	handler := requestid.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCtxID = requestid.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(requestid.Header, clientID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotCtxID != clientID {
		t.Fatalf("context ID = %q, want %q", gotCtxID, clientID)
	}
	if rec.Header().Get(requestid.Header) != clientID {
		t.Fatalf("response header = %q, want %q", rec.Header().Get(requestid.Header), clientID)
	}
}

func TestFromContext_emptyWhenUnset(t *testing.T) {
	t.Parallel()

	if id := requestid.FromContext(t.Context()); id != "" {
		t.Fatalf("FromContext() = %q, want empty", id)
	}
}
