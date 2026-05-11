package cmd

import (
	"context"
	"fmt"

	"cantillo.dev/kidsboard/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// migrateCmd applies pending schema migrations and exits. Designed to be run
// as a Kubernetes init container before the main `serve` container starts.
// Idempotent — runs already-applied migrations as no-ops via goose.
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply pending database migrations and exit",
	Long: `Opens the SQLite database at --db (or $KIDSBOARD_DB), runs any
pending goose migrations, checkpoints the WAL, and exits.
Idempotent: a no-op when nothing is pending. Intended for init containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		db, err := storage.OpenSQLite(ctx, viper.GetString("db"))
		if err != nil {
			return fmt.Errorf("open db: %w", err)
		}
		return storage.Close(db)
	},
}

func init() {
	migrateCmd.Flags().String("db", "kidsboard.db", "SQLite database file path")
	// `db` is shared with `serve` and `seed` — the same viper key powers all
	// three. Re-binding is harmless: viper overwrites the previous flag bind.
	must(viper.BindPFlag("db", migrateCmd.Flags().Lookup("db")))
	rootCmd.AddCommand(migrateCmd)
}
