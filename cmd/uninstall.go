package cmd

import (
	"fmt"

	"github.com/peterfox/claude2-d2/internal/launchd"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Stop and remove the r2 launchd user agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := launchd.Unload(); err != nil {
			return fmt.Errorf("failed to stop daemon: %w", err)
		}

		if err := launchd.RemovePlist(); err != nil {
			return fmt.Errorf("failed to remove plist: %w", err)
		}

		fmt.Println("Daemon stopped and removed.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
