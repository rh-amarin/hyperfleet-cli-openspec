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
)


// ---- test fixtures ----

const nodepoolID = "019dc049-e76c-7be1-b201-0db50e2c8ecb"

var nodepoolJSON = `{
  "id": "` + nodepoolID + `",
  "kind": "NodePool",
  "name": "test-nodepool",
  "generation": 1,
  "labels": {"counter": "1"},
  "spec": {"counter": "1", "platform": {"type": "n2-standard-4"}, "replicas": 2},
  "status": {
    "conditions": [
      {"type": "Available",  "status": "True", "last_transition_time": "2026-05-10T00:00:00Z", "observed_generation": 1},
      {"type": "Reconciled", "status": "True", "last_transition_time": "2026-05-10T00:00:00Z", "observed_generation": 1}
    ]
  },
  "owner_references": {"id": "` + clusterID + `", "kind": "Cluster", "href": "/api/hyperfleet/v1/clusters/` + clusterID + `"},
  "created_by": "user@example.com",
  "created_time": "2026-05-10T00:00:00Z",
  "href": "/api/hyperfleet/v1/nodepools/` + nodepoolID + `"
}`

var nodepoolListJSON = `{
  "items": [` + nodepoolJSON + `],
  "kind": "NodePoolList",
  "page": 1,
  "size": 1,
  "total": 1
}`

var emptyNPListJSON = `{"items": [], "kind": "NodePoolList", "page": 1, "size": 0, "total": 0}`

