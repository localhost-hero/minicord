package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/localhost-hero/minicord/backend/internal/adminsession"
	"github.com/localhost-hero/minicord/backend/internal/rooms"
)

type adminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *server) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	var payload adminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "request body must be valid JSON")
		return
	}

	if !adminsession.CheckCredentials(payload.Username, payload.Password, s.config.AdminUsername, s.config.AdminPasswordHash) {
		writeJSONError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookie,
		Value:    adminsession.NewToken(s.config.SessionSecret, adminsession.DefaultSessionTTL),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(adminsession.DefaultSessionTTL.Seconds()),
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) adminLogoutHandler(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

type createInviteRequest struct {
	TTLSeconds int64 `json:"ttlSeconds"`
}

type inviteResponse struct {
	ID        string  `json:"id"`
	Room      string  `json:"room"`
	Token     string  `json:"token"`
	ExpiresAt *string `json:"expiresAt"`
}

func (s *server) createInviteHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminSession(r, s.config) {
		writeJSONError(w, http.StatusUnauthorized, "admin session required")
		return
	}

	room := r.PathValue("room")
	if err := rooms.ValidateRoomName(room); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var payload createInviteRequest
	if r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSONError(w, http.StatusBadRequest, "request body must be valid JSON")
			return
		}
	}

	invite, token, err := s.inviteStore.Create(room, time.Duration(payload.TTLSeconds)*time.Second)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create invite")
		return
	}

	response := inviteResponse{ID: invite.ID, Room: invite.Room, Token: token}
	if invite.ExpiresAt != nil {
		formatted := invite.ExpiresAt.Format(timeRFC3339)
		response.ExpiresAt = &formatted
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func (s *server) revokeInviteHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminSession(r, s.config) {
		writeJSONError(w, http.StatusUnauthorized, "admin session required")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "invite id must not be empty")
		return
	}

	if err := s.inviteStore.Revoke(id); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to revoke invite")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"
