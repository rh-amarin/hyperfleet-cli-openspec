// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"github.com/spf13/cobra"
)

// nodepoolCmd is the top-level group for all node pool operations.
var nodepoolCmd = &cobra.Command{
	Use:   "nodepool",
	Short: "Manage HyperFleet node pools",
	Long: `Manage HyperFleet node pools.

Subcommands: create, get, list, search, patch, delete, id, conditions, statuses, adapter.`,
}

func init() {
	rootCmd.AddCommand(nodepoolCmd)
}