var npAdapterStatusesJSON = `{
  "items": [
    {
      "adapter": "np-configmap",
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

// setBothIDsInState writes state.yaml with active-environment, cluster-id, and nodepool-id.
func setBothIDsInState(t *testing.T, dir, envName, cid, npid string) {
	t.Helper()
	statePath := filepath.Join(dir, "state.yaml")
	content := fmt.Sprintf("active-environment: %s\ncluster-id: %s\nnodepool-id: %s\n", envName, cid, npid)
	if err := os.WriteFile(statePath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

// setNodepoolIDInState writes state.yaml with the active env, cluster-id, and nodepool-id.
func setNodepoolIDInState(t *testing.T, dir, envName, id string) {
	t.Helper()
	setBothIDsInState(t, dir, envName, clusterID, id)
}

// setupNodepoolEnv creates a temp dir with an env pointing to ts.URL, activates it,
// and writes cluster-id to state (required for cluster-scoped API paths).
func setupNodepoolEnv(t *testing.T, ts *httptest.Server) string {
	t.Helper()
	dir := t.TempDir()
	makeEnv(t, dir, "test", ts.URL)
	setActiveEnv(t, dir, "test")
	setBothIDsInState(t, dir, "test", clusterID, "")
	return dir
}

// resetNodepoolFlags resets all nodepool and global flag vars to defaults.
func resetNodepoolFlags() {
	outputFmt = "json"
	noColor = false
	verbose = false
	nodepoolCreateName = ""
	nodepoolCreateFile = ""
	nodepoolCreateType = ""
	nodepoolCreateReplicas = 0
	nodepoolUpdateName = ""
	nodepoolUpdateReplicas = 0
	nodepoolListWatch = false
	nodepoolListWatchSecs = 5
	nodepoolInteractive = false
	nodepoolListSearch = ""
	nodepoolIDInteractive = false
	nodepoolDeleteForce = false
	nodepoolDeleteReason = ""
}

// runNodepoolCmd wraps runCmd and resets all nodepool flag state before each invocation.
func runNodepoolCmd(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	resetNodepoolFlags()
	return runCmd(t, dir, args...)
}

// ---- nodepool list ----

func TestNodepoolList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "list")
	if err != nil {
		t.Fatalf("nodepool list: %v", err)
	}
	if !strings.Contains(out, nodepoolID) {
		t.Errorf("expected nodepool ID in output, got: %q", out)
	}
	if !strings.Contains(out, `"items"`) {
		t.Errorf("expected JSON list response, got: %q", out)
	}
}

func TestNodepoolList_Table(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, nodepoolListJSON)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "list", "--output", "table")
	if err != nil {
		t.Fatalf("nodepool list --output table: %v", err)
	}
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") {
		t.Errorf("expected table headers, got: %q", out)
	}
	if !strings.Contains(out, "test-nodepool") {
		t.Errorf("expected nodepool name in table, got: %q", out)
	}
	if !strings.Contains(out, "TYPE") || !strings.Contains(out, "REPLICAS") {
		t.Errorf("expected TYPE and REPLICAS columns, got: %q", out)
	}
}

// ---- nodepool table ----

func TestNodepoolTable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "table")
	if err != nil {
		t.Fatalf("nodepool table: %v", err)
	}
	for _, header := range []string{"ID", "NAME", "TYPE", "GEN", "REPLICAS", "STATUS"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output, got: %q", header, out)
		}
	}
	if !strings.Contains(out, "test-nodepool") {
		t.Errorf("expected nodepool name in table output, got: %q", out)
	}
}

// ---- nodepool list --search ----

func TestNodepoolList_Search_NoFlag(t *testing.T) {
	var receivedQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, nodepoolListJSON)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	_, err := runNodepoolCmd(t, dir, "nodepool", "list")
	if err != nil {
		t.Fatalf("nodepool list (no flag): %v", err)
	}
	if receivedQuery != "" {
		t.Errorf("expected no query string, got: %q", receivedQuery)
	}
}

func TestNodepoolList_Search_OwnerID(t *testing.T) {
	var receivedQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, nodepoolListJSON)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	expr := "owner_id='" + clusterID + "'"
	_, err := runNodepoolCmd(t, dir, "nodepool", "list", "--search", expr)
	if err != nil {
		t.Fatalf("nodepool list --search owner_id: %v", err)
	}
	if !strings.HasPrefix(receivedQuery, "search=") {
		t.Errorf("expected query to start with 'search=', got: %q", receivedQuery)
	}
	if strings.Contains(receivedQuery, "'") {
		t.Errorf("single quotes should be URL-encoded, got: %q", receivedQuery)
	}
}

func TestNodepoolList_Search_CompoundExpr(t *testing.T) {
	var receivedQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, nodepoolListJSON)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	expr := "labels.role='worker' and name='test-nodepool'"
	_, err := runNodepoolCmd(t, dir, "nodepool", "list", "--search", expr)
	if err != nil {
		t.Fatalf("nodepool list --search compound: %v", err)
	}
	if receivedQuery == "" || receivedQuery == "search="+expr {
		t.Errorf("expected URL-encoded query, got: %q", receivedQuery)
	}
	if !strings.HasPrefix(receivedQuery, "search=") {
		t.Errorf("expected query to start with 'search=', got: %q", receivedQuery)
	}
}

func TestNodepoolList_Search_APIError400(t *testing.T) {
	errBody := `{"type":"about:blank","title":"Validation Error","status":400,"detail":"invalid search expression","code":"HYPERFLEET-VAL-001"}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errBody)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "list", "--search", "not status.conditions.Ready='True'")
	if err != nil {
		t.Fatalf("nodepool list --search API 400 should exit 0, got error: %v", err)
	}
	if !strings.Contains(out, "400") && !strings.Contains(out, "Validation Error") {
		t.Errorf("expected error JSON in output, got: %q", out)
	}
}

// ---- nodepool get ----

func TestNodepoolGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "get", nodepoolID)
	if err != nil {
		t.Fatalf("nodepool get: %v", err)
	}
	if !strings.Contains(out, nodepoolID) {
		t.Errorf("expected nodepool ID in output, got: %q", out)
	}
	if !strings.Contains(out, `"kind"`) {
		t.Errorf("expected JSON nodepool object, got: %q", out)
	}
}

func TestNodepoolGet_FromState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID)

	out, err := runNodepoolCmd(t, dir, "nodepool", "get")
	if err != nil {
		t.Fatalf("nodepool get (from state): %v", err)
	}
	if !strings.Contains(out, nodepoolID) {
		t.Errorf("expected nodepool ID in output, got: %q", out)
	}
}

