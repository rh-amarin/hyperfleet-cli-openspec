// Package cmd tests for the config command tree and active-env guard.
package cmd

import (
	"bytes"
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

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)

	return buf.String(), err
}

// makeEnv creates a named environment file in dir/environments/<name>.yaml.
func makeEnv(t *testing.T, dir, name, apiURL string) {
	t.Helper()
	makeEnvRaw(t, dir, name, "hyperfleet:\n  api-url: "+apiURL+"\n")
}

// makeEnvRaw creates a named environment file with arbitrary content.
func makeEnvRaw(t *testing.T, dir, name, content string) {
	t.Helper()
	envDir := filepath.Join(dir, "environments")
	if err := os.MkdirAll(envDir, 0700); err != nil {
		t.Fatal(err)
	}
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

// ---- hf config (no subcommand) ----

func TestConfigNoArgs_ShowsActiveConfig(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	setActiveEnv(t, dir, "dev")

	out, err := runCmd(t, dir, "config")
	if err != nil {
		t.Fatalf("hf config: %v", err)
	}
	if !strings.Contains(out, "hyperfleet:") {
		t.Errorf("expected config output, got: %q", out)
	}
}

func TestConfigNoArgs_NoActiveEnv_Errors(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "config")
	if err == nil {
		t.Fatal("expected error when no active env")
	}
	if !strings.Contains(err.Error(), "no active environment") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- active-env guard ----

func TestActiveEnvGuard_BlocksConfigShow(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "config", "show")
	if err == nil {
		t.Fatal("expected error when no active env for 'config show'")
	}
	if !strings.Contains(err.Error(), "no active environment") {
		t.Errorf("error message: got %q", err.Error())
	}
}

func TestActiveEnvGuard_BlocksConfigSet(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "config", "set", "hyperfleet.api-url", "http://x:8000")
	if err == nil {
		t.Fatal("expected error when no active env for 'config set'")
	}
	if !strings.Contains(err.Error(), "no active environment") {
		t.Errorf("error message: got %q", err.Error())
	}
}

func TestActiveEnvGuard_BypassEnvList(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "env", "list")
	if err != nil {
		t.Errorf("env list should bypass guard, got: %v", err)
	}
}

func TestActiveEnvGuard_BypassEnvCreate(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "env", "create", "myenv")
	if err != nil {
		t.Errorf("env create should bypass guard, got: %v", err)
	}
}

