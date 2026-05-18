package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/selector"
)

// ---- test fixtures ----

const clusterID = "019dc049-43a8-7a42-b44a-8d7f89e9e10f"

var clusterJSON = `{
  "id": "` + clusterID + `",
  "kind": "Cluster",
  "name": "test-cluster",
  "generation": 1,
  "labels": {"counter": "1"},
  "spec": {"region": "us-east-1", "version": "4.15.0"},
  "status": {
    "conditions": [
      {"type": "Available", "status": "True", "last_transition_time": "2026-05-10T00:00:00Z", "observed_generation": 1},
      {"type": "Reconciled", "status": "True", "last_transition_time": "2026-05-10T00:00:00Z", "observed_generation": 1}
    ]
  },
  "created_by": "user@example.com",
  "created_time": "2026-05-10T00:00:00Z",
  "href": "/api/hyperfleet/v1/clusters/` + clusterID + `"
}`

var clusterListJSON = `{
  "items": [` + clusterJSON + `],
  "kind": "ClusterList",
  "page": 1,
  "size": 1,
  "total": 1
}`

var emptyListJSON = `{"items": [], "kind": "ClusterList", "page": 1, "size": 0, "total": 0}`

var adapterStatusesJSON = `{
  "items": [
    {
      "adapter": "cl-deployment",
      "observed_generation": 1,
      "conditions": [
        {"type": "Available", "status": "True", "last_transition_time": "2026-05-10T00:00:00Z"}
      ],
      "created_time": "2026-05-10T00:00:00Z",
      "last_report_time": "2026-05-10T00:00:00Z"
    }
  ],
  "kind": "AdapterStatusList",
  "page": 1,
  "size": 1,
  "total": 1
}`

// ---- helpers ----

// setClusterIDInState writes state.yaml with the active env and cluster-id.
func setClusterIDInState(t *testing.T, dir, envName, id string) {
	t.Helper()
	statePath := filepath.Join(dir, "state.yaml")
	content := fmt.Sprintf("active-environment: %s\ncluster-id: %s\n", envName, id)
	if err := os.WriteFile(statePath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

// setupClusterEnv creates a temp dir with an env pointing to ts.URL and activates it.
func setupClusterEnv(t *testing.T, ts *httptest.Server) string {
	t.Helper()
	dir := t.TempDir()
	makeEnv(t, dir, "test", ts.URL)
	setActiveEnv(t, dir, "test")
	return dir
}

// resetClusterFlags resets all cluster and global flag vars to defaults.
// Call before each test to prevent state carry-over from previous test runs.
func resetClusterFlags() {
	// Global flags
	outputFmt = "json"
	noColor = false
	verbose = false
	// Cluster-specific flags
	clusterCreateName = ""
	clusterCreateFile = ""
	clusterCreateReplicas = 0
	clusterCreateNPID = ""
	clusterUpdateName = ""
	clusterUpdateReplicas = 0
	clusterListWatch = false
	clusterListWatchSecs = 5
	clusterInteractive = false
	clusterListSearch = ""
	clusterIDInteractive = false
}

// mockSel is a test double for selector.Selector shared across cluster and nodepool tests.
type mockSel struct {
	idx int
	err error
}

func (m mockSel) Select(_ []selector.Item) (int, error) { return m.idx, m.err }

// runClusterCmd wraps runCmd and resets all flag state before each invocation
// so tests are independent regardless of execution order.
func runClusterCmd(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	resetClusterFlags()
	return runCmd(t, dir, args...)
}

// apiPrefix is the URL path prefix for all API calls in tests.
const apiPrefix = "/api/hyperfleet/v1"

// ---- cluster list ----

func TestClusterList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "list")
	if err != nil {
		t.Fatalf("cluster list: %v", err)
	}
	if !strings.Contains(out, clusterID) {
		t.Errorf("expected cluster ID in output, got: %q", out)
	}
	// Verify it's valid JSON
	if !strings.Contains(out, `"items"`) {
		t.Errorf("expected JSON list response, got: %q", out)
	}
}

func TestClusterList_Table(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, clusterListJSON)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "list", "--output", "table")
	if err != nil {
		t.Fatalf("cluster list --output table: %v", err)
	}
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") {
		t.Errorf("expected table headers, got: %q", out)
	}
	if !strings.Contains(out, "test-cluster") {
		t.Errorf("expected cluster name in table, got: %q", out)
	}
}

