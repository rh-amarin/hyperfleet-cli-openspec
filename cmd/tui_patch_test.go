package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/tui"
)

func TestTuiPatchClusterSpec(t *testing.T) {
	const id = "c1"
	var patched map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+id:
			json.NewEncoder(w).Encode(map[string]any{
				"id": id, "spec": map[string]any{"counter": "1"},
			})
		case r.Method == http.MethodPatch && r.URL.Path == apiPrefix+"/clusters/"+id:
			json.NewDecoder(r.Body).Decode(&patched)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"id": id})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := setupClusterEnv(t, srv)
	t.Setenv("HF_CONFIG_DIR", dir)

	info, err := tuiPatchResource(tui.PatchTarget{ClusterID: id}, "spec")
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	if !strings.Contains(info, "1 -> 2") {
		t.Errorf("info = %q", info)
	}
	spec, _ := patched["spec"].(map[string]any)
	if spec["counter"] != "2" {
		t.Errorf("patched counter = %v", spec["counter"])
	}
}

func TestTuiPatchNodePoolLabels(t *testing.T) {
	const clusterID = "c1"
	const npID = "np1"
	var patched map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+npID:
			json.NewEncoder(w).Encode(map[string]any{
				"id": npID, "labels": map[string]any{"counter": "3"},
			})
		case r.Method == http.MethodPatch && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+npID:
			json.NewDecoder(r.Body).Decode(&patched)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"id": npID})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := setupClusterEnv(t, srv)
	t.Setenv("HF_CONFIG_DIR", dir)

	info, err := tuiPatchResource(tui.PatchTarget{
		IsNodePool: true, ClusterID: clusterID, NodePoolID: npID,
	}, "labels")
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	if !strings.Contains(info, "3 -> 4") {
		t.Errorf("info = %q", info)
	}
	labels, _ := patched["labels"].(map[string]any)
	if labels["counter"] != "4" {
		t.Errorf("patched counter = %v", labels["counter"])
	}
}
