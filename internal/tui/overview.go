package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

func tableCell(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func formatConditionStatus(status string, noColor bool) string {
	if status == "" {
		return output.StatusDot(status, noColor)
	}
	return output.StatusDot(status, noColor) + " " + status
}

func buildResourceConditionsTable(conditions []resource.ResourceCondition, noColor bool) ([]string, [][]string) {
	headers := []string{
		"TYPE", "STATUS", "LAST_TRANSITION", "OBS_GEN", "CREATED", "UPDATED", "REASON", "MESSAGE",
	}
	var rows [][]string
	for _, c := range conditions {
		rows = append(rows, []string{
			c.Type,
			formatConditionStatus(c.Status, noColor),
			tableCell(c.LastTransitionTime),
			tableCell(strconv.Itoa(int(c.ObservedGeneration))),
			tableCell(c.CreatedTime),
			tableCell(c.LastUpdatedTime),
			tableCell(c.Reason),
			tableCell(c.Message),
		})
	}
	return headers, rows
}

func buildAllAdapterConditionsTable(statuses []resource.AdapterStatus, noColor bool) ([]string, [][]string) {
	headers := []string{"ADAPTER", "TYPE", "STATUS", "LAST_TRANSITION", "REASON", "MESSAGE"}
	var rows [][]string
	for _, s := range statuses {
		for _, c := range s.Conditions {
			rows = append(rows, []string{
				s.Adapter,
				c.Type,
				formatConditionStatus(c.Status, noColor),
				tableCell(c.LastTransitionTime),
				tableCell(c.Reason),
				tableCell(c.Message),
			})
		}
	}
	return headers, rows
}

func renderFormattedTable(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return "(none)\n"
	}
	ft := output.FormatTable(headers, rows)
	var lines []string
	lines = append(lines, ft.HeaderLines...)
	lines = append(lines, ft.DataLines...)
	return strings.Join(lines, "\n") + "\n"
}

func renderDetailOverview(cluster *resource.Cluster, nodepool *resource.NodePool, statuses []resource.AdapterStatus, noColor bool) string {
	var sb strings.Builder

	if nodepool != nil {
		sb.WriteString(formatNodePoolSummary(*nodepool))
	} else if cluster != nil {
		sb.WriteString(formatClusterSummary(*cluster))
	}

	var conditions []resource.ResourceCondition
	if nodepool != nil {
		conditions = nodepool.Status.Conditions
	} else if cluster != nil {
		conditions = cluster.Status.Conditions
	}

	sb.WriteString("\nResource conditions:\n")
	if len(conditions) == 0 {
		sb.WriteString("(none)\n")
	} else {
		h, r := buildResourceConditionsTable(conditions, noColor)
		sb.WriteString(renderFormattedTable(h, r))
	}

	sb.WriteString("\nAdapter statuses:\n")
	if len(statuses) == 0 {
		sb.WriteString("(none)\n")
	} else {
		h, r := buildAllAdapterConditionsTable(statuses, noColor)
		sb.WriteString(renderFormattedTable(h, r))
	}

	return sb.String()
}

func formatClusterSummary(cl resource.Cluster) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ID:          %s\n", cl.ID))
	sb.WriteString(fmt.Sprintf("Name:        %s\n", cl.Name))
	sb.WriteString(fmt.Sprintf("Generation:  %d\n", cl.Generation))
	if cl.DeletedTime != "" {
		sb.WriteString(fmt.Sprintf("Deleted:     %s\n", cl.DeletedTime))
	}
	return sb.String()
}

func formatNodePoolSummary(np resource.NodePool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ID:          %s\n", np.ID))
	sb.WriteString(fmt.Sprintf("Name:        %s\n", np.Name))
	sb.WriteString(fmt.Sprintf("Generation:  %d\n", np.Generation))
	sb.WriteString(fmt.Sprintf("Cluster:     %s\n", np.OwnerReferences.ID))
	if np.DeletedTime != "" {
		sb.WriteString(fmt.Sprintf("Deleted:     %s\n", np.DeletedTime))
	}
	return sb.String()
}
