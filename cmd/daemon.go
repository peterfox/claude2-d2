package cmd

import (
	"fmt"
	"os/signal"
	"syscall"

	"github.com/peterfox/claude2-d2/internal/daemon"
	"github.com/peterfox/claude2-d2/internal/r2"
	"github.com/spf13/cobra"
)

var debugMode bool

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Connect to R2-D2 and listen for HTTP events on :2187",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := adapter.Enable(); err != nil {
			return fmt.Errorf("bluetooth unavailable: %w", err)
		}

		cfg, err := r2.LoadConfig()
		if err != nil {
			return err
		}

		fmt.Printf("Connecting to %s (%s)...\n", cfg.DeviceName, cfg.DeviceAddress)
		client, err := r2.ConnectByAddress(adapter, cfg.DeviceAddress)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
		defer client.Disconnect()

		fmt.Printf("Connected. Listening on :%d\n", daemon.Port)

		reconnect := func() (*r2.Client, error) {
			return r2.ConnectByAddress(adapter, cfg.DeviceAddress)
		}
		machine := daemon.NewMachine(client, reconnect)

		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		defer machine.Stop()

		daemon.ServeEvents(ctx, machine, debugMode)

		return nil
	},
}

func init() {
	daemonCmd.Flags().BoolVarP(&debugMode, "debug", "d", false, "Log received events with timestamps")
	rootCmd.AddCommand(daemonCmd)
}
