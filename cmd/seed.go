package cmd

import (
	"context"
	"fmt"

	"cantillo.dev/kidsboard/internal/seed"
	"cantillo.dev/kidsboard/internal/storage"
	"github.com/spf13/cobra"
)

var seedDBPath string

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Upsert the curated categories, activity types, and achievements",
	Long: `Applies the in-code seed data (defined in internal/seed/data.go) to the
SQLite database. Idempotent — re-running updates fields by slug without
changing row IDs, so kid_achievements and activity FKs are preserved.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		db, err := storage.OpenSQLite(ctx, seedDBPath)
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
		fmt.Printf("Seeded: %d categorías, %d tipos de actividad, %d logros (%d reglas)\n",
			stats.Categories, stats.ActivityTypes, stats.Achievements, stats.AchievementRules)
		return nil
	},
}

func init() {
	seedCmd.Flags().StringVar(&seedDBPath, "db", "kidsboard.db", "SQLite database file path")
	rootCmd.AddCommand(seedCmd)
}
