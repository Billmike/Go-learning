package handlers

import (
	"net/http"

	"github.com/kayodeayelegun/ai-gateway/pkg/response"
)

// BuildVersion is the gateway release version. Override at link time with:
//
//	go build -ldflags "-X github.com/kayodeayelegun/ai-gateway/internal/handlers.BuildVersion=1.0.0"
var BuildVersion = "0.1.0"

type versionResponse struct {
	Version string `json:"version"`
}

// Version handles GET /version.
func Version(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, versionResponse{Version: BuildVersion})
}
