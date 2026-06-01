package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/selector"
)

type mockPreviewSel struct {
	idx int
	err error
}

func (m mockPreviewSel) SelectWithPreview(_ []selector.Item, _ func(int) string, _ string) (int, error) {
	return m.idx, m.err
}

// ---- hf env (interactive picker) ----

func TestEnvPickerNoArgs_NoEnvironments(t *testing.T) {
	dir := t.TempDir()
	out, err := runCmd(t, dir, "env")
	if err != nil {
		t.Fatalf("hf env with no envs: %v", err)
	}
	if !strings.Contains(out, "No environments found") {
		t.Errorf("expected 'No environments found' message: %q", out)
	}
	if !strings.Contains(out, "hf env create") {
		t.Errorf("expected hint to run 'hf env create': %q", out)
	}
}

func TestEnvPickerNoArgs_Abort(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")

	old := envSel
	envSel = mockPreviewSel{idx: -1}
	defer func() { envSel = old }()

	_, err := runCmd(t, dir, "env")
	if err != nil {
		t.Errorf("abort (idx=-1) should return nil error, got: %v", err)
	}
}

func TestEnvPickerNoArgs_ActivatesAndShowsConfig(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	makeEnv(t, dir, "prod", "http://prod:8000")

	old := envSel
	envSel = mockPreviewSel{idx: 0} // selects first env alphabetically ("dev")
	defer func() { envSel = old }()

	out, err := runCmd(t, dir, "env")
	if err != nil {
		t.Fatalf("hf env picker: %v", err)
	}
	if !strings.Contains(out, "Activated") {
		t.Errorf("expected activation message in output: %q", out)
	}
	if !strings.Contains(out, "hyperfleet:") {
		t.Errorf("expected config sections after activation: %q", out)
	}
}

// ---- hf env list ----

func TestEnvList(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	makeEnv(t, dir, "prod", "http://prod:8000")
	setActiveEnv(t, dir, "dev")

	out, err := runCmd(t, dir, "env", "list")
	if err != nil {
		t.Fatalf("hf env list: %v", err)
	}
	if !strings.Contains(out, "dev") {
		t.Errorf("missing dev in list: %q", out)
	}
	if !strings.Contains(out, "prod") {
		t.Errorf("missing prod in list: %q", out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("active marker missing: %q", out)
	}
}

func TestEnvList_Empty(t *testing.T) {
	dir := t.TempDir()
	out, err := runCmd(t, dir, "env", "list")
	if err != nil {
		t.Fatalf("hf env list (empty): %v", err)
	}
	if !strings.Contains(out, "hf env create") {
		t.Errorf("empty list should reference create command: %q", out)
	}
}

func TestEnvList_Alias(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "env", "ls")
	if err != nil {
		t.Fatalf("env ls alias: %v", err)
	}
}

// ---- hf env create ----

func TestEnvCreate_CreatesFileAndActivates(t *testing.T) {
	dir := t.TempDir()
	out, err := runCmd(t, dir, "env", "create", "newenv")
	if err != nil {
		t.Fatalf("hf env create: %v", err)
	}
	if !strings.Contains(out, "created and activated") {
		t.Errorf("expected success message: %q", out)
	}
}

func TestEnvCreate_Duplicate(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "myenv", "http://x:8000")

	_, err := runCmd(t, dir, "env", "create", "myenv")
	if err == nil {
		t.Fatal("expected error for duplicate env")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- hf env activate ----

func TestEnvActivate(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "prod", "http://prod:8000")

	out, err := runCmd(t, dir, "env", "activate", "prod")
	if err != nil {
		t.Fatalf("hf env activate: %v", err)
	}
	if !strings.Contains(out, "prod") {
		t.Errorf("success message missing env name: %q", out)
	}
}

// ---- hf env delete ----

func TestEnvDelete(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "old", "http://old:8000")

	_, err := runCmd(t, dir, "env", "delete", "old")
	if err != nil {
		t.Fatalf("hf env delete: %v", err)
	}
}