func TestActiveEnvGuard_BypassEnvActivate(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	_, err := runCmd(t, dir, "env", "activate", "dev")
	if err != nil {
		t.Errorf("env activate should bypass guard, got: %v", err)
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
	if !strings.Contains(out, "active-environment: staging") {
		t.Errorf("output missing active-environment: %q", out)
	}
	if !strings.Contains(out, "hyperfleet:") {
		t.Errorf("output missing hyperfleet section: %q", out)
	}
	if strings.Contains(out, "foobar-bizz-buzz") {
		t.Errorf("password leaked in output: %q", out)
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

func TestConfigShow_EnvFilePath(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "prod", "http://prod:8000")
	setActiveEnv(t, dir, "prod")

	out, err := runCmd(t, dir, "config", "show")
	if err != nil {
		t.Fatalf("config show: %v", err)
	}
	want := filepath.Join(dir, "environments", "prod.yaml")
	if !strings.Contains(out, want) {
		t.Errorf("output missing env file path %q: got %q", want, out)
	}
}

func TestConfigShow_NoActiveEnv_Error(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "config", "show")
	if err == nil {
		t.Fatal("expected error when no active env")
	}
	if !strings.Contains(err.Error(), "no active environment") {
		t.Errorf("error should mention 'no active environment': got %q", err.Error())
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
	if !strings.Contains(out, "http://dev:8000") {
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

func TestConfigSet_DottedNotation(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "hyperfleet:\n  api-url: http://dev:8000\n  api-version: v1\n")
	setActiveEnv(t, dir, "dev")

	_, err := runCmd(t, dir, "config", "set", "hyperfleet.api-version", "v2")
	if err != nil {
		t.Fatalf("config set: %v", err)
	}

	out, err := runCmd(t, dir, "config", "get", "hyperfleet.api-version")
	if err != nil {
		t.Fatalf("config get after set: %v", err)
	}
	if !strings.Contains(out, "v2") {
		t.Errorf("value not persisted: got %q", out)
	}
}

func TestConfigSet_WritesToEnvFile(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "hyperfleet:\n  api-url: http://dev:8000\n")
	setActiveEnv(t, dir, "dev")

	_, err := runCmd(t, dir, "config", "set", "hyperfleet.api-url", "http://new:9000")
	if err != nil {
		t.Fatalf("config set: %v", err)
	}

	raw, _ := os.ReadFile(filepath.Join(dir, "environments", "dev.yaml"))
	if !strings.Contains(string(raw), "http://new:9000") {
		t.Errorf("env file not updated: %s", raw)
	}
}

func TestConfigSet_InvalidKeyFormat(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "hyperfleet:\n  api-url: http://dev:8000\n")
	setActiveEnv(t, dir, "dev")

	_, err := runCmd(t, dir, "config", "set", "nodotkey", "value")
	if err == nil {
		t.Fatal("expected error for key without dot")
	}
	if !strings.Contains(err.Error(), "section.key format") {
		t.Errorf("error message: got %q", err.Error())
	}
}

func TestConfigSet_InvalidSection(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "hyperfleet:\n  api-url: http://dev:8000\n")
	setActiveEnv(t, dir, "dev")

	_, err := runCmd(t, dir, "config", "set", "badSection.someKey", "value")
	if err == nil {
		t.Fatal("expected error for unknown section")
	}
	if !strings.Contains(err.Error(), "unknown config section") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- env list ----

func TestConfigEnvList(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	makeEnv(t, dir, "prod", "http://prod:8000")
	setActiveEnv(t, dir, "dev")

	out, err := runCmd(t, dir, "env", "list")
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
}

func TestConfigEnvList_Empty(t *testing.T) {
	dir := t.TempDir()
	out, err := runCmd(t, dir, "env", "list")
	if err != nil {
		t.Fatalf("config env list (empty): %v", err)
	}
	if !strings.Contains(out, "hf env create") {
		t.Errorf("empty list should reference create command: %q", out)
	}
}

func TestConfigEnvList_Alias(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "env", "ls")
	if err != nil {
		t.Fatalf("config env ls alias: %v", err)
	}
}

// ---- config env create ----

func TestConfigEnvCreate_CreatesFileAndActivates(t *testing.T) {
	dir := t.TempDir()
	out, err := runCmd(t, dir, "env", "create", "myenv")
	if err != nil {
		t.Fatalf("config env create: %v", err)
	}
	if !strings.Contains(out, "created and activated") {
		t.Errorf("expected success message: %q", out)
	}

	profPath := filepath.Join(dir, "environments", "myenv.yaml")
	if _, err := os.Stat(profPath); os.IsNotExist(err) {
		t.Error("environment file was not created")
	}

	// Verify activated
	stateRaw, _ := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if !strings.Contains(string(stateRaw), "myenv") {
		t.Errorf("state.yaml should reference new env: %s", stateRaw)
	}
}

func TestConfigEnvCreate_PrintsFilePath(t *testing.T) {
	dir := t.TempDir()
	out, err := runCmd(t, dir, "env", "create", "myenv")
	if err != nil {
		t.Fatalf("config env create: %v", err)
	}
	if !strings.Contains(out, "environments/myenv.yaml") {
		t.Errorf("output should contain file path: %q", out)
	}
}

func TestConfigEnvCreate_SeedsFromTemplate(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "env", "create", "myenv")
	if err != nil {
		t.Fatalf("config env create: %v", err)
	}

	raw, _ := os.ReadFile(filepath.Join(dir, "environments", "myenv.yaml"))
	if !strings.Contains(string(raw), "http://localhost:8000") {
		t.Errorf("env file should contain template defaults: %s", raw)
	}
}

func TestConfigEnvCreate_Duplicate(t *testing.T) {
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

func TestConfigEnvCreate_NoArg_ShowsHelp(t *testing.T) {
	dir := t.TempDir()
	out, err := runCmd(t, dir, "env", "create")
	if err == nil {
		t.Fatal("expected error when no name given")
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected help output: %q", out)
	}
}

// ---- config env activate ----

func TestConfigEnvActivate(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "prod", "http://prod:8000")

	out, err := runCmd(t, dir, "env", "activate", "prod")
	if err != nil {
		t.Fatalf("config env activate: %v", err)
	}
	if !strings.Contains(out, "prod") {
		t.Errorf("success message missing env name: %q", out)
	}

	stateRaw, _ := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if !strings.Contains(string(stateRaw), "prod") {
		t.Errorf("state.yaml does not contain 'prod': %q", string(stateRaw))
	}
}