// ---- cluster table ----

func TestClusterTable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "table")
	if err != nil {
		t.Fatalf("cluster table: %v", err)
	}
	for _, header := range []string{"ID", "NAME", "GEN", "STATUS"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output, got: %q", header, out)
		}
	}
	if !strings.Contains(out, "test-cluster") {
		t.Errorf("expected cluster name in table output, got: %q", out)
	}
}

// ---- cluster list --search ----

func TestClusterList_Search_NoFlag(t *testing.T) {
	var receivedQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, clusterListJSON)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	_, err := runClusterCmd(t, dir, "cluster", "list")
	if err != nil {
		t.Fatalf("cluster list (no flag): %v", err)
	}
	if receivedQuery != "" {
		t.Errorf("expected no query string, got: %q", receivedQuery)
	}
}

func TestClusterList_Search_LabelFilter(t *testing.T) {
	var receivedQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, clusterListJSON)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	_, err := runClusterCmd(t, dir, "cluster", "list", "--search", "labels.environment='prod'")
	if err != nil {
		t.Fatalf("cluster list --search: %v", err)
	}
	want := "search=labels.environment%3D%27prod%27"
	if receivedQuery != want {
		t.Errorf("query string: got %q, want %q", receivedQuery, want)
	}
}

func TestClusterList_Search_CompoundExpr(t *testing.T) {
	var receivedQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, clusterListJSON)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	expr := "labels.team='core' and generation>1"
	_, err := runClusterCmd(t, dir, "cluster", "list", "--search", expr)
	if err != nil {
		t.Fatalf("cluster list --search compound: %v", err)
	}
	if receivedQuery == "" || receivedQuery == "search="+expr {
		t.Errorf("expected URL-encoded query, got: %q", receivedQuery)
	}
	if !strings.HasPrefix(receivedQuery, "search=") {
		t.Errorf("expected query to start with 'search=', got: %q", receivedQuery)
	}
}

func TestClusterList_Search_APIError400(t *testing.T) {
	errBody := `{"type":"about:blank","title":"Validation Error","status":400,"detail":"invalid search expression","code":"HYPERFLEET-VAL-001"}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errBody)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "list", "--search", "not status.conditions.Ready='True'")
	if err != nil {
		t.Fatalf("cluster list --search API 400 should exit 0, got error: %v", err)
	}
	if !strings.Contains(out, "400") && !strings.Contains(out, "Validation Error") {
		t.Errorf("expected error JSON in output, got: %q", out)
	}
}

// ---- cluster get ----

func TestClusterGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "get", clusterID)
	if err != nil {
		t.Fatalf("cluster get: %v", err)
	}
	if !strings.Contains(out, clusterID) {
		t.Errorf("expected cluster ID in output, got: %q", out)
	}
	if !strings.Contains(out, `"kind"`) {
		t.Errorf("expected JSON cluster object, got: %q", out)
	}
}

func TestClusterGet_FromState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	out, err := runClusterCmd(t, dir, "cluster", "get")
	if err != nil {
		t.Fatalf("cluster get (from state): %v", err)
	}
	if !strings.Contains(out, clusterID) {
		t.Errorf("expected cluster ID in output, got: %q", out)
	}
}

func TestClusterGet_NotFound(t *testing.T) {
	notFoundJSON := `{"type":"https://api.hyperfleet.io/errors/not-found","title":"Resource Not Found","status":404,"detail":"Cluster not found","code":"HYPERFLEET-NTF-001"}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, notFoundJSON)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "get", "00000000-0000-0000-0000-000000000000")
	// API errors exit 0 (err == nil)
	if err != nil {
		t.Fatalf("cluster get 404 should exit 0, got error: %v", err)
	}
	// Output should contain the error JSON
	if !strings.Contains(out, "Not Found") && !strings.Contains(out, "404") {
		t.Errorf("expected error JSON in output, got: %q", out)
	}
}

// ---- cluster create ----

