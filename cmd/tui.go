package cmd

import (
	"fmt"

	"github.com/rh-amarin/hyperfleet-cli/internal/tui"
	"github.com/spf13/cobra"
)

var tuiRefreshSecs int

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive terminal dashboard for HyperFleet clusters",
	Long: `Launch a k9s-style terminal UI for monitoring clusters and nodepools.

Main view mirrors hf table --watch in a single full-width panel (k9s-style).
Use arrow keys to select rows, Enter to describe a resource, d to delete (with
confirmation), s/a to filter adapter statuses, c to patch, p to toggle
port-forwards, and Esc to return.`,
	RunE: runTUI,
}

func runTUI(cmd *cobra.Command, _ []string) error {
	if tuiRefreshSecs < 1 {
		return fmt.Errorf("[ERROR] refresh interval must be at least 1 second")
	}

	fetcher := func() ([]tui.ClusterEntry, error) {
		entries, _, err := fetchResourceEntries(cmd)
		if err != nil {
			return nil, err
		}
		return toTUIEntries(entries), nil
	}

	return tui.Run(tui.Options{
		Fetcher:            fetcher,
		Patcher:            tuiPatchResource,
		Deleter:            tuiDeleteResource,
		ContextProvider:    tuiContextSnapshot,
		PortForwardToggler: tuiTogglePortForwards,
		RefreshSecs:        tuiRefreshSecs,
		NoColor:            noColor,
	})
}

func toTUIEntries(entries []clusterEntry) []tui.ClusterEntry {
	out := make([]tui.ClusterEntry, len(entries))
	for i, e := range entries {
		out[i] = tui.ClusterEntry{
			Cluster:         e.cluster,
			AdapterStatuses: e.adapterStatuses,
			Nodepools:       e.nodepools,
			NPStatuses:      e.npStatuses,
		}
	}
	return out
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	tuiCmd.Flags().IntVarP(&tuiRefreshSecs, "seconds", "s", 5, "refresh interval in seconds")
}
