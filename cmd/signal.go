package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

var signalCmd = &cobra.Command{
	Use:   "signal <event>",
	Short: "Send an event to the daemon HTTP server (prompt, thinking, stop)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		event := args[0]
		switch event {
		case "prompt", "thinking", "stop", "session_start", "stop_failure", "permission_request", "power_off":
		default:
			return fmt.Errorf("unknown event %q — valid events: prompt, thinking, stop, session_start, stop_failure, permission_request, power_off", event)
		}

		resp, err := http.Post(
			fmt.Sprintf("http://localhost:%d/event", 2187),
			"text/plain",
			strings.NewReader(event),
		)
		if err != nil {
			if errors.Is(err, errors.New("connection refused")) || isConnectionRefused(err) {
				return nil
			}
			return nil
		}
		resp.Body.Close()
		return nil
	},
}

func isConnectionRefused(err error) bool {
	return err != nil && strings.Contains(err.Error(), "connection refused")
}

func init() {
	rootCmd.AddCommand(signalCmd)
}