func TestClusterCreate(t *testing.T) {
	resetClusterFlags()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// Duplicate check — returns empty list
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyListJSON)
		// POST create
		case r.Method == http.MethodPost && r.URL.Path == apiPrefix+"/clusters":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, clusterJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "create", "--name", "test-cluster")
	if err != nil {
		t.Fatalf("cluster create: %v", err)
	}
	if !strings.Contains(out, clusterID) {
		t.Errorf("expected cluster ID in output, got: %q", out)
	}

	// Verify cluster-id persisted to state.yaml
	stateRaw, _ := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if !strings.Contains(string(stateRaw), clusterID) {
		t.Errorf("cluster-id not persisted to state.yaml: %q", string(stateRaw))
	}
}

func TestClusterCreate_DuplicateGuard(t *testing.T) {
	resetClusterFlags()
	postCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters":
			// Return existing cluster — triggers duplicate guard
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterListJSON)
		case r.Method == http.MethodPost:
			postCalled = true
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, clusterJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	_, err := runClusterCmd(t, dir, "cluster", "create", "--name", "test-cluster")
	if err != nil {
		t.Fatalf("cluster create duplicate: %v", err)
	}
	if postCalled {
		t.Error("POST should not have been called for duplicate cluster")
	}
}

func TestClusterCreate_Defaults(t *testing.T) {
	resetClusterFlags()
	var capturedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyListJSON)
		case r.Method == http.MethodPost:
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, clusterJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	// No --name flag → should use default "my-cluster"
	_, err := runClusterCmd(t, dir, "cluster", "create")
	if err != nil {
		t.Fatalf("cluster create defaults: %v", err)
	}
	if capturedBody != nil {
		if name, _ := capturedBody["name"].(string); name != "my-cluster" {
			t.Errorf("expected default name 'my-cluster', got %q", name)
		}
	}
}

// ---- cluster update ----

func TestClusterUpdate(t *testing.T) {
	resetClusterFlags()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch && r.URL.Path == apiPrefix+"/clusters/"+clusterID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "update", clusterID, "--name", "new-name")
	if err != nil {
		t.Fatalf("cluster update: %v", err)
	}
	if !strings.Contains(out, clusterID) {
		t.Errorf("expected cluster ID in output, got: %q", out)
	}
}

// ---- cluster delete ----

func TestClusterDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == apiPrefix+"/clusters/"+clusterID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "delete", clusterID)
	if err != nil {
		t.Fatalf("cluster delete: %v", err)
	}
	// Spec: output the deleted cluster object.
	if !strings.Contains(out, clusterID) {
		t.Errorf("cluster delete: expected deleted cluster JSON in output, got: %q", out)
	}
}

func TestClusterDelete_FromState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == apiPrefix+"/clusters/"+clusterID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	out, err := runClusterCmd(t, dir, "cluster", "delete")
	if err != nil {
		t.Fatalf("cluster delete (from state): %v", err)
	}
	if !strings.Contains(out, clusterID) {
		t.Errorf("cluster delete from state: expected cluster JSON in output, got: %q", out)
	}
}

