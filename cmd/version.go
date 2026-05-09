// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"fmt"

	"github.com/rh-amarin/hyperfleet-cli/internal/version"
	"github.com/spf13/cobra"
)

// versionCmd prints the CLI version and exits.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the hf CLI version",
	Long:  `Print the hf CLI version string and exit.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(version.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
