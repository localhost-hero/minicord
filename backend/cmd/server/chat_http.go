package main

import (
	"encoding/json"
	"net/http"

	"github.com/localhost-hero/minicord/backend/internal/rooms"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// authorizeChatAccess requires a valid, unexpired, unrevoked invite for the
// room. Deliberately does not accept the admin session cookie: this endpoint
// is reached via a plain GET (including the WebSocket handshake), which
// browsers can trigger cross-site with ambient cookies attached, so relying
// on a cookie here would expose it to CSRF. The invite token must be read
// from the URL by legitimate frontend code and is not attached automatically.
func (s *server) authorizeChatAccess(r *http.Request, room string) bool {
	token := r.URL.Query().Get("token")
	if token == "" {
		return false
	}
	invite, err := s.inviteStore.Redeem(token)
	if err != nil {
		return false
	}
	return invite.Room == room
}

func (s *server) roomMessagesHandler(w http.ResponseWriter, r *http.Request) {
	room := r.PathValue("room")
	if err := rooms.ValidateRoomName(room); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !s.authorizeChatAccess(r, room) {
		writeJSONError(w, http.StatusForbidden, "a valid invite token is required")
		return
	}

	messages, err := s.chatStore.ListMessages(room, 50)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load messages")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(messages)
}

type chatIncomingMessage struct {
	Body string `json:"body"`
}

func (s *server) roomChatHandler(w http.ResponseWriter, r *http.Request) {
	room := r.PathValue("room")
	if err := rooms.ValidateRoomName(room); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	identity := r.URL.Query().Get("identity")
	if err := rooms.ValidateIdentity(identity); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !s.authorizeChatAccess(r, room) {
		writeJSONError(w, http.StatusForbidden, "a valid invite token is required")
		return
	}

	// OriginPatterns is intentionally permissive: access is already gated by
	// the invite token above rather than by an ambient cookie, so relaxing
	// the same-origin check does not weaken authorization.
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()
	subscription := s.chatHub.Subscribe(room)
	defer s.chatHub.Unsubscribe(room, subscription)

	readErrors := make(chan error, 1)
	go func() {
		for {
			var incoming chatIncomingMessage
			if err := wsjson.Read(ctx, conn, &incoming); err != nil {
				readErrors <- err
				return
			}

			message, err := s.chatStore.SaveMessage(room, identity, incoming.Body)
			if err != nil {
				continue
			}
			s.chatHub.Publish(room, message)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-readErrors:
			return
		case message, ok := <-subscription:
			if !ok {
				return
			}
			if err := wsjson.Write(ctx, conn, message); err != nil {
				return
			}
		}
	}
}