func TestEnvDelete_Alias(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "todelete", "http://x:8000")

	_, err := runCmd(t, dir, "env", "rm", "todelete")
	if err != nil {
		t.Fatalf("env rm alias: %v", err)
	}
}

// ---- hf env show ----

func TestEnvShow(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "staging", "http://staging:8000")

	out, err := runCmd(t, dir, "env", "show", "staging")
	if err != nil {
		t.Fatalf("hf env show: %v", err)
	}
	if !strings.Contains(out, "hyperfleet:") {
		t.Errorf("output missing config sections: %q", out)
	}
	want := filepath.Join(dir, "environments", "staging.yaml")
	if !strings.Contains(out, want) {
		t.Errorf("output should contain env file path %q: got %q", want, out)
	}
	if !strings.Contains(out, "Edit these files") {
		t.Errorf("output should mention editing files: %q", out)
	}
	hyperIdx := strings.Index(out, "hyperfleet:")
	pathIdx := strings.Index(out, "Environment file:")
	if hyperIdx < 0 || pathIdx < 0 || hyperIdx > pathIdx {
		t.Errorf("expected config before file paths: hyperIdx=%d pathIdx=%d", hyperIdx, pathIdx)
	}
}

func TestEnvShow_ActiveEnvNoArg(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	setActiveEnv(t, dir, "dev")

	out, err := runCmd(t, dir, "env", "show", "--no-color")
	if err != nil {
		t.Fatalf("hf env show (active): %v", err)
	}
	if !strings.Contains(out, "http://dev:8000") {
		t.Errorf("expected active env config: %q", out)
	}
	if !strings.Contains(out, "[active]") {
		t.Errorf("expected [active] marker on env file path: %q", out)
	}
}

func TestEnvShow_ActivePrefix(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "myenv", "http://x:8000")
	setActiveEnv(t, dir, "myenv")

	out, err := runCmd(t, dir, "env", "show", "myenv")
	if err != nil {
		t.Fatalf("hf env show: %v", err)
	}
	if !strings.Contains(out, "[active]") {
		t.Errorf("expected [active] marker when showing active env: %q", out)
	}
}

func TestEnvShow_StateVariables(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	stateContent := "active-environment: dev\ncluster-id: cl-123\ncluster-name: my-cluster\n"
	if err := os.WriteFile(filepath.Join(dir, "state.yaml"), []byte(stateContent), 0600); err != nil {
		t.Fatal(err)
	}

	out, err := runCmd(t, dir, "env", "show", "--no-color")
	if err != nil {
		t.Fatalf("hf env show with state: %v", err)
	}
	if !strings.Contains(out, "state:") {
		t.Errorf("output missing state section: %q", out)
	}
	if !strings.Contains(out, "cluster-id: cl-123") {
		t.Errorf("output missing cluster-id from state: %q", out)
	}
	if strings.Contains(out, "foobar-bizz-buzz") {
		t.Errorf("password leaked in output: %q", out)
	}
}

func TestEnvShow_NoActiveEnv_Error(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "env", "show")
	if err == nil {
		t.Fatal("expected error when no active env")
	}
	if !strings.Contains(err.Error(), "no active environment") {
		t.Errorf("error should mention 'no active environment': got %q", err.Error())
	}
}

func TestEnvShow_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "env", "show", "ghost")
	if err == nil {
		t.Fatal("expected error for non-existent env")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- active-env guard ----

func TestActiveEnvGuard_BypassEnvShow(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	_, err := runCmd(t, dir, "env", "show", "dev")
	if err != nil {
		t.Errorf("env show should bypass guard, got: %v", err)
	}
}

func TestActiveEnvGuard_BypassEnvList(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "env", "list")
	if err != nil {
		t.Errorf("env list should bypass guard, got: %v", err)
	}
}

func TestActiveEnvGuard_BypassVersion(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "version")
	if err != nil {
		t.Errorf("version should bypass guard, got: %v", err)
	}
}
