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

// setNodepoolIDInState writes state.yaml with the active env and nodepool-id.
func setNodepoolIDInState(t *testing.T, dir, envName, id string) {
	t.Helper()
	statePath := filepath.Join(dir, "state.yaml")
	content := fmt.Sprintf("active-environment: %s\nnodepool-id: %s\n", envName, id)
	if err := os.WriteFile(statePath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

// setupNodepoolEnv creates a temp dir with an env pointing to ts.URL and activates it.
func setupNodepoolEnv(t *testing.T, ts *httptest.Server) string {
	t.Helper()
	dir := t.TempDir()
	makeEnv(t, dir, "test", ts.URL)
	setActiveEnv(t, dir, "test")
	return dir
}

// resetNodepoolFlags resets all nodepool and global flag vars to defaults.
func resetNodepoolFlags() {
	outputFmt = "json"
	noColor = false
	verbose = false
	nodepoolCreateName = ""
	nodepoolCreateType = ""
	nodepoolCreateReplicas = 0
	nodepoolUpdateName = ""
	nodepoolUpdateReplicas = 0
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
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/nodepools" {
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

// ---- nodepool get ----

func TestNodepoolGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/nodepools/"+nodepoolID {
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
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/nodepools/"+nodepoolID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setNodepoolIDInState(t, dir, "test", nodepoolID)

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
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/nodepools":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyNPListJSON)
		case r.Method == http.MethodPost && r.URL.Path == apiPrefix+"/nodepools":
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
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/nodepools":
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
		if r.Method == http.MethodPatch && r.URL.Path == apiPrefix+"/nodepools/"+nodepoolID {
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
		if r.Method == http.MethodDelete && r.URL.Path == apiPrefix+"/nodepools/"+nodepoolID {
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
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/nodepools/"+nodepoolID {
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
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/nodepools/"+nodepoolID {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodepoolJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setNodepoolIDInState(t, dir, "test", nodepoolID)

	_, err := runNodepoolCmd(t, dir, "nodepool", "conditions")
	if err != nil {
		t.Fatalf("nodepool conditions (from state): %v", err)
	}
}

func TestNodepoolConditions_Table(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/nodepools/"+nodepoolID {
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
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/nodepools/"+nodepoolID+"/statuses" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, npAdapterStatusesJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setNodepoolIDInState(t, dir, "test", nodepoolID)

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
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/nodepools/"+nodepoolID+"/statuses" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, npAdapterStatusesJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupNodepoolEnv(t, ts)
	setNodepoolIDInState(t, dir, "test", nodepoolID)

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
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/nodepools/"+nodepoolID+"/statuses" {
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