func TestNodepoolGet_NotFound(t *testing.T) {
	notFoundJSON := `{"type":"https://api.hyperfleet.io/errors/not-found","title":"Resource Not Found","status":404,"detail":"NodePool not found","code":"HYPERFLEET-NTF-001"}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, notFoundJSON)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "get", "00000000-0000-0000-0000-000000000000")
	// API errors exit 0
	if err != nil {
		t.Fatalf("nodepool get 404 should exit 0, got error: %v", err)
	}
	if !strings.Contains(out, "Not Found") && !strings.Contains(out, "404") {
		t.Errorf("expected error JSON in output, got: %q", out)
	}
}

// ---- nodepool create ----

func TestNodepoolCreate(t *testing.T) {
	resetNodepoolFlags()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyNPListJSON)
		case r.Method == http.MethodPost && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, nodepoolJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "create", "--name", "test-nodepool")
	if err != nil {
		t.Fatalf("nodepool create: %v", err)
	}
	if !strings.Contains(out, nodepoolID) {
		t.Errorf("expected nodepool ID in output, got: %q", out)
	}

	// Verify nodepool-id persisted to state.yaml
	stateRaw, _ := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if !strings.Contains(string(stateRaw), nodepoolID) {
		t.Errorf("nodepool-id not persisted to state.yaml: %q", string(stateRaw))
	}
}

func TestNodepoolCreate_DuplicateGuard(t *testing.T) {
	resetNodepoolFlags()
	postCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolListJSON)
		case r.Method == http.MethodPost:
			postCalled = true
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, nodepoolJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	_, err := runNodepoolCmd(t, dir, "nodepool", "create", "--name", "test-nodepool")
	if err != nil {
		t.Fatalf("nodepool create duplicate: %v", err)
	}
	if postCalled {
		t.Error("POST should not have been called for duplicate nodepool")
	}
}

func TestNodepoolCreate_Defaults(t *testing.T) {
	resetNodepoolFlags()
	var capturedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyNPListJSON)
		case r.Method == http.MethodPost:
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, nodepoolJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	_, err := runNodepoolCmd(t, dir, "nodepool", "create")
	if err != nil {
		t.Fatalf("nodepool create defaults: %v", err)
	}
	if capturedBody != nil {
		if name, _ := capturedBody["name"].(string); name != "my-nodepool" {
			t.Errorf("expected default name 'my-nodepool', got %q", name)
		}
	}
}

// ---- nodepool update ----

func TestNodepoolUpdate(t *testing.T) {
	resetNodepoolFlags()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "update", nodepoolID, "--name", "new-name")
	if err != nil {
		t.Fatalf("nodepool update: %v", err)
	}
	if !strings.Contains(out, nodepoolID) {
		t.Errorf("expected nodepool ID in output, got: %q", out)
	}
}

// ---- nodepool delete ----

func TestNodepoolDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "delete", nodepoolID)
	if err != nil {
		t.Fatalf("nodepool delete: %v", err)
	}
	// Silent success — no output expected
	if strings.TrimSpace(out) != "" {
		t.Errorf("nodepool delete should be silent, got: %q", out)
	}
}

func TestNodepoolDelete_NotFound(t *testing.T) {
	notFoundJSON := `{"type":"https://api.hyperfleet.io/errors/not-found","title":"Resource Not Found","status":404,"detail":"NodePool not found"}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, notFoundJSON)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	_, err := runNodepoolCmd(t, dir, "nodepool", "delete", "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nodepool delete 404")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- nodepool conditions ----

func TestNodepoolConditions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "conditions", nodepoolID)
	if err != nil {
		t.Fatalf("nodepool conditions: %v", err)
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

func TestNodepoolConditions_FromState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID)

	_, err := runNodepoolCmd(t, dir, "nodepool", "conditions")
	if err != nil {
		t.Fatalf("nodepool conditions (from state): %v", err)
	}
}

func TestNodepoolConditions_Table(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "conditions", nodepoolID, "--output", "table")
	if err != nil {
		t.Fatalf("nodepool conditions --output table: %v", err)
	}
	if !strings.Contains(out, "TYPE") || !strings.Contains(out, "STATUS") {
		t.Errorf("expected table headers, got: %q", out)
	}
	if !strings.Contains(out, "Available") {
		t.Errorf("expected Available condition in table, got: %q", out)
	}
}

// ---- nodepool statuses ----

func TestNodepoolStatuses(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID+"/statuses" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, npAdapterStatusesJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID)

	out, err := runNodepoolCmd(t, dir, "nodepool", "statuses")
	if err != nil {
		t.Fatalf("nodepool statuses: %v", err)
	}
	if !strings.Contains(out, "np-configmap") {
		t.Errorf("expected adapter name in output, got: %q", out)
	}
	if !strings.Contains(out, `"AdapterStatusList"`) {
		t.Errorf("expected AdapterStatusList kind, got: %q", out)
	}
}

