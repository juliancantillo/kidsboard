package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var (
	healthcheckURL     string
	healthcheckTimeout time.Duration
)

// healthcheckCmd is the binary's own HTTP health prober. Lets the distroless
// container run a real Docker HEALTHCHECK without needing curl/wget. Exits 0
// if the URL returns 2xx within the timeout, non-zero otherwise.
var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Probe an HTTP endpoint and exit 0 (healthy) or 1 (unhealthy)",
	Long: `Used by Docker HEALTHCHECK and ad-hoc CLI checks. Defaults to
http://127.0.0.1:8080/healthz with a 3s timeout. Exit codes:
  0  endpoint returned 2xx
  1  network error, timeout, or non-2xx response`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), healthcheckTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthcheckURL, nil)
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("User-Agent", "kidsboard-healthcheck/1")

		client := &http.Client{Timeout: healthcheckTimeout}
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
	healthcheckCmd.Flags().StringVar(&healthcheckURL, "url", "http://127.0.0.1:8080/healthz", "URL to probe")
	healthcheckCmd.Flags().DurationVar(&healthcheckTimeout, "timeout", 3*time.Second, "Request timeout")
	rootCmd.AddCommand(healthcheckCmd)
}
