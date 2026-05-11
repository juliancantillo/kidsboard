package testutil

import (
	"context"
	"database/sql"
	"testing"

	"cantillo.dev/kidsboard/internal/storage"
	"github.com/stretchr/testify/require"
)

// NewDB opens a fresh in-memory SQLite database, applies all migrations,
// and registers a Cleanup to close it. Each test gets its own DB — no sharing.
func NewDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := storage.OpenSQLite(context.Background(), ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}
