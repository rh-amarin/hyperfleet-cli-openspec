package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

func TestModelDetailToggle(t *testing.T) {
	m := NewModel(Options{RefreshSecs: 5, NoColor: true})
	m.snapshot = Snapshot{
		Rows: [][]string{{"c1", "name", "1"}},
		Meta: []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{
			{Cluster: mustCluster()},
		},
	}
	m.width = 120
	m.height = 40

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := updated.(Model)
	if !m2.detailOpen {
		t.Fatal("Enter should open detail panel")
	}

	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := updated.(Model)
	if m3.detailOpen {
		t.Fatal("Esc should close detail panel")
	}
}

func TestModelFormatToggle(t *testing.T) {
	m := NewModel(Options{RefreshSecs: 5, NoColor: true})
	m.detailOpen = true
	m.snapshot = Snapshot{
		Rows: [][]string{{"c1"}},
		Meta: []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{{Cluster: mustCluster()}},
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m2 := updated.(Model)
	if m2.detailFormat != DetailOverview {
		t.Errorf("format = %v, want overview after first V", m2.detailFormat)
	}

	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m3 := updated.(Model)
	if m3.detailFormat != DetailJSON {
		t.Errorf("format = %v, want json after second V", m3.detailFormat)
	}

	updated, _ = m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m4 := updated.(Model)
	if m4.detailFormat != DetailYAML {
		t.Errorf("format = %v, want yaml after Y", m4.detailFormat)
	}
}

func TestModelPatchMode(t *testing.T) {
	patched := false
	fetches := 0
	m := NewModel(Options{
		RefreshSecs: 5,
		NoColor:     true,
		Fetcher: func() ([]ClusterEntry, error) {
			fetches++
			cl := mustCluster()
			if patched {
				cl.Generation = 2
			}
			return []ClusterEntry{{Cluster: cl}}, nil
		},
		Patcher: func(target PatchTarget, section string) (string, error) {
			patched = true
			if section != "spec" || target.ClusterID != "c1" {
				t.Fatalf("unexpected patch: %+v %q", target, section)
			}
			return "[INFO] ok", nil
		},
	})
	m.snapshot = Snapshot{
		Rows:    [][]string{{"c1"}},
		Meta:    []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{{Cluster: mustCluster()}},
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m1 := updated.(Model)
	if !m1.patchMode {
		t.Fatal("c should enter patch mode")
	}

	updated, cmd := m1.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m2 := updated.(Model)
	if m2.patchMode || cmd == nil {
		t.Fatal("s should run patch and exit patch mode")
	}

	msg := cmd()
	if _, ok := msg.(patchResultMsg); !ok {
		t.Fatalf("expected patchResultMsg, got %T", msg)
	}
	if !patched {
		t.Fatal("patcher not called")
	}

	updated, _ = m2.Update(msg.(patchResultMsg))
	m3 := updated.(Model)
	if m3.snapshot.Entries[0].Cluster.Generation != 2 {
		t.Fatalf("expected immediate refresh after patch, gen=%d", m3.snapshot.Entries[0].Cluster.Generation)
	}
	if m3.secsLeft != 5 {
		t.Fatalf("expected refresh countdown reset, secsLeft=%d", m3.secsLeft)
	}
	if fetches != 1 {
		t.Fatalf("expected one immediate fetch after patch, got %d", fetches)
	}
}

func TestModelSinglePanelView(t *testing.T) {
	m := NewModel(Options{RefreshSecs: 5, NoColor: true})
	m.snapshot = Snapshot{
		Rows:    [][]string{{"c1", "name", "1"}},
		Meta:    []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{{Cluster: mustCluster()}},
	}
	m.width = 120
	m.height = 40
	m.detailOpen = true

	view := m.View()
	if strings.Contains(view, " │ ") {
		t.Error("describe view must not use split-panel layout")
	}
	if !strings.Contains(view, "View: Describe") {
		t.Error("expected describe header in view")
	}
}

func TestModelConditionTypeFilter(t *testing.T) {
	m := NewModel(Options{RefreshSecs: 5, NoColor: true})
	m.snapshot = Snapshot{
		Rows: [][]string{{"c1"}},
		Meta: []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{{
			Cluster: mustCluster(),
			AdapterStatuses: []resource.AdapterStatus{
				{
					Adapter: "sentinel",
					Conditions: []resource.AdapterCondition{
						{Type: "Health", Status: "True", Reason: "ok"},
						{Type: "Available", Status: "False"},
					},
				},
				{
					Adapter: "maestro",
					Conditions: []resource.AdapterCondition{
						{Type: "Health", Status: "True"},
					},
				},
			},
		}},
	}
	m.width = 120
	m.height = 40

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m1 := updated.(Model)
	if m1.viewMode != ViewPicker || m1.pickerKind != FilterByCondition {
		t.Fatalf("s should open condition picker, got mode=%v kind=%v", m1.viewMode, m1.pickerKind)
	}
	if len(m1.pickerItems) != 2 {
		t.Fatalf("picker items = %v, want 2 condition types", m1.pickerItems)
	}

	// Picker is sorted; select "Health" (second item).
	updated, _ = m1.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m1 = updated.(Model)
	updated, _ = m1.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := updated.(Model)
	if m2.viewMode != ViewFilter || m2.filterKey == "" {
		t.Fatalf("Enter should confirm picker, got mode=%v key=%q", m2.viewMode, m2.filterKey)
	}

	view := m2.View()
	if !strings.Contains(view, "View: Condition") {
		t.Errorf("expected condition filter header: %s", view)
	}
	if !strings.Contains(view, "sentinel") || !strings.Contains(view, "maestro") {
		t.Errorf("expected adapter rows in filter table: %s", view)
	}
}

func TestModelAdapterFilter(t *testing.T) {
	m := NewModel(Options{RefreshSecs: 5, NoColor: true})
	m.snapshot = Snapshot{
		Rows: [][]string{{"c1"}},
		Meta: []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{{
			Cluster: mustCluster(),
			AdapterStatuses: []resource.AdapterStatus{
				{
					Adapter: "sentinel",
					Conditions: []resource.AdapterCondition{
						{Type: "Health", Status: "True"},
						{Type: "Available", Status: "False"},
					},
				},
			},
		}},
	}
	m.width = 120
	m.height = 40

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m1 := updated.(Model)
	if m1.viewMode != ViewPicker || m1.pickerKind != FilterByAdapter {
		t.Fatalf("a should open adapter picker, got mode=%v kind=%v", m1.viewMode, m1.pickerKind)
	}

	updated, _ = m1.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := updated.(Model)
	if m2.viewMode != ViewFilter {
		t.Fatalf("Enter should open adapter filter view, got mode=%v", m2.viewMode)
	}

	view := m2.View()
	if !strings.Contains(view, "View: Adapter sentinel") {
		t.Errorf("expected adapter filter header: %s", view)
	}
	if !strings.Contains(view, "Health") || !strings.Contains(view, "Available") {
		t.Errorf("expected condition rows: %s", view)
	}
}

func TestModelDetailRefresh(t *testing.T) {
	fetches := 0
	m := NewModel(Options{
		RefreshSecs: 5,
		NoColor:     true,
		Fetcher: func() ([]ClusterEntry, error) {
			fetches++
			return []ClusterEntry{{Cluster: mustCluster()}}, nil
		},
	})
	m.detailOpen = true
	m.snapshot = Snapshot{
		Rows:    [][]string{{"c1"}},
		Meta:    []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{{Cluster: mustCluster()}},
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m1 := updated.(Model)
	if m1.statusMsg != "Refreshing…" {
		t.Fatalf("statusMsg = %q, want Refreshing…", m1.statusMsg)
	}
	if cmd == nil {
		t.Fatal("expected refresh cmd")
	}

	updated, _ = m1.Update(entriesFetchedMsg{entries: []ClusterEntry{{Cluster: mustCluster()}}})
	m2 := updated.(Model)
	if m2.statusMsg != "" {
		t.Fatalf("statusMsg should clear after refresh, got %q", m2.statusMsg)
	}
	if fetches != 0 {
		t.Fatalf("fetcher should run via cmd, not synchronously; fetches=%d", fetches)
	}
}

func TestModelFilterRefresh(t *testing.T) {
	m := NewModel(Options{
		RefreshSecs: 5,
		NoColor:     true,
		Fetcher: func() ([]ClusterEntry, error) {
			return []ClusterEntry{{Cluster: mustCluster()}}, nil
		},
	})
	m.viewMode = ViewFilter
	m.filterKind = FilterByCondition
	m.filterKey = "Health"
	m.snapshot = Snapshot{
		Rows:    [][]string{{"c1"}},
		Meta:    []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{{Cluster: mustCluster()}},
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m1 := updated.(Model)
	if m1.statusMsg != "Refreshing…" {
		t.Fatalf("statusMsg = %q, want Refreshing…", m1.statusMsg)
	}
	if cmd == nil {
		t.Fatal("expected refresh cmd")
	}

	updated, _ = m1.Update(entriesFetchedMsg{entries: []ClusterEntry{{Cluster: mustCluster()}}})
	m2 := updated.(Model)
	if m2.statusMsg != "" {
		t.Fatalf("statusMsg should clear after refresh, got %q", m2.statusMsg)
	}
}

func TestModelDeleteConfirm(t *testing.T) {
	var deleted, force bool
	fetches := 0
	m := NewModel(Options{
		RefreshSecs: 5,
		NoColor:     true,
		Fetcher: func() ([]ClusterEntry, error) {
			fetches++
			cl := mustCluster()
			if deleted {
				cl.DeletedTime = "2024-01-02T00:00:00Z"
			}
			return []ClusterEntry{{Cluster: cl}}, nil
		},
		Deleter: func(target PatchTarget, f bool) (string, error) {
			deleted = true
			force = f
			if target.ClusterID != "c1" || f {
				t.Fatalf("unexpected delete: %+v force=%v", target, f)
			}
			return "[INFO] deleted", nil
		},
	})
	m.snapshot = Snapshot{
		Rows:    [][]string{{"c1"}},
		Meta:    []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{{Cluster: mustCluster()}},
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m1 := updated.(Model)
	if m1.deletePhase != DeletePhaseConfirm || m1.deleteForce {
		t.Fatalf("d on active resource should confirm normal delete, phase=%v force=%v", m1.deletePhase, m1.deleteForce)
	}

	updated, cmd := m1.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m2 := updated.(Model)
	if m2.deletePhase != DeletePhaseNone || cmd == nil {
		t.Fatal("y should run delete and exit delete mode")
	}

	msg := cmd()
	if _, ok := msg.(deleteResultMsg); !ok {
		t.Fatalf("expected deleteResultMsg, got %T", msg)
	}
	if !deleted || force {
		t.Fatalf("deleter called with force=%v", force)
	}

	updated, _ = m2.Update(msg.(deleteResultMsg))
	m3 := updated.(Model)
	if m3.snapshot.Entries[0].Cluster.DeletedTime == "" {
		t.Fatal("expected immediate refresh after delete")
	}
	if fetches != 1 {
		t.Fatalf("expected one immediate fetch after delete, got %d", fetches)
	}
}

func TestModelForceDeleteConfirm(t *testing.T) {
	var force bool
	m := NewModel(Options{
		RefreshSecs: 5,
		NoColor:     true,
		Fetcher: func() ([]ClusterEntry, error) {
			return []ClusterEntry{{Cluster: mustCluster()}}, nil
		},
		Deleter: func(_ PatchTarget, f bool) (string, error) {
			force = f
			return "[INFO] force-deleted", nil
		},
	})
	m.snapshot = Snapshot{
		Rows: [][]string{{"c1"}},
		Meta: []RowMeta{{Kind: RowCluster, ClusterIdx: 0}},
		Entries: []ClusterEntry{{Cluster: resource.Cluster{
			ID: "c1", Name: "test", Generation: 1, DeletedTime: "2024-01-01T00:00:00Z",
		}}},
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m1 := updated.(Model)
	if m1.deletePhase != DeletePhaseConfirm || !m1.deleteForce {
		t.Fatalf("d on deleted resource should confirm force-delete, phase=%v force=%v", m1.deletePhase, m1.deleteForce)
	}

	updated, cmd := m1.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected force-delete cmd")
	}
	cmd()

	if !force {
		t.Fatal("expected force delete")
	}
}

func TestModelPortForwardToggle(t *testing.T) {
	toggled := false
	m := NewModel(Options{
		RefreshSecs: 5,
		NoColor:     true,
		PortForwardToggler: func() (string, error) {
			toggled = true
			return "[INFO] started", nil
		},
	})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m2 := updated.(Model)
	if !m2.portForwarding {
		t.Fatal("p should start port-forward toggle")
	}
	if cmd == nil {
		t.Fatal("expected async cmd for port-forward toggle")
	}

	msg := cmd()
	if _, ok := msg.(portForwardResultMsg); !ok {
		t.Fatalf("expected portForwardResultMsg, got %T", msg)
	}
	if !toggled {
		t.Fatal("toggler not called")
	}
}

func TestModelSpinnerTickRefreshesAdapterCells(t *testing.T) {
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

	m := NewModel(Options{RefreshSecs: 5, NoColor: true})
	m.snapshot = BuildSnapshot(entries, 0, 5, true)
	adIdx := len(m.snapshot.Headers) - 1
	before := m.snapshot.Rows[0][adIdx]

	updated, _ := m.Update(spinnerTickMsg{})
	m1 := updated.(Model)
	after := m1.snapshot.Rows[0][adIdx]
	if before == after {
		t.Fatalf("adapter cell should update on spinner tick: %q", before)
	}
	if m1.tick != 1 {
		t.Fatalf("tick = %d, want 1", m1.tick)
	}
}

func mustCluster() resource.Cluster {
	return resource.Cluster{ID: "c1", Name: "test", Generation: 1}
}
