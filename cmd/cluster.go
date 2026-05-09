// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"github.com/spf13/cobra"
)

// clusterCmd is the top-level group for all cluster operations.
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage HyperFleet clusters",
	Long: `Manage HyperFleet clusters.

Subcommands: create, get, list, search, patch, delete, id, conditions, statuses, adapter.`,
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
