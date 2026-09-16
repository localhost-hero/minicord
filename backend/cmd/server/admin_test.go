package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminLoginRejectsWrongPassword(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))

	body, _ := json.Marshal(adminLoginRequest{Username: "admin", Password: "wrong"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewReader(body))

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestAdminLogoutClearsCookie(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))
	cookie := adminCookie(t, handler)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/admin/logout", nil)
	request.AddCookie(cookie)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) == 0 || cookies[0].MaxAge >= 0 {
		t.Fatal("expected logout to clear the session cookie")
	}
}

func TestCreateInviteRequiresAdminSession(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/admin/rooms/team-standup/invites", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestCreateInviteReturnsUsableToken(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))
	cookie := adminCookie(t, handler)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/admin/rooms/team-standup/invites", nil)
	request.AddCookie(cookie)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}

	var response inviteResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if response.Token == "" || response.Room != "team-standup" {
		t.Fatalf("unexpected invite response: %+v", response)
	}

	inspectRecorder := httptest.NewRecorder()
	inspectRequest := httptest.NewRequest(http.MethodGet, "/api/invites/"+response.Token, nil)
	handler.ServeHTTP(inspectRecorder, inspectRequest)

	if inspectRecorder.Code != http.StatusOK {
		t.Fatalf("expected the created invite to be redeemable, got status %d", inspectRecorder.Code)
	}
}

func TestRevokeInviteMakesTokenUnusable(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))
	cookie := adminCookie(t, handler)

	createRecorder := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/admin/rooms/team-standup/invites", nil)
	createRequest.AddCookie(cookie)
	handler.ServeHTTP(createRecorder, createRequest)

	var invite inviteResponse
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &invite); err != nil {
		t.Fatalf("failed to unmarshal invite: %v", err)
	}

	revokeRecorder := httptest.NewRecorder()
	revokeRequest := httptest.NewRequest(http.MethodDelete, "/api/admin/invites/"+invite.ID, nil)
	revokeRequest.AddCookie(cookie)
	handler.ServeHTTP(revokeRecorder, revokeRequest)

	if revokeRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, revokeRecorder.Code)
	}

	inspectRecorder := httptest.NewRecorder()
	inspectRequest := httptest.NewRequest(http.MethodGet, "/api/invites/"+invite.Token, nil)
	handler.ServeHTTP(inspectRecorder, inspectRequest)

	if inspectRecorder.Code != http.StatusGone {
		t.Fatalf("expected a revoked invite to return status %d, got %d", http.StatusGone, inspectRecorder.Code)
	}
}