func TestNodepoolStatuses_Table(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID+"/statuses" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, npAdapterStatusesJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID)

	out, err := runNodepoolCmd(t, dir, "nodepool", "statuses", "--output", "table")
	if err != nil {
		t.Fatalf("nodepool statuses --output table: %v", err)
	}
	if !strings.Contains(out, "ADAPTER") || !strings.Contains(out, "GEN") {
		t.Errorf("expected table headers, got: %q", out)
	}
	if !strings.Contains(out, "np-configmap") {
		t.Errorf("expected adapter name in table, got: %q", out)
	}
}

func TestNodepoolStatuses_ExplicitID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID+"/statuses" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, npAdapterStatusesJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "statuses", nodepoolID)
	if err != nil {
		t.Fatalf("nodepool statuses (explicit id): %v", err)
	}
	if !strings.Contains(out, "np-configmap") {
		t.Errorf("expected adapter name in output, got: %q", out)
	}
}

// ---- nodepool search ----

func TestNodepoolSearch_ByName_Found(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "search", "test-nodepool")
	if err != nil {
		t.Fatalf("nodepool search: %v", err)
	}
	if !strings.Contains(out, nodepoolID) {
		t.Errorf("expected nodepool ID in search output, got: %q", out)
	}
	// Verify nodepool-id persisted to state.
	stateRaw, _ := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if !strings.Contains(string(stateRaw), nodepoolID) {
		t.Errorf("nodepool-id not persisted after search: %q", string(stateRaw))
	}
}

func TestNodepoolSearch_ByName_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, emptyNPListJSON)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "search", "nonexistent")
	if err != nil {
		t.Fatalf("nodepool search not-found should exit 0, got: %v", err)
	}
	if !strings.Contains(out, "[]") {
		t.Errorf("expected empty JSON array for not-found search, got: %q", out)
	}
}

func TestNodepoolSearch_NoArgs_WithState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID)

	out, err := runNodepoolCmd(t, dir, "nodepool", "search")
	if err != nil {
		t.Fatalf("nodepool search (no args, with state): %v", err)
	}
	if !strings.Contains(out, nodepoolID) {
		t.Errorf("expected nodepool JSON in output, got: %q", out)
	}
}

func TestNodepoolSearch_NoArgs_WithoutState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no HTTP request should be made when state is empty")
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	_, err := runNodepoolCmd(t, dir, "nodepool", "search")
	if err == nil {
		t.Fatal("expected error when no nodepool-id in state")
	}
	if !strings.Contains(err.Error(), "No nodepool-id set in state") {
		t.Errorf("error message: got %q", err.Error())
	}
}

// ---- nodepool patch ----

func TestNodepoolPatch_Spec(t *testing.T) {
	var capturedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
		case r.Method == http.MethodPatch && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID:
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID)

	_, err := runNodepoolCmd(t, dir, "nodepool", "patch", "spec")
	if err != nil {
		t.Fatalf("nodepool patch spec: %v", err)
	}
	if capturedBody == nil {
		t.Fatal("expected PATCH request body")
	}
	spec, ok := capturedBody["spec"].(map[string]any)
	if !ok {
		t.Fatalf("expected spec in PATCH body, got: %v", capturedBody)
	}
	if spec["counter"] != "2" {
		t.Errorf("expected counter '2', got %v", spec["counter"])
	}
}

func TestNodepoolPatch_Labels(t *testing.T) {
	var capturedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
		case r.Method == http.MethodPatch && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID:
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID)

	_, err := runNodepoolCmd(t, dir, "nodepool", "patch", "labels")
	if err != nil {
		t.Fatalf("nodepool patch labels: %v", err)
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

func TestNodepoolPatch_NoArgs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no HTTP request should be made for patch with no args")
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	out, err := runNodepoolCmd(t, dir, "nodepool", "patch")
	if err == nil {
		t.Fatal("expected error for nodepool patch with no args")
	}
	if !strings.Contains(out, "Usage: hf nodepool patch") {
		t.Errorf("expected usage message in stdout, got: %q", out)
	}
}

// ---- nodepool adapter post-status ----

