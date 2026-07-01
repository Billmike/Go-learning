package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestRecovery_returns500OnPanic(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	handler := middleware.Chain(
		middleware.Recovery(logger),
		requestid.Middleware,
	)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("something went wrong")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["error"] != "internal server error" {
		t.Fatalf("error = %q, want %q", body["error"], "internal server error")
	}

	if !bytes.Contains(buf.Bytes(), []byte("panic recovered")) {
		t.Fatalf("expected panic log, got: %s", buf.String())
	}
}

func TestRecovery_allowsNormalRequests(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelError}))

	handler := middleware.Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTimeout_returns504WhenHandlerExceedsDeadline(t *testing.T) {
	t.Parallel()

	handler := middleware.Timeout(10 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/slow", nil))

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusGatewayTimeout)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["error"] != "gateway timeout" {
		t.Fatalf("error = %q, want %q", body["error"], "gateway timeout")
	}
}

func TestTimeout_allowsFastHandlers(t *testing.T) {
	t.Parallel()

	handler := middleware.Timeout(time.Second)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTimeout_propagatesContextToHandler(t *testing.T) {
	t.Parallel()

	handler := middleware.Timeout(10 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ctx", nil))

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusGatewayTimeout)
	}
}

func TestTimeout_contextCarriesDeadline(t *testing.T) {
	t.Parallel()

	var gotDeadline bool
	handler := middleware.Timeout(10 * time.Millisecond)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, gotDeadline = r.Context().Deadline()
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !gotDeadline {
		t.Fatal("expected request context to carry a deadline")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAuth_returns401WhenTokenMissing(t *testing.T) {
	t.Parallel()

	handler := middleware.Auth("secret-token")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["error"] != "missing authorization token" {
		t.Fatalf("error = %q, want %q", body["error"], "missing authorization token")
	}
}

func TestAuth_returns401WhenTokenInvalid(t *testing.T) {
	t.Parallel()

	handler := middleware.Auth("secret-token")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["error"] != "invalid authorization token" {
		t.Fatalf("error = %q, want %q", body["error"], "invalid authorization token")
	}
}

func TestAuth_allowsValidToken(t *testing.T) {
	t.Parallel()

	handler := middleware.Auth("secret-token")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

type stubLimiter struct {
	allowed bool
}

func (s stubLimiter) Allow(string) bool {
	return s.allowed
}

func TestRateLimit_returns429WhenExceeded(t *testing.T) {
	t.Parallel()

	handler := middleware.RateLimit(stubLimiter{allowed: false})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["error"] != "rate limit exceeded" {
		t.Fatalf("error = %q, want %q", body["error"], "rate limit exceeded")
	}
}

func TestRateLimit_allowsWhenUnderLimit(t *testing.T) {
	t.Parallel()

	handler := middleware.RateLimit(stubLimiter{allowed: true})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
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
