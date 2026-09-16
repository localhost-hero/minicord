package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/localhost-hero/minicord/backend/internal/invites"
	"github.com/localhost-hero/minicord/backend/internal/rooms"
)

type inviteInspectResponse struct {
	Room string `json:"room"`
}

func (s *server) inspectInviteHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	invite, err := s.inviteStore.Redeem(token)
	if err != nil {
		writeInviteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(inviteInspectResponse{Room: invite.Room})
}

func (s *server) inviteTokenHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	invite, err := s.inviteStore.Redeem(token)
	if err != nil {
		writeInviteError(w, err)
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

	issueRoomToken(w, s.config, invite.Room, payload.Identity)
}

func writeInviteError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, invites.ErrNotFound):
		writeJSONError(w, http.StatusNotFound, "invite not found")
	case errors.Is(err, invites.ErrExpiredOrRevoked):
		writeJSONError(w, http.StatusGone, "invite has expired or been revoked")
	default:
		writeJSONError(w, http.StatusInternalServerError, "failed to validate invite")
	}
}
