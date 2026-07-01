package gateway

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kayodeayelegun/ai-gateway/internal/config"
	"github.com/kayodeayelegun/ai-gateway/internal/handlers"
	"github.com/kayodeayelegun/ai-gateway/internal/metrics"
	"github.com/kayodeayelegun/ai-gateway/internal/middleware"
	"github.com/kayodeayelegun/ai-gateway/internal/ratelimit"
	"github.com/kayodeayelegun/ai-gateway/internal/requestid"
	"github.com/kayodeayelegun/ai-gateway/internal/router"
)

const readHeaderTimeout = 5 * time.Second

// ShutdownTimeout is the maximum time allowed to drain in-flight requests.
const ShutdownTimeout = 10 * time.Second

// Server holds the configured HTTP server and shared runtime dependencies.
type Server struct {
	HTTP    *http.Server
	Metrics metrics.Collector
}

// New assembles the gateway mux, middleware chain, and http.Server from configuration.
func New(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	collector := metrics.New()

	rt, err := router.New(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("router: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("GET /version", handlers.Version)
	mux.HandleFunc("GET /metrics", handlers.Metrics(collector))
	rt.Register(mux)

	limiter := ratelimit.New(cfg.RateLimit)

	handler := middleware.Chain(
		middleware.Recovery(logger),
		requestid.Middleware,
		middleware.Logging(logger),
		middleware.Metrics(collector),
		middleware.Timeout(cfg.RequestTimeout),
		middleware.Auth(cfg.APIToken),
		middleware.RateLimit(limiter),
	)(mux)

	addr := fmt.Sprintf(":%s", cfg.Port)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	return &Server{
		HTTP:    httpServer,
		Metrics: collector,
	}, nil
}
