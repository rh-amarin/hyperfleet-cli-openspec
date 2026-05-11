// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"strconv"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/spf13/cobra"
)

// resourcesCmd displays a combined overview of all clusters and their node pools.
var resourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Show a combined overview of all clusters and their node pools",
	RunE:  runResources,
}

// tableCmd is an alias for resourcesCmd.
var tableCmd = &cobra.Command{
	Use:   "table",
	Short: "Alias for hf resources — combined cluster+nodepool table",
	RunE:  runResources,
}

// watch flags for resources / table commands.
var (
	resourcesWatchMode bool
	resourcesWatchSecs int
)

// clusterEntry holds a fetched cluster with its adapter statuses and nodepools.
type clusterEntry struct {
	cluster         resource.Cluster
	adapterStatuses []resource.AdapterStatus
	nodepools       []resource.NodePool
	npStatuses      map[string][]resource.AdapterStatus
}

func runResources(cmd *cobra.Command, _ []string) error {
	// Default to table when --output was not explicitly provided.
	effectiveFmt := outputFmt
	if !rootCmd.PersistentFlags().Changed("output") {
		effectiveFmt = "table"
	}

	if resourcesWatchMode && effectiveFmt == "table" {
		ctx, cancel := watchContext(context.Background())
		defer cancel()
		return runWatch(ctx, cmd.OutOrStdout(), resourcesWatchSecs, func(tick int) error {
			return fetchAndRenderResources(cmd, effectiveFmt, tick, resourcesWatchSecs)
		})
	}

	return fetchAndRenderResources(cmd, effectiveFmt, 0, 0)
}

func fetchAndRenderResources(cmd *cobra.Command, effectiveFmt string, tick, frequencySecs int) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	client := newAPIClient(s)
	p := output.NewPrinter(effectiveFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

	clusterList, err := api.Get[resource.ListResponse[resource.Cluster]](context.Background(), client, "clusters")
	if err != nil {
		return handleAPIError(p, err)
	}

	if effectiveFmt != "table" {
		return p.Print(clusterList)
	}

	// Fetch adapter statuses and nodepools for each cluster.
	entries := make([]clusterEntry, 0, len(clusterList.Items))
	for _, cl := range clusterList.Items {
		var adStatuses resource.ListResponse[resource.AdapterStatus]
		adStatuses, _ = api.Get[resource.ListResponse[resource.AdapterStatus]](
			context.Background(), client, "clusters/"+cl.ID+"/statuses",
		)

		var npList resource.ListResponse[resource.NodePool]
		npList, _ = api.Get[resource.ListResponse[resource.NodePool]](
			context.Background(), client, "clusters/"+cl.ID+"/nodepools",
		)

		npStatuses := make(map[string][]resource.AdapterStatus)
		for _, np := range npList.Items {
			var npAdStatus resource.ListResponse[resource.AdapterStatus]
			npAdStatus, _ = api.Get[resource.ListResponse[resource.AdapterStatus]](
				context.Background(), client,
				"clusters/"+cl.ID+"/nodepools/"+np.ID+"/statuses",
			)
			npStatuses[np.ID] = npAdStatus.Items
		}

		entries = append(entries, clusterEntry{
			cluster:         cl,
			adapterStatuses: adStatuses.Items,
			nodepools:       npList.Items,
			npStatuses:      npStatuses,
		})
	}

	condCols := collectConditionCols(entries)
	adapterCols := collectAdapterCols(entries)

	headers := make([]string, 0, 3+len(condCols)+len(adapterCols))
	headers = append(headers, "ID", "NAME", "GEN")
	headers = append(headers, condCols...)
	headers = append(headers, adapterCols...)

	rows := make([][]string, 0)
	for _, e := range entries {
		rows = append(rows, buildClusterRow(e.cluster, e.adapterStatuses, condCols, adapterCols, tick, frequencySecs))
		for _, np := range e.nodepools {
			rows = append(rows, buildNodePoolRow(np, e.npStatuses[np.ID], condCols, adapterCols, tick, frequencySecs))
		}
	}

	return p.PrintTable(headers, rows)
}

