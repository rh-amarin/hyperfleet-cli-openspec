// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/spf13/cobra"
)

// legacyResourcesCmd is a top-level alias for hf rs (combined operational overview).
var legacyResourcesCmd = &cobra.Command{
	Use:        "resources",
	Short:      "Alias for hf rs — combined operational overview",
	RunE:       runResourceOverview,
	Deprecated: "use hf rs instead",
}

// legacyTableCmd is a top-level alias for hf rs (adapter-rich cluster overview).
var legacyTableCmd = &cobra.Command{
	Use:        "table",
	Short:      "Alias for hf rs — combined operational overview",
	RunE:       runResourceOverview,
	Deprecated: "use hf rs instead",
}

// resourcesCmd displays a combined overview via the legacy resources command path.
var resourcesCmd = legacyResourcesCmd

// tableCmd is the per-type list/table subcommand name inside hf rs <type>.

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
		if curlMode {
			return fetchAndRenderResources(cmd, effectiveFmt, 0, 0)
		}
		ctx, cancel := watchContext(context.Background())
		defer cancel()

		var (
			cachedEntries []clusterEntry
			cachedAdCols  []string
			nextRefresh   time.Time
		)
		return runWatchFast(ctx, cmd.OutOrStdout(), resourcesWatchSecs, func(tick int, refresh bool) error {
			if refresh || cachedEntries == nil {
				entries, adCols, err := fetchResourceEntries(cmd)
				if err != nil {
					p := output.NewPrinter("table", noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
					return handleAPIError(p, err)
				}
				cachedEntries, cachedAdCols = entries, adCols
				nextRefresh = time.Now().Add(time.Duration(resourcesWatchSecs) * time.Second)
			}
			return renderResourcesTable(cmd, cachedEntries, cachedAdCols, tick, resourcesWatchSecs, secsUntil(nextRefresh))
		})
	}

	return fetchAndRenderResources(cmd, effectiveFmt, 0, 0)
}

// fetchResourceEntries fetches all clusters with their adapter statuses and nodepools
// for table rendering. Returns raw errors without printing them.
func fetchResourceEntries(cmd *cobra.Command) (entries []clusterEntry, adCols []string, err error) {
	s, err := loadConfig()
	if err != nil {
		return
	}
	client := newAPIClient(s)

	clusterList, err := api.Get[resource.ListResponse[resource.Cluster]](context.Background(), client, "clusters")
	if err != nil {
		return
	}

	entries = make([]clusterEntry, 0, len(clusterList.Items))
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

	adCols = collectAdapterCols(entries)
	return
}

// renderResourcesTable renders the cluster+nodepool table from pre-fetched data.
// In watch mode (frequencySecs > 0) a countdown line is printed above the table.
func renderResourcesTable(cmd *cobra.Command, entries []clusterEntry, adapterCols []string, tick, frequencySecs, secsLeft int) error {
	if frequencySecs > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "↻ %ds  %s\n", secsLeft, output.SpinnerFrame(tick))
		printWatchFooter(cmd, frequencySecs)
	}
	p := output.NewPrinter("table", noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

	headers := make([]string, 0, 3+len(fixedCondCols)+len(adapterCols))
	headers = append(headers, "ID", "NAME", "GEN")
	headers = append(headers, fixedCondCols...)
	headers = append(headers, adapterCols...)

	rows := make([][]string, 0)
	for _, e := range entries {
		rows = append(rows, buildClusterRow(e.cluster, e.adapterStatuses, adapterCols, tick, frequencySecs))
		for i, np := range e.nodepools {
			idPrefix := "  "
			namePrefix := "  "
			if treeOverview {
				idPrefix = treeLinePrefix(1, i == len(e.nodepools)-1)
				namePrefix = idPrefix
			}
			rows = append(rows, buildNodePoolRowPrefixed(np, e.npStatuses[np.ID], adapterCols, tick, frequencySecs, idPrefix, namePrefix))
		}
	}

	return p.PrintTable(headers, rows)
}

var treeOverview bool

func buildNodePoolRowPrefixed(np resource.NodePool, statuses []resource.AdapterStatus, adapterCols []string, tick, frequencySecs int, idPrefix, namePrefix string) []string {
	row := buildNodePoolRow(np, statuses, adapterCols, tick, frequencySecs)
	row[0] = idPrefix + np.ID
	row[1] = namePrefix + np.Name
	return row
}

