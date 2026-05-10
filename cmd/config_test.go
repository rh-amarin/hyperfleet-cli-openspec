// Package cmd tests for the config command tree and active-env guard.
// White-box tests: package cmd so we can inspect package-level state.
package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCmd executes the root command with the given args, setting HF_CONFIG_DIR to
// a temp dir. Returns stdout content and the command error.
func runCmd(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	t.Setenv("HF_CONFIG_DIR", dir)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	// Reset cobra state after each run.
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)

	return buf.String(), err
}

// makeEnv creates a named environment file in dir/environments/<name>.yaml.
func makeEnv(t *testing.T, dir, name, apiURL string) {
	t.Helper()
	envDir := filepath.Join(dir, "environments")
	if err := os.MkdirAll(envDir, 0700); err != nil {
		t.Fatal(err)
	}
	content := "hyperfleet:\n  api-url: " + apiURL + "\n"
	if err := os.WriteFile(filepath.Join(envDir, name+".yaml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

// setActiveEnv writes state.yaml in dir to set the active environment.
func setActiveEnv(t *testing.T, dir, name string) {
	t.Helper()
	statePath := filepath.Join(dir, "state.yaml")
	if err := os.WriteFile(statePath, []byte("active-environment: "+name+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
}

// ---- active-env guard tests ----

func TestActiveEnvGuard_BlocksConfigShow(t *testing.T) {
	dir := t.TempDir()
	// No active env set — guard should fire.
	_, err := runCmd(t, dir, "config", "show")
	if err == nil {
		t.Fatal("expected error when no active env for 'config show'")
	}
	if !strings.Contains(err.Error(), "No active environment") {
		t.Errorf("error message: got %q, want 'No active environment'", err.Error())
	}
}

func TestActiveEnvGuard_BlocksConfigSet(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "config", "set", "hyperfleet", "api-url", "http://x:8000")
	if err == nil {
		t.Fatal("expected error when no active env for 'config set'")
	}
	if !strings.Contains(err.Error(), "No active environment") {
		t.Errorf("error message: got %q", err.Error())
	}
}

func TestActiveEnvGuard_BlocksConfigGet(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "config", "get", "hyperfleet.api-url")
	if err == nil {
		t.Fatal("expected error when no active env for 'config get'")
	}
	if !strings.Contains(err.Error(), "No active environment") {
		t.Errorf("error message: got %q", err.Error())
	}
}

func TestActiveEnvGuard_BypassEnvList(t *testing.T) {
	dir := t.TempDir()
	// No active env — env list should still work.
	_, err := runCmd(t, dir, "config", "env", "list")
	if err != nil {
		t.Errorf("config env list should bypass guard, got: %v", err)
	}
}

func TestActiveEnvGuard_BypassEnvActivate(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	_, err := runCmd(t, dir, "config", "env", "activate", "dev")
	if err != nil {
		t.Errorf("config env activate should bypass guard, got: %v", err)
	}
}

func TestActiveEnvGuard_BypassVersion(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "version")
	if err != nil {
		t.Errorf("version should bypass guard, got: %v", err)
	}
}

// ---- config show ----

func TestConfigShow(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "staging", "http://staging:8000")
	setActiveEnv(t, dir, "staging")

	out, err := runCmd(t, dir, "config", "show")
	if err != nil {
		t.Fatalf("config show: %v", err)
	}
	if !strings.Contains(out, "state:") {
		t.Errorf("output missing state section: %q", out)
	}
	if !strings.Contains(out, "active-environment: staging") {
		t.Errorf("output missing active-environment in state section: %q", out)
	}
	if !strings.Contains(out, "hyperfleet:") {
		t.Errorf("output missing hyperfleet section: %q", out)
	}
	// Secrets must be masked
	if strings.Contains(out, "foobar-bizz-buzz") {
		t.Errorf("password leaked in output: %q", out)
	}
	if !strings.Contains(out, "<set>") && !strings.Contains(out, "<not set>") {
		t.Errorf("secrets should be masked as <set>/<not set>: %q", out)
	}
}

func TestConfigShow_StateVariables(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	stateContent := "active-environment: dev\ncluster-id: cl-123\ncluster-name: my-cluster\n"
	if err := os.WriteFile(filepath.Join(dir, "state.yaml"), []byte(stateContent), 0600); err != nil {
		t.Fatal(err)
	}

	out, err := runCmd(t, dir, "config", "show")
	if err != nil {
		t.Fatalf("config show with state: %v", err)
	}
	if !strings.Contains(out, "cluster-id: cl-123") {
		t.Errorf("output missing cluster-id from state: %q", out)
	}
	if !strings.Contains(out, "cluster-name: my-cluster") {
		t.Errorf("output missing cluster-name from state: %q", out)
	}
}

// ---- config get ----

func TestConfigGet_Found(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	setActiveEnv(t, dir, "dev")

	out, err := runCmd(t, dir, "config", "get", "hyperfleet.api-url")
	if err != nil {
		t.Fatalf("config get: %v", err)
	}
	if !strings.Contains(out, "http://localhost:8000") && !strings.Contains(out, "http://dev:8000") {
		t.Errorf("unexpected api-url: %q", out)
	}
}

func TestConfigGet_NotFound(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	setActiveEnv(t, dir, "dev")

	_, err := runCmd(t, dir, "config", "get", "hyperfleet.nonexistent-key")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message: got %q", err.Error())
	}
}

