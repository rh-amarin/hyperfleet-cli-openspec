package tui

import (
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

func TestOverlayModalCentersDialog(t *testing.T) {
	base := strings.Repeat("x\n", 10)
	base = strings.TrimSuffix(base, "\n")
	modal := "┌─────┐\n│ ?   │\n└─────┘"

	out := overlayModal(base, modal, 20, 10)
	lines := strings.Split(out, "\n")
	if len(lines) != 10 {
		t.Fatalf("lines = %d, want 10", len(lines))
	}
	center := lines[4]
	if !strings.Contains(center, "?") {
		t.Fatalf("expected modal near vertical center, line 4 = %q", center)
	}
	if strings.HasPrefix(strings.TrimSpace(center), "x") {
		t.Fatalf("center line should be modal content: %q", center)
	}
}

func TestRenderDeleteModal(t *testing.T) {
	m := NewModel(Options{RefreshSecs: 5, NoColor: true})
	m.snapshot = Snapshot{
		Meta:    []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{{Cluster: resource.Cluster{ID: "c1", Name: "test", Generation: 1}}},
	}
	m.deleteForce = false

	modal := m.renderDeleteModal()
	if !strings.Contains(modal, "Delete") {
		t.Fatalf("modal = %q", modal)
	}
	if !strings.Contains(modal, "test") {
		t.Fatalf("modal should name resource: %q", modal)
	}

	m.deleteForce = true
	modal = m.renderDeleteModal()
	if !strings.Contains(modal, "Force-delete") {
		t.Fatalf("force modal = %q", modal)
	}
}

func TestViewShowsDeleteModal(t *testing.T) {
	m := NewModel(Options{RefreshSecs: 5, NoColor: true})
	m.width = 80
	m.height = 24
	m.deletePhase = DeletePhaseConfirm
	m.snapshot = Snapshot{
		Rows:    [][]string{{"c1", "name", "1"}},
		Meta:    []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{{Cluster: resource.Cluster{ID: "c1", Name: "test", Generation: 1}}},
	}

	view := m.View()
	if !strings.Contains(view, "y  confirm") {
		t.Fatalf("view should include modal hints: %s", view)
	}
}
