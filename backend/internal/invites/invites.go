// Package invites manages room invite links: unauthenticated guests join a
// room only by presenting a valid, unexpired, unrevoked invite token.
package invites

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when an invite token does not match any stored invite.
var ErrNotFound = errors.New("invites: not found")

// ErrExpiredOrRevoked is returned when an invite exists but is no longer usable.
var ErrExpiredOrRevoked = errors.New("invites: expired or revoked")

// Invite describes a stored room invitation.
type Invite struct {
	ID        string
	Room      string
	CreatedAt time.Time
	ExpiresAt *time.Time
	RevokedAt *time.Time
}

// Store persists invites in the database.
type Store struct {
	db *sql.DB
}

// NewStore returns an invite store backed by db.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Create returns an active invite token for room, reusing an existing active
// invite if one already exists. When ttl is zero the invite never expires
// until explicitly revoked; a negative ttl produces an already-expired invite,
// which is useful for tests.
func (s *Store) Create(room string, ttl time.Duration) (Invite, string, error) {
	nowUnix := time.Now().UTC().Unix()

	if ttl >= 0 {
		var (
			existingInvite Invite
			existingToken  string
			createdAt      int64
			expiresAt      sql.NullInt64
		)
		row := s.db.QueryRow(
			`SELECT id, room, token, created_at, expires_at FROM invites
			 WHERE room = ? AND revoked_at IS NULL AND token != '' AND (expires_at IS NULL OR expires_at > ?)
			 ORDER BY created_at DESC LIMIT 1`,
			room, nowUnix,
		)
		err := row.Scan(&existingInvite.ID, &existingInvite.Room, &existingToken, &createdAt, &expiresAt)
		if err == nil && existingToken != "" {
			existingInvite.CreatedAt = time.Unix(createdAt, 0).UTC()
			if expiresAt.Valid {
				t := time.Unix(expiresAt.Int64, 0).UTC()
				existingInvite.ExpiresAt = &t
			}
			return existingInvite, existingToken, nil
		}
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return Invite{}, "", fmt.Errorf("invites: generate token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	invite := Invite{
		ID:        uuid.NewString(),
		Room:      room,
		CreatedAt: time.Now().UTC(),
	}
	if ttl != 0 {
		expiresAt := invite.CreatedAt.Add(ttl)
		invite.ExpiresAt = &expiresAt
	}

	_, err := s.db.Exec(
		`INSERT INTO invites (id, room, token, token_hash, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?)`,
		invite.ID, invite.Room, token, hashToken(token), invite.CreatedAt.Unix(), nullableUnix(invite.ExpiresAt),
	)
	if err != nil {
		return Invite{}, "", fmt.Errorf("invites: insert: %w", err)
	}

	return invite, token, nil
}

// Redeem validates a plaintext invite token and returns the invite if it is
// usable: it exists, is not expired, and has not been revoked.
func (s *Store) Redeem(token string) (Invite, error) {
	row := s.db.QueryRow(
		`SELECT id, room, created_at, expires_at, revoked_at FROM invites WHERE token_hash = ?`,
		hashToken(token),
	)

	var (
		invite    Invite
		createdAt int64
		expiresAt sql.NullInt64
		revokedAt sql.NullInt64
	)
	if err := row.Scan(&invite.ID, &invite.Room, &createdAt, &expiresAt, &revokedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Invite{}, ErrNotFound
		}
		return Invite{}, fmt.Errorf("invites: query: %w", err)
	}

	invite.CreatedAt = time.Unix(createdAt, 0).UTC()
	if expiresAt.Valid {
		t := time.Unix(expiresAt.Int64, 0).UTC()
		invite.ExpiresAt = &t
	}
	if revokedAt.Valid {
		t := time.Unix(revokedAt.Int64, 0).UTC()
		invite.RevokedAt = &t
		return invite, ErrExpiredOrRevoked
	}
	if invite.ExpiresAt != nil && time.Now().After(*invite.ExpiresAt) {
		return invite, ErrExpiredOrRevoked
	}

	return invite, nil
}

// Revoke marks an invite as no longer usable. It is idempotent.
func (s *Store) Revoke(id string) error {
	_, err := s.db.Exec(
		`UPDATE invites SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL`,
		time.Now().UTC().Unix(), id,
	)
	if err != nil {
		return fmt.Errorf("invites: revoke: %w", err)
	}
	return nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func nullableUnix(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Unix()
}
