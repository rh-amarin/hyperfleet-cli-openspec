// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"github.com/spf13/cobra"
)

// completionCmd generates shell completion scripts.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for hf.

To load completions in the current session:

  Bash:        source <(hf completion bash)
  Zsh:         source <(hf completion zsh)
  Fish:        hf completion fish | source
  PowerShell:  hf completion powershell | Out-String | Invoke-Expression

To install completions permanently:

  Bash:   hf completion bash > /etc/bash_completion.d/hf
  Zsh:    hf completion zsh > "${fpath[1]}/_hf"
  Fish:   hf completion fish > ~/.config/fish/completions/hf.fish`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletionV2(cmd.OutOrStdout(), true)
		case "zsh":
			return rootCmd.GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
