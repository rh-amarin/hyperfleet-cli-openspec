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
  versions:
    parent: channels
    path: "channels/{channel_id}/versions"
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

	ch, err := s.ResourceType("channels")
	if err != nil || ch.StateKey != "channels" {
		t.Fatalf("channels StateKey: %+v err=%v", ch, err)
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

	if err := s.SetState("channels", "abc-123"); err != nil {
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

func TestResolveResourcePath_ThreeLevelChain(t *testing.T) {
	dir := t.TempDir()
	content := `resource-types:
  channels:
    path: channels
  versions:
    parent: channels
    path: "channels/{channel_id}/versions"
  releases:
    parent: versions
    path: "channels/{channel_id}/versions/{version_id}/releases"
`
	writeEnv(t, dir, "dev", content)

	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("active-environment", "dev"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("channels", "chan-1"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("versions", "ver-1"); err != nil {
		t.Fatal(err)
	}

	path, err := s.ResolveResourcePath("releases")
	if err != nil {
		t.Fatal(err)
	}
	want := "channels/chan-1/versions/ver-1/releases"
	if path != want {
		t.Fatalf("releases path: got %q want %q", path, want)
	}
}

func TestParseResourceTypes_UnknownParent(t *testing.T) {
	dir := t.TempDir()
	content := `resource-types:
  versions:
    parent: missing
    path: versions
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

func TestParseResourceTypes_PathParamOverride(t *testing.T) {
	dir := t.TempDir()
	content := `resource-types:
  widgets:
    path: widgets
    path-param: widget_uuid
  parts:
    parent: widgets
    path: "widgets/{widget_uuid}/parts"
`
	writeEnv(t, dir, "dev", content)
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("active-environment", "dev"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("widgets", "w-1"); err != nil {
		t.Fatal(err)
	}

	path, err := s.ResolveResourcePath("parts")
	if err != nil {
		t.Fatal(err)
	}
	if path != "widgets/w-1/parts" {
		t.Fatalf("parts path: got %q", path)
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

func TestResolveResourceStatusPath(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", `resource-types:
  channels:
    path: channels
  versions:
    parent: channels
    path: "channels/{channel_id}/versions"
`)
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("active-environment", "dev"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("channels", "chan-42"); err != nil {
		t.Fatal(err)
	}

	path, err := s.ResolveResourceStatusPath("channels", "chan-99")
	if err != nil {
		t.Fatal(err)
	}
	if path != "channels/chan-99/statuses" {
		t.Fatalf("channels statuses path: got %q", path)
	}

	path, err = s.ResolveResourceStatusPath("versions", "ver-1")
	if err != nil {
		t.Fatal(err)
	}
	if path != "channels/chan-42/versions/ver-1/statuses" {
		t.Fatalf("versions statuses path: got %q", path)
	}
}

func TestResolveListPath_ReleasesWithVersionPathParam(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", `resource-types:
  channels:
    path: channels
  versions:
    parent: channels
    path: "channels/{channel_id}/versions"
    path-param: channel_id
  releases:
    parent: versions
    path: "channels/{channel_id}/versions/{version_id}/releases"
    path-param: version_id
`)
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("active-environment", "dev"); err != nil {
		t.Fatal(err)
	}

	path, err := s.ResolveListPath("releases", map[string]string{
		"channels": "chan-1",
		"versions": "ver-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "channels/chan-1/versions/ver-1/releases"
	if path != want {
		t.Fatalf("releases list path: got %q want %q", path, want)
	}
}

func TestResolveListPath_WithAncestorIDs(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", `resource-types:
  channels:
    path: channels
  versions:
    parent: channels
    path: "channels/{channel_id}/versions"
`)
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetState("active-environment", "dev"); err != nil {
		t.Fatal(err)
	}

	path, err := s.ResolveListPath("versions", map[string]string{"channels": "chan-42"})
	if err != nil {
		t.Fatal(err)
	}
	if path != "channels/chan-42/versions" {
		t.Fatalf("got %q", path)
	}
}

func TestRootAndChildResourceTypes(t *testing.T) {
	types := []config.ResourceTypeDef{
		{Name: "versions", Parent: "channels", StateKey: "versions"},
		{Name: "channels", StateKey: "channels"},
		{Name: "releases", Parent: "versions", StateKey: "releases"},
	}
	roots := config.RootResourceTypes(types)
	if len(roots) != 1 || roots[0].Name != "channels" {
		t.Fatalf("roots: %+v", roots)
	}
	children := config.ChildResourceTypes(types, "channels")
	if len(children) != 1 || children[0].Name != "versions" {
		t.Fatalf("children: %+v", children)
	}
}

func TestResourceID_ExplicitWins(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", `resource-types:
  channels:
    path: channels
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

func TestDerivePathParamFromTypeName(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"clusters":  "cluster_id",
		"nodepools": "nodepool_id",
		"channels":  "channel_id",
		"widget":    "widget_id",
	}
	for name, want := range cases {
		if got := config.DerivePathParamFromTypeName(name); got != want {
			t.Errorf("%s: got %q want %q", name, got, want)
		}
	}
}
