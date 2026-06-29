package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kayodeayelegun/ai-gateway/internal/config"
	"github.com/kayodeayelegun/ai-gateway/internal/handlers"
	"github.com/kayodeayelegun/ai-gateway/internal/requestid"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("GET /version", handlers.Version)

	handler := requestid.Middleware(mux)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("gateway listening on %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