func TestNodePoolAdapterPostStatus(t *testing.T) {
	var capturedBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := apiPrefix + "/clusters/" + clusterID + "/nodepools/" + nodepoolID + "/statuses"
		if r.Method == http.MethodPut && r.URL.Path == wantPath {
			capturedBody, _ = io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"adapter": "np-configmap",
				"observed_generation": 2,
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

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID)

	out, err := runNodepoolCmd(t, dir, "nodepool", "adapter", "post-status", "np-configmap", "True", "2")
	if err != nil {
		t.Fatalf("nodepool adapter post-status: %v", err)
	}
	if !strings.Contains(out, "np-configmap") {
		t.Errorf("expected adapter name in output, got: %q", out)
	}
	if len(capturedBody) == 0 {
		t.Fatal("expected request body to be captured")
	}
	var body map[string]any
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("captured body is not JSON: %v", err)
	}
	if body["adapter"] != "np-configmap" {
		t.Errorf("body adapter: got %v, want np-configmap", body["adapter"])
	}
	conds, ok := body["conditions"].([]any)
	if !ok || len(conds) != 4 {
		t.Errorf("expected 4 conditions, got %v", body["conditions"])
	}
}

// ---- nodepool create: template-driven ----

func TestNodepoolCreate_Template_NameFromPositionalArg(t *testing.T) {
	resetNodepoolFlags()
	var capturedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyNPListJSON)
		case r.Method == http.MethodPost:
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, nodepoolJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	_, err := runNodepoolCmd(t, dir, "nodepool", "create", "workers")
	if err != nil {
		t.Fatalf("nodepool create with name arg: %v", err)
	}
	if capturedBody["name"] != "workers" {
		t.Errorf("expected name=workers, got %v", capturedBody["name"])
	}
}

func TestNodepoolCreate_FileFlag(t *testing.T) {
	resetNodepoolFlags()
	var capturedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyNPListJSON)
		case r.Method == http.MethodPost:
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, nodepoolJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)

	tplPath := filepath.Join(dir, "custom-nodepool.json")
	if err := os.WriteFile(tplPath,
		[]byte(`{"kind":"NodePool","name":"from-file","labels":{"env":"ci"},"spec":{"counter":"1","platform":{"type":"n4"},"replicas":3}}`),
		0600); err != nil {
		t.Fatal(err)
	}

	_, err := runNodepoolCmd(t, dir, "nodepool", "create", "-f", tplPath)
	if err != nil {
		t.Fatalf("nodepool create -f: %v", err)
	}
	if capturedBody["name"] != "from-file" {
		t.Errorf("expected name=from-file, got %v", capturedBody["name"])
	}
	spec, _ := capturedBody["spec"].(map[string]any)
	platform, _ := spec["platform"].(map[string]any)
	if platform["type"] != "n4" {
		t.Errorf("expected platform.type=n4, got %v", platform["type"])
	}

	// Config-dir template should NOT have been created.
	if _, err := os.Stat(filepath.Join(dir, "nodepool-template.json")); !os.IsNotExist(err) {
		t.Error("config-dir nodepool-template.json should not be created when -f is used")
	}
}

