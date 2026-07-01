package handlers

import (
	"net/http"

	"github.com/kayodeayelegun/ai-gateway/internal/metrics"
	"github.com/kayodeayelegun/ai-gateway/pkg/response"
)

// Metrics returns a handler for GET /metrics using the given collector.
func Metrics(collector metrics.Collector) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		response.JSON(w, http.StatusOK, collector.Snapshot())
	}
}
