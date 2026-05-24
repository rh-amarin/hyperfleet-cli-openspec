package tui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

func TestEntryFetcherFromAPI(t *testing.T) {
	clusterList := resource.ListResponse[resource.Cluster]{
		Items: []resource.Cluster{
			{ID: "c1", Name: "cluster-1", Generation: 1},
		},
		Kind: "ClusterList",
	}
	statusList := resource.ListResponse[resource.AdapterStatus]{
		Items: []resource.AdapterStatus{{Adapter: "sentinel"}},
	}
	npList := resource.ListResponse[resource.NodePool]{Items: nil}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/hyperfleet/v1/clusters":
			json.NewEncoder(w).Encode(clusterList)
		case "/api/hyperfleet/v1/clusters/c1/statuses":
			json.NewEncoder(w).Encode(statusList)
		case "/api/hyperfleet/v1/clusters/c1/nodepools":
			json.NewEncoder(w).Encode(npList)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := api.NewClient(srv.URL+"/api/hyperfleet/v1/", "", false, false)
	fetcher := NewAPIFetcher(client)
	entries, err := fetcher()
	if err != nil {
		t.Fatalf("fetcher error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if entries[0].Cluster.ID != "c1" {
		t.Errorf("cluster id = %q", entries[0].Cluster.ID)
	}
	if len(entries[0].AdapterStatuses) != 1 {
		t.Errorf("adapter statuses = %d, want 1", len(entries[0].AdapterStatuses))
	}
}
