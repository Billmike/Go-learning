package main

import (
	"log"

	"github.com/kayodeayelegun/ai-gateway/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("gateway config loaded (port=%s, timeout=%s, rate_limit=%d)",
		cfg.Port, cfg.RequestTimeout, cfg.RateLimit)
}
