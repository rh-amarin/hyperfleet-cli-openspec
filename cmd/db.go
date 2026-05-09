// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"github.com/spf13/cobra"
)

// dbCmd is the top-level group for database operations.
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Run database operations against the HyperFleet PostgreSQL database",
	Long: `Run database operations against the HyperFleet PostgreSQL database.

Subcommands: query, delete, config.`,
}

func init() {
	rootCmd.AddCommand(dbCmd)
}
