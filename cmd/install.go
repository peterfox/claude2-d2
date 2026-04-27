package cmd

import (
	"fmt"
	"os"

	"github.com/peterfox/claude2-d2/internal/launchd"
	"github.com/peterfox/claude2-d2/internal/r2"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the r2 daemon as a launchd user agent (auto-starts on login)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := r2.LoadConfig(); err != nil {
			return err
		}

		binaryPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("could not resolve binary path: %w", err)
		}

		if err := launchd.WritePlist(binaryPath); err != nil {
			return fmt.Errorf("failed to write plist: %w", err)
		}

		if err := launchd.Load(); err != nil {
			return fmt.Errorf("failed to load daemon: %w", err)
		}

		fmt.Println("Daemon installed and started.")
		fmt.Println("Logs: /tmp/r2d2.log")
		fmt.Printf("Status: launchctl list | grep %s\n", launchd.Label)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
