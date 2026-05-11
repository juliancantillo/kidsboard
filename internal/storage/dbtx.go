package storage

import "cantillo.dev/kidsboard/internal/storage/sqldb"

// DBTX is the connection-handle interface used by every repository.
// Both *sql.DB and *sql.Tx satisfy it, so callers control whether work
// runs in its own transaction or joins one already open. Aliased to
// sqlc's generated DBTX so we have a single canonical interface.
type DBTX = sqldb.DBTX
