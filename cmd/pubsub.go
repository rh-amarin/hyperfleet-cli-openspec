// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"github.com/spf13/cobra"
)

// pubsubCmd is the top-level group for GCP Pub/Sub operations.
var pubsubCmd = &cobra.Command{
	Use:   "pubsub",
	Short: "Interact with GCP Pub/Sub topics",
	Long: `Interact with GCP Pub/Sub topics.

Subcommands: list, publish.`,
}

func init() {
	rootCmd.AddCommand(pubsubCmd)
}