func TestConfigEnvActivate_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "env", "activate", "nonexistent")
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

	_, err := runCmd(t, dir, "env", "delete", "old")
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

	_, err := runCmd(t, dir, "env", "delete", "active-one")
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

	_, err := runCmd(t, dir, "env", "rm", "todelete")
	if err != nil {
		t.Fatalf("config env rm alias: %v", err)
	}
}

func TestConfigEnvDelete_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "env", "delete", "ghost")
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

	out, err := runCmd(t, dir, "env", "show", "staging")
	if err != nil {
		t.Fatalf("config env show: %v", err)
	}
	if !strings.Contains(out, "environments/staging.yaml") {
		t.Errorf("output should contain env file path: %q", out)
	}
}

func TestConfigEnvShow_ActivePrefix(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "myenv", "http://x:8000")
	setActiveEnv(t, dir, "myenv")

	out, err := runCmd(t, dir, "env", "show", "myenv")
	if err != nil {
		t.Fatalf("config env show: %v", err)
	}
	if !strings.Contains(out, "[active]") {
		t.Errorf("expected [active] prefix when showing active env: %q", out)
	}
}

func TestConfigEnvShow_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, dir, "env", "show", "ghost")
	if err == nil {
		t.Fatal("expected error for non-existent env")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- new UX tests ----

func TestConfigNoArgs_ShowsHelpBeforeConfig(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "http://dev:8000")
	setActiveEnv(t, dir, "dev")

	out, err := runCmd(t, dir, "config")
	if err != nil {
		t.Fatalf("hf config: %v", err)
	}
	usageIdx := strings.Index(out, "Usage:")
	configIdx := strings.Index(out, "hyperfleet:")
	if usageIdx < 0 {
		t.Errorf("expected help block with 'Usage:' in output, got: %q", out)
	}
	if configIdx < 0 {
		t.Errorf("expected config section 'hyperfleet:' in output, got: %q", out)
	}
	if usageIdx > configIdx {
		t.Errorf("expected help block before config output: usageIdx=%d configIdx=%d", usageIdx, configIdx)
	}
}

func TestConfigSet_ShowsConfigAfterSet(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "dev", "hyperfleet:\n  api-url: http://dev:8000\n  api-version: v1\n")
	setActiveEnv(t, dir, "dev")

	out, err := runCmd(t, dir, "config", "set", "hyperfleet.api-version", "v2")
	if err != nil {
		t.Fatalf("config set: %v", err)
	}
	if !strings.Contains(out, "hyperfleet:") {
		t.Errorf("expected config sections in output after set, got: %q", out)
	}
	if !strings.Contains(out, "api-version") {
		t.Errorf("expected api-version key in config output, got: %q", out)
	}
}

func TestConfigSet_Interactive(t *testing.T) {
	dir := t.TempDir()
	makeEnvRaw(t, dir, "dev", "hyperfleet:\n  api-url: http://dev:8000\n  api-version: v1\n")
	setActiveEnv(t, dir, "dev")

	old := configSetSel
	configSetSel = mockPreviewSel{idx: 0} // selects hyperfleet.api-url (first item)
	defer func() { configSetSel = old }()

	rootCmd.SetIn(strings.NewReader("http://new:9000\n"))
	defer rootCmd.SetIn(nil)

	out, err := runCmd(t, dir, "config", "set")
	if err != nil {
		t.Fatalf("config set interactive: %v", err)
	}
	if !strings.Contains(out, "hyperfleet:") {
		t.Errorf("expected config output after interactive set, got: %q", out)
	}

	// Verify the value was actually persisted
	out2, err := runCmd(t, dir, "config", "get", "hyperfleet.api-url")
	if err != nil {
		t.Fatalf("config get after interactive set: %v", err)
	}
	if !strings.Contains(out2, "http://new:9000") {
		t.Errorf("expected new value persisted, got: %q", out2)
	}
}
