package tui

import (
	"sort"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

// FilterKind identifies adapter-status filter mode.
type FilterKind int

const (
	FilterByCondition FilterKind = iota
	FilterByAdapter
)

// collectConditionTypes returns unique condition types from adapter statuses.
func collectConditionTypes(statuses []resource.AdapterStatus) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range statuses {
		for _, c := range s.Conditions {
			if !seen[c.Type] {
				seen[c.Type] = true
				out = append(out, c.Type)
			}
		}
	}
	sort.Strings(out)
	return out
}

// collectAdapterNames returns unique adapter names from adapter statuses.
func collectAdapterNames(statuses []resource.AdapterStatus) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range statuses {
		if !seen[s.Adapter] {
			seen[s.Adapter] = true
			out = append(out, s.Adapter)
		}
	}
	sort.Strings(out)
	return out
}

// buildConditionTypeTable builds rows for one condition type across all adapters.
func buildConditionTypeTable(statuses []resource.AdapterStatus, condType string, noColor bool) ([]string, [][]string) {
	headers := []string{"ADAPTER", "STATUS", "LAST_TRANSITION", "REASON", "MESSAGE"}
	var rows [][]string
	for _, s := range statuses {
		for _, c := range s.Conditions {
			if c.Type != condType {
				continue
			}
			rows = append(rows, []string{
				s.Adapter,
				formatConditionStatus(c.Status, noColor),
				tableCell(c.LastTransitionTime),
				tableCell(c.Reason),
				tableCell(c.Message),
			})
			break
		}
	}
	return headers, rows
}

// buildAdapterConditionsTable builds rows for all conditions on one adapter.
func buildAdapterConditionsTable(statuses []resource.AdapterStatus, adapter string, noColor bool) ([]string, [][]string) {
	headers := []string{"TYPE", "STATUS", "LAST_TRANSITION", "REASON", "MESSAGE"}
	var rows [][]string
	for _, s := range statuses {
		if s.Adapter != adapter {
			continue
		}
		for _, c := range s.Conditions {
			rows = append(rows, []string{
				c.Type,
				formatConditionStatus(c.Status, noColor),
				tableCell(c.LastTransitionTime),
				tableCell(c.Reason),
				tableCell(c.Message),
			})
		}
		break
	}
	return headers, rows
}
