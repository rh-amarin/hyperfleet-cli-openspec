package cmd

import (
	"fmt"
	"strings"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

// runStatusFilterUI opens an interactive split-screen filter for adapter statuses.
// Left panel: unique adapter names (prefixed [adapter]) and condition types (prefixed [cond]).
// Right panel: preview updates as the user navigates — showing full adapter status or
// cross-adapter condition view depending on whether the highlighted item is an adapter name
// (lowercase-prefixed) or condition type (uppercase-prefixed).
// Pressing q, Esc, or Enter exits cleanly with code 0.
func runStatusFilterUI(statuses []resource.AdapterStatus, noColor bool) error {
	if len(statuses) == 0 {
		fmt.Println("(no statuses)")
		return nil
	}

	// Collect unique adapter names and condition types, preserving insertion order.
	adapterSeen := map[string]bool{}
	condTypeSeen := map[string]bool{}
	var adapters []string
	var condTypes []string
	for _, s := range statuses {
		if !adapterSeen[s.Adapter] {
			adapterSeen[s.Adapter] = true
			adapters = append(adapters, s.Adapter)
		}
		for _, c := range s.Conditions {
			if !condTypeSeen[c.Type] {
				condTypeSeen[c.Type] = true
				condTypes = append(condTypes, c.Type)
			}
		}
	}

	type filterItem struct {
		display string
		key     string
		isAdapt bool
	}

	var items []filterItem
	for _, a := range adapters {
		items = append(items, filterItem{
			display: fmt.Sprintf("[adapter]  %s", a),
			key:     a,
			isAdapt: true,
		})
	}
	for _, ct := range condTypes {
		items = append(items, filterItem{
			display: fmt.Sprintf("[cond]     %s", ct),
			key:     ct,
			isAdapt: false,
		})
	}

	header := "Adapter statuses  ·  Type to filter  ·  ↑↓ navigate  ·  q / Esc to quit"

	opts := []fuzzyfinder.Option{
		fuzzyfinder.WithPreviewWindow(func(i, _, _ int) string {
			if i < 0 || i >= len(items) {
				return ""
			}
			it := items[i]
			if it.isAdapt {
				return statusAdapterPreview(it.key, statuses, noColor)
			}
			return statusConditionPreview(it.key, statuses, noColor)
		}),
		fuzzyfinder.WithHeader(header),
	}

	_, err := fuzzyfinder.Find(items, func(i int) string { return items[i].display }, opts...)
	if err == fuzzyfinder.ErrAbort || err == nil {
		return nil
	}
	return err
}

// statusAdapterPreview renders the full status report for a single adapter.
func statusAdapterPreview(adapterName string, statuses []resource.AdapterStatus, noColor bool) string {
	var sb strings.Builder
	for _, s := range statuses {
		if s.Adapter != adapterName {
			continue
		}
		sb.WriteString(fmt.Sprintf("Adapter:    %s\n", s.Adapter))
		sb.WriteString(fmt.Sprintf("Generation: %d\n", s.ObservedGeneration))
		if s.LastReportTime != "" {
			sb.WriteString(fmt.Sprintf("Reported:   %s\n", s.LastReportTime))
		}
		sb.WriteString("\nConditions:\n")
		for _, c := range s.Conditions {
			dot := output.StatusDot(c.Status, noColor)
			line := fmt.Sprintf("  %s  %-20s", dot, c.Type)
			if c.Reason != "" {
				line += "  " + c.Reason
			}
			sb.WriteString(line + "\n")
		}
		break
	}
	return sb.String()
}

// statusConditionPreview renders a specific condition type across all adapters.
// Each line is formatted as: <adapter>  <dot>  <status>  <reason>
func statusConditionPreview(condType string, statuses []resource.AdapterStatus, noColor bool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Condition: %s\n\n", condType))
	found := false
	for _, s := range statuses {
		for _, c := range s.Conditions {
			if c.Type != condType {
				continue
			}
			found = true
			dot := output.StatusDot(c.Status, noColor)
			line := fmt.Sprintf("%s  %s  %s", s.Adapter, dot, c.Status)
			if c.Reason != "" {
				line += "  " + c.Reason
			}
			sb.WriteString(line + "\n")
		}
	}
	if !found {
		sb.WriteString("(no results)\n")
	}
	return sb.String()
}
