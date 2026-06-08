package tui

import (
	"sort"
	"strconv"

	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

var fixedCondCols = []string{"Reconciled", "LastKnownReconciled"}

// BuildSnapshot constructs table headers, display rows, and row metadata from entries.
func BuildSnapshot(entries []ClusterEntry, tick, frequencySecs int, noColor bool) Snapshot {
	adapterCols := collectAdapterCols(entries)
	headers := make([]string, 0, 3+len(fixedCondCols)+len(adapterCols))
	headers = append(headers, "ID", "NAME", "GEN")
	headers = append(headers, fixedCondCols...)
	headers = append(headers, adapterCols...)

	var rows [][]string
	var meta []RowMeta

	for ci, e := range entries {
		rows = append(rows, buildClusterRow(e.Cluster, e.AdapterStatuses, adapterCols, tick, frequencySecs, noColor))
		meta = append(meta, RowMeta{Kind: RowCluster, ClusterIdx: ci, NodePoolIdx: -1})
		for ni, np := range e.Nodepools {
			rows = append(rows, buildNodePoolRow(np, e.NPStatuses[np.ID], adapterCols, tick, frequencySecs, noColor))
			meta = append(meta, RowMeta{Kind: RowNodePool, ClusterIdx: ci, NodePoolIdx: ni})
		}
	}

	return Snapshot{
		Headers: headers,
		Rows:    rows,
		Meta:    meta,
		Entries: entries,
	}
}

func collectAdapterCols(entries []ClusterEntry) []string {
	earliest := map[string]string{}
	consider := func(as resource.AdapterStatus) {
		t := as.CreatedTime
		prev, seen := earliest[as.Adapter]
		if !seen || (t != "" && (prev == "" || t < prev)) {
			earliest[as.Adapter] = t
		}
	}
	for _, e := range entries {
		for _, as := range e.AdapterStatuses {
			consider(as)
		}
		for _, statuses := range e.NPStatuses {
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

func condDot(conditions []resource.ResourceCondition, condType string, noColor bool) string {
	for _, c := range conditions {
		if c.Type == condType {
			return output.StatusDot(c.Status, noColor) + " " + strconv.Itoa(int(c.ObservedGeneration))
		}
	}
	return "-"
}

func adapterDot(statuses []resource.AdapterStatus, adName, condKey string, tick, frequencySecs int, noColor bool) string {
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

func buildClusterRow(cl resource.Cluster, statuses []resource.AdapterStatus, adapterCols []string, tick, frequencySecs int, noColor bool) []string {
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
		row = append(row, condDot(cl.Status.Conditions, ct, noColor))
	}
	for _, adName := range adapterCols {
		row = append(row, adapterDot(statuses, adName, condKey, tick, frequencySecs, noColor))
	}
	return row
}

func buildNodePoolRow(np resource.NodePool, statuses []resource.AdapterStatus, adapterCols []string, tick, frequencySecs int, noColor bool) []string {
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
		row = append(row, condDot(np.Status.Conditions, ct, noColor))
	}
	for _, adName := range adapterCols {
		row = append(row, adapterDot(statuses, adName, condKey, tick, frequencySecs, noColor))
	}
	return row
}
