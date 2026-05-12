package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---- maestro test fixtures ----

const bundleID = "b1d2c3e4-0000-0000-0000-000000000001"

var bundleJSON = `{
  "id": "` + bundleID + `",
  "name": "mw-cluster1",
  "consumer_name": "cluster1",
  "version": 3,
  "manifest_count": 1,
  "manifests": [{"kind": "Deployment", "name": "my-app", "namespace": "default"}],
  "conditions": [{"type": "Applied", "status": "True", "reason": "Applied"}]
}`

var bundleListJSON = `{
  "items": [` + bundleJSON + `],
  "kind": "ResourceBundleList",
  "total": 1
}`

var consumerListJSON = `{
  "items": [{"id": "c-1", "kind": "Consumer", "name": "cluster1"}],
  "kind": "ConsumerList",
  "total": 1
}`

// ---- helpers ----

// setupMaestroEnv creates a temp config dir with an env that configures the
// Maestro HTTP endpoint to ts.URL and activates that env.
func setupMaestroEnv(t *testing.T, apiURL, maestroURL string) string {
	t.Helper()
	dir := t.TempDir()

	// Write environment file with both hyperfleet and maestro endpoints.
	envDir := filepath.Join(dir, "environments")
	if err := os.MkdirAll(envDir, 0700); err != nil {
		t.Fatal(err)
	}
	content := fmt.Sprintf("hyperfleet:\n  api-url: %s\nmaestro:\n  http-endpoint: %s\n  consumer: cluster1\n", apiURL, maestroURL)
	if err := os.WriteFile(filepath.Join(envDir, "test.yaml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	setActiveEnv(t, dir, "test")
	return dir
}

func resetMaestroFlags() {
	outputFmt = "json"
	noColor = false
	verbose = false
	deleteAllBundles = false
}

func runMaestroCmd(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	resetMaestroFlags()
	return runCmd(t, dir, args...)
}

func runMaestroCmdWithStdin(t *testing.T, dir, stdin string, args ...string) (string, error) {
	t.Helper()
	resetMaestroFlags()
	rootCmd.SetIn(strings.NewReader(stdin))
	t.Cleanup(func() { rootCmd.SetIn(nil) })
	return runCmd(t, dir, args...)
}

// ---- maestro list ----

func TestMaestroList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/maestro/v1/resource-bundles" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, bundleListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	// Use a dummy API URL; maestro calls go to srv.
	dir := setupMaestroEnv(t, "http://localhost:9999", srv.URL)
	out, err := runMaestroCmd(t, dir, "maestro", "list")
	if err != nil {
		t.Fatalf("maestro list: %v", err)
	}
	if !strings.Contains(out, bundleID) {
		t.Errorf("expected bundle ID in output, got: %q", out)
	}
	if !strings.Contains(out, "mw-cluster1") {
		t.Errorf("expected bundle name in output, got: %q", out)
	}
}

func TestMaestroListTable_WithManifests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/maestro/v1/resource-bundles" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, bundleListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	dir := setupMaestroEnv(t, "http://localhost:9999", srv.URL)
	out, err := runMaestroCmd(t, dir, "maestro", "list", "--output", "table")
	if err != nil {
		t.Fatalf("maestro list --output table: %v", err)
	}
	if !strings.Contains(out, bundleID) {
		t.Errorf("expected bundle ID in output, got: %q", out)
	}
	if !strings.Contains(out, "mw-cluster1") {
		t.Errorf("expected bundle name in output, got: %q", out)
	}
	if !strings.Contains(out, "v3") {
		t.Errorf("expected version in output, got: %q", out)
	}
	if !strings.Contains(out, "  Deployment/my-app") {
		t.Errorf("expected indented manifest line in output, got: %q", out)
	}
	if !strings.Contains(out, "default") {
		t.Errorf("expected namespace in output, got: %q", out)
	}
}

