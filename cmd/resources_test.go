package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
}

func runResourcesCmd(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	resetResourcesFlags()
	return runCmd(t, dir, args...)
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
	// Dynamic adapter columns (headers are uppercased by PrintTable)
	if !strings.Contains(out, "CL-DEPLOYMENT") {
		t.Errorf("expected CL-DEPLOYMENT adapter column header, got: %q", out)
	}
	// Nodepool row indented
	if !strings.Contains(out, "  "+nodepoolID) {
		t.Errorf("expected indented nodepool row, got: %q", out)
	}
	if !strings.Contains(out, "NP-CONFIGMAP") {
		t.Errorf("expected NP-CONFIGMAP adapter column header, got: %q", out)
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
