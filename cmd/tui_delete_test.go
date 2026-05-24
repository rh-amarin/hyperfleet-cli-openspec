package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/tui"
)

func TestTuiDeleteCluster(t *testing.T) {
	const id = "c1"
	deleteCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == apiPrefix+"/clusters/"+id {
			deleteCalled = true
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":"c1","name":"test","generation":1}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	dir := setupClusterEnv(t, srv)
	t.Setenv("HF_CONFIG_DIR", dir)

	info, err := tuiDeleteResource(tui.PatchTarget{ClusterID: id}, false)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !deleteCalled {
		t.Fatal("DELETE not called")
	}
	if !strings.Contains(info, id) {
		t.Errorf("info = %q", info)
	}
}

func TestTuiForceDeleteCluster(t *testing.T) {
	const id = "c1"
	var body map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == apiPrefix+"/clusters/"+id+"/force-delete" {
			json.NewDecoder(r.Body).Decode(&body)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":"c1","name":"test","generation":1}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	dir := setupClusterEnv(t, srv)
	t.Setenv("HF_CONFIG_DIR", dir)

	info, err := tuiDeleteResource(tui.PatchTarget{ClusterID: id}, true)
	if err != nil {
		t.Fatalf("force delete: %v", err)
	}
	if body["reason"] != tuiForceDeleteReason {
		t.Fatalf("reason = %q", body["reason"])
	}
	if !strings.Contains(info, "force-deleted") {
		t.Errorf("info = %q", info)
	}
}

func TestTuiDeleteNodePool(t *testing.T) {
	const clusterID = "c1"
	const npID = "np1"
	deleteCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+npID {
			deleteCalled = true
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":"np1","name":"np","generation":1}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	dir := setupClusterEnv(t, srv)
	t.Setenv("HF_CONFIG_DIR", dir)

	info, err := tuiDeleteResource(tui.PatchTarget{
		IsNodePool: true, ClusterID: clusterID, NodePoolID: npID,
	}, false)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !deleteCalled {
		t.Fatal("DELETE not called")
	}
	if !strings.Contains(info, npID) {
		t.Errorf("info = %q", info)
	}
}

func TestTuiForceDeleteNodePool(t *testing.T) {
	const clusterID = "c1"
	const npID = "np1"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == apiPrefix+"/clusters/"+clusterID+"/nodepools/"+npID+"/force-delete" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":"np1","name":"np","generation":1}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	dir := setupClusterEnv(t, srv)
	t.Setenv("HF_CONFIG_DIR", dir)

	info, err := tuiDeleteResource(tui.PatchTarget{
		IsNodePool: true, ClusterID: clusterID, NodePoolID: npID,
	}, true)
	if err != nil {
		t.Fatalf("force delete: %v", err)
	}
	if !strings.Contains(info, "force-deleted") {
		t.Errorf("info = %q", info)
	}
}