// collectConditionCols returns unique condition types (excluding *Successful) in insertion order.
func collectConditionCols(entries []clusterEntry) []string {
	seen := map[string]bool{}
	var cols []string
	add := func(t string) {
		if !strings.HasSuffix(t, "Successful") && !seen[t] {
			seen[t] = true
			cols = append(cols, t)
		}
	}
	for _, e := range entries {
		for _, c := range e.cluster.Status.Conditions {
			add(c.Type)
		}
		for _, np := range e.nodepools {
			for _, c := range np.Status.Conditions {
				add(c.Type)
			}
		}
	}
	return cols
}

// collectAdapterCols returns unique adapter names across all adapter statuses in insertion order.
func collectAdapterCols(entries []clusterEntry) []string {
	seen := map[string]bool{}
	var cols []string
	add := func(name string) {
		if !seen[name] {
			seen[name] = true
			cols = append(cols, name)
		}
	}
	for _, e := range entries {
		for _, as := range e.adapterStatuses {
			add(as.Adapter)
		}
		for _, statuses := range e.npStatuses {
			for _, as := range statuses {
				add(as.Adapter)
			}
		}
	}
	return cols
}

// adapterDot returns the status cell for a named adapter column.
// When the adapter has a recent last_report_time (within 2×frequencySecs), a spinner frame is prepended.
func adapterDot(statuses []resource.AdapterStatus, adName, condKey string, tick, frequencySecs int) string {
	for _, as := range statuses {
		if as.Adapter == adName {
			for _, c := range as.Conditions {
				if c.Type == condKey {
					cell := output.StatusDot(c.Status, noColor) + " " + strconv.Itoa(int(as.ObservedGeneration))
					if output.IsActive(as.LastReportTime, frequencySecs) {
						cell = output.SpinnerFrame(tick) + " " + cell
					}
					return cell
				}
			}
			return "-"
		}
	}
	return "-"
}

func buildClusterRow(cl resource.Cluster, statuses []resource.AdapterStatus, condCols, adapterCols []string, tick, frequencySecs int) []string {
	isDeleted := cl.DeletedTime != ""
	gen := strconv.Itoa(int(cl.Generation))
	if isDeleted {
		gen += " ❌"
	}
	condKey := "Available"
	if isDeleted {
		condKey = "Finalized"
	}

	row := []string{cl.ID, cl.Name, gen}
	for _, col := range condCols {
		val := "-"
		for _, c := range cl.Status.Conditions {
			if c.Type == col {
				val = output.StatusDot(c.Status, noColor) + " " + strconv.Itoa(int(c.ObservedGeneration))
				break
			}
		}
		row = append(row, val)
	}
	for _, adName := range adapterCols {
		row = append(row, adapterDot(statuses, adName, condKey, tick, frequencySecs))
	}
	return row
}

func buildNodePoolRow(np resource.NodePool, statuses []resource.AdapterStatus, condCols, adapterCols []string, tick, frequencySecs int) []string {
	isDeleted := np.DeletedTime != ""
	gen := strconv.Itoa(int(np.Generation))
	if isDeleted {
		gen += " ❌"
	}
	condKey := "Available"
	if isDeleted {
		condKey = "Finalized"
	}

	row := []string{"  " + np.ID, "  " + np.Name, gen}
	for _, col := range condCols {
		val := "-"
		for _, c := range np.Status.Conditions {
			if c.Type == col {
				val = output.StatusDot(c.Status, noColor) + " " + strconv.Itoa(int(c.ObservedGeneration))
				break
			}
		}
		row = append(row, val)
	}
	for _, adName := range adapterCols {
		row = append(row, adapterDot(statuses, adName, condKey, tick, frequencySecs))
	}
	return row
}

func init() {
	rootCmd.AddCommand(resourcesCmd)
	rootCmd.AddCommand(tableCmd)

	for _, cmd := range []*cobra.Command{resourcesCmd, tableCmd} {
		cmd.Flags().BoolVar(&resourcesWatchMode, "watch", false, "continuously refresh the table")
		cmd.Flags().IntVarP(&resourcesWatchSecs, "seconds", "s", 5, "refresh interval in seconds (used with --watch)")
	}
}
