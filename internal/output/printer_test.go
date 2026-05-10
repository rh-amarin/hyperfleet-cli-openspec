package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/output"
)

func newPrinter(t *testing.T, format string) (*output.Printer, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	p := output.NewPrinter(format, true, buf, nil)
	return p, buf
}

func TestJSONOutput(t *testing.T) {
	p, buf := newPrinter(t, "json")
	data := map[string]any{"id": "1", "name": "test", "count": 42}
	if err := p.Print(data); err != nil {
		t.Fatalf("Print: %v", err)
	}

	// Should be valid JSON with 2-space indent and trailing newline
	out := buf.String()
	if !strings.HasSuffix(out, "\n") {
		t.Error("JSON output missing trailing newline")
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &parsed); err != nil {
		t.Errorf("invalid JSON output: %v\nOutput: %s", err, out)
	}
	if parsed["id"] != "1" {
		t.Errorf("id: got %v", parsed["id"])
	}
}

func TestTableOutput(t *testing.T) {
	p, buf := newPrinter(t, "table")
	headers := []string{"id", "name", "status"}
	rows := [][]string{
		{"1", "cluster-a", "running"},
		{"2", "cluster-b", "stopped"},
	}
	if err := p.PrintTable(headers, rows); err != nil {
		t.Fatalf("PrintTable: %v", err)
	}

	out := buf.String()
	// Headers should be uppercase
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") || !strings.Contains(out, "STATUS") {
		t.Errorf("uppercase headers missing:\n%s", out)
	}
	// Rows should be present
	if !strings.Contains(out, "cluster-a") || !strings.Contains(out, "cluster-b") {
		t.Errorf("rows missing:\n%s", out)
	}
}

// TestTableHeaderWrapping verifies that headers longer than 10 characters are
// split across multiple lines and that short headers remain on a single line.
func TestTableHeaderWrapping(t *testing.T) {
	p, buf := newPrinter(t, "table")
	// "last transition" uppercases to "LAST TRANSITION" (15 chars) → must wrap.
	// "type" and "reason" are short → single-line.
	headers := []string{"type", "last transition", "reason"}
	rows := [][]string{{"Degraded", "2024-01-01T00:00:00Z", "SomeReason"}}

	if err := p.PrintTable(headers, rows); err != nil {
		t.Fatalf("PrintTable: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")

	// "LAST TRANSITION" splits into "LAST" (line 0) and "TRANSITION" (line 1),
	// so there should be at least 3 output lines (2 header + 1 data).
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d:\n%s", len(lines), out)
	}
	if !strings.Contains(lines[0], "LAST") {
		t.Errorf("first header line should contain LAST:\n%s", lines[0])
	}
	if !strings.Contains(lines[1], "TRANSITION") {
		t.Errorf("second header line should contain TRANSITION:\n%s", lines[1])
	}
	// Data row must still be present
	if !strings.Contains(out, "SomeReason") {
		t.Errorf("data row missing:\n%s", out)
	}
}