func TestNodepoolCreate_MalformedTemplate(t *testing.T) {
	resetNodepoolFlags()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no HTTP request should be made for malformed template")
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	tplPath := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(tplPath, []byte(`{not json`), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := runNodepoolCmd(t, dir, "nodepool", "create", "-f", tplPath)
	if err == nil {
		t.Fatal("expected error for malformed template")
	}
	if !strings.Contains(err.Error(), "loading template") {
		t.Errorf("expected 'loading template' in error, got: %v", err)
	}
}

// ---- nodepool list watch flags ----

func TestNodepoolListWatchFlagRegistered(t *testing.T) {
	f := nodepoolListCmd.Flags().Lookup("watch")
	if f == nil {
		t.Fatal("--watch flag not registered on nodepoolListCmd")
	}
	if f.DefValue != "false" {
		t.Errorf("--watch default = %q, want %q", f.DefValue, "false")
	}
}

func TestNodepoolListSecondsFlagRegistered(t *testing.T) {
	f := nodepoolListCmd.Flags().Lookup("seconds")
	if f == nil {
		t.Fatal("-s/--seconds flag not registered on nodepoolListCmd")
	}
	if f.DefValue != "5" {
		t.Errorf("-s default = %q, want %q", f.DefValue, "5")
	}
}

// ---- nodepool id ----

func TestNodepoolID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected API call: %s %s", r.Method, r.URL.Path)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)

	t.Run("prints ID when set", func(t *testing.T) {
		setBothIDsInState(t, dir, "test", clusterID, nodepoolID)
		out, err := runNodepoolCmd(t, dir, "nodepool", "id")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.TrimSpace(out) != nodepoolID {
			t.Errorf("got %q, want %q", strings.TrimSpace(out), nodepoolID)
		}
	})

	t.Run("errors when not set", func(t *testing.T) {
		setBothIDsInState(t, dir, "test", clusterID, "")
		_, err := runNodepoolCmd(t, dir, "nodepool", "id")
		if err == nil {
			t.Fatal("expected error when nodepool-id is not set")
		}
		if !strings.Contains(err.Error(), "No nodepool-id set in state") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

// ---- nodepool id --interactive ----

const secondNodepoolID = "bbbbbbbb-cccc-dddd-eeee-ffffffffffff"

var nodepoolListTwoJSON = `{
  "items": [` + nodepoolJSON + `, {
    "id": "` + secondNodepoolID + `",
    "kind": "NodePool",
    "name": "second-nodepool",
    "generation": 1,
    "labels": {},
    "spec": {},
    "status": {"conditions": []},
    "owner_references": {"id": "` + clusterID + `", "kind": "Cluster", "href": ""},
    "created_by": "user@example.com",
    "created_time": "2026-05-10T00:00:00Z",
    "href": "/api/hyperfleet/v1/clusters/` + clusterID + `/nodepools/` + secondNodepoolID + `"
  }],
  "kind": "NodePoolList",
  "page": 1,
  "size": 2,
  "total": 2
}`

// ---- nodepool get/delete --interactive ----

func TestPickNodepoolInteractive_Select(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolListTwoJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+secondNodepoolID:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"id":%q,"kind":"NodePool","name":"second-nodepool","generation":1,"labels":{},"spec":{},"status":{"conditions":[]},"owner_references":{"id":%q,"kind":"Cluster","href":""},"created_by":"u","created_time":"2026-05-10T00:00:00Z","href":""}`, secondNodepoolID, clusterID)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	orig := nodepoolIDSel
	nodepoolIDSel = mockSel{idx: 1} // pick second nodepool
	t.Cleanup(func() { nodepoolIDSel = orig })

	_, err := runNodepoolCmd(t, dir, "nodepool", "get", "-i")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	state, _ := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if !strings.Contains(string(state), secondNodepoolID) {
		t.Errorf("nodepool-id %q not found in state.yaml:\n%s", secondNodepoolID, state)
	}
}

func TestPickNodepoolInteractive_Abort(t *testing.T) {
	listCalled := false
	getCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools" {
			listCalled = true
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolListJSON)
			return
		}
		getCalled = true
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	orig := nodepoolIDSel
	nodepoolIDSel = mockSel{idx: -1} // abort
	t.Cleanup(func() { nodepoolIDSel = orig })

	_, err := runNodepoolCmd(t, dir, "nodepool", "get", "-i")
	if err != nil {
		t.Fatalf("unexpected error on abort: %v", err)
	}
	if !listCalled {
		t.Error("expected nodepool list API to be called")
	}
	if getCalled {
		t.Error("expected no nodepool GET after abort")
	}
}

func TestNodepoolGetInteractive(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolListJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	orig := nodepoolIDSel
	nodepoolIDSel = mockSel{idx: 0}
	t.Cleanup(func() { nodepoolIDSel = orig })

	out, err := runNodepoolCmd(t, dir, "nodepool", "get", "-i")
	if err != nil {
		t.Fatalf("nodepool get -i: %v", err)
	}
	if !strings.Contains(out, nodepoolID) {
		t.Errorf("expected nodepool ID in output, got: %q", out)
	}
}

func TestNodepoolDeleteInteractive(t *testing.T) {
	deleteCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolListJSON)
		case r.Method == http.MethodDelete && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID:
			deleteCalled = true
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	orig := nodepoolIDSel
	nodepoolIDSel = mockSel{idx: 0}
	t.Cleanup(func() { nodepoolIDSel = orig })

	_, err := runNodepoolCmd(t, dir, "nodepool", "delete", "-i")
	if err != nil {
		t.Fatalf("nodepool delete -i: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called with picked nodepool ID")
	}
}

func TestNodepoolIDInteractive_Select(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolListTwoJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	orig := nodepoolIDSel
	nodepoolIDSel = mockSel{idx: 1} // pick second nodepool
	t.Cleanup(func() { nodepoolIDSel = orig })

	out, err := runNodepoolCmd(t, dir, "nodepool", "id", "--interactive")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Active nodepool set to:") {
		t.Errorf("expected confirmation message, got: %q", out)
	}
	if !strings.Contains(out, secondNodepoolID) {
		t.Errorf("expected selected nodepool ID in output, got: %q", out)
	}
	state, err := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if err != nil {
		t.Fatalf("reading state.yaml: %v", err)
	}
	if !strings.Contains(string(state), secondNodepoolID) {
		t.Errorf("nodepool-id %q not found in state.yaml:\n%s", secondNodepoolID, state)
	}
}

func TestNodepoolIDInteractive_Abort(t *testing.T) {
	apiCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools" {
			apiCalled = true
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID) // seed an existing ID

	orig := nodepoolIDSel
	nodepoolIDSel = mockSel{idx: -1} // abort
	t.Cleanup(func() { nodepoolIDSel = orig })

	out, err := runNodepoolCmd(t, dir, "nodepool", "id", "--interactive")
	if err != nil {
		t.Fatalf("unexpected error on abort: %v", err)
	}
	if out != "" {
		t.Errorf("expected no output on abort, got: %q", out)
	}
	if !apiCalled {
		t.Error("expected API to be called even on abort")
	}
	// Original nodepool-id must be preserved.
	state, _ := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if !strings.Contains(string(state), nodepoolID) {
		t.Errorf("original nodepool-id should be preserved in state.yaml:\n%s", state)
	}
}

func TestNodepoolIDInteractive_NoCluster(t *testing.T) {
	apiCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		http.NotFound(w, r)
	}))
	defer ts.Close()

	// Setup env without cluster-id in state.
	dir := t.TempDir()
	makeEnv(t, dir, "test", ts.URL)
	setActiveEnv(t, dir, "test")

	orig := nodepoolIDSel
	nodepoolIDSel = mockSel{idx: 0}
	t.Cleanup(func() { nodepoolIDSel = orig })

	_, err := runNodepoolCmd(t, dir, "nodepool", "id", "--interactive")
	if err == nil {
		t.Fatal("expected error when cluster-id is not set")
	}
	if !strings.Contains(err.Error(), "No cluster-id set in state") {
		t.Errorf("unexpected error message: %v", err)
	}
	if apiCalled {
		t.Error("API should not be called when cluster-id is missing")
	}
}

// ---- nodepool delete --force and state fallback ----

func TestNodepoolDelete_FromState(t *testing.T) {
	var gotMethod, gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if r.Method == http.MethodDelete && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID)

	_, err := runNodepoolCmd(t, dir, "nodepool", "delete")
	if err != nil {
		t.Fatalf("nodepool delete (from state): %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if gotPath != apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID {
		t.Errorf("unexpected path: %s", gotPath)
	}
}

func TestNodepoolDelete_NoID_NoState_Errors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	// state has cluster-id but no nodepool-id

	_, err := runNodepoolCmd(t, dir, "nodepool", "delete")
	if err == nil {
		t.Fatal("expected error when no nodepool-id in state and no arg")
	}
	if !strings.Contains(err.Error(), "no active nodepool") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNodepoolDelete_Force(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if r.Method == http.MethodPost && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID+"/force-delete" {
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID)

	_, err := runNodepoolCmd(t, dir, "nodepool", "delete", "--force")
	if err != nil {
		t.Fatalf("nodepool delete --force: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotPath != apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID+"/force-delete" {
		t.Errorf("unexpected path: %s", gotPath)
	}
}

func TestNodepoolDelete_Force_WithReason(t *testing.T) {
	var gotBody map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID+"/force-delete" {
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID)

	_, err := runNodepoolCmd(t, dir, "nodepool", "delete", "--force", "--reason", "stuck in finalizing")
	if err != nil {
		t.Fatalf("nodepool delete --force --reason: %v", err)
	}
	if gotBody["reason"] != "stuck in finalizing" {
		t.Errorf("expected reason 'stuck in finalizing', got %q", gotBody["reason"])
	}
}

func TestNodepoolDelete_NoForce_StillCallsDelete(t *testing.T) {
	var gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		if r.Method == http.MethodDelete && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setBothIDsInState(t, dir, "test", clusterID, nodepoolID)

	_, err := runNodepoolCmd(t, dir, "nodepool", "delete")
	if err != nil {
		t.Fatalf("nodepool delete (no force): %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
}
