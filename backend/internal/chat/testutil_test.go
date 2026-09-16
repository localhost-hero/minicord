package chat

import (
	"database/sql"
	"testing"

	"github.com/localhost-hero/minicord/backend/internal/store"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
