package http

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"cantillo.dev/kidsboard/internal/service"
	"cantillo.dev/kidsboard/internal/view"
)

// Deps bundles everything a controller needs. Wired once at boot.
// Every controller dependency is an interface — no sqldb leaks above storage.
type Deps struct {
	DB            *sql.DB
	Renderer      *view.Renderer
	Kids          service.KidService
	Categories    service.CategoryService
	ActivityTypes service.ActivityTypeService
	Activities    service.ActivityService
	Achievements  service.AchievementService
	Profile       service.ProfileService
}

// Options groups the HTTP server's runtime knobs. Keep additions here rather
// than growing Serve's parameter list.
type Options struct {
	Addr            string
	ShutdownTimeout time.Duration // how long to drain in-flight requests
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
}

// Serve binds the HTTP server and blocks until ctx is cancelled or the server
// errors. When ctx is cancelled (typically SIGINT/SIGTERM), the server stops
// accepting new connections and drains in-flight requests up to
// opts.ShutdownTimeout, then forces close. Returns the first non-trivial
// error encountered (boot, runtime, or shutdown).
func Serve(ctx context.Context, opts Options, deps Deps) error {
	if opts.ShutdownTimeout <= 0 {
		opts.ShutdownTimeout = 30 * time.Second
	}
	if opts.ReadTimeout <= 0 {
		opts.ReadTimeout = 30 * time.Second
	}
	if opts.IdleTimeout <= 0 {
		opts.IdleTimeout = 60 * time.Second
	}

	srv := &http.Server{
		Addr:              opts.Addr,
		Handler:           NewRouter(deps),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       opts.ReadTimeout,
		WriteTimeout:      opts.WriteTimeout, // 0 = unlimited (safe for HTMX/SSE later)
		IdleTimeout:       opts.IdleTimeout,
	}

	// Run the server in its own goroutine so the main flow can race ctx
	// cancellation against an unexpected server error.
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("kidsboard listening on http://%s", opts.Addr)
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			serverErr <- nil // expected on graceful shutdown
			return
		}
		serverErr <- err
	}()

	// Block until either the server exits on its own (panic during boot,
	// listener died) or the caller cancels ctx (signal).
	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	case <-ctx.Done():
		log.Printf("shutdown signal received, draining in-flight requests (timeout %s)", opts.ShutdownTimeout)
	}

	// Detach from the parent ctx — that one is already cancelled. Use a fresh
	// timeout-bounded context for the drain.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), opts.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		// Drain didn't finish in time. Force-close any leftover connections
		// so the process can exit; the timeout is the caller's promise about
		// "this is too long to wait."
		if forceErr := srv.Close(); forceErr != nil {
			log.Printf("force close: %v", forceErr)
		}
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	// Reap the listener goroutine's final exit (it should be nil here since
	// Shutdown returned cleanly). Bounded wait so we never hang on a bug.
	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
	case <-time.After(time.Second):
		log.Printf("listener goroutine didn't exit within 1s after shutdown")
	}

	log.Printf("kidsboard shutdown complete")
	return nil
}
