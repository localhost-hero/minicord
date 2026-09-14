package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createTestInvite(t *testing.T, handler http.Handler, room string) inviteResponse {
	t.Helper()
	cookie := adminCookie(t, handler)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/admin/rooms/"+room+"/invites", nil)
	request.AddCookie(cookie)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected invite creation to succeed, got status %d: %s", recorder.Code, recorder.Body.String())
	}

	var invite inviteResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &invite); err != nil {
		t.Fatalf("failed to unmarshal invite: %v", err)
	}
	return invite
}

func TestInspectInviteRejectsUnknownToken(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/invites/unknown-token", nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestInviteTokenHandlerIssuesLiveKitTokenForValidInvite(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))
	invite := createTestInvite(t, handler, "team-standup")

	body, _ := json.Marshal(tokenRequest{Identity: "guest-1"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/invites/"+invite.Token+"/token", bytes.NewReader(body))
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	var response tokenResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if response.Token == "" {
		t.Fatal("expected a non-empty token")
	}
}

func TestInviteTokenHandlerRejectsInvalidIdentity(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))
	invite := createTestInvite(t, handler, "team-standup")

	body, _ := json.Marshal(tokenRequest{Identity: "no"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/invites/"+invite.Token+"/token", bytes.NewReader(body))
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}
