// Package cmd contains the Cobra command definitions for the hf CLI.
// Each domain has its own file (cluster.go, nodepool.go, etc.) that registers
// its commands with the root command via init().
package cmd

import (
	"fmt"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/spf13/cobra"
)

// Global flag values — populated by PersistentPreRunE before any RunE fires.
var (
	cfgFile   string
	outputFmt string
	noColor   bool
	verbose   bool
	curlMode  bool
	apiURL    string
	apiToken  string
)

// rootCmd is the base command for the hf CLI.
var rootCmd = &cobra.Command{
	Use:   "hf",
	Short: "HyperFleet CLI — manage HyperFleet clusters",
	Long: `hf is a self-contained CLI for managing HyperFleet clusters.
It replaces a suite of bash scripts with a single binary — no external tools required.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if isBypassCommand(cmd) {
			return nil
		}
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}
		if _, err := s.RequireActiveEnvironment(); err != nil {
			return fmt.Errorf("[ERROR] no active environment\n  → run 'hf config env create <name>' to create one\n  → run 'hf config env activate <name>' to activate an existing one")
		}
		return nil
	},
}

// isBypassCommand returns true for commands that may run without an active environment.
// Bypassed: all `config env *` subcommands, version, completion, help,
// and cobra's built-in completion helpers.
func isBypassCommand(cmd *cobra.Command) bool {
	path := cmd.CommandPath()
	if strings.Contains(path, "config env") {
		return true
	}
	leaf := cmd.Name()
	switch leaf {
	case "version", "completion", "help":
		return true
	}
	if strings.HasPrefix(leaf, "__complete") {
		return true
	}
	// Hidden daemon commands (port-forward background processes) bypass env check.
	if strings.HasPrefix(leaf, "_") {
		return true
	}
	return false
}

// Execute runs the root command and returns any error to main.
func Execute() error {
	return rootCmd.Execute()
}

// helpOnNoArgs returns an Args validator that prints the command's full help
// and fails silently when zero arguments are supplied, falling back to
// cobra.ExactArgs(n) otherwise. Relies on rootCmd.SilenceErrors=true so the
// blank error string is not echoed to stderr.
func helpOnNoArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			_ = cmd.Help()
			return fmt.Errorf("")
		}
		return cobra.ExactArgs(n)(cmd, args)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/hf/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "json", "output format: json, table, yaml")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose/debug logging")
	rootCmd.PersistentFlags().BoolVar(&curlMode, "curl", false, "print equivalent curl command for each API request")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "override HyperFleet API URL for this invocation")
	rootCmd.PersistentFlags().StringVar(&apiToken, "api-token", "", "override API bearer token for this invocation")

	_ = rootCmd.RegisterFlagCompletionFunc("output", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "table", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	})
}
