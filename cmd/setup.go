package cmd

import (
	"fmt"

	"github.com/peterfox/claude2-d2/internal/r2"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Scan for your R2-D2 and save its address to ~/.claude2-d2",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := adapter.Enable(); err != nil {
			return fmt.Errorf("bluetooth unavailable: %w", err)
		}

		fmt.Println("Scanning for R2-D2 (up to 15 seconds)...")
		address, name, err := r2.FindDevice(adapter)
		if err != nil {
			return err
		}

		fmt.Printf("Found: %s (%s)\n", name, address)

		cfg := &r2.Config{
			DeviceAddress: address,
			DeviceName:    name,
		}
		if err := r2.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Saved to ~/.claude2-d2 — run `claude2-d2 install` to set up the daemon.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
