package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/output"
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

func TestBuildSnapshot_AdapterSpinnerChangesWithTick(t *testing.T) {
	recent := time.Now().UTC().Format(time.RFC3339)
	entries := []ClusterEntry{{
		Cluster: resource.Cluster{ID: "c1", Name: "cluster-1", Generation: 1},
		AdapterStatuses: []resource.AdapterStatus{{
			Adapter:            "sentinel",
			ObservedGeneration: 1,
			LastReportTime:     recent,
			Conditions:         []resource.AdapterCondition{{Type: "Available", Status: "True"}},
		}},
	}}

	snap0 := BuildSnapshot(entries, 0, 5, true)
	snap1 := BuildSnapshot(entries, 1, 5, true)
	if len(snap0.Rows) != 1 || len(snap1.Rows) != 1 {
		t.Fatalf("expected one row, got %d and %d", len(snap0.Rows), len(snap1.Rows))
	}

	adIdx := len(snap0.Headers) - 1
	cell0 := snap0.Rows[0][adIdx]
	cell1 := snap1.Rows[0][adIdx]
	if cell0 == cell1 {
		t.Fatalf("spinner cell should change with tick: %q vs %q", cell0, cell1)
	}
	if !strings.HasPrefix(cell0, output.SpinnerFrame(0)) {
		t.Fatalf("expected spinner prefix in %q", cell0)
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
