package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// clusterBetaID is a second cluster used in combined table tests.
const clusterBetaID = "019dc049-5096-7f33-af06-8efe296e9e25"

// nodepoolBetaID is a nodepool belonging to clusterBetaID.
const nodepoolBetaID = "019dc049-e79e-72a9-94f8-0056a11193cd"

// resourcesListJSON returns two clusters: alpha and beta.
var resourcesClusterListJSON = `{
  "items": [
    {
      "id": "` + clusterID + `",
      "kind": "Cluster",
      "name": "test-cluster-alpha",
      "generation": 3,
      "labels": {},
      "spec": {},
      "status": {
        "conditions": [
          {"type": "Available",  "status": "False", "last_transition_time": "2026-05-10T00:00:00Z", "observed_generation": 3},
          {"type": "Reconciled", "status": "False", "last_transition_time": "2026-05-10T00:00:00Z", "observed_generation": 3}
        ]
      },
      "created_time": "2026-05-10T00:00:00Z",
      "href": "/api/hyperfleet/v1/clusters/` + clusterID + `"
    },
    {
      "id": "` + clusterBetaID + `",
      "kind": "Cluster",
      "name": "test-cluster-beta",
      "generation": 1,
      "labels": {},
      "spec": {},
      "status": {
        "conditions": [
          {"type": "Available",  "status": "False", "last_transition_time": "2026-05-10T00:00:00Z", "observed_generation": 1},
          {"type": "Reconciled", "status": "False", "last_transition_time": "2026-05-10T00:00:00Z", "observed_generation": 1}
        ]
      },
      "created_time": "2026-05-10T00:00:00Z",
      "href": "/api/hyperfleet/v1/clusters/` + clusterBetaID + `"
    }
  ],
  "kind": "ClusterList",
  "page": 1,
  "size": 2,
  "total": 2
}`

