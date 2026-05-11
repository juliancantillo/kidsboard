package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Probe an HTTP endpoint and exit 0 (healthy) or 1 (unhealthy)",
	Long: `Used by Docker HEALTHCHECK and ad-hoc CLI checks. Defaults to
http://127.0.0.1:8080/healthz with a 3s timeout. Exit codes:
  0  endpoint returned 2xx
  1  network error, timeout, or non-2xx response

Env: KIDSBOARD_HEALTHCHECK_URL, KIDSBOARD_HEALTHCHECK_TIMEOUT`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		url := viper.GetString("healthcheck-url")
		timeout := viper.GetDuration("healthcheck-timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("User-Agent", "kidsboard-healthcheck/1")

		client := &http.Client{Timeout: timeout}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("unhealthy status: %s", resp.Status)
		}
		return nil
	},
}

func init() {
	healthcheckCmd.Flags().String("url", "http://127.0.0.1:8080/healthz", "URL to probe")
	healthcheckCmd.Flags().Duration("timeout", 3*time.Second, "Request timeout")
	// Prefixed viper keys so they don't collide with `serve --read-timeout`.
	must(viper.BindPFlag("healthcheck-url", healthcheckCmd.Flags().Lookup("url")))
	must(viper.BindPFlag("healthcheck-timeout", healthcheckCmd.Flags().Lookup("timeout")))
	rootCmd.AddCommand(healthcheckCmd)
}
