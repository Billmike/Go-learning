package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kayodeayelegun/ai-gateway/internal/config"
	"github.com/kayodeayelegun/ai-gateway/internal/handlers"
	"github.com/kayodeayelegun/ai-gateway/internal/logging"
	"github.com/kayodeayelegun/ai-gateway/internal/middleware"
	"github.com/kayodeayelegun/ai-gateway/internal/ratelimit"
	"github.com/kayodeayelegun/ai-gateway/internal/requestid"
	"github.com/kayodeayelegun/ai-gateway/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := logging.New(cfg.LogLevel)

	rt, err := router.New(cfg, logger)
	if err != nil {
		log.Fatalf("failed to create router: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("GET /version", handlers.Version)
	rt.Register(mux)

	limiter := ratelimit.New(cfg.RateLimit)

	handler := middleware.Chain(
		middleware.Recovery(logger),
		requestid.Middleware,
		middleware.Logging(logger),
		middleware.Timeout(cfg.RequestTimeout),
		middleware.Auth(cfg.APIToken),
		middleware.RateLimit(limiter),
	)(mux)

	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info("gateway listening", "addr", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Error("server error", "error", err)
		log.Fatalf("server error: %v", err)
	}
}
