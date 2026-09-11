package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/localhost-hero/minicord/backend/internal/config"
	"github.com/localhost-hero/minicord/backend/internal/livekit"
	"github.com/localhost-hero/minicord/backend/internal/rooms"
)

type healthResponse struct {
	Status string `json:"status"`
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}

type tokenRequest struct {
	Identity string `json:"identity"`
}

type tokenResponse struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func roomTokenHandler(configuration config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		room := r.PathValue("room")
		if err := rooms.ValidateRoomName(room); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		var payload tokenRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSONError(w, http.StatusBadRequest, "request body must be valid JSON")
			return
		}

		if err := rooms.ValidateIdentity(payload.Identity); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		token, err := livekit.NewAccessToken(
			configuration.LiveKitAPIKey,
			configuration.LiveKitAPISecret,
			payload.Identity,
			room,
			livekit.DefaultTokenTTL,
		)
		if err != nil {
			log.Printf("failed to issue livekit token: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to issue token")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(tokenResponse{Token: token, URL: configuration.LiveKitURL})
	}
}

func newServer(configuration config.Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /api/rooms/{room}/token", roomTokenHandler(configuration))
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
	if err := http.ListenAndServe(configuration.BackendAddr, newServer(configuration)); err != nil {
		log.Fatal(err)
	}
}