func TestMaestroListTable_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"items": [], "kind": "ResourceBundleList", "total": 0}`)
	}))
	defer srv.Close()

	dir := setupMaestroEnv(t, "http://localhost:9999", srv.URL)
	out, err := runMaestroCmd(t, dir, "maestro", "list", "--output", "table")
	if err != nil {
		t.Fatalf("maestro list --output table empty: %v", err)
	}
	if strings.TrimSpace(out) != "No resource bundles." {
		t.Errorf("expected 'No resource bundles.', got: %q", out)
	}
}

// ---- maestro bundles ----

func TestMaestroBundles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/maestro/v1/resource-bundles" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, bundleListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	dir := setupMaestroEnv(t, "http://localhost:9999", srv.URL)
	out, err := runMaestroCmd(t, dir, "maestro", "bundles")
	if err != nil {
		t.Fatalf("maestro bundles: %v", err)
	}
	if !strings.Contains(out, "ResourceBundleList") {
		t.Errorf("expected ResourceBundleList in output, got: %q", out)
	}
}

// ---- maestro consumers ----

func TestMaestroConsumers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/maestro/v1/consumers" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, consumerListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	dir := setupMaestroEnv(t, "http://localhost:9999", srv.URL)
	out, err := runMaestroCmd(t, dir, "maestro", "consumers")
	if err != nil {
		t.Fatalf("maestro consumers: %v", err)
	}
	if !strings.Contains(out, "ConsumerList") {
		t.Errorf("expected ConsumerList in output, got: %q", out)
	}
	if !strings.Contains(out, "cluster1") {
		t.Errorf("expected consumer name in output, got: %q", out)
	}
}

// ---- maestro get ----

func TestMaestroGet_ByName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/maestro/v1/resource-bundles" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, bundleListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	dir := setupMaestroEnv(t, "http://localhost:9999", srv.URL)
	out, err := runMaestroCmd(t, dir, "maestro", "get", "mw-cluster1")
	if err != nil {
		t.Fatalf("maestro get: %v", err)
	}
	// Verify it's valid JSON of a single bundle.
	var rb map[string]any
	if jerr := json.Unmarshal([]byte(out), &rb); jerr != nil {
		t.Fatalf("expected JSON object output, got: %q", out)
	}
	if rb["id"] != bundleID {
		t.Errorf("expected bundle ID %s, got %v", bundleID, rb["id"])
	}
}

func TestMaestroGet_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"items": [], "kind": "ResourceBundleList", "total": 0}`)
	}))
	defer srv.Close()

	dir := setupMaestroEnv(t, "http://localhost:9999", srv.URL)
	_, err := runMaestroCmd(t, dir, "maestro", "get", "nonexistent")
	// Should exit 0 — warning written to stderr, not an error return.
	if err != nil {
		t.Fatalf("maestro get not-found should exit 0, got: %v", err)
	}
}

// ---- maestro delete ----

func TestMaestroDelete_ByName(t *testing.T) {
	var gotDeletePath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/maestro/v1/resource-bundles":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, bundleListJSON)
		case r.Method == http.MethodDelete:
			gotDeletePath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := setupMaestroEnv(t, "http://localhost:9999", srv.URL)
	_, err := runMaestroCmd(t, dir, "maestro", "delete", "mw-cluster1")
	if err != nil {
		t.Fatalf("maestro delete: %v", err)
	}
	if gotDeletePath != "/api/maestro/v1/resource-bundles/"+bundleID {
		t.Errorf("unexpected DELETE path: %s", gotDeletePath)
	}
}

func TestMaestroDelete_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"items": [], "kind": "ResourceBundleList", "total": 0}`)
	}))
	defer srv.Close()

	dir := setupMaestroEnv(t, "http://localhost:9999", srv.URL)
	_, err := runMaestroCmd(t, dir, "maestro", "delete", "nonexistent")
	// Not found → exit 0 with WARN message.
	if err != nil {
		t.Fatalf("maestro delete not-found should exit 0, got: %v", err)
	}
}

func TestMaestroDelete_All_Confirmed(t *testing.T) {
	var deletedIDs []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/maestro/v1/resource-bundles":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"items":[
				{"id":"id-1","name":"bundle-a","kind":"ResourceBundle"},
				{"id":"id-2","name":"bundle-b","kind":"ResourceBundle"}
			],"kind":"ResourceBundleList","total":2}`)
		case r.Method == http.MethodDelete:
			deletedIDs = append(deletedIDs, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := setupMaestroEnv(t, "http://localhost:9999", srv.URL)
	out, err := runMaestroCmdWithStdin(t, dir, "yes\n", "maestro", "delete", "--all")
	if err != nil {
		t.Fatalf("maestro delete --all: %v", err)
	}
	if len(deletedIDs) != 2 {
		t.Errorf("expected 2 DELETE requests, got %d: %v", len(deletedIDs), deletedIDs)
	}
	if !strings.Contains(out, "2 resource bundle(s) will be deleted") {
		t.Errorf("expected count in output, got: %q", out)
	}
	if !strings.Contains(out, "bundle-a") || !strings.Contains(out, "bundle-b") {
		t.Errorf("expected bundle names in output, got: %q", out)
	}
}

func TestMaestroDelete_All_Aborted(t *testing.T) {
	var deleteCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, bundleListJSON)
		case r.Method == http.MethodDelete:
			deleteCount++
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := setupMaestroEnv(t, "http://localhost:9999", srv.URL)
	out, err := runMaestroCmdWithStdin(t, dir, "no\n", "maestro", "delete", "--all")
	if err != nil {
		t.Fatalf("unexpected error on abort: %v", err)
	}
	if deleteCount != 0 {
		t.Errorf("expected no DELETE requests after abort, got %d", deleteCount)
	}
	if !strings.Contains(out, "Aborted") {
		t.Errorf("expected 'Aborted' in output, got: %q", out)
	}
}

func TestMaestroDelete_All_MutuallyExclusiveWithName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	dir := setupMaestroEnv(t, "http://localhost:9999", srv.URL)
	_, err := runMaestroCmd(t, dir, "maestro", "delete", "--all", "some-name")
	if err == nil {
		t.Fatal("expected error when --all and name are both provided")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' error, got: %v", err)
	}
}
