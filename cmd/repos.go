// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"github.com/spf13/cobra"
)

// reposCmd shows a GitHub registry overview for HyperFleet repositories.
var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "Show an overview of HyperFleet GitHub repositories",
	Long:  `Show an overview of HyperFleet GitHub repositories including package versions and digests.`,
}

func init() {
	rootCmd.AddCommand(reposCmd)
}
