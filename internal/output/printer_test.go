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

func TestWrapHeader_ShortString(t *testing.T) {
	got := output.WrapHeader("ID", 10)
	if len(got) != 1 || got[0] != "ID" {
		t.Errorf("short string: got %v", got)
	}
}

func TestWrapHeader_ExactlyTen(t *testing.T) {
	got := output.WrapHeader("ABCDEFGHIJ", 10) // exactly 10 chars
	if len(got) != 1 || got[0] != "ABCDEFGHIJ" {
		t.Errorf("exactly 10: got %v", got)
	}
}

func TestWrapHeader_UnderscoreSplit(t *testing.T) {
	got := output.WrapHeader("CLUSTER_NAME", 10)
	want := []string{"CLUSTER", "NAME"}
	if !equalSlice(got, want) {
		t.Errorf("underscore split: got %v, want %v", got, want)
	}
}

func TestWrapHeader_MultiSegment(t *testing.T) {
	// "OBSERVED" (8) + "_GENERATION" would be 19 > 10 → split
	// "GENERATION" is exactly 10 → fits on its own line
	got := output.WrapHeader("OBSERVED_GENERATION", 10)
	want := []string{"OBSERVED", "GENERATION"}
	if !equalSlice(got, want) {
		t.Errorf("multi-segment: got %v, want %v", got, want)
	}
}

func TestWrapHeader_LongSingleToken(t *testing.T) {
	// "PROVISIONING" is 12 chars, no underscore → hard-break at 10
	got := output.WrapHeader("PROVISIONING", 10)
	want := []string{"PROVISIONI", "NG"}
	if !equalSlice(got, want) {
		t.Errorf("long single token: got %v, want %v", got, want)
	}
}

func TestWrapHeader_GreedyPack(t *testing.T) {
	// "A" (1) + "_B" = 3 ≤ 10; "A_B" (3) + "_C" = 5 ≤ 10 → all fit in one line
	got := output.WrapHeader("A_B_C", 10)
	if len(got) != 1 || got[0] != "A_B_C" {
		t.Errorf("greedy pack: got %v", got)
	}

	// "LONGTOKEN" (9) + "_X" = 11 > 10 → wrap; X alone on next line
	got2 := output.WrapHeader("LONGTOKEN_X", 10)
	want2 := []string{"LONGTOKEN", "X"}
	if !equalSlice(got2, want2) {
		t.Errorf("greedy pack split: got %v, want %v", got2, want2)
	}
}

func TestPrintTable_LongHeader(t *testing.T) {
	p, buf := newPrinter(t, "table")
	headers := []string{"id", "cluster_name", "status"}
	rows := [][]string{
		{"1", "my-cluster", "running"},
	}
	if err := p.PrintTable(headers, rows); err != nil {
		t.Fatalf("PrintTable: %v", err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	// Should have 3 lines: header line 1, header line 2, data row
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (2 header + 1 data), got %d:\n%s", len(lines), out)
	}
	// First header line contains "CLUSTER" (first wrap segment)
	if !strings.Contains(lines[0], "CLUSTER") {
		t.Errorf("line 0 missing CLUSTER: %q", lines[0])
	}
	// Second header line contains "NAME" (second wrap segment)
	if !strings.Contains(lines[1], "NAME") {
		t.Errorf("line 1 missing NAME: %q", lines[1])
	}
	// Data row is present
	if !strings.Contains(lines[2], "my-cluster") {
		t.Errorf("line 2 missing data: %q", lines[2])
	}
}

func TestPrintTable_ShortHeaders(t *testing.T) {
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
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	// Short headers: exactly 1 header line + 2 data rows = 3 lines total
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (1 header + 2 data), got %d:\n%s", len(lines), out)
	}
	if !strings.Contains(lines[0], "ID") || !strings.Contains(lines[0], "NAME") || !strings.Contains(lines[0], "STATUS") {
		t.Errorf("header line missing columns: %q", lines[0])
	}
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestPrintTable_ANSIAlignment(t *testing.T) {
	// Inject raw ANSI codes into a data cell to test display-width-aware alignment.
	// "name" col: header "NAME" (4 chars), data "my-cluster" (10 chars) → col width 10.
	// "status" col: header "STATUS" (6 chars), data "\033[32mOK\033[0m" (2 display chars) → col width 6.
	// STATUS header should start at byte offset 10+2=12 in the header line.
	// ANSI prefix in the data line should also start at byte offset 12.
	buf := &bytes.Buffer{}
	p := output.NewPrinter("table", true, buf, nil)
	headers := []string{"name", "status"}
	rows := [][]string{
		{"my-cluster", "\033[32mOK\033[0m"},
	}
	if err := p.PrintTable(headers, rows); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d:\n%s", len(lines), out)
	}
	// STATUS header starts at position 12 (10-char name col + 2-space separator).
	statusPos := strings.Index(lines[0], "STATUS")
	if statusPos != 12 {
		t.Errorf("STATUS header at byte %d, want 12:\n%q", statusPos, lines[0])
	}
	// ANSI escape in the data row also starts at byte 12.
	ansiPos := strings.Index(lines[1], "\033")
	if ansiPos != 12 {
		t.Errorf("ANSI prefix in data at byte %d, want 12:\n%q", ansiPos, lines[1])
	}
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
