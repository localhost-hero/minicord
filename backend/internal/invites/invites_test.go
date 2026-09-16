package invites

import (
	"testing"
	"time"

	"github.com/localhost-hero/minicord/backend/internal/store"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test store: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewStore(db)
}

func TestCreateAndRedeem(t *testing.T) {
	invites := newTestStore(t)

	invite, token, err := invites.Create("team-standup", time.Hour)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected a non-empty token")
	}

	redeemed, err := invites.Redeem(token)
	if err != nil {
		t.Fatalf("expected redeem to succeed, got %v", err)
	}
	if redeemed.ID != invite.ID || redeemed.Room != "team-standup" {
		t.Fatalf("unexpected redeemed invite: %+v", redeemed)
	}
}

func TestRedeemRejectsUnknownToken(t *testing.T) {
	invites := newTestStore(t)

	if _, err := invites.Redeem("unknown-token"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRedeemRejectsExpiredInvite(t *testing.T) {
	invites := newTestStore(t)

	_, token, err := invites.Create("team-standup", -time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := invites.Redeem(token); err != ErrExpiredOrRevoked {
		t.Fatalf("expected ErrExpiredOrRevoked, got %v", err)
	}
}

func TestRedeemRejectsRevokedInvite(t *testing.T) {
	invites := newTestStore(t)

	invite, token, err := invites.Create("team-standup", time.Hour)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := invites.Revoke(invite.ID); err != nil {
		t.Fatalf("expected revoke to succeed, got %v", err)
	}

	if _, err := invites.Redeem(token); err != ErrExpiredOrRevoked {
		t.Fatalf("expected ErrExpiredOrRevoked, got %v", err)
	}
}

func TestCreateWithoutTTLNeverExpires(t *testing.T) {
	invites := newTestStore(t)

	_, token, err := invites.Create("team-standup", 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := invites.Redeem(token); err != nil {
		t.Fatalf("expected a no-TTL invite to redeem successfully, got %v", err)
	}
}

func TestCreateReusesExistingActiveInvite(t *testing.T) {
	invites := newTestStore(t)

	invite1, token1, err := invites.Create("team-standup", time.Hour)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	invite2, token2, err := invites.Create("team-standup", time.Hour)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if invite1.ID != invite2.ID || token1 != token2 {
		t.Fatalf("expected existing active invite to be reused, got different invites: %+v vs %+v", invite1, invite2)
	}

	if err := invites.Revoke(invite1.ID); err != nil {
		t.Fatalf("expected revoke to succeed, got %v", err)
	}

	invite3, token3, err := invites.Create("team-standup", time.Hour)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if invite3.ID == invite1.ID || token3 == token1 {
		t.Fatalf("expected new invite to be created after revoking previous one")
	}
}
