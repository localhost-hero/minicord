package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/localhost-hero/minicord/backend/internal/chat"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func TestRoomMessagesRequiresInviteToken(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/rooms/team-standup/messages", nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
}

func TestRoomMessagesReturnsEmptyHistoryWithValidInvite(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))
	invite := createTestInvite(t, handler, "team-standup")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/rooms/team-standup/messages?token="+invite.Token, nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	var messages []chat.Message
	if err := json.Unmarshal(recorder.Body.Bytes(), &messages); err != nil {
		t.Fatalf("failed to unmarshal messages: %v", err)
	}
	if len(messages) != 0 {
		t.Fatalf("expected no messages yet, got %d", len(messages))
	}
}

func TestRoomChatBroadcastsMessagesBetweenConnections(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))
	invite := createTestInvite(t, handler, "team-standup")

	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") +
		"/api/rooms/team-standup/chat?token=" + invite.Token + "&identity=sender"

	ctx := context.Background()
	senderConn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial sender connection: %v", err)
	}
	defer senderConn.CloseNow()

	receiverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") +
		"/api/rooms/team-standup/chat?token=" + invite.Token + "&identity=receiver"
	receiverConn, _, err := websocket.Dial(ctx, receiverURL, nil)
	if err != nil {
		t.Fatalf("failed to dial receiver connection: %v", err)
	}
	defer receiverConn.CloseNow()

	// Dial returns once the client-side handshake completes, but the
	// server-side handler subscribes to the room hub asynchronously
	// afterwards; give it a moment before publishing.
	time.Sleep(100 * time.Millisecond)

	if err := wsjson.Write(ctx, senderConn, map[string]string{"body": "hello there"}); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	var received chat.Message
	if err := wsjson.Read(ctx, receiverConn, &received); err != nil {
		t.Fatalf("failed to read broadcast message: %v", err)
	}

	if received.Body != "hello there" || received.Identity != "sender" || received.Room != "team-standup" {
		t.Fatalf("unexpected broadcast message: %+v", received)
	}
}

func TestRoomChatRejectsMissingInviteToken(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))

	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/api/rooms/team-standup/chat?identity=guest-1"

	_, resp, err := websocket.Dial(context.Background(), wsURL, nil)
	if err == nil {
		t.Fatal("expected the connection without an invite token to be rejected")
	}
	if resp != nil && resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, resp.StatusCode)
	}
}