var alphaAdapterStatusesJSON = `{
  "items": [
    {
      "adapter": "cl-deployment",
      "observed_generation": 3,
      "conditions": [
        {"type": "Available", "status": "True", "last_transition_time": "2026-05-10T00:00:00Z"},
        {"type": "Finalized", "status": "True", "last_transition_time": "2026-05-10T00:00:00Z"}
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

var emptyAdapterStatusesJSON = `{"items": [], "kind": "AdapterStatusList", "page": 1, "size": 0, "total": 0}`

var alphaNodepoolListJSON = `{
  "items": [
    {
      "id": "` + nodepoolID + `",
      "kind": "NodePool",
      "name": "workers-1",
      "generation": 2,
      "spec": {"platform": {"type": "n2-standard-4"}, "replicas": 1},
      "status": {
        "conditions": [
          {"type": "Available",  "status": "True", "last_transition_time": "2026-05-10T00:00:00Z", "observed_generation": 2},
          {"type": "Reconciled", "status": "True", "last_transition_time": "2026-05-10T00:00:00Z", "observed_generation": 2}
        ]
      },
      "owner_references": {"id": "` + clusterID + `", "kind": "Cluster", "href": ""},
      "created_time": "2026-05-10T00:00:00Z"
    }
  ],
  "kind": "NodePoolList",
  "page": 1,
  "size": 1,
  "total": 1
}`

var emptyNodepoolListJSON = `{"items": [], "kind": "NodePoolList", "page": 1, "size": 0, "total": 0}`

var workers1AdapterStatusesJSON = `{
  "items": [
    {
      "adapter": "np-configmap",
      "observed_generation": 2,
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

// setupResourcesServer builds an httptest server for combined resources tests.
func setupResourcesServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters":
			fmt.Fprint(w, resourcesClusterListJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/statuses":
			fmt.Fprint(w, alphaAdapterStatusesJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterBetaID+"/statuses":
			fmt.Fprint(w, emptyAdapterStatusesJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools":
			fmt.Fprint(w, alphaNodepoolListJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterBetaID+"/nodepools":
			fmt.Fprint(w, emptyNodepoolListJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+nodepoolID+"/statuses":
			fmt.Fprint(w, workers1AdapterStatusesJSON)
		default:
			http.NotFound(w, r)
		}
	}))
}

// resetResourcesFlags resets global flags before each resources test.
func resetResourcesFlags() {
	outputFmt = "json"
	noColor = false
	verbose = false
	resourcesWatchMode = false
	resourcesWatchSecs = 5
	if f := rootCmd.PersistentFlags().Lookup("output"); f != nil {
		f.Changed = false
	}
}

func runResourcesCmd(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	resetResourcesFlags()
	resetGenericFlags()
	resetResourceRegistrationForTest()
	full := []string{"rs"}
	for i, a := range args {
		if i == 0 && (a == "resources" || a == "table") {
			continue
		}
		full = append(full, a)
	}
	return runCmd(t, dir, full...)
}

func TestResourcesTable(t *testing.T) {
	ts := setupResourcesServer(t)
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	out, err := runResourcesCmd(t, dir, "resources", "--output", "table")
	if err != nil {
		t.Fatalf("resources --output table: %v", err)
	}

	// Headers present
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") || !strings.Contains(out, "GEN") {
		t.Errorf("expected fixed headers in table, got: %q", out)
	}
	// Cluster names
	if !strings.Contains(out, "test-cluster-alpha") {
		t.Errorf("expected alpha cluster name, got: %q", out)
	}
	if !strings.Contains(out, "test-cluster-beta") {
		t.Errorf("expected beta cluster name, got: %q", out)
	}
	// Dynamic adapter columns — headers > 10 chars are wrapped across two header lines.
	// CL-DEPLOYMENT (13) → "CL-DEPLOYM" / "ENT"
	if !strings.Contains(out, "CL-DEPLOYM") {
		t.Errorf("expected CL-DEPLOYM (wrapped header line 1), got: %q", out)
	}
	if !strings.Contains(out, "ENT") {
		t.Errorf("expected ENT (wrapped header line 2 of CL-DEPLOYMENT), got: %q", out)
	}
	// Nodepool row uses tree prefix under cluster
	if !strings.Contains(out, nodepoolID) {
		t.Errorf("expected nodepool id in table, got: %q", out)
	}
	// NP-CONFIGMAP (12) → "NP-CONFIGM" / "AP"
	if !strings.Contains(out, "NP-CONFIGM") {
		t.Errorf("expected NP-CONFIGM (wrapped header line 1), got: %q", out)
	}
	if !strings.Contains(out, "AP") {
		t.Errorf("expected AP (wrapped header line 2 of NP-CONFIGMAP), got: %q", out)
	}
}

func TestResourcesTable_ViaAlias(t *testing.T) {
	ts := setupResourcesServer(t)
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	out, err := runResourcesCmd(t, dir, "table", "--output", "table")
	if err != nil {
		t.Fatalf("table alias --output table: %v", err)
	}
	if !strings.Contains(out, "test-cluster-alpha") {
		t.Errorf("expected cluster name via table alias, got: %q", out)
	}
}

func TestResourcesJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, resourcesClusterListJSON)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	out, err := runResourcesCmd(t, dir, "resources", "--output", "json")
	if err != nil {
		t.Fatalf("resources --output json: %v", err)
	}
	if !strings.Contains(out, `"items"`) {
		t.Errorf("expected JSON list with 'items', got: %q", out)
	}
	if !strings.Contains(out, "test-cluster-alpha") {
		t.Errorf("expected cluster name in JSON, got: %q", out)
	}
}

// ---- resources / table watch flags and default format ----

func TestResourcesDefaultsToTable(t *testing.T) {
	ts := setupResourcesServer(t)
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	// No --output flag: should render as table
	out, err := runResourcesCmd(t, dir, "resources")
	if err != nil {
		t.Fatalf("resources (no --output): %v", err)
	}
	if strings.Contains(out, `"items"`) {
		t.Errorf("expected table output (not JSON) when no --output flag, got: %q", out)
	}
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") {
		t.Errorf("expected table headers, got: %q", out)
	}
}

