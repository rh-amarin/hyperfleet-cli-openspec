// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"github.com/spf13/cobra"
)

// configCmd is the top-level group for configuration management.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage hf configuration",
	Long: `Manage hf configuration.

Subcommands: show, set, env.`,
}

func init() {
	rootCmd.AddCommand(configCmd)
}