func TestConfigGet_StateKey(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	stateContent := "active-environment: dev\ncluster-id: cl-456\n"
	if err := os.WriteFile(filepath.Join(dir, "state.yaml"), []byte(stateContent), 0600); err != nil {
		t.Fatal(err)
	}

	out, err := runCmd(t, dir, "config", "get", "cluster-id")
	if err != nil {
		t.Fatalf("config get cluster-id: %v", err)
	}
	if !strings.Contains(out, "cl-456") {
		t.Errorf("unexpected cluster-id: %q", out)
	}
}

func TestConfigGet_NoArgs_ShowsHelp(t *testing.T) {
	dir := t.TempDir()
	out, err := runCmd(t, dir, "config", "get")
	if err == nil {
		t.Fatal("expected non-nil error when no args given")
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected help output when no args given, got stdout: %q", out)
	}
}

// ---- config set ----

func TestConfigSet_Valid(t *testing.T) {
	dir := t.TempDir()
	// Use an env profile that does NOT set api-version so we can verify config.yaml write.
	makeEnv(t, dir, "dev", "http://dev:8000")
	setActiveEnv(t, dir, "dev")

	_, err := runCmd(t, dir, "config", "set", "hyperfleet", "api-version", "v2")
	if err != nil {
		t.Fatalf("config set: %v", err)
	}

	// Read back — profile doesn't override api-version, so config.yaml value wins.
	out, err := runCmd(t, dir, "config", "get", "hyperfleet.api-version")
	if err != nil {
		t.Fatalf("config get after set: %v", err)
	}
	if !strings.Contains(out, "v2") {
		t.Errorf("value not persisted: got %q", out)
	}
}

func TestConfigSet_InvalidSection(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	setActiveEnv(t, dir, "dev")

	_, err := runCmd(t, dir, "config", "set", "badSection", "someKey", "value")
	if err == nil {
		t.Fatal("expected error for unknown section")
	}
	if !strings.Contains(err.Error(), "Unknown config section") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- config env list ----

func TestConfigEnvList(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	makeEnv(t, dir, "prod", "http://prod:8000")
	setActiveEnv(t, dir, "dev")

	out, err := runCmd(t, dir, "config", "env", "list")
	if err != nil {
		t.Fatalf("config env list: %v", err)
	}
	if !strings.Contains(out, "dev") {
		t.Errorf("missing dev env: %q", out)
	}
	if !strings.Contains(out, "prod") {
		t.Errorf("missing prod env: %q", out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("active marker missing: %q", out)
	}
	if !strings.Contains(out, "NAME") {
		t.Errorf("table header missing: %q", out)
	}
}

func TestConfigEnvList_Alias(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "config", "env", "ls")
	if err != nil {
		t.Fatalf("config env ls alias: %v", err)
	}
}

// ---- config env create ----

func TestConfigEnvCreate(t *testing.T) {
	dir := t.TempDir()
	// env create must work without active env
	out, err := runCmd(t, dir, "config", "env", "create", "myenv", "--api-url", "http://myenv:8000")
	if err != nil {
		t.Fatalf("config env create: %v", err)
	}
	if !strings.Contains(out, "created") {
		t.Errorf("expected success message: %q", out)
	}

	// File must exist
	profPath := filepath.Join(dir, "environments", "myenv.yaml")
	if _, err := os.Stat(profPath); os.IsNotExist(err) {
		t.Error("environment file was not created")
	}
}

func TestConfigEnvCreate_Duplicate(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "myenv", "http://x:8000")

	_, err := runCmd(t, dir, "config", "env", "create", "myenv")
	if err == nil {
		t.Fatal("expected error for duplicate env")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- config env activate ----

func TestConfigEnvActivate(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "prod", "http://prod:8000")

	out, err := runCmd(t, dir, "config", "env", "activate", "prod")
	if err != nil {
		t.Fatalf("config env activate: %v", err)
	}
	if !strings.Contains(out, "prod") {
		t.Errorf("success message missing env name: %q", out)
	}

	// Verify state.yaml
	stateRaw, _ := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if !strings.Contains(string(stateRaw), "prod") {
		t.Errorf("state.yaml does not contain 'prod': %q", string(stateRaw))
	}
}

func TestConfigEnvActivate_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "config", "env", "activate", "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent env")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- config env delete ----

