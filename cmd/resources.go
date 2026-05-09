// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"github.com/spf13/cobra"
)

// resourcesCmd displays a combined overview of all clusters and their node pools.
var resourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Show a combined overview of all clusters and their node pools",
	Long: `Show a combined overview of all clusters and their node pools.

This command does not support --table; it always outputs a combined summary.`,
}

func init() {
	rootCmd.AddCommand(resourcesCmd)
}
