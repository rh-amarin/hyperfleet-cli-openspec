package cmd

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/spf13/cobra"
)

func effectiveClusterListSearch() string {
	if genericListSearch != "" {
		return genericListSearch
	}
	return clusterListSearch
}

func effectiveNodepoolListSearch() string {
	if genericListSearch != "" {
		return genericListSearch
	}
	return nodepoolListSearch
}

func effectiveListWatch() (bool, int) {
	if genericListWatch {
		return true, genericListWatchSecs
	}
	if clusterListWatch || nodepoolListWatch {
		if clusterListWatch {
			return true, clusterListWatchSecs
		}
		return true, nodepoolListWatchSecs
	}
	return false, 5
}

// fetchAndRenderReconciledClusterList renders clusters with condition and adapter columns.
func fetchAndRenderReconciledClusterList(cmd *cobra.Command, tick, frequencySecs int) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	client := newAPIClient(s)
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

	path := "clusters"
	if search := effectiveClusterListSearch(); search != "" {
		path = "clusters?search=" + url.QueryEscape(search)
	}
	list, err := api.Get[resource.ListResponse[resource.Cluster]](context.Background(), client, path)
	if err != nil {
		return handleAPIError(p, err)
	}

	if outputFmt != "table" {
		return p.Print(list)
	}

	if frequencySecs > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "↻ %ds  %s\n", 0, output.SpinnerFrame(tick))
	}

	entries := make([]clusterEntry, 0, len(list.Items))
	for _, cl := range list.Items {
		var adStatuses resource.ListResponse[resource.AdapterStatus]
		adStatuses, _ = api.Get[resource.ListResponse[resource.AdapterStatus]](
			context.Background(), client, "clusters/"+cl.ID+"/statuses",
		)
		entries = append(entries, clusterEntry{
			cluster:         cl,
			adapterStatuses: adStatuses.Items,
		})
	}
	adapterCols := collectAdapterCols(entries)

	headers := make([]string, 0, 3+len(fixedCondCols)+len(adapterCols))
	headers = append(headers, "ID", "NAME", "GEN")
	headers = append(headers, fixedCondCols...)
	headers = append(headers, adapterCols...)

	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, buildClusterRow(e.cluster, e.adapterStatuses, adapterCols, tick, frequencySecs))
	}
	return p.PrintTable(headers, rows)
}

// fetchAndRenderReconciledNodepoolList renders nodepools with TYPE, REPLICAS, conditions, and adapters.
func fetchAndRenderReconciledNodepoolList(cmd *cobra.Command, tick, frequencySecs int) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	clusterID, err := s.ResourceID("clusters", "")
	if err != nil {
		return err
	}
	client := newAPIClient(s)
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

	npPath := npBase(clusterID)
	if search := effectiveNodepoolListSearch(); search != "" {
		npPath = npBase(clusterID) + "?search=" + url.QueryEscape(search)
	}
	list, err := api.Get[resource.ListResponse[resource.NodePool]](context.Background(), client, npPath)
	if err != nil {
		return handleAPIError(p, err)
	}

	if outputFmt != "table" {
		return p.Print(list)
	}

	if frequencySecs > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "↻ %ds  %s\n", 0, output.SpinnerFrame(tick))
	}

	npStatuses := make(map[string][]resource.AdapterStatus)
	var entries []clusterEntry
	for _, np := range list.Items {
		var adStatuses resource.ListResponse[resource.AdapterStatus]
		adStatuses, _ = api.Get[resource.ListResponse[resource.AdapterStatus]](
			context.Background(), client,
			npBase(clusterID)+"/"+np.ID+"/statuses",
		)
		npStatuses[np.ID] = adStatuses.Items
		entries = append(entries, clusterEntry{
			nodepools:  []resource.NodePool{np},
			npStatuses: npStatuses,
		})
	}
	adapterCols := collectAdapterCols(entries)

	headers := make([]string, 0, 5+len(fixedCondCols)+len(adapterCols))
	headers = append(headers, "ID", "NAME", "TYPE", "GEN", "REPLICAS")
	headers = append(headers, fixedCondCols...)
	headers = append(headers, adapterCols...)

	rows := make([][]string, 0, len(list.Items))
	for _, np := range list.Items {
		rows = append(rows, buildReconciledNodePoolRow(np, npStatuses[np.ID], adapterCols, tick, frequencySecs))
	}
	return p.PrintTable(headers, rows)
}

func buildReconciledNodePoolRow(np resource.NodePool, statuses []resource.AdapterStatus, adapterCols []string, tick, frequencySecs int) []string {
	isDeleted := np.DeletedTime != ""
	gen := strconv.Itoa(int(np.Generation))
	if isDeleted {
		gen += " ❌"
	}
	condKey := "Available"
	if isDeleted {
		condKey = "Finalized"
	}

	row := []string{np.ID, np.Name, nodepoolType(np), gen, nodepoolReplicas(np)}
	for _, ct := range fixedCondCols {
		row = append(row, condDot(np.Status.Conditions, ct))
	}
	for _, adName := range adapterCols {
		row = append(row, adapterDot(statuses, adName, condKey, tick, frequencySecs))
	}
	return row
}
