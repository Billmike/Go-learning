package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kayodeayelegun/ai-gateway/pkg/response"
)

type embeddingsResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float64 `json:"embeddings"`
}

type audioResponse struct {
	Model   string `json:"model"`
	Message string `json:"message"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9002"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /embeddings", func(w http.ResponseWriter, _ *http.Request) {
		response.JSON(w, http.StatusOK, embeddingsResponse{
			Model:      "mock-b",
			Embeddings: [][]float64{{0.1, 0.2, 0.3}},
		})
	})
	mux.HandleFunc("POST /audio", func(w http.ResponseWriter, _ *http.Request) {
		response.JSON(w, http.StatusOK, audioResponse{
			Model:   "mock-b",
			Message: "ok",
		})
	})

	addr := ":" + port
	log.Printf("mock-model-b listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
