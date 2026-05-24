package tui

import (
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

func TestRenderHeaderShowsViewAndConfig(t *testing.T) {
	m := NewModel(Options{RefreshSecs: 5, NoColor: true})
	m.width = 100
	m.height = 30
	m.context = ContextInfo{
		ActiveEnv:   "gke",
		APIURL:      "http://localhost:8000",
		KubeContext: "kind-kind",
		PortForwards: []PortForwardLine{
			{Name: "hyperfleet-api", LocalPort: 8000, Connected: true},
			{Name: "postgresql", LocalPort: 5432, Connected: false},
		},
	}

	out := m.renderHeader()
	if !strings.Contains(out, "View: Clusters/Nodepools") {
		t.Errorf("missing view title: %s", out)
	}
	if !strings.Contains(out, "env:") || !strings.Contains(out, "gke") {
		t.Errorf("missing env: %s", out)
	}
	if !strings.Contains(out, "api:") || !strings.Contains(out, "localhost:8000") {
		t.Errorf("missing api url: %s", out)
	}
	if !strings.Contains(out, "pf:") || !strings.Contains(out, "hyperfleet-api:8000") {
		t.Errorf("missing port-forward line: %s", out)
	}
}

func TestRenderHeaderSeparator(t *testing.T) {
	m := NewModel(Options{RefreshSecs: 5, NoColor: true})
	m.width = 40

	sep := m.renderHeaderSeparator()
	if len([]rune(sep)) != 40 {
		t.Fatalf("separator width = %d, want 40: %q", len([]rune(sep)), sep)
	}

	view := m.View()
	headerEnd := strings.Index(view, strings.Repeat("─", 40))
	if headerEnd < 0 {
		t.Fatalf("expected separator in view:\n%s", view)
	}
}

func TestRenderHeaderDescribeView(t *testing.T) {
	m := NewModel(Options{RefreshSecs: 5, NoColor: true})
	m.detailOpen = true
	m.snapshot = Snapshot{
		Meta:    []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{{Cluster: resource.Cluster{ID: "abc-123", Name: "my-cluster"}}},
	}

	out := m.renderHeader()
	if !strings.Contains(out, "View: Describe") {
		t.Errorf("expected describe view in header: %s", out)
	}
	if !strings.Contains(out, "cluster:") || !strings.Contains(out, "my-cluster") {
		t.Errorf("expected cluster resource line: %s", out)
	}
	if !strings.Contains(out, "abc-123") {
		t.Errorf("expected cluster id in header: %s", out)
	}
}
