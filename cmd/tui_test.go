package cmd

import (
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

func TestTUICmdRegistered(t *testing.T) {
	if tuiCmd == nil {
		t.Fatal("tuiCmd is nil")
	}
	if tuiCmd.Use != "tui" {
		t.Errorf("Use = %q, want tui", tuiCmd.Use)
	}
}

func TestTUIRefreshFlagDefault(t *testing.T) {
	f := tuiCmd.Flags().Lookup("seconds")
	if f == nil {
		t.Fatal("-s flag not registered")
	}
	if f.DefValue != "5" {
		t.Errorf("default = %q, want 5", f.DefValue)
	}
}

func TestToTUIEntries(t *testing.T) {
	entries := []clusterEntry{
		{
			cluster:    resource.Cluster{ID: "c1", Name: "test"},
			npStatuses: map[string][]resource.AdapterStatus{},
		},
	}
	got := toTUIEntries(entries)
	if len(got) != 1 || got[0].Cluster.ID != "c1" {
		t.Fatalf("conversion failed: %+v", got)
	}
}
