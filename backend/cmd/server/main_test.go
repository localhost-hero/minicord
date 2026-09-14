package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/localhost-hero/minicord/backend/internal/config"
	"github.com/localhost-hero/minicord/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

const testAdminPassword = "test-admin-password"

func testConfig(t *testing.T) config.Config {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(testAdminPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash test admin password: %v", err)
	}

	return config.Config{
		BackendAddr:       ":8080",
		LiveKitURL:        "ws://localhost:7880",
		LiveKitAPIKey:     "test-key",
		LiveKitAPISecret:  "test-secret",
		AdminUsername:     "admin",
		AdminPasswordHash: string(hash),
		SessionSecret:     "01234567890123456789012345678901",
	}
}

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// adminCookie logs in against the given server and returns the resulting
// admin session cookie for use in subsequent authenticated test requests.
func adminCookie(t *testing.T, handler http.Handler) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(adminLoginRequest{Username: "admin", Password: testAdminPassword})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewReader(body))
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected admin login to succeed, got status %d: %s", recorder.Code, recorder.Body.String())
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected an admin session cookie to be set")
	}
	return cookies[0]
}

func TestHealthHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	newServer(testConfig(t), testDB(t)).ServeHTTP(recorder, request)

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

	newServer(testConfig(t), testDB(t)).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, recorder.Code)
	}
}

func TestRoomTokenHandlerRejectsWithoutAdminSession(t *testing.T) {
	body, _ := json.Marshal(tokenRequest{Identity: "user-1"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/rooms/team-standup/token", bytes.NewReader(body))

	newServer(testConfig(t), testDB(t)).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestRoomTokenHandlerIssuesTokenForAdminSession(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))
	cookie := adminCookie(t, handler)

	body, _ := json.Marshal(tokenRequest{Identity: "user-1"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/rooms/team-standup/token", bytes.NewReader(body))
	request.AddCookie(cookie)

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
	if response.URL != "ws://localhost:7880" {
		t.Fatalf("unexpected livekit url: %q", response.URL)
	}
}

func TestRoomTokenHandlerRejectsInvalidRoomName(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))
	cookie := adminCookie(t, handler)

	body, _ := json.Marshal(tokenRequest{Identity: "user-1"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/rooms/ab/token", bytes.NewReader(body))
	request.AddCookie(cookie)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestRoomTokenHandlerRejectsInvalidIdentity(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))
	cookie := adminCookie(t, handler)

	body, _ := json.Marshal(tokenRequest{Identity: "user with spaces"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/rooms/team-standup/token", bytes.NewReader(body))
	request.AddCookie(cookie)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestRoomTokenHandlerRejectsMalformedBody(t *testing.T) {
	handler := newServer(testConfig(t), testDB(t))
	cookie := adminCookie(t, handler)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/rooms/team-standup/token", strings.NewReader("not-json"))
	request.AddCookie(cookie)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}
