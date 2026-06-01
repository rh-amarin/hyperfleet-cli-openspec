package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"gopkg.in/yaml.v3"
)

func newTestStore(t *testing.T) *config.Store {
	t.Helper()
	dir := t.TempDir()
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	return s
}

// writeEnv creates a named environment file in dir/environments/<name>.yaml.
func writeEnv(t *testing.T, dir, name, content string) {
	t.Helper()
	envDir := filepath.Join(dir, "environments")
	if err := os.MkdirAll(envDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envDir, name+".yaml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

// ---- Load ----

func TestLoad_NoConfigYAMLCreated(t *testing.T) {
	dir := t.TempDir()
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config.yaml")); !os.IsNotExist(err) {
		t.Error("Load must not create config.yaml")
	}
}

func TestLoad_StateYAMLCreated(t *testing.T) {
	dir := t.TempDir()
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "state.yaml")); err != nil {
		t.Errorf("Load must create state.yaml: %v", err)
	}
}

func TestLoad_EnvironmentsDirCreated(t *testing.T) {
	dir := t.TempDir()
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "environments")); err != nil {
		t.Errorf("Load must create environments/ dir: %v", err)
	}
}

// ---- Get precedence ----

func TestGet_Defaults(t *testing.T) {
	s := newTestStore(t)

	if got := s.Get("hyperfleet", "api-url"); got != "http://localhost:8000" {
		t.Errorf("api-url default: got %q", got)
	}
	if got := s.Get("hyperfleet", "api-version"); got != "v1" {
		t.Errorf("api-version default: got %q", got)
	}
	if got := s.Get("database", "host"); got != "localhost" {
		t.Errorf("database.host default: got %q", got)
	}
	if got := s.Get("database", "port"); got != "5432" {
		t.Errorf("database.port default: got %q", got)
	}
}

func TestGet_EnvVarOverridesDefault(t *testing.T) {
	s := newTestStore(t)

	t.Setenv("HF_API_URL", "http://custom-api:9000")
	t.Setenv("HF_TOKEN", "my-token")

	if got := s.Get("hyperfleet", "api-url"); got != "http://custom-api:9000" {
		t.Errorf("HF_API_URL override: got %q", got)
	}
	if got := s.Get("hyperfleet", "token"); got != "my-token" {
		t.Errorf("HF_TOKEN override: got %q", got)
	}
}

func TestGet_ActiveEnvFileOverridesDefault(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "myenv", "hyperfleet:\n  api-url: http://profile-api:8888\n")
	if err := os.WriteFile(filepath.Join(dir, "state.yaml"), []byte("active-environment: myenv\n"), 0600); err != nil {
		t.Fatal(err)
	}

	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	if got := s.Get("hyperfleet", "api-url"); got != "http://profile-api:8888" {
		t.Errorf("env file override: got %q, want http://profile-api:8888", got)
	}
	// Keys not in env file still fall back to defaults
	if got := s.Get("hyperfleet", "api-version"); got != "v1" {
		t.Errorf("default fallback: got %q", got)
	}
}

func TestGet_EnvVarOverridesActiveEnvFile(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "myenv", "hyperfleet:\n  api-url: http://profile-api:8888\n")
	if err := os.WriteFile(filepath.Join(dir, "state.yaml"), []byte("active-environment: myenv\n"), 0600); err != nil {
		t.Fatal(err)
	}

	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HF_API_URL", "http://envvar-api:7777")
	if got := s.Get("hyperfleet", "api-url"); got != "http://envvar-api:7777" {
		t.Errorf("env var over active env file: got %q, want http://envvar-api:7777", got)
	}
}

func TestGet_KubeconfigFromEnvFile(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", "kubernetes:\n  kubeconfig: /path/to/kubeconfig\n")
	if err := os.WriteFile(filepath.Join(dir, "state.yaml"), []byte("active-environment: dev\n"), 0600); err != nil {
		t.Fatal(err)
	}

	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if got := s.Get("kubernetes", "kubeconfig"); got != "/path/to/kubeconfig" {
		t.Errorf("kubeconfig from env file: got %q", got)
	}
}

func TestGet_HF_KUBECONFIGOverridesEnvFile(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", "kubernetes:\n  kubeconfig: /from/file\n")
	if err := os.WriteFile(filepath.Join(dir, "state.yaml"), []byte("active-environment: dev\n"), 0600); err != nil {
		t.Fatal(err)
	}

	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HF_KUBECONFIG", "/from/env")
	if got := s.Get("kubernetes", "kubeconfig"); got != "/from/env" {
		t.Errorf("HF_KUBECONFIG override: got %q, want /from/env", got)
	}
}

