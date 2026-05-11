package cmd

import (
	"context"
	"fmt"

	"cantillo.dev/kidsboard/internal/storage"
	"github.com/spf13/cobra"
)

var migrateDBPath string

// migrateCmd applies pending schema migrations and exits. Designed to be run
// as a Kubernetes init container before the main `serve` container starts.
// Idempotent — runs already-applied migrations as no-ops via goose.
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply pending database migrations and exit",
	Long: `Opens the SQLite database at --db, runs any pending goose migrations,
checkpoints the WAL, and exits. Idempotent: running with no pending
migrations is a fast no-op. Intended for init containers and CI deploys.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		db, err := storage.OpenSQLite(ctx, migrateDBPath)
		if err != nil {
			return fmt.Errorf("open db: %w", err)
		}
		// OpenSQLite already ran migrations on open. Close cleanly so the WAL
		// is checkpointed back into the main file before the container exits.
		return storage.Close(db)
	},
}

func init() {
	migrateCmd.Flags().StringVar(&migrateDBPath, "db", "kidsboard.db", "SQLite database file path")
	rootCmd.AddCommand(migrateCmd)
}
