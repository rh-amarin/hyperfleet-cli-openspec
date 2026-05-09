// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"github.com/spf13/cobra"
)

// maestroCmd is the top-level group for Maestro API operations.
var maestroCmd = &cobra.Command{
	Use:   "maestro",
	Short: "Interact with the Maestro API",
	Long: `Interact with the Maestro API.

Subcommands: list, get, delete, bundles, consumers.`,
}

func init() {
	rootCmd.AddCommand(maestroCmd)
}
