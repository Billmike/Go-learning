package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kayodeayelegun/ai-gateway/internal/config"
	"github.com/kayodeayelegun/ai-gateway/internal/gateway"
	"github.com/kayodeayelegun/ai-gateway/internal/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := logging.New(cfg.LogLevel)

	srv, err := gateway.New(cfg, logger)
	if err != nil {
		log.Fatalf("failed to create gateway: %v", err)
	}

	logger.Info("gateway listening", "addr", srv.HTTP.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.HTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		logger.Error("server error", "error", err)
		log.Fatalf("server error: %v", err)
	case sig := <-quit:
		logger.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), gateway.ShutdownTimeout)
	defer cancel()

	if err := srv.HTTP.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
		log.Fatalf("shutdown error: %v", err)
	}

	logger.Info("server stopped")
}
