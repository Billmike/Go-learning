package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kayodeayelegun/ai-gateway/internal/config"
	"github.com/kayodeayelegun/ai-gateway/internal/handlers"
	"github.com/kayodeayelegun/ai-gateway/internal/logging"
	"github.com/kayodeayelegun/ai-gateway/internal/middleware"
	"github.com/kayodeayelegun/ai-gateway/internal/requestid"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := logging.New(cfg.LogLevel)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("GET /version", handlers.Version)

	handler := middleware.Chain(
		requestid.Middleware,
		middleware.Logging(logger),
	)(mux)

	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info("gateway listening", "addr", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Error("server error", "error", err)
		log.Fatalf("server error: %v", err)
	}
}
