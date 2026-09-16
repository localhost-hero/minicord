// Package store owns the embedded SQLite database used for invites and chat history.
package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Open opens (creating if necessary) the SQLite database at path and applies
// the schema migrations.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("store: open %s: %w", path, err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: migrate: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS invites (
			id TEXT PRIMARY KEY,
			room TEXT NOT NULL,
			token TEXT NOT NULL DEFAULT '',
			token_hash TEXT NOT NULL UNIQUE,
			created_at INTEGER NOT NULL,
			expires_at INTEGER,
			revoked_at INTEGER
		)`,
		`CREATE INDEX IF NOT EXISTS idx_invites_room ON invites(room)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			room TEXT NOT NULL,
			identity TEXT NOT NULL,
			body TEXT NOT NULL,
			created_at INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_room_created ON messages(room, created_at)`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("exec %q: %w", statement, err)
		}
	}

	// Add token column if migrating from an older schema version
	_, _ = db.Exec(`ALTER TABLE invites ADD COLUMN token TEXT NOT NULL DEFAULT ''`)

	return nil
}
