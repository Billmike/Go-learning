package handlers

import (
	"net/http"

	"github.com/kayodeayelegun/ai-gateway/pkg/response"
)

type healthResponse struct {
	Status string `json:"status"`
}

// Health handles GET /health.
func Health(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
