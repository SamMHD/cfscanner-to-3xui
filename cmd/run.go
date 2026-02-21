package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run scan, then generate, then update",
	RunE: func(cmd *cobra.Command, args []string) error {
		root := cmd.Root()
		for _, name := range []string{"scan", "generate", "update"} {
			c, _, err := root.Find([]string{name})
			if err != nil {
				return fmt.Errorf("find %s: %w", name, err)
			}
			if name == "scan" {
				c.Run(c, nil)
				continue
			}
			if err := c.RunE(c, nil); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			fmt.Printf("[%s] completed\n", name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