func TestClusterDelete_NotFound(t *testing.T) {
	notFoundJSON := `{"type":"https://api.hyperfleet.io/errors/not-found","title":"Resource Not Found","status":404,"detail":"Cluster not found"}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, notFoundJSON)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	_, err := runClusterCmd(t, dir, "cluster", "delete", "00000000-0000-0000-0000-000000000000")
	// 404 on delete returns an error (exit 1)
	if err == nil {
		t.Fatal("expected error for cluster delete 404")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- cluster conditions ----

func TestClusterConditions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "conditions", clusterID)
	if err != nil {
		t.Fatalf("cluster conditions: %v", err)
	}
	if !strings.Contains(out, `"generation"`) {
		t.Errorf("expected generation in output, got: %q", out)
	}
	if !strings.Contains(out, `"conditions"`) {
		t.Errorf("expected conditions in output, got: %q", out)
	}
	if !strings.Contains(out, "Available") {
		t.Errorf("expected Available condition, got: %q", out)
	}
}

func TestClusterConditions_FromState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	_, err := runClusterCmd(t, dir, "cluster", "conditions")
	if err != nil {
		t.Fatalf("cluster conditions (from state): %v", err)
	}
}

// ---- cluster statuses ----

func TestClusterStatuses(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/statuses" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, adapterStatusesJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	out, err := runClusterCmd(t, dir, "cluster", "statuses")
	if err != nil {
		t.Fatalf("cluster statuses: %v", err)
	}
	if !strings.Contains(out, "cl-deployment") {
		t.Errorf("expected adapter name in output, got: %q", out)
	}
	if !strings.Contains(out, `"AdapterStatusList"`) {
		t.Errorf("expected AdapterStatusList kind, got: %q", out)
	}
}

func TestClusterStatuses_Table(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/statuses" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, adapterStatusesJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	out, err := runClusterCmd(t, dir, "cluster", "statuses", "--output", "table")
	if err != nil {
		t.Fatalf("cluster statuses --output table: %v", err)
	}
	if !strings.Contains(out, "ADAPTER") || !strings.Contains(out, "GEN") {
		t.Errorf("expected table headers, got: %q", out)
	}
	if !strings.Contains(out, "cl-deployment") {
		t.Errorf("expected adapter name in table, got: %q", out)
	}
}

// ---- cluster adapter post-status ----

// ---- cluster search ----

func TestClusterSearch_ByName_Found(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "search", "test-cluster")
	if err != nil {
		t.Fatalf("cluster search: %v", err)
	}
	if !strings.Contains(out, clusterID) {
		t.Errorf("expected cluster ID in search output, got: %q", out)
	}
	// Verify cluster-id persisted to state.
	stateRaw, _ := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if !strings.Contains(string(stateRaw), clusterID) {
		t.Errorf("cluster-id not persisted after search: %q", string(stateRaw))
	}
}

func TestClusterSearch_ByName_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, emptyListJSON)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "search", "nonexistent")
	if err != nil {
		t.Fatalf("cluster search not-found should exit 0, got: %v", err)
	}
	if !strings.Contains(out, "[]") {
		t.Errorf("expected empty JSON array for not-found search, got: %q", out)
	}
}

func TestClusterSearch_ByName_Multiple(t *testing.T) {
	multiJSON := `{"items":[` + clusterJSON + `,` + clusterJSON + `],"kind":"ClusterList","page":1,"size":2,"total":2}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, multiJSON)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "search", "test-cluster")
	if err != nil {
		t.Fatalf("cluster search multiple: %v", err)
	}
	if !strings.Contains(out, clusterID) {
		t.Errorf("expected cluster ID in output, got: %q", out)
	}
}

func TestClusterSearch_NoArgs_WithState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	out, err := runClusterCmd(t, dir, "cluster", "search")
	if err != nil {
		t.Fatalf("cluster search (no args, with state): %v", err)
	}
	if !strings.Contains(out, clusterID) {
		t.Errorf("expected cluster JSON in output, got: %q", out)
	}
}

func TestClusterSearch_NoArgs_WithoutState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no HTTP request should be made when state is empty")
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	// No cluster-id in state.
	_, err := runClusterCmd(t, dir, "cluster", "search")
	if err == nil {
		t.Fatal("expected error when no cluster-id in state")
	}
	if !strings.Contains(err.Error(), "No cluster-id set in state") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- cluster patch ----

func TestClusterPatch_Spec(t *testing.T) {
	var capturedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
		case r.Method == http.MethodPatch && r.URL.Path == apiPrefix+"/clusters/"+clusterID:
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	_, err := runClusterCmd(t, dir, "cluster", "patch", "spec")
	if err != nil {
		t.Fatalf("cluster patch spec: %v", err)
	}
	if capturedBody == nil {
		t.Fatal("expected PATCH request body")
	}
	spec, ok := capturedBody["spec"].(map[string]any)
	if !ok {
		t.Fatalf("expected spec in PATCH body, got: %v", capturedBody)
	}
	// Fixture spec has no counter (treated as 0), so first patch sends "1".
	if spec["counter"] != "1" {
		t.Errorf("expected counter '1' (0→1), got %v", spec["counter"])
	}
}

func TestClusterPatch_Labels(t *testing.T) {
	var capturedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
		case r.Method == http.MethodPatch && r.URL.Path == apiPrefix+"/clusters/"+clusterID:
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	_, err := runClusterCmd(t, dir, "cluster", "patch", "labels")
	if err != nil {
		t.Fatalf("cluster patch labels: %v", err)
	}
	if capturedBody == nil {
		t.Fatal("expected PATCH request body")
	}
	labels, ok := capturedBody["labels"].(map[string]any)
	if !ok {
		t.Fatalf("expected labels in PATCH body, got: %v", capturedBody)
	}
	if labels["counter"] != "2" {
		t.Errorf("expected counter '2', got %v", labels["counter"])
	}
}

