package router

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kayodeayelegun/ai-gateway/internal/config"
	"github.com/kayodeayelegun/ai-gateway/internal/proxy"
	"github.com/kayodeayelegun/ai-gateway/pkg/response"
)

// RouteTable resolves a method and path pair to an http.Handler.
type RouteTable interface {
	Handler(method, path string) (http.Handler, bool)
}

// Router maps gateway paths to upstream reverse proxies.
type Router struct {
	routes map[string]http.Handler
}

func routeKey(method, path string) string {
	return method + " " + path
}

// New builds the Phase 1 route table from configuration.
func New(cfg *config.Config, logger *slog.Logger) (*Router, error) {
	chatHandler, err := proxy.New(cfg.ModelAURL, "/chat", logger)
	if err != nil {
		return nil, fmt.Errorf("chat route: %w", err)
	}

	embeddingsHandler, err := proxy.New(cfg.ModelBURL, "/embeddings", logger)
	if err != nil {
		return nil, fmt.Errorf("embeddings route: %w", err)
	}

	audioHandler, err := proxy.New(cfg.ModelBURL, "/audio", logger)
	if err != nil {
		return nil, fmt.Errorf("audio route: %w", err)
	}

	return &Router{
		routes: map[string]http.Handler{
			routeKey("POST", "/v1/chat/completions"): chatHandler,
			routeKey("POST", "/v1/embeddings"):       embeddingsHandler,
			routeKey("POST", "/v1/audio"):            audioHandler,
		},
	}, nil
}

// Handler returns the handler for method and path, or false when no route matches.
func (rt *Router) Handler(method, path string) (http.Handler, bool) {
	h, ok := rt.routes[routeKey(method, path)]
	return h, ok
}

// ServeHTTP dispatches to a registered route or returns 404 for unknown paths.
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, ok := rt.Handler(r.Method, r.URL.Path)
	if !ok {
		response.Error(w, http.StatusNotFound, "not found")
		return
	}
	h.ServeHTTP(w, r)
}

// Register attaches all proxy routes to mux using Go 1.22 method-aware patterns.
func (rt *Router) Register(mux *http.ServeMux) {
	for pattern, handler := range rt.routes {
		mux.Handle(pattern, handler)
	}
}
