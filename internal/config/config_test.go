package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
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

func TestDefaults(t *testing.T) {
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

func TestEnvVarOverride(t *testing.T) {
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

func TestEnvVarHigherThanProfile(t *testing.T) {
	dir := t.TempDir()
	// Create an env profile that sets api-url
	if err := os.MkdirAll(filepath.Join(dir, "environments"), 0700); err != nil {
		t.Fatal(err)
	}
	profContent := "hyperfleet:\n  api-url: http://profile-api:8888\n"
	if err := os.WriteFile(filepath.Join(dir, "environments", "myenv.yaml"), []byte(profContent), 0600); err != nil {
		t.Fatal(err)
	}

	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	// Activate the profile
	if err := s.SetState("active-environment", "myenv"); err != nil {
		t.Fatal(err)
	}
	// Reload so the profile is applied
	s2 := config.New(dir)
	if err := s2.Load(); err != nil {
		t.Fatal(err)
	}

	// Profile should override file
	if got := s2.Get("hyperfleet", "api-url"); got != "http://profile-api:8888" {
		t.Errorf("profile override: got %q, want %q", got, "http://profile-api:8888")
	}

	// Env var should override profile
	t.Setenv("HF_API_URL", "http://envvar-api:7777")
	if got := s2.Get("hyperfleet", "api-url"); got != "http://envvar-api:7777" {
		t.Errorf("env var over profile: got %q, want %q", got, "http://envvar-api:7777")
	}
}

func TestSetAndPersist(t *testing.T) {
	dir := t.TempDir()
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	if err := s.Set("hyperfleet", "api-url", "http://new-api:1234"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Reload from disk — value should persist
	s2 := config.New(dir)
	if err := s2.Load(); err != nil {
		t.Fatal(err)
	}
	if got := s2.Get("hyperfleet", "api-url"); got != "http://new-api:1234" {
		t.Errorf("persisted value: got %q", got)
	}
}

func TestAtomicStateWrite(t *testing.T) {
	s := newTestStore(t)

	if err := s.SetState("cluster-id", "cl-abc-123"); err != nil {
		t.Fatalf("SetState: %v", err)
	}

	if got := s.GetState("cluster-id"); got != "cl-abc-123" {
		t.Errorf("GetState: got %q", got)
	}

	// Verify file was written with correct permissions
	statePath := filepath.Join(s.ConfigDir(), "state.yaml")
	info, err := os.Stat(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("state.yaml perm: got %o, want 0600", perm)
	}
}

func TestEnvProfileDeepMerge(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "environments"), 0700); err != nil {
		t.Fatal(err)
	}

	// Profile only overrides api-url; other keys should still come from defaults
	profContent := "hyperfleet:\n  api-url: http://env-api:8888\ndatabase:\n  host: db.prod.internal\n"
	if err := os.WriteFile(filepath.Join(dir, "environments", "prod.yaml"), []byte(profContent), 0600); err != nil {
		t.Fatal(err)
	}
	// Write state with active-environment
	stateContent := "active-environment: prod\n"
	if err := os.WriteFile(filepath.Join(dir, "state.yaml"), []byte(stateContent), 0600); err != nil {
		t.Fatal(err)
	}

	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	// Overridden by profile
	if got := s.Get("hyperfleet", "api-url"); got != "http://env-api:8888" {
		t.Errorf("profile api-url: got %q", got)
	}
	if got := s.Get("database", "host"); got != "db.prod.internal" {
		t.Errorf("profile database.host: got %q", got)
	}
	// Not overridden — should still be default
	if got := s.Get("database", "port"); got != "5432" {
		t.Errorf("default database.port: got %q", got)
	}
	if got := s.Get("hyperfleet", "api-version"); got != "v1" {
		t.Errorf("default api-version: got %q", got)
	}
}

func TestRequireActiveEnvironment(t *testing.T) {
	s := newTestStore(t)

	// No active env → error
	_, err := s.RequireActiveEnvironment()
	if err == nil {
		t.Fatal("expected error when no active environment")
	}

	// Set active env → no error
	if err := s.SetState("active-environment", "dev"); err != nil {
		t.Fatal(err)
	}
	env, err := s.RequireActiveEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env != "dev" {
		t.Errorf("RequireActiveEnvironment: got %q, want %q", env, "dev")
	}
}

func TestClusterIDResolution(t *testing.T) {
	s := newTestStore(t)

	// No state, no explicit → error
	_, err := s.ClusterID("")
	if err == nil {
		t.Fatal("expected error when no cluster ID")
	}

	// Explicit arg wins
	id, err := s.ClusterID("explicit-id")
	if err != nil || id != "explicit-id" {
		t.Errorf("explicit: got %q, %v", id, err)
	}

	// State fallback
	if err := s.SetState("cluster-id", "state-id"); err != nil {
		t.Fatal(err)
	}
	id, err = s.ClusterID("")
	if err != nil || id != "state-id" {
		t.Errorf("state fallback: got %q, %v", id, err)
	}
}

func TestConfigFilePermissions(t *testing.T) {
	dir := t.TempDir()
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(dir, "config.yaml")
	info, err := os.Stat(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("config.yaml perm: got %o, want 0600", perm)
	}
}