func TestClusterPatch_NoArgs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no HTTP request should be made for patch with no args")
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	out, err := runClusterCmd(t, dir, "cluster", "patch")
	if err == nil {
		t.Fatal("expected error for cluster patch with no args")
	}
	if !strings.Contains(out, "Usage: hf cluster patch") {
		t.Errorf("expected usage message in stdout, got: %q", out)
	}
}

func TestClusterCreate_PositionalArgs(t *testing.T) {
	resetClusterFlags()
	var capturedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyListJSON)
		case r.Method == http.MethodPost && r.URL.Path == apiPrefix+"/clusters":
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, clusterJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	_, err := runClusterCmd(t, dir, "cluster", "create", "my-named-cluster", "eu-west1", "1.30")
	if err != nil {
		t.Fatalf("cluster create with positional args: %v", err)
	}
	if capturedBody == nil {
		t.Fatal("expected POST request body")
	}
	if capturedBody["name"] != "my-named-cluster" {
		t.Errorf("expected name 'my-named-cluster', got %v", capturedBody["name"])
	}
	spec, _ := capturedBody["spec"].(map[string]any)
	if spec["region"] != "eu-west1" {
		t.Errorf("expected region 'eu-west1', got %v", spec["region"])
	}
	if spec["version"] != "1.30" {
		t.Errorf("expected version '1.30', got %v", spec["version"])
	}
}

func TestClusterAdapterPostStatus(t *testing.T) {
	var capturedBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/statuses" {
			capturedBody, _ = io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"adapter": "cl-deployment",
				"observed_generation": 3,
				"observed_time": "2026-05-10T00:00:00Z",
				"conditions": [
					{"type": "Available", "status": "True", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"},
					{"type": "Applied",   "status": "True", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"},
					{"type": "Health",    "status": "True", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"},
					{"type": "Finalized", "status": "True", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"}
				],
				"created_time": "2026-05-10T00:00:00Z",
				"last_report_time": "2026-05-10T00:00:00Z"
			}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	out, err := runClusterCmd(t, dir, "cluster", "adapter", "post-status", "cl-deployment", "True", "3")
	if err != nil {
		t.Fatalf("cluster adapter post-status: %v", err)
	}
	if !strings.Contains(out, "cl-deployment") {
		t.Errorf("expected adapter name in output, got: %q", out)
	}
	if len(capturedBody) == 0 {
		t.Fatal("expected request body to be captured")
	}
	var body map[string]any
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("captured body is not JSON: %v", err)
	}
	if body["adapter"] != "cl-deployment" {
		t.Errorf("body adapter: got %v, want cl-deployment", body["adapter"])
	}
	conds, ok := body["conditions"].([]any)
	if !ok || len(conds) != 4 {
		t.Errorf("expected 4 conditions, got %v", body["conditions"])
	}
}

func TestClusterAdapterPostStatus_InvalidStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("HTTP request should not be made for invalid status")
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	_, err := runClusterCmd(t, dir, "cluster", "adapter", "post-status", "cl-deployment", "INVALID", "3")
	if err == nil {
		t.Fatal("expected error for invalid status value")
	}
	if !strings.Contains(err.Error(), "Invalid status value") {
		t.Errorf("expected 'Invalid status value' error, got: %v", err)
	}
}

// ---- cluster create: template-driven ----

func TestClusterCreate_Template_NameFromPositionalArg(t *testing.T) {
	resetClusterFlags()
	var capturedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyListJSON)
		case r.Method == http.MethodPost:
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, clusterJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	_, err := runClusterCmd(t, dir, "cluster", "create", "prod-cluster")
	if err != nil {
		t.Fatalf("cluster create with name arg: %v", err)
	}
	if capturedBody["name"] != "prod-cluster" {
		t.Errorf("expected name=prod-cluster, got %v", capturedBody["name"])
	}
}

