package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kayodeayelegun/ai-gateway/internal/middleware"
	"github.com/kayodeayelegun/ai-gateway/internal/requestid"
)

func TestChain_appliesMiddlewareInOrder(t *testing.T) {
	t.Parallel()

	var order []string
	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "outer-before")
			next.ServeHTTP(w, r)
			order = append(order, "outer-after")
		})
	}
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "inner-before")
			next.ServeHTTP(w, r)
			order = append(order, "inner-after")
		})
	}

	handler := middleware.Chain(mw1, mw2)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	want := []string{"outer-before", "inner-before", "handler", "inner-after", "outer-after"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i, step := range want {
		if order[i] != step {
			t.Fatalf("order[%d] = %q, want %q (full order: %v)", i, order[i], step, order)
		}
	}
}

func TestLogging_middlewareCapturesStatusAndRequestID(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	handler := middleware.Chain(
		requestid.Middleware,
		middleware.Logging(logger),
	)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal log: %v\nraw: %s", err, buf.String())
	}

	if entry["msg"] != "request completed" {
		t.Fatalf("msg = %v, want %q", entry["msg"], "request completed")
	}
	if entry["method"] != "POST" {
		t.Fatalf("method = %v, want POST", entry["method"])
	}
	if entry["path"] != "/v1/chat/completions" {
		t.Fatalf("path = %v, want /v1/chat/completions", entry["path"])
	}
	if status, ok := entry["status"].(float64); !ok || int(status) != http.StatusCreated {
		t.Fatalf("status = %v, want %d", entry["status"], http.StatusCreated)
	}
	if entry["client_ip"] != "127.0.0.1" {
		t.Fatalf("client_ip = %v, want 127.0.0.1", entry["client_ip"])
	}
	if entry["request_id"] == "" {
		t.Fatal("expected non-empty request_id in log")
	}
	if _, ok := entry["latency_ms"]; !ok {
		t.Fatal("expected latency_ms in log")
	}
}

func TestLogging_defaultsStatusTo200WhenWriteOnly(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	handler := middleware.Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal log: %v", err)
	}
	if status, ok := entry["status"].(float64); !ok || int(status) != http.StatusOK {
		t.Fatalf("status = %v, want %d", entry["status"], http.StatusOK)
	}
}
