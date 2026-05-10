// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"os"

	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/repos"
	"github.com/spf13/cobra"
)

// reposCmd shows a GitHub + Quay status overview for HyperFleet repositories.
var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "Show an overview of HyperFleet GitHub repositories",
	Long: `Show a status overview for all HyperFleet GitHub repositories.

Displays the latest commit, open PR, and Quay container image tag for each
tracked repository in the openshift-hyperfleet organization.

Set GITHUB_TOKEN (or registry.github-token in config) for authenticated
GitHub API access (5000 req/h vs 60 req/h unauthenticated).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}

		token := s.Get("registry", "github-token")
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}
		quayNS := s.Get("registry", "quay-namespace")

		client := repos.New(token, "", quayNS)
		statuses := client.FetchAll(context.Background())

		p := output.NewPrinter(outputFmt, noColor, nil, nil)
		if outputFmt == "table" {
			headers := []string{"repository", "commit", "pr url", "pr branch", "quay tag", "quay aliases"}
			rows := make([][]string, len(statuses))
			for i, rs := range statuses {
				rows[i] = []string{rs.Repository, rs.Commit, rs.PRURL, rs.PRBranch, rs.QuayTag, rs.QuayAliases}
			}
			return p.PrintTable(headers, rows)
		}
		return p.Print(statuses)
	},
}

func init() {
	rootCmd.AddCommand(reposCmd)
}
