package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ---- mock Querier ----

type mockQuerier struct {
	queryHeaders []string
	queryRows    [][]string
	queryErr     error
	execN        int64
	execErr      error
}

func (m *mockQuerier) Query(_ context.Context, _ *pgxpool.Pool, _ string, _ ...any) ([]string, [][]string, error) {
	return m.queryHeaders, m.queryRows, m.queryErr
}

func (m *mockQuerier) Exec(_ context.Context, _ *pgxpool.Pool, _ string, _ ...any) (int64, error) {
	return m.execN, m.execErr
}

// ---- helpers ----

func setupDBEnv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	makeEnv(t, dir, "test", "http://localhost:8000")
	setActiveEnv(t, dir, "test")
	return dir
}

func resetDBFlags() {
	outputFmt = "table"
	noColor = true
	verbose = false
	dbQueryFile = ""
	dbDeleteAll = false
}

func runDBCmd(t *testing.T, dir string, mock *mockQuerier, stdin string, args ...string) (string, error) {
	t.Helper()
	resetDBFlags()
	prev := dbQuerier
	dbQuerier = mock
	t.Cleanup(func() { dbQuerier = prev })

	if stdin != "" {
		rootCmd.SetIn(strings.NewReader(stdin))
		t.Cleanup(func() { rootCmd.SetIn(nil) })
	}

	return runCmd(t, dir, args...)
}

// ---- hf db query ----

func TestDBQuery_TableOutput(t *testing.T) {
	dir := setupDBEnv(t)
	mock := &mockQuerier{
		queryHeaders: []string{"id", "name"},
		queryRows:    [][]string{{"1", "alpha"}, {"2", "beta"}},
	}
	out, err := runDBCmd(t, dir, mock, "", "db", "query", "SELECT id, name FROM clusters")
	if err != nil {
		t.Fatalf("db query: %v", err)
	}
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") {
		t.Errorf("expected uppercase headers in table output, got: %q", out)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Errorf("expected row data in output, got: %q", out)
	}
}

func TestDBQuery_JSONOutput(t *testing.T) {
	dir := setupDBEnv(t)
	mock := &mockQuerier{
		queryHeaders: []string{"id"},
		queryRows:    [][]string{{"42"}},
	}
	resetDBFlags()
	outputFmt = "json"
	prev := dbQuerier
	dbQuerier = mock
	defer func() { dbQuerier = prev }()

	out, err := runCmd(t, dir, "db", "query", "SELECT id FROM clusters")
	if err != nil {
		t.Fatalf("db query json: %v", err)
	}
	if !strings.Contains(out, `"id"`) || !strings.Contains(out, `"42"`) {
		t.Errorf("expected JSON array with id field, got: %q", out)
	}
}

func TestDBQuery_NoRows(t *testing.T) {
	dir := setupDBEnv(t)
	mock := &mockQuerier{queryHeaders: []string{"id"}, queryRows: nil}
	out, err := runDBCmd(t, dir, mock, "", "db", "query", "SELECT id FROM clusters WHERE 1=0")
	if err != nil {
		t.Fatalf("db query no-rows: %v", err)
	}
	if !strings.Contains(out, "[INFO] No rows returned.") {
		t.Errorf("expected no-rows info message, got: %q", out)
	}
}

func TestDBQuery_FileReadError(t *testing.T) {
	dir := setupDBEnv(t)
	mock := &mockQuerier{}
	resetDBFlags()
	dbQueryFile = "/nonexistent/path/query.sql"
	prev := dbQuerier
	dbQuerier = mock
	defer func() {
		dbQuerier = prev
		dbQueryFile = ""
	}()

	_, err := runCmd(t, dir, "db", "query")
	if err == nil {
		t.Fatal("expected error reading nonexistent file")
	}
	if !strings.Contains(err.Error(), "[ERROR]") {
		t.Errorf("expected [ERROR] prefix, got: %q", err.Error())
	}
}

// ---- hf db exec ----

func TestDBExec_RowsAffected(t *testing.T) {
	dir := setupDBEnv(t)
	mock := &mockQuerier{execN: 3}
	out, err := runDBCmd(t, dir, mock, "", "db", "exec", "DELETE FROM clusters WHERE id='x'")
	if err != nil {
		t.Fatalf("db exec: %v", err)
	}
	if !strings.Contains(out, "Rows affected: 3") {
		t.Errorf("expected 'Rows affected: 3', got: %q", out)
	}
}

// ---- hf db delete ----

func TestDBDelete_UnknownTarget(t *testing.T) {
	dir := setupDBEnv(t)
	// Unknown target fails before any DB interaction.
	_, err := runDBCmd(t, dir, &mockQuerier{}, "", "db", "delete", "foobar")
	if err == nil {
		t.Fatal("expected error for unknown target")
	}
	if !strings.Contains(err.Error(), "[ERROR] Unknown target 'foobar'") || !strings.Contains(err.Error(), "resources") {
		t.Errorf("unexpected error: %q", err.Error())
	}
}

func TestDBDelete_ResourcesTarget(t *testing.T) {
	dir := setupDBEnv(t)
	mock := &mockQuerier{
		queryHeaders: []string{"count"},
		queryRows:    [][]string{{"3"}},
	}
	out, err := runDBCmd(t, dir, mock, "no\n", "db", "delete", "resources")
	if err != nil {
		t.Fatalf("db delete resources: %v", err)
	}
	if !strings.Contains(out, "resources: 3 rows") {
		t.Fatalf("expected resources row count, got: %q", out)
	}
}

func TestDBDelete_AllShowsTable(t *testing.T) {
	dir := setupDBEnv(t)
	mock := &mockQuerier{
		queryHeaders: []string{"count"},
		queryRows:    [][]string{{"1"}},
	}
	out, err := runDBCmd(t, dir, mock, "no\n", "db", "delete", "--all")
	if err != nil {
		t.Fatalf("db delete --all: %v", err)
	}
	for _, want := range []string{"TARGET", "TABLE", "ROWS", "resources", "adapter_statuses", "nodepools", "clusters"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	// resources must appear before clusters in dependency order table
	res := strings.Index(out, "resources")
	cl := strings.Index(out, "clusters")
	if res < 0 || cl < 0 || res > cl {
		t.Fatalf("expected resources before clusters in --all preview:\n%s", out)
	}
}

func TestDBDelete_ConfirmationDenied(t *testing.T) {
	dir := setupDBEnv(t)
	mock := &mockQuerier{
		// COUNT query returns 1 row with value "5"
		queryHeaders: []string{"count"},
		queryRows:    [][]string{{"5"}},
	}
	out, err := runDBCmd(t, dir, mock, "no\n", "db", "delete", "clusters")
	if err != nil {
		t.Fatalf("db delete denied: %v", err)
	}
	if !strings.Contains(out, "Aborted") {
		t.Errorf("expected 'Aborted', got: %q", out)
	}
}

