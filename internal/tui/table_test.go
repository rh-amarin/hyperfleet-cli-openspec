package tui

import (
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

func TestBuildSnapshotRows(t *testing.T) {
	entries := []ClusterEntry{
		{
			Cluster: resource.Cluster{ID: "c1", Name: "cluster-1", Generation: 1},
			Nodepools: []resource.NodePool{
				{ID: "np1", Name: "pool-1", Generation: 1},
			},
			NPStatuses: map[string][]resource.AdapterStatus{"np1": {}},
		},
	}
	snap := BuildSnapshot(entries, 0, 5, true)
	if len(snap.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(snap.Rows))
	}
	if snap.Meta[1].Kind != RowNodePool {
		t.Errorf("second row kind = %v, want RowNodePool", snap.Meta[1].Kind)
	}
	if len(snap.Headers) < 3 {
		t.Fatalf("headers too short: %v", snap.Headers)
	}
}

func TestSelectedResource(t *testing.T) {
	entries := []ClusterEntry{
		{
			Cluster:         resource.Cluster{ID: "c1", Name: "cluster-1"},
			AdapterStatuses: []resource.AdapterStatus{{Adapter: "a1"}},
			Nodepools: []resource.NodePool{
				{ID: "np1", Name: "pool-1"},
			},
			NPStatuses: map[string][]resource.AdapterStatus{
				"np1": {{Adapter: "np-ad"}},
			},
		},
	}
	snap := BuildSnapshot(entries, 0, 0, true)

	cl, np, statuses, ok := snap.SelectedResource(0)
	if !ok || cl == nil || np != nil || len(statuses) != 1 {
		t.Fatalf("cluster selection failed")
	}

	cl, np, statuses, ok = snap.SelectedResource(1)
	if !ok || cl == nil || np == nil || statuses[0].Adapter != "np-ad" {
		t.Fatalf("nodepool selection failed")
	}
}
