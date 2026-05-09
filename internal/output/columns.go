package output

import "sort"

// DynamicColumns builds a column list for condition-based resource tables.
// Ordering: fixed columns → Available (if present) → other conditions (alpha) → Reconciled (if present).
func DynamicColumns(fixed []string, conditionTypes []string) []string {
	seen := map[string]bool{}
	for _, ct := range conditionTypes {
		seen[ct] = true
	}

	cols := make([]string, 0, len(fixed)+len(conditionTypes))
	cols = append(cols, fixed...)

	if seen["Available"] {
		cols = append(cols, "Available")
	}

	// Collect and sort all conditions except Available and Reconciled
	var others []string
	for ct := range seen {
		if ct != "Available" && ct != "Reconciled" {
			others = append(others, ct)
		}
	}
	sort.Strings(others)
	cols = append(cols, others...)

	if seen["Reconciled"] {
		cols = append(cols, "Reconciled")
	}

	return cols
}
