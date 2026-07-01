package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kayodeayelegun/ai-gateway/pkg/response"
)

type chatResponse struct {
	Model   string `json:"model"`
	Message string `json:"message"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9001"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /chat", func(w http.ResponseWriter, _ *http.Request) {
		response.JSON(w, http.StatusOK, chatResponse{
			Model:   "mock-a",
			Message: "ok",
		})
	})

	addr := ":" + port
	log.Printf("mock-model-a listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
