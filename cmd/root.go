/*
Copyright © 2026 Julian Cantillo <julian@cantillo.dev>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"cantillo.dev/kidsboard/internal/applog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// envPrefix scopes every environment-variable binding. A flag named --addr
// is read from KIDSBOARD_ADDR; --shutdown-timeout from KIDSBOARD_SHUTDOWN_TIMEOUT.
const envPrefix = "KIDSBOARD"

// cfgFile is the optional --config path, populated by the persistent flag.
var cfgFile string

// rootCmd is the entry point. All subcommands hang off this. Persistent
// flags here apply everywhere: --config (file), --log-level, --log-format.
var rootCmd = &cobra.Command{
	Use:   "kidsboard",
	Short: "RPG-style household activity tracker",
	Long: `Kidsboard tracks chores, school, faith, meals, and other household
activities as XP-and-points game progression for kids. Single Go binary,
embedded SQLite, server-rendered Tailwind UI.`,
	SilenceUsage: true,
	// PersistentPreRunE runs before every subcommand's RunE. By the time
	// it fires, viper has already loaded env + config + flag values, so
	// we can apply log settings consistently across the binary.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		applog.Setup(viper.GetString("log-level"), viper.GetString("log-format"))
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Optional config file (default: search $HOME/.kidsboard.yaml then ./kidsboard.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level: debug, info, warn, error")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format: text (dev) or json (prod)")

	// Bind persistent flags to viper so KIDSBOARD_LOG_LEVEL etc. work.
	must(viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level")))
	must(viper.BindPFlag("log-format", rootCmd.PersistentFlags().Lookup("log-format")))
}

// initConfig configures viper's env + file behavior. Called once per
// process via cobra.OnInitialize. 12-factor §III: env vars are the
// primary config source; the optional config file is a convenience for
// local dev only.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("kidsboard")
		viper.SetConfigType("yaml")
		if home, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(home)
		}
		viper.AddConfigPath(".")
	}

	viper.SetEnvPrefix(envPrefix)
	// Map flag-name conventions (kebab-case + dots) to env-var conventions
	// (UPPER_SNAKE). So --shutdown-timeout reads KIDSBOARD_SHUTDOWN_TIMEOUT.
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

	// Missing config file is fine — env + flag defaults suffice.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "using config file: %s\n", viper.ConfigFileUsed())
	}
}

// must panics on a non-nil error during init wiring. Used for cobra/viper
// binding calls that should never fail in correct code.
func must(err error) {
	if err != nil {
		panic(err)
	}
}
