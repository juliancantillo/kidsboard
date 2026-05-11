package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// version is the build-time stamped string from main.version, threaded
// through SetVersion at startup. Falls back to debug.BuildInfo when run
// without ldflags (e.g. `go run .`) so the value is always meaningful.
var version = "dev"

// SetVersion is called from main() to inject the build-stamped version.
// Kept as a setter (vs. importing main.version) to avoid a cycle.
func SetVersion(v string) {
	if v != "" {
		version = v
	}
	rootCmd.Version = version
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print build version and runtime info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kidsboard %s\n", version)
		fmt.Printf("  go:       %s\n", runtime.Version())
		fmt.Printf("  platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		// debug.BuildInfo carries the VCS commit + module info when
		// available — useful when investigating "what's actually running."
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision", "vcs.time", "vcs.modified":
					fmt.Printf("  %s: %s\n", setting.Key, setting.Value)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
