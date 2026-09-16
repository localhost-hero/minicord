package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/localhost-hero/minicord/backend/internal/adminsession"
	"github.com/localhost-hero/minicord/backend/internal/chat"
	"github.com/localhost-hero/minicord/backend/internal/config"
	"github.com/localhost-hero/minicord/backend/internal/invites"
	"github.com/localhost-hero/minicord/backend/internal/livekit"
	"github.com/localhost-hero/minicord/backend/internal/rooms"
	"github.com/localhost-hero/minicord/backend/internal/store"
)

const adminSessionCookie = "minicord_admin_session"

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

// isAdminSession reports whether the request carries a valid admin session cookie.
func isAdminSession(r *http.Request, configuration config.Config) bool {
	cookie, err := r.Cookie(adminSessionCookie)
	if err != nil {
		return false
	}
	return adminsession.ValidateToken(configuration.SessionSecret, cookie.Value)
}

func issueRoomToken(w http.ResponseWriter, configuration config.Config, room, identity string) {
	token, err := livekit.NewAccessToken(
		configuration.LiveKitAPIKey,
		configuration.LiveKitAPISecret,
		identity,
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

// roomTokenHandler issues a LiveKit token for the admin only. Guests must use
// the invite-gated endpoints in invites_http.go.
func roomTokenHandler(configuration config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAdminSession(r, configuration) {
			writeJSONError(w, http.StatusUnauthorized, "admin session required")
			return
		}

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

		issueRoomToken(w, configuration, room, payload.Identity)
	}
}

type server struct {
	config      config.Config
	inviteStore *invites.Store
	chatStore   *chat.Store
	chatHub     *chat.Hub
}

func newServer(configuration config.Config, db *sql.DB) http.Handler {
	s := &server{
		config:      configuration,
		inviteStore: invites.NewStore(db),
		chatStore:   chat.NewStore(db),
		chatHub:     chat.NewHub(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /api/rooms/{room}/token", roomTokenHandler(configuration))

	mux.HandleFunc("POST /api/admin/login", s.adminLoginHandler)
	mux.HandleFunc("POST /api/admin/logout", s.adminLogoutHandler)
	mux.HandleFunc("POST /api/admin/rooms/{room}/invites", s.createInviteHandler)
	mux.HandleFunc("DELETE /api/admin/invites/{id}", s.revokeInviteHandler)

	mux.HandleFunc("GET /api/invites/{token}", s.inspectInviteHandler)
	mux.HandleFunc("POST /api/invites/{token}/token", s.inviteTokenHandler)

	mux.HandleFunc("GET /api/rooms/{room}/messages", s.roomMessagesHandler)
	mux.HandleFunc("GET /api/rooms/{room}/chat", s.roomChatHandler)

	return mux
}

func main() {
	configuration, err := config.Load(func(key string) string {
		return os.Getenv(key)
	})
	if err != nil {
		log.Fatal(err)
	}

	db, err := store.Open(configuration.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Printf("backend listening on %s", configuration.BackendAddr)
	if err := http.ListenAndServe(configuration.BackendAddr, newServer(configuration, db)); err != nil {
		log.Fatal(err)
	}
}

