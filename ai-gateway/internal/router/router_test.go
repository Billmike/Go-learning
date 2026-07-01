package router_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kayodeayelegun/ai-gateway/internal/config"
	"github.com/kayodeayelegun/ai-gateway/internal/requestid"
	"github.com/kayodeayelegun/ai-gateway/internal/router"
)

func TestRouterRoutesToCorrectUpstream(t *testing.T) {
	t.Parallel()

	modelA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat" {
			t.Errorf("model A path = %q, want /chat", r.URL.Path)
		}
		if got := r.Header.Get(requestid.Header); got != "test-id" {
			t.Errorf("model A request id = %q, want test-id", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"model":"a"}`))
	}))
	defer modelA.Close()

	modelB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" && r.URL.Path != "/audio" {
			t.Errorf("model B unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"model":"b"}`))
	}))
	defer modelB.Close()

	cfg := &config.Config{
		ModelAURL: modelA.URL,
		ModelBURL: modelB.URL,
	}

	rt, err := router.New(cfg, slog.Default())
	if err != nil {
		t.Fatalf("router.New: %v", err)
	}

	tests := []struct {
		method string
		path   string
		want   string
	}{
		{http.MethodPost, "/v1/chat/completions", `{"model":"a"}`},
		{http.MethodPost, "/v1/embeddings", `{"model":"b"}`},
		{http.MethodPost, "/v1/audio", `{"model":"b"}`},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
			req.Header.Set(requestid.Header, "test-id")

			rec := httptest.NewRecorder()
			requestid.Middleware(rt).ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			body, _ := io.ReadAll(rec.Body)
			if strings.TrimSpace(string(body)) != tc.want {
				t.Fatalf("body = %q, want %q", string(body), tc.want)
			}
		})
	}
}

func TestRouterUnknownRouteReturns404(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		ModelAURL: "http://127.0.0.1:1",
		ModelBURL: "http://127.0.0.1:2",
	}

	rt, err := router.New(cfg, slog.Default())
	if err != nil {
		t.Fatalf("router.New: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	rec := httptest.NewRecorder()
	rt.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