func fetchAndRenderReconciledOverview(cmd *cobra.Command, effectiveFmt string, tick, frequencySecs int) error {
	treeOverview = effectiveFmt == "table"
	defer func() { treeOverview = false }()
	return fetchAndRenderResources(cmd, effectiveFmt, tick, frequencySecs)
}

func fetchAndRenderResources(cmd *cobra.Command, effectiveFmt string, tick, frequencySecs int) error {
	if effectiveFmt != "table" {
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
		return p.Print(clusterList)
	}

	entries, adCols, err := fetchResourceEntries(cmd)
	if err != nil {
		p := output.NewPrinter(effectiveFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		return handleAPIError(p, err)
	}
	return renderResourcesTable(cmd, entries, adCols, tick, frequencySecs, 0)
}

// collectAdapterCols returns unique adapter names sorted by the earliest
// created_time seen for each adapter across all entries. RFC 3339 strings
// compare lexicographically in chronological order, so no parsing is needed.
// Adapters with the same (or absent) timestamp are sorted alphabetically.
func collectAdapterCols(entries []clusterEntry) []string {
	earliest := map[string]string{} // adapter name → earliest created_time

	consider := func(as resource.AdapterStatus) {
		t := as.CreatedTime
		prev, seen := earliest[as.Adapter]
		if !seen || (t != "" && (prev == "" || t < prev)) {
			earliest[as.Adapter] = t
		}
	}

	for _, e := range entries {
		for _, as := range e.adapterStatuses {
			consider(as)
		}
		for _, statuses := range e.npStatuses {
			for _, as := range statuses {
				consider(as)
			}
		}
	}

	names := make([]string, 0, len(earliest))
	for name := range earliest {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		ti, tj := earliest[names[i]], earliest[names[j]]
		if ti != tj {
			return ti < tj
		}
		return names[i] < names[j]
	})
	return names
}

// fixedCondCols are the two aggregated condition columns always shown before adapter columns.
var fixedCondCols = []string{"Reconciled", "LastKnownReconciled"}

// condDot returns the status cell for a fixed condition column.
func condDot(conditions []resource.ResourceCondition, condType string) string {
	for _, c := range conditions {
		if c.Type == condType {
			return output.StatusDot(c.Status, noColor) + " " + strconv.Itoa(int(c.ObservedGeneration))
		}
	}
	return "-"
}

// adapterDot returns the status cell for a named adapter column.
// Every cell reserves a 2-char activity prefix so watch and one-shot tables align.
// Active adapters show the spinner frame; inactive ones show two spaces.
func adapterDot(statuses []resource.AdapterStatus, adName, condKey string, tick, frequencySecs int) string {
	const emptyCell = "  -"
	for _, as := range statuses {
		if as.Adapter == adName {
			for _, c := range as.Conditions {
				if c.Type == condKey {
					cell := output.StatusDot(c.Status, noColor) + " " + strconv.Itoa(int(as.ObservedGeneration))
					return output.AdapterActivityPrefix(as.LastReportTime, tick, frequencySecs) + cell
				}
			}
			return emptyCell
		}
	}
	return emptyCell
}

func buildClusterRow(cl resource.Cluster, statuses []resource.AdapterStatus, adapterCols []string, tick, frequencySecs int) []string {
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
	for _, ct := range fixedCondCols {
		row = append(row, condDot(cl.Status.Conditions, ct))
	}
	for _, adName := range adapterCols {
		row = append(row, adapterDot(statuses, adName, condKey, tick, frequencySecs))
	}
	return row
}

func buildNodePoolRow(np resource.NodePool, statuses []resource.AdapterStatus, adapterCols []string, tick, frequencySecs int) []string {
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
	for _, ct := range fixedCondCols {
		row = append(row, condDot(np.Status.Conditions, ct))
	}
	for _, adName := range adapterCols {
		row = append(row, adapterDot(statuses, adName, condKey, tick, frequencySecs))
	}
	return row
}

// secsUntil returns the ceiling of the number of seconds until t, clamped to 0.
func secsUntil(t time.Time) int {
	d := time.Until(t)
	if d <= 0 {
		return 0
	}
	return int((d + time.Second - 1) / time.Second)
}

// runResources remains for tests that invoke the legacy resources command table path.
