package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/localhost-hero/minicord/backend/internal/config"
)

type healthResponse struct {
	Status string `json:"status"`
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}

func newServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	return mux
}

func main() {
	configuration, err := config.Load(func(key string) string {
		return os.Getenv(key)
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("backend listening on %s", configuration.BackendAddr)
	if err := http.ListenAndServe(configuration.BackendAddr, newServer()); err != nil {
		log.Fatal(err)
	}
}
