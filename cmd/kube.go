// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"github.com/spf13/cobra"
)

// kubeCmd is the top-level group for Kubernetes operations.
var kubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "Perform Kubernetes operations without requiring kubectl",
	Long: `Perform Kubernetes operations without requiring kubectl.

Subcommands: port-forward, curl, debug.`,
}

func init() {
	rootCmd.AddCommand(kubeCmd)
}
