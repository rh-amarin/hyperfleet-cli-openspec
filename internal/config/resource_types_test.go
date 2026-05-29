package config_test

import (
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
)

func TestParseResourceTypes_RootAndChild(t *testing.T) {
	dir := t.TempDir()
	content := `hyperfleet:
  api-url: http://localhost:8000
resource-types:
  channels:
    path: channels
    state-key: channel-id
  versions:
    parent: channels
    path: "channels/{channel_id}/versions"
    state-key: version-id
`
	writeEnv(t, dir, "dev", content)

	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("active-environment", "dev"); err != nil {
		t.Fatal(err)
	}

	types, err := s.ResourceTypes()
	if err != nil {
		t.Fatal(err)
	}
	if len(types) != 2 {
		t.Fatalf("expected 2 types, got %d", len(types))
	}

	path, err := s.ResolveResourcePath("channels")
	if err != nil {
		t.Fatal(err)
	}
	if path != "channels" {
		t.Fatalf("channels path: got %q", path)
	}

	_, err = s.ResolveResourcePath("versions")
	if err == nil {
		t.Fatal("expected error without parent state")
	}

	if err := s.SetState("channel-id", "abc-123"); err != nil {
		t.Fatal(err)
	}
	path, err = s.ResolveResourcePath("versions")
	if err != nil {
		t.Fatal(err)
	}
	if path != "channels/abc-123/versions" {
		t.Fatalf("versions path: got %q", path)
	}
}

func TestParseResourceTypes_UnknownParent(t *testing.T) {
	dir := t.TempDir()
	content := `resource-types:
  versions:
    parent: missing
    path: versions
    state-key: version-id
`
	writeEnv(t, dir, "dev", content)
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("active-environment", "dev"); err != nil {
		t.Fatal(err)
	}
	_, err := s.ResourceTypes()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestParseResourceTypes_DuplicateStateKey(t *testing.T) {
	dir := t.TempDir()
	content := `resource-types:
  a:
    path: a
    state-key: same-id
  b:
    path: b
    state-key: same-id
`
	writeEnv(t, dir, "dev", content)
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("active-environment", "dev"); err != nil {
		t.Fatal(err)
	}
	_, err := s.ResourceTypes()
	if err == nil {
		t.Fatal("expected duplicate state-key error")
	}
}

func TestLoad_MixedEnvProfileWithResourceTypes(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", `hyperfleet:
  api-url: http://example:9000
  api-version: v1
resource-types:
  channels:
    path: channels
    state-key: channel-id
`)
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("active-environment", "dev"); err != nil {
		t.Fatal(err)
	}
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if got := s.Get("hyperfleet", "api-url"); got != "http://example:9000" {
		t.Fatalf("api-url: got %q", got)
	}
}

func TestResourceID_ExplicitWins(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", `resource-types:
  channels:
    path: channels
    state-key: channel-id
`)
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("active-environment", "dev"); err != nil {
		t.Fatal(err)
	}
	id, err := s.ResourceID("channels", "explicit")
	if err != nil || id != "explicit" {
		t.Fatalf("ResourceID: got %q err=%v", id, err)
	}
}