// TestTableHeaderNoWrapShort verifies headers ≤ 10 chars produce a single header line.
func TestTableHeaderNoWrapShort(t *testing.T) {
	p, buf := newPrinter(t, "table")
	headers := []string{"id", "name", "status"} // all ≤ 10 chars
	rows := [][]string{{"1", "foo", "ok"}}

	if err := p.PrintTable(headers, rows); err != nil {
		t.Fatalf("PrintTable: %v", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	// Exactly 1 header line + 1 data line = 2 lines
	if len(lines) != 2 {
		t.Errorf("expected 2 lines for short headers, got %d:\n%s", len(lines), buf.String())
	}
}

func TestYAMLOutput(t *testing.T) {
	p, buf := newPrinter(t, "yaml")
	data := map[string]any{"cluster_id": "abc", "node_count": 3}
	if err := p.Print(data); err != nil {
		t.Fatalf("Print: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "cluster_id") || !strings.Contains(out, "abc") {
		t.Errorf("YAML output missing fields:\n%s", out)
	}
}

func TestStatusDotTrue(t *testing.T) {
	dot := output.StatusDot("True", true)
	if dot != "True" {
		t.Errorf("True no-color: got %q", dot)
	}
}

func TestStatusDotFalse(t *testing.T) {
	dot := output.StatusDot("False", true)
	if dot != "False" {
		t.Errorf("False no-color: got %q", dot)
	}
}

func TestStatusDotUnknown(t *testing.T) {
	dot := output.StatusDot("Unknown", true)
	if dot != "Unknown" {
		t.Errorf("Unknown no-color: got %q", dot)
	}
}

func TestStatusDotAbsent(t *testing.T) {
	dot := output.StatusDot("", true)
	if dot != "-" {
		t.Errorf("absent: got %q", dot)
	}
}

func TestStatusDotColorTrue(t *testing.T) {
	dot := output.StatusDot("True", false)
	if !strings.Contains(dot, "●") {
		t.Errorf("True with color: got %q", dot)
	}
	if !strings.Contains(dot, "\033[32m") {
		t.Errorf("True missing green ANSI: got %q", dot)
	}
}

func TestStatusDotColorFalse(t *testing.T) {
	dot := output.StatusDot("False", false)
	if !strings.Contains(dot, "\033[31m") {
		t.Errorf("False missing red ANSI: got %q", dot)
	}
}

func TestStatusDotColorUnknown(t *testing.T) {
	dot := output.StatusDot("Unknown", false)
	if !strings.Contains(dot, "\033[33m") {
		t.Errorf("Unknown missing yellow ANSI: got %q", dot)
	}
}

func TestDynamicColumnsOrdering(t *testing.T) {
	fixed := []string{"ID", "NAME"}
	conditions := []string{"Reconciled", "Available", "Degraded", "Bootstrapping"}

	cols := output.DynamicColumns(fixed, conditions)

	// Fixed first
	if cols[0] != "ID" || cols[1] != "NAME" {
		t.Errorf("fixed cols wrong: %v", cols[:2])
	}
	// Available after fixed
	if cols[2] != "Available" {
		t.Errorf("Available not at index 2: %v", cols)
	}
	// Reconciled last
	if cols[len(cols)-1] != "Reconciled" {
		t.Errorf("Reconciled not last: %v", cols)
	}
	// Alpha others between Available and Reconciled
	// Bootstrapping < Degraded alphabetically
	bootstrapIdx := indexOf(cols, "Bootstrapping")
	degradedIdx := indexOf(cols, "Degraded")
	reconcileIdx := indexOf(cols, "Reconciled")
	availableIdx := indexOf(cols, "Available")
	if bootstrapIdx > degradedIdx {
		t.Errorf("Bootstrapping should come before Degraded: %v", cols)
	}
	if bootstrapIdx < availableIdx {
		t.Errorf("Bootstrapping should come after Available: %v", cols)
	}
	if degradedIdx > reconcileIdx {
		t.Errorf("Degraded should come before Reconciled: %v", cols)
	}
}

func TestDynamicColumnsNoConditions(t *testing.T) {
	fixed := []string{"ID", "NAME"}
	cols := output.DynamicColumns(fixed, nil)
	if len(cols) != 2 || cols[0] != "ID" || cols[1] != "NAME" {
		t.Errorf("empty conditions: got %v", cols)
	}
}

func TestDynamicColumnsOnlyFixed(t *testing.T) {
	cols := output.DynamicColumns([]string{"ID"}, []string{})
	if len(cols) != 1 || cols[0] != "ID" {
		t.Errorf("only fixed: got %v", cols)
	}
}

func indexOf(slice []string, s string) int {
	for i, v := range slice {
		if v == s {
			return i
		}
	}
	return -1
}

func TestWarnInfoError(t *testing.T) {
	errBuf := &bytes.Buffer{}
	outBuf := &bytes.Buffer{}
	p := output.NewPrinter("json", true, outBuf, errBuf)

	p.Warn("something is wrong")
	p.Info("just FYI")
	p.Error("fatal problem")

	errOut := errBuf.String()
	if !strings.Contains(errOut, "[WARN] something is wrong") {
		t.Errorf("Warn: got %q", errOut)
	}
	if !strings.Contains(errOut, "[INFO] just FYI") {
		t.Errorf("Info: got %q", errOut)
	}
	if !strings.Contains(errOut, "[ERROR] fatal problem") {
		t.Errorf("Error: got %q", errOut)
	}
	// stdout should be empty
	if outBuf.Len() != 0 {
		t.Errorf("stdout should be empty: got %q", outBuf.String())
	}
}