// ---- Set ----

func TestSet_WritesToActiveEnvFile(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", "hyperfleet:\n  api-url: http://dev:8000\n")
	if err := os.WriteFile(filepath.Join(dir, "state.yaml"), []byte("active-environment: dev\n"), 0600); err != nil {
		t.Fatal(err)
	}

	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("hyperfleet", "api-version", "v2"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Reload and verify persistence
	s2 := config.New(dir)
	if err := s2.Load(); err != nil {
		t.Fatal(err)
	}
	if got := s2.Get("hyperfleet", "api-version"); got != "v2" {
		t.Errorf("persisted value: got %q, want v2", got)
	}
}

func TestSet_ErrorWhenNoActiveEnv(t *testing.T) {
	s := newTestStore(t)
	err := s.Set("hyperfleet", "api-url", "http://x:8000")
	if err == nil {
		t.Fatal("expected error when no active environment")
	}
	if !contains(err.Error(), "no active environment") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- State ----

func TestAtomicStateWrite(t *testing.T) {
	s := newTestStore(t)

	if err := s.SetState("cluster-id", "cl-abc-123"); err != nil {
		t.Fatalf("SetState: %v", err)
	}
	if got := s.GetState("cluster-id"); got != "cl-abc-123" {
		t.Errorf("GetState: got %q", got)
	}

	statePath := filepath.Join(s.ConfigDir(), "state.yaml")
	info, err := os.Stat(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("state.yaml perm: got %o, want 0600", perm)
	}
}

// ---- RequireActiveEnvironment ----

func TestRequireActiveEnvironment(t *testing.T) {
	s := newTestStore(t)

	_, err := s.RequireActiveEnvironment()
	if err == nil {
		t.Fatal("expected error when no active environment")
	}

	if err := s.SetState("active-environment", "dev"); err != nil {
		t.Fatal(err)
	}
	env, err := s.RequireActiveEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env != "dev" {
		t.Errorf("RequireActiveEnvironment: got %q, want dev", env)
	}
}

// ---- ClusterID / NodePoolID ----

func TestClusterIDResolution(t *testing.T) {
	s := newTestStore(t)

	_, err := s.ClusterID("")
	if err == nil {
		t.Fatal("expected error when no cluster ID")
	}

	id, err := s.ClusterID("explicit-id")
	if err != nil || id != "explicit-id" {
		t.Errorf("explicit: got %q, %v", id, err)
	}

	if err := s.SetState("cluster-id", "state-id"); err != nil {
		t.Fatal(err)
	}
	id, err = s.ClusterID("")
	if err != nil || id != "state-id" {
		t.Errorf("state fallback: got %q, %v", id, err)
	}
}

// ---- ConfigTemplateYAML ----

func TestConfigTemplateYAML_NonEmptyAndParses(t *testing.T) {
	if len(config.ConfigTemplateYAML) == 0 {
		t.Fatal("ConfigTemplateYAML must not be empty")
	}
	var doc map[string]any
	if err := yaml.Unmarshal(config.ConfigTemplateYAML, &doc); err != nil {
		t.Fatalf("ConfigTemplateYAML does not parse as YAML: %v", err)
	}
	if len(doc) == 0 {
		t.Fatal("ConfigTemplateYAML parsed to an empty map")
	}
	rt, ok := doc["resource-types"].(map[string]any)
	if !ok {
		t.Fatal("resource-types section missing from template")
	}
	for _, name := range []string{"clusters", "nodepools"} {
		if _, ok := rt[name]; !ok {
			t.Fatalf("resource-types.%s missing from template", name)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

func TestHFNamespace_EnvVar(t *testing.T) {
	s := newTestStore(t)
	t.Setenv("HF_NAMESPACE", "test-ns")
	if got := s.Get("hyperfleet", "namespace"); got != "test-ns" {
		t.Fatalf("expected test-ns, got %q", got)
	}
}

func TestHFNamespace_Profile(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "myenv", "hyperfleet:\n  namespace: my-ns\n")
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := s.SetState("active-environment", "myenv"); err != nil {
		t.Fatalf("SetState: %v", err)
	}
	if err := s.Load(); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := s.Get("hyperfleet", "namespace"); got != "my-ns" {
		t.Fatalf("expected my-ns, got %q", got)
	}
}
