package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

var rootCmd = &cobra.Command{
	Use:   "r2",
	Short: "Control your Sphero R2-D2 and integrate it with Claude Code",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
