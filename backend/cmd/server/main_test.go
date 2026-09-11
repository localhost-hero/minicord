package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/localhost-hero/minicord/backend/internal/config"
)

func testConfig() config.Config {
	return config.Config{
		BackendAddr:      ":8080",
		LiveKitURL:       "ws://localhost:7880",
		LiveKitAPIKey:    "test-key",
		LiveKitAPISecret: "test-secret",
	}
}

func TestHealthHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	newServer(testConfig()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if body := recorder.Body.String(); body != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected response body: %q", body)
	}
}

func TestHealthHandlerRejectsOtherMethods(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/health", nil)

	newServer(testConfig()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, recorder.Code)
	}
}

func TestRoomTokenHandlerIssuesTokenForValidRequest(t *testing.T) {
	body, _ := json.Marshal(tokenRequest{Identity: "user-1"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/rooms/team-standup/token", bytes.NewReader(body))

	newServer(testConfig()).ServeHTTP(recorder, request)

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
	if response.URL != "ws://localhost:7880" {
		t.Fatalf("unexpected livekit url: %q", response.URL)
	}
}

func TestRoomTokenHandlerRejectsInvalidRoomName(t *testing.T) {
	body, _ := json.Marshal(tokenRequest{Identity: "user-1"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/rooms/ab/token", bytes.NewReader(body))

	newServer(testConfig()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestRoomTokenHandlerRejectsInvalidIdentity(t *testing.T) {
	body, _ := json.Marshal(tokenRequest{Identity: "user with spaces"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/rooms/team-standup/token", bytes.NewReader(body))

	newServer(testConfig()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestRoomTokenHandlerRejectsMalformedBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/rooms/team-standup/token", strings.NewReader("not-json"))

	newServer(testConfig()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}