func TestClusterCreate_FileFlag(t *testing.T) {
	resetClusterFlags()
	var capturedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyListJSON)
		case r.Method == http.MethodPost:
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, clusterJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)

	// Write a custom template file.
	tplPath := filepath.Join(dir, "custom-cluster.json")
	if err := os.WriteFile(tplPath,
		[]byte(`{"kind":"Cluster","name":"from-file","labels":{"env":"ci"},"spec":{"counter":"1","region":"ap-east-1","version":"5.0.0"}}`),
		0600); err != nil {
		t.Fatal(err)
	}

	_, err := runClusterCmd(t, dir, "cluster", "create", "-f", tplPath)
	if err != nil {
		t.Fatalf("cluster create -f: %v", err)
	}
	if capturedBody["name"] != "from-file" {
		t.Errorf("expected name=from-file, got %v", capturedBody["name"])
	}
	labels, _ := capturedBody["labels"].(map[string]any)
	if labels["env"] != "ci" {
		t.Errorf("expected labels.env=ci, got %v", labels["env"])
	}

	// Config-dir template should NOT have been created.
	if _, err := os.Stat(filepath.Join(dir, "cluster-template.json")); !os.IsNotExist(err) {
		t.Error("config-dir cluster-template.json should not be created when -f is used")
	}
}

func TestClusterCreate_MalformedTemplate(t *testing.T) {
	resetClusterFlags()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no HTTP request should be made for malformed template")
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	tplPath := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(tplPath, []byte(`{not json`), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := runClusterCmd(t, dir, "cluster", "create", "-f", tplPath)
	if err == nil {
		t.Fatal("expected error for malformed template")
	}
	if !strings.Contains(err.Error(), "loading template") {
		t.Errorf("expected 'loading template' in error, got: %v", err)
	}
}

// ---- cluster list watch flags ----

func TestClusterListWatchFlagRegistered(t *testing.T) {
	f := clusterListCmd.Flags().Lookup("watch")
	if f == nil {
		t.Fatal("--watch flag not registered on clusterListCmd")
	}
	if f.DefValue != "false" {
		t.Errorf("--watch default = %q, want %q", f.DefValue, "false")
	}
}

func TestClusterListSecondsFlagRegistered(t *testing.T) {
	f := clusterListCmd.Flags().Lookup("seconds")
	if f == nil {
		t.Fatal("-s/--seconds flag not registered on clusterListCmd")
	}
	if f.DefValue != "5" {
		t.Errorf("-s default = %q, want %q", f.DefValue, "5")
	}
}

// ---- cluster id ----

func TestClusterID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected API call: %s %s", r.Method, r.URL.Path)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	clusterID := "019dc049-e79e-72a9-94f8-0056a11193cd"

	t.Run("prints ID when set", func(t *testing.T) {
		setClusterIDInState(t, dir, "test", clusterID)
		out, err := runClusterCmd(t, dir, "cluster", "id")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.TrimSpace(out) != clusterID {
			t.Errorf("got %q, want %q", strings.TrimSpace(out), clusterID)
		}
	})

	t.Run("errors when not set", func(t *testing.T) {
		setClusterIDInState(t, dir, "test", "")
		_, err := runClusterCmd(t, dir, "cluster", "id")
		if err == nil {
			t.Fatal("expected error when cluster-id is not set")
		}
		if !strings.Contains(err.Error(), "No cluster-id set in state") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

// ---- cluster id --interactive ----

const secondClusterID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

var clusterListTwoJSON = `{
  "items": [` + clusterJSON + `, {
    "id": "` + secondClusterID + `",
    "kind": "Cluster",
    "name": "second-cluster",
    "generation": 1,
    "labels": {},
    "spec": {},
    "status": {"conditions": []},
    "created_by": "user@example.com",
    "created_time": "2026-05-10T00:00:00Z",
    "href": "/api/hyperfleet/v1/clusters/` + secondClusterID + `"
  }],
  "kind": "ClusterList",
  "page": 1,
  "size": 2,
  "total": 2
}`

func TestClusterIDInteractive_Select(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterListTwoJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	orig := clusterIDSel
	clusterIDSel = mockSel{idx: 1} // pick second cluster
	t.Cleanup(func() { clusterIDSel = orig })

	out, err := runClusterCmd(t, dir, "cluster", "id", "--interactive")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Active cluster set to:") {
		t.Errorf("expected confirmation message, got: %q", out)
	}
	if !strings.Contains(out, secondClusterID) {
		t.Errorf("expected selected cluster ID in output, got: %q", out)
	}
	state, err := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if err != nil {
		t.Fatalf("reading state.yaml: %v", err)
	}
	if !strings.Contains(string(state), secondClusterID) {
		t.Errorf("cluster-id %q not found in state.yaml:\n%s", secondClusterID, state)
	}
}

func TestClusterIDInteractive_Abort(t *testing.T) {
	apiCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters" {
			apiCalled = true
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID) // seed an existing ID

	orig := clusterIDSel
	clusterIDSel = mockSel{idx: -1} // abort
	t.Cleanup(func() { clusterIDSel = orig })

	out, err := runClusterCmd(t, dir, "cluster", "id", "--interactive")
	if err != nil {
		t.Fatalf("unexpected error on abort: %v", err)
	}
	if out != "" {
		t.Errorf("expected no output on abort, got: %q", out)
	}
	if !apiCalled {
		t.Error("expected API to be called even on abort")
	}
	// Original cluster-id must be preserved.
	state, _ := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if !strings.Contains(string(state), clusterID) {
		t.Errorf("original cluster-id should be preserved in state.yaml:\n%s", state)
	}
}

// ---- cluster get/delete --interactive ----

func TestPickClusterInteractive_Select(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterListTwoJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+secondClusterID:
			w.Header().Set("Content-Type", "application/json")
			// Return a minimal cluster for the get call that follows picker selection
			fmt.Fprintf(w, `{"id":%q,"kind":"Cluster","name":"second-cluster","generation":1,"labels":{},"spec":{},"status":{"conditions":[]},"created_by":"u","created_time":"2026-05-10T00:00:00Z","href":""}`, secondClusterID)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	orig := clusterIDSel
	clusterIDSel = mockSel{idx: 1} // pick second cluster
	t.Cleanup(func() { clusterIDSel = orig })

	_, err := runClusterCmd(t, dir, "cluster", "get", "-i")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	state, _ := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if !strings.Contains(string(state), secondClusterID) {
		t.Errorf("cluster-id %q not found in state.yaml:\n%s", secondClusterID, state)
	}
}

func TestPickClusterInteractive_Abort(t *testing.T) {
	listCalled := false
	getCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters" {
			listCalled = true
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterListJSON)
			return
		}
		getCalled = true
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	orig := clusterIDSel
	clusterIDSel = mockSel{idx: -1} // abort
	t.Cleanup(func() { clusterIDSel = orig })

	_, err := runClusterCmd(t, dir, "cluster", "get", "-i")
	if err != nil {
		t.Fatalf("unexpected error on abort: %v", err)
	}
	if !listCalled {
		t.Error("expected list API to be called")
	}
	if getCalled {
		t.Error("expected no cluster GET after abort")
	}
}

