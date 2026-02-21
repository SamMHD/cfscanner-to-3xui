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
		n, _ := cmd.Flags().GetInt("minutes")
		if n < 1 {
			return fmt.Errorf("minutes must be >= 1")
		}
		interval := time.Duration(n) * time.Minute
		runOnce := func() error {
			return runCmd.RunE(runCmd, nil)
		}
		for {
			if err := runOnce(); err != nil {
				return err
			}
			time.Sleep(interval)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCronCmd)
	runCronCmd.Flags().IntP("minutes", "n", 60, "Interval in minutes")
}
