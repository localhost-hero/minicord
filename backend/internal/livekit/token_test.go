package livekit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewAccessTokenRejectsMissingCredentials(t *testing.T) {
	if _, err := NewAccessToken("", "secret", "user-1", "room-1", time.Minute); err == nil {
		t.Fatal("expected an error for a missing api key")
	}
	if _, err := NewAccessToken("key", "", "user-1", "room-1", time.Minute); err == nil {
		t.Fatal("expected an error for a missing api secret")
	}
}

func TestNewAccessTokenRejectsMissingIdentityOrRoom(t *testing.T) {
	if _, err := NewAccessToken("key", "secret", "", "room-1", time.Minute); err == nil {
		t.Fatal("expected an error for a missing identity")
	}
	if _, err := NewAccessToken("key", "secret", "user-1", "", time.Minute); err == nil {
		t.Fatal("expected an error for a missing room")
	}
}

func TestNewAccessTokenProducesVerifiableSignature(t *testing.T) {
	token, err := NewAccessToken("key", "secret", "user-1", "room-1", time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT segments, got %d", len(parts))
	}

	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if parts[2] != expectedSignature {
		t.Fatal("token signature does not match the expected HMAC-SHA256 signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode claims: %v", err)
	}

	var decoded claims
	if err := json.Unmarshal(claimsJSON, &decoded); err != nil {
		t.Fatalf("failed to unmarshal claims: %v", err)
	}

	if decoded.Issuer != "key" {
		t.Fatalf("unexpected issuer: %q", decoded.Issuer)
	}
	if decoded.Subject != "user-1" {
		t.Fatalf("unexpected subject: %q", decoded.Subject)
	}
	if decoded.Video.Room != "room-1" {
		t.Fatalf("unexpected room: %q", decoded.Video.Room)
	}
	if !decoded.Video.RoomJoin || !decoded.Video.CanPublish || !decoded.Video.CanSubscribe {
		t.Fatal("expected the video grant to allow join, publish, and subscribe")
	}
	if decoded.Expiration <= decoded.IssuedAt {
		t.Fatal("expected expiration to be after issued-at time")
	}
}

func TestNewAccessTokenAppliesDefaultTTLWhenNonPositive(t *testing.T) {
	token, err := NewAccessToken("key", "secret", "user-1", "room-1", 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	parts := strings.Split(token, ".")
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode claims: %v", err)
	}

	var decoded claims
	if err := json.Unmarshal(claimsJSON, &decoded); err != nil {
		t.Fatalf("failed to unmarshal claims: %v", err)
	}

	if decoded.Expiration-decoded.IssuedAt != int64(DefaultTokenTTL.Seconds()) {
		t.Fatalf("expected default TTL of %v seconds", DefaultTokenTTL.Seconds())
	}
}