func TestClusterGetInteractive(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterListJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	orig := clusterIDSel
	clusterIDSel = mockSel{idx: 0}
	t.Cleanup(func() { clusterIDSel = orig })

	out, err := runClusterCmd(t, dir, "cluster", "get", "-i")
	if err != nil {
		t.Fatalf("cluster get -i: %v", err)
	}
	if !strings.Contains(out, clusterID) {
		t.Errorf("expected cluster ID in output, got: %q", out)
	}
}

func TestClusterDeleteInteractive(t *testing.T) {
	deleteCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterListJSON)
		case r.Method == http.MethodDelete && r.URL.Path == apiPrefix+"/clusters/"+clusterID:
			deleteCalled = true
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, clusterJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	orig := clusterIDSel
	clusterIDSel = mockSel{idx: 0}
	t.Cleanup(func() { clusterIDSel = orig })

	_, err := runClusterCmd(t, dir, "cluster", "delete", "-i")
	if err != nil {
		t.Fatalf("cluster delete -i: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called with picked cluster ID")
	}
}

func TestClusterIDInteractive_Empty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	orig := clusterIDSel
	clusterIDSel = mockSel{idx: 0}
	t.Cleanup(func() { clusterIDSel = orig })

	_, err := runClusterCmd(t, dir, "cluster", "id", "--interactive")
	if err == nil {
		t.Fatal("expected error when no clusters available")
	}
	if !strings.Contains(err.Error(), "no clusters available") {
		t.Errorf("unexpected error message: %v", err)
	}
}
