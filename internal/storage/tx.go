package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// WithTx runs fn inside a write transaction (BEGIN IMMEDIATE via Serializable
// isolation) and commits if fn returns nil, otherwise rolls back. The tx is
// exposed to fn as DBTX so callees don't depend on *sql.Tx directly.
func WithTx[T any](ctx context.Context, db *sql.DB, fn func(DBTX) (T, error)) (T, error) {
	var zero T
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return zero, fmt.Errorf("begin tx: %w", err)
	}
	result, err := fn(tx)
	if err != nil {
		_ = tx.Rollback()
		return zero, err
	}
	if err := tx.Commit(); err != nil {
		return zero, fmt.Errorf("commit tx: %w", err)
	}
	return result, nil
}