func TestTableDefaultsToTable(t *testing.T) {
	ts := setupResourcesServer(t)
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	out, err := runResourcesCmd(t, dir, "table")
	if err != nil {
		t.Fatalf("table (no --output): %v", err)
	}
	if strings.Contains(out, `"items"`) {
		t.Errorf("expected table output (not JSON) when no --output flag on 'table' cmd, got: %q", out)
	}
	if !strings.Contains(out, "ID") {
		t.Errorf("expected table headers, got: %q", out)
	}
}

func TestResourcesOutputJSON(t *testing.T) {
	ts := setupResourcesServer(t)
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	out, err := runResourcesCmd(t, dir, "resources", "--output", "json")
	if err != nil {
		t.Fatalf("resources --output json: %v", err)
	}
	if !strings.Contains(out, `"items"`) {
		t.Errorf("expected JSON output with --output json, got: %q", out)
	}
}

func TestResourcesWatchFlagRegistered(t *testing.T) {
	f := resourceCmd.Flags().Lookup("watch")
	if f == nil {
		t.Fatal("--watch flag not registered on resourceCmd")
	}
	if f.DefValue != "false" {
		t.Errorf("--watch default = %q, want %q", f.DefValue, "false")
	}
}

func TestResourcesSecondsFlagRegistered(t *testing.T) {
	f := resourceCmd.Flags().Lookup("seconds")
	if f == nil {
		t.Fatal("-s/--seconds flag not registered on resourceCmd")
	}
	if f.DefValue != "5" {
		t.Errorf("-s default = %q, want %q", f.DefValue, "5")
	}
}

func TestResourcesSpinnerForActiveAdapter(t *testing.T) {
	// Build a server that returns an adapter with a very recent last_report_time.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters":
			fmt.Fprint(w, resourcesClusterListJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/statuses":
			// last_report_time = now → adapter is active
			activeJSON := `{"items":[{"adapter":"cl-active","observed_generation":3,` +
				`"conditions":[{"type":"Available","status":"True","last_transition_time":"2026-05-10T00:00:00Z"}],` +
				`"created_time":"2026-05-10T00:00:00Z","last_report_time":"` + time.Now().UTC().Format(time.RFC3339) + `"}],` +
				`"kind":"AdapterStatusList","page":1,"size":1,"total":1}`
			fmt.Fprint(w, activeJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterBetaID+"/statuses":
			fmt.Fprint(w, emptyAdapterStatusesJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools":
			fmt.Fprint(w, emptyNodepoolListJSON)
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterBetaID+"/nodepools":
			fmt.Fprint(w, emptyNodepoolListJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := setupClusterEnv(t, ts)
	setClusterIDInState(t, dir, "test", clusterID)

	// Non-watch render with frequencySecs=0 → IsActive returns false → no spinner
	out, err := runResourcesCmd(t, dir, "resources")
	if err != nil {
		t.Fatalf("resources: %v", err)
	}
	if !strings.Contains(out, "CL-ACTIVE") {
		t.Errorf("expected CL-ACTIVE adapter column header in output, got: %q", out)
	}
	// No spinner when not in watch mode (frequencySecs=0 → IsActive=false)
	for _, frame := range []string{"⠋", "⠙", "⠹", "⠸"} {
		if strings.Contains(out, frame) {
			t.Errorf("unexpected spinner frame %q in non-watch output", frame)
		}
	}
}

func TestRenderResourcesTable_CountdownLine(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	err := renderResourcesTable(cmd, nil, nil, 3, 5, 3)
	if err != nil {
		t.Fatalf("renderResourcesTable: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "↻ 3s") {
		t.Errorf("expected countdown line with '↻ 3s' in watch mode output, got: %q", out)
	}
}

func TestRenderResourcesTable_NoCountdownInNonWatchMode(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	err := renderResourcesTable(cmd, nil, nil, 0, 0, 0)
	if err != nil {
		t.Fatalf("renderResourcesTable: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "↻") {
		t.Errorf("expected no countdown line in non-watch mode, got: %q", out)
	}
}