func TestConfigEnvDelete(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "old", "http://old:8000")

	_, err := runCmd(t, dir, "config", "env", "delete", "old")
	if err != nil {
		t.Fatalf("config env delete: %v", err)
	}

	profPath := filepath.Join(dir, "environments", "old.yaml")
	if _, err := os.Stat(profPath); !os.IsNotExist(err) {
		t.Error("environment file should have been deleted")
	}
}

func TestConfigEnvDelete_ClearsActiveEnv(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "active-one", "http://x:8000")
	setActiveEnv(t, dir, "active-one")

	_, err := runCmd(t, dir, "config", "env", "delete", "active-one")
	if err != nil {
		t.Fatalf("config env delete: %v", err)
	}

	stateRaw, _ := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if strings.Contains(string(stateRaw), "active-one") {
		t.Errorf("active-environment should be cleared: %q", string(stateRaw))
	}
}

func TestConfigEnvDelete_Alias(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "todelete", "http://x:8000")

	_, err := runCmd(t, dir, "config", "env", "rm", "todelete")
	if err != nil {
		t.Fatalf("config env rm alias: %v", err)
	}
}

func TestConfigEnvDelete_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "config", "env", "delete", "ghost")
	if err == nil {
		t.Fatal("expected error for non-existent env")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- config env show ----

func TestConfigEnvShow(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "staging", "http://staging:8000")

	out, err := runCmd(t, dir, "config", "env", "show", "staging")
	if err != nil {
		t.Fatalf("config env show: %v", err)
	}
	if !strings.Contains(out, "environments/staging.yaml") {
		t.Errorf("output should contain env file path: %q", out)
	}
	if !strings.Contains(out, "staging") {
		t.Errorf("output should contain env name or api-url: %q", out)
	}
}

func TestConfigEnvShow_ActivePrefix(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "myenv", "http://x:8000")
	setActiveEnv(t, dir, "myenv")

	out, err := runCmd(t, dir, "config", "env", "show", "myenv")
	if err != nil {
		t.Fatalf("config env show: %v", err)
	}
	if !strings.Contains(out, "[active]") {
		t.Errorf("expected [active] prefix when showing active env: %q", out)
	}
}

func TestConfigEnvShow_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "config", "env", "show", "ghost")
	if err == nil {
		t.Fatal("expected error for non-existent env")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- config doctor ----

func TestConfigDoctor_Reachable(t *testing.T) {
	// Start a test HTTP server that responds 200 to /healthz.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	dir := t.TempDir()
	// Write config pointing to test server.
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	cfgContent := "hyperfleet:\n  api-url: " + ts.URL + "\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(cfgContent), 0600); err != nil {
		t.Fatal(err)
	}

	out, err := runCmd(t, dir, "config", "doctor")
	if err != nil {
		t.Fatalf("config doctor: %v", err)
	}
	if !strings.Contains(out, "[OK]") {
		t.Errorf("expected [OK] in output: %q", out)
	}
}

func TestConfigDoctor_Unreachable(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "hyperfleet:\n  api-url: http://127.0.0.1:19999\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(cfgContent), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := runCmd(t, dir, "config", "doctor")
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
	if !strings.Contains(err.Error(), "Cannot reach") {
		t.Errorf("error message: got %q", err.Error())
	}
}
