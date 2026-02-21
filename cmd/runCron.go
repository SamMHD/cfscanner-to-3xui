package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var runCronCmd = &cobra.Command{
	Use:   "run-cron",
	Short: "Run scan/generate/update every N minutes",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Interval in minutes between runs (default 60)
		n := envInt("CRON_MINUTES", 60)
		if cmd.Flags().Changed("minutes") {
			n, _ = cmd.Flags().GetInt("minutes")
		}
		if n < 1 {
			return fmt.Errorf("minutes must be >= 1")
		}
		interval := time.Duration(n) * time.Minute
		runOnce := func() error {
			return runCmd.RunE(runCmd, nil)
		}
		for {
			var err error
			panicked := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("[run-cron] panic recovered: %v\n", r)
						panicked = true
					}
				}()
				err = runOnce()
			}()
			if panicked {
				// continue to next cycle
			} else if err != nil {
				return err
			} else {
				fmt.Println("[run-cron] cycle completed")
			}
			time.Sleep(interval)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCronCmd)
	runCronCmd.Flags().IntP("minutes", "n", 60, "Interval in minutes (overrides CRON_MINUTES)")
}
