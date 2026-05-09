// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"github.com/spf13/cobra"
)

// logsCmd streams pod logs matching a pattern (replaces stern).
var logsCmd = &cobra.Command{
	Use:   "logs <pattern>",
	Short: "Stream pod logs matching a pattern",
	Long: `Stream pod logs from one or more pods matching a pattern.

Subcommands: adapter.`,
}

func init() {
	rootCmd.AddCommand(logsCmd)
}
