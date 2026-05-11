package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	khttp "cantillo.dev/kidsboard/internal/http"
	"cantillo.dev/kidsboard/internal/service"
	"cantillo.dev/kidsboard/internal/storage"
	"cantillo.dev/kidsboard/internal/view"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the kidsboard HTTP server",
	Long: `Starts the HTTP server. Handles SIGINT/SIGTERM by draining
in-flight requests within --shutdown-timeout before closing the database.

All flags are also readable via env vars (KIDSBOARD_<FLAG_UPPER>):
  KIDSBOARD_ADDR              → --addr
  KIDSBOARD_DB                → --db
  KIDSBOARD_SHUTDOWN_TIMEOUT  → --shutdown-timeout
  KIDSBOARD_READ_TIMEOUT      → --read-timeout
  KIDSBOARD_IDLE_TIMEOUT      → --idle-timeout`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		dbPath := viper.GetString("db")
		addr := viper.GetString("addr")
		shutdownTimeout := viper.GetDuration("shutdown-timeout")
		readTimeout := viper.GetDuration("read-timeout")
		idleTimeout := viper.GetDuration("idle-timeout")

		db, err := storage.OpenSQLite(ctx, dbPath)
		if err != nil {
			return fmt.Errorf("open db: %w", err)
		}
		defer func() {
			if err := storage.Close(db); err != nil {
				slog.Error("db close", "err", err)
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
			Addr:            addr,
			ShutdownTimeout: shutdownTimeout,
			ReadTimeout:     readTimeout,
			IdleTimeout:     idleTimeout,
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
	serveCmd.Flags().String("addr", ":8080", "Address to listen on")
	serveCmd.Flags().String("db", "kidsboard.db", "SQLite database file path")
	serveCmd.Flags().Duration("shutdown-timeout", 30*time.Second, "Max time to drain in-flight requests on shutdown")
	serveCmd.Flags().Duration("read-timeout", 30*time.Second, "HTTP read timeout")
	serveCmd.Flags().Duration("idle-timeout", 60*time.Second, "HTTP idle timeout for keep-alive connections")

	must(viper.BindPFlag("addr", serveCmd.Flags().Lookup("addr")))
	must(viper.BindPFlag("db", serveCmd.Flags().Lookup("db")))
	must(viper.BindPFlag("shutdown-timeout", serveCmd.Flags().Lookup("shutdown-timeout")))
	must(viper.BindPFlag("read-timeout", serveCmd.Flags().Lookup("read-timeout")))
	must(viper.BindPFlag("idle-timeout", serveCmd.Flags().Lookup("idle-timeout")))

	rootCmd.AddCommand(serveCmd)
}
