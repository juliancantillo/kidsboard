// Package applog wires the process's logging. 12-factor §XI: logs are
// event streams written to stdout; the platform decides what to do with
// them. We never open a log file.
//
// Format is selectable: `text` (human-friendly for local) or `json`
// (machine-readable for production / Loki / Datadog).
package applog

import (
	stdlog "log"
	"log/slog"
	"os"
	"strings"
)

// Setup configures slog as the default logger AND bridges stdlib `log`
// through it so third-party libs (and our own legacy log.Printf calls)
// flow into the same handler. Idempotent — safe to call from any
// subcommand's PersistentPreRun.
func Setup(level, format string) {
	lvl := parseLevel(level)
	opts := &slog.HandlerOptions{Level: lvl}

	var h slog.Handler
	switch strings.ToLower(format) {
	case "json":
		h = slog.NewJSONHandler(os.Stdout, opts)
	default:
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(h))

	// Bridge: route every stdlib log.Printf through slog. Stdlib log loses
	// its own prefix/flags so the slog formatter is the single source of
	// truth for log line format.
	stdlog.SetFlags(0)
	stdlog.SetPrefix("")
	stdlog.SetOutput(&slogWriter{})
}

// slogWriter adapts io.Writer (which stdlib `log` expects) to slog.Info.
type slogWriter struct{}

func (slogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\n")
	slog.Default().Info(msg)
	return len(p), nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "err":
		return slog.LevelError
	case "info", "":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}
