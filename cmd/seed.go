package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"cantillo.dev/kidsboard/internal/seed"
	"cantillo.dev/kidsboard/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Upsert the curated categories, activity types, and achievements",
	Long: `Applies the in-code seed data to the SQLite database. Idempotent —
re-running updates fields by slug without changing row IDs, so
kid_achievements and activity FKs are preserved.

Reads --db / KIDSBOARD_DB for the database path.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dbPath := viper.GetString("db")

		db, err := storage.OpenSQLite(ctx, dbPath)
		if err != nil {
			return fmt.Errorf("open db: %w", err)
		}
		defer db.Close()

		if err := seed.Run(ctx, db); err != nil {
			return fmt.Errorf("seed: %w", err)
		}
		stats, err := seed.Count(ctx, db)
		if err != nil {
			return fmt.Errorf("count: %w", err)
		}
		slog.Info("seed applied",
			"categories", stats.Categories,
			"activity_types", stats.ActivityTypes,
			"achievements", stats.Achievements,
			"rules", stats.AchievementRules,
		)
		return nil
	},
}

func init() {
	seedCmd.Flags().String("db", "kidsboard.db", "SQLite database file path")
	must(viper.BindPFlag("db", seedCmd.Flags().Lookup("db")))
	rootCmd.AddCommand(seedCmd)
}
