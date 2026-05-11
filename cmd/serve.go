package cmd

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	khttp "cantillo.dev/kidsboard/internal/http"
	"cantillo.dev/kidsboard/internal/service"
	"cantillo.dev/kidsboard/internal/storage"
	"cantillo.dev/kidsboard/internal/view"
	"github.com/spf13/cobra"
)

var (
	serveAddr            string
	serveDBPath          string
	serveShutdownTimeout time.Duration
	serveReadTimeout     time.Duration
	serveIdleTimeout     time.Duration
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the kidsboard HTTP server",
	Long: `Starts the HTTP server. Handles SIGINT/SIGTERM by draining
in-flight requests within --shutdown-timeout before closing the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		db, err := storage.OpenSQLite(ctx, serveDBPath)
		if err != nil {
			return fmt.Errorf("open db: %w", err)
		}
		// Close happens AFTER Serve returns — by then srv.Shutdown has finished
		// and no requests are using the DB, so the WAL checkpoint is safe.
		defer func() {
			if err := storage.Close(db); err != nil {
				log.Printf("db close: %v", err)
			}
		}()

		renderer, err := view.NewRenderer()
		if err != nil {
			return fmt.Errorf("build renderer: %w", err)
		}

		balance := service.NewBalanceService()
		achievements := service.NewAchievementService(balance)
		activityTypes := service.NewActivityTypeService()
		activities := service.NewActivityService(activityTypes, achievements)
		categories := service.NewCategoryService()
		kids := service.NewKidService(view.AvatarSlugs())
		profile := service.NewProfileService(balance)

		return khttp.Serve(ctx, khttp.Options{
			Addr:            serveAddr,
			ShutdownTimeout: serveShutdownTimeout,
			ReadTimeout:     serveReadTimeout,
			IdleTimeout:     serveIdleTimeout,
		}, khttp.Deps{
			DB:            db,
			Renderer:      renderer,
			Kids:          kids,
			Categories:    categories,
			ActivityTypes: activityTypes,
			Activities:    activities,
			Achievements:  achievements,
			Profile:       profile,
		})
	},
}

func init() {
	serveCmd.Flags().StringVar(&serveAddr, "addr", ":8080", "Address to listen on")
	serveCmd.Flags().StringVar(&serveDBPath, "db", "kidsboard.db", "SQLite database file path")
	serveCmd.Flags().DurationVar(&serveShutdownTimeout, "shutdown-timeout", 30*time.Second, "Max time to drain in-flight requests on shutdown")
	serveCmd.Flags().DurationVar(&serveReadTimeout, "read-timeout", 30*time.Second, "HTTP read timeout")
	serveCmd.Flags().DurationVar(&serveIdleTimeout, "idle-timeout", 60*time.Second, "HTTP idle timeout for keep-alive connections")
	rootCmd.AddCommand(serveCmd)
}
