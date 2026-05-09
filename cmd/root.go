// Package cmd contains the Cobra command definitions for the hf CLI.
// Each domain has its own file (cluster.go, nodepool.go, etc.) that registers
// its commands with the root command via init().
package cmd

import (
	"github.com/spf13/cobra"
)

// Global flag values — populated by PersistentPreRunE before any RunE fires.
var (
	cfgFile     string
	outputFmt   string
	noColor     bool
	verbose     bool
	apiURL      string
	apiToken    string
)

// rootCmd is the base command for the hf CLI.
var rootCmd = &cobra.Command{
	Use:   "hf",
	Short: "HyperFleet CLI — manage HyperFleet clusters",
	Long: `hf is a self-contained CLI for managing HyperFleet clusters.
It replaces a suite of bash scripts with a single binary — no external tools required.

Run 'hf config doctor' to verify connectivity to the HyperFleet API.`,
	// SilenceUsage prevents the usage message from being printed on every error.
	SilenceUsage: true,
	// SilenceErrors prevents duplicate error printing (the caller handles it).
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// All global flag values are already bound to the package-level vars
		// via BindPFlags / the flag definitions below. Nothing extra needed here
		// for the scaffold phase — individual commands will extend this when
		// they need to initialise the config or API client.
		return nil
	},
}

// Execute runs the root command and returns any error to main.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// --config: override config file location
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/hf/config.yaml)")

	// --output: output format; default varies per command — root default is json
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "json", "output format: json, table, yaml")

	// --no-color: disable ANSI colour output
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")

	// --verbose / -v: enable debug logging to stderr
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose/debug logging")

	// --api-url: override the HyperFleet API base URL for this invocation
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "override HyperFleet API URL for this invocation")

	// --api-token: override the API bearer token for this invocation
	rootCmd.PersistentFlags().StringVar(&apiToken, "api-token", "", "override API bearer token for this invocation")
}
