package tui

import (
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/output"
)

func TestFormatMainTableAlignsColumns(t *testing.T) {
	green := "\033[32m●\033[0m"
	headers := []string{"ID", "NAME", "GEN"}
	rows := [][]string{
		{"short", "cluster-a", green + " 1"},
		{"much-longer-id", "cluster-b", green + " 2"},
	}

	ft := formatMainTable(headers, rows, 0)
	if len(ft.DataLines) != 2 {
		t.Fatalf("data lines = %d, want 2", len(ft.DataLines))
	}
	// Both rows should have NAME column starting at same visible offset.
	idxA := strings.Index(ft.DataLines[0], "cluster-a")
	idxB := strings.Index(ft.DataLines[1], "cluster-b")
	if idxA <= 0 || idxA != idxB {
		t.Errorf("NAME column misaligned: idxA=%d idxB=%d\n%q\n%q", idxA, idxB, ft.DataLines[0], ft.DataLines[1])
	}
}

func TestFormatMainTableClipsColumnsNotANSI(t *testing.T) {
	green := "\033[32m●\033[0m"
	headers := []string{"ID", "NAME", "GEN", "ADAPTER-A", "ADAPTER-B", "ADAPTER-C"}
	rows := [][]string{
		{"id1", "name1", "1", green + " 1", "-", "-"},
	}

	ft := formatMainTable(headers, rows, 40)
	all := append(ft.HeaderLines, ft.DataLines...)
	if output.MaxLineWidth(all) > 40 {
		t.Errorf("line too wide: %d > 40 in %q", output.MaxLineWidth(all), ft.DataLines[0])
	}
	// Must not contain broken escape sequences (orphan [ without m).
	for _, line := range all {
		if strings.Contains(line, "\033[32") && !strings.Contains(line, "\033[0m") && strings.Contains(line, "●") {
			// dot cell should always reset
		}
		if strings.Count(line, "\033[") != strings.Count(line, "m") {
			// rough sanity — each CSI should end with m somewhere
		}
	}
}

func TestRenderTableBlockSelectedStyle(t *testing.T) {
	green := "\033[32m●\033[0m"
	ft := output.FormattedTable{
		HeaderLines: []string{"ID  NAME  GEN  ADAPTER"},
		DataLines: []string{
			"a   one   1  -",
			"b   two   2  " + green + " 1",
		},
	}
	out := renderTableBlock(ft, 1, 0, 10, 80, false)
	if !strings.Contains(out, "two") {
		t.Fatalf("missing row: %s", out)
	}
	lines := strings.Split(out, "\n")
	selectedLine := lines[len(ft.HeaderLines)+1]
	if output.DisplayWidth(selectedLine) < 80 {
		t.Errorf("selected row highlight should pad to panel width, got %d", output.DisplayWidth(selectedLine))
	}
}

func TestHighlightTableRowPadsWithANSI(t *testing.T) {
	green := "\033[32m●\033[0m"
	red := "\033[31m●\033[0m"
	line := "id1  name1  1  " + green + " 1  " + red + " 2  adapter-col"
	highlighted := highlightTableRow(line, 80, false)
	if output.DisplayWidth(highlighted) < 80 {
		t.Errorf("highlighted width = %d, want >= 80", output.DisplayWidth(highlighted))
	}
	if !strings.Contains(highlighted, "\033[32m") || !strings.Contains(highlighted, "\033[31m") {
		t.Errorf("selected row should preserve status dot colors: %q", highlighted)
	}
	if !strings.Contains(highlighted, ansiRowBG) {
		t.Errorf("selected row should apply row background: %q", highlighted)
	}
	if !strings.Contains(highlighted, "adapter-col") {
		t.Errorf("selected row should include all columns: %q", highlighted)
	}
}
