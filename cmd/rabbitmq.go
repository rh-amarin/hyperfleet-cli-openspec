// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"github.com/spf13/cobra"
)

// rabbitmqCmd is the top-level group for RabbitMQ operations.
var rabbitmqCmd = &cobra.Command{
	Use:   "rabbitmq",
	Short: "Publish events to RabbitMQ exchanges",
	Long: `Publish events to RabbitMQ exchanges.

Subcommands: publish.`,
}

func init() {
	rootCmd.AddCommand(rabbitmqCmd)
}
