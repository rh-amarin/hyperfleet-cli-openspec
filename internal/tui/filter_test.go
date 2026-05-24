package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

func TestCollectConditionTypes(t *testing.T) {
	statuses := []resource.AdapterStatus{
		{Adapter: "a", Conditions: []resource.AdapterCondition{{Type: "Health"}, {Type: "Available"}}},
		{Adapter: "b", Conditions: []resource.AdapterCondition{{Type: "Health"}}},
	}
	got := collectConditionTypes(statuses)
	if len(got) != 2 || got[0] != "Available" || got[1] != "Health" {
		t.Fatalf("collectConditionTypes = %v", got)
	}
}

func TestCollectAdapterNames(t *testing.T) {
	statuses := []resource.AdapterStatus{
		{Adapter: "sentinel-cluster"},
		{Adapter: "maestro"},
		{Adapter: "Sentinel-NodePool"},
	}
	got := collectAdapterNames(statuses)
	if len(got) != 3 {
		t.Fatalf("collectAdapterNames len = %d, want 3", len(got))
	}
}

func TestBuildConditionTypeTable(t *testing.T) {
	statuses := []resource.AdapterStatus{
		{
			Adapter: "sentinel",
			Conditions: []resource.AdapterCondition{
				{Type: "Health", Status: "True", Reason: "ok", Message: "healthy", LastTransitionTime: "2024-01-01T00:00:00Z"},
				{Type: "Available", Status: "False"},
			},
		},
		{
			Adapter: "maestro",
			Conditions: []resource.AdapterCondition{
				{Type: "Health", Status: "True"},
			},
		},
	}

	headers, rows := buildConditionTypeTable(statuses, "Health", true)
	if len(headers) != 5 || len(rows) != 2 {
		t.Fatalf("headers=%v rows=%d", headers, len(rows))
	}
	if rows[0][0] != "sentinel" || rows[0][3] != "ok" || rows[0][4] != "healthy" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestBuildAdapterConditionsTable(t *testing.T) {
	statuses := []resource.AdapterStatus{
		{
			Adapter: "sentinel",
			Conditions: []resource.AdapterCondition{
				{Type: "Health", Status: "True", Reason: "ok", LastTransitionTime: "2024-01-01T00:00:00Z"},
				{Type: "Available", Status: "False", Message: "waiting"},
			},
		},
	}

	headers, rows := buildAdapterConditionsTable(statuses, "sentinel", true)
	if len(headers) != 5 || len(rows) != 2 {
		t.Fatalf("headers=%v rows=%d", headers, len(rows))
	}
	if rows[0][0] != "Health" || rows[1][4] != "waiting" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestRenderDetailJSON(t *testing.T) {
	cl := resource.Cluster{ID: "abc", Name: "test", Generation: 1}
	out := RenderDetail(&cl, nil, nil, DetailJSON, true)
	if !strings.Contains(out, "abc") {
		t.Errorf("expected JSON content: %s", out)
	}
}

func TestRenderDetailOverview(t *testing.T) {
	cl := resource.Cluster{
		ID: "abc", Name: "test", Generation: 3,
		Status: resource.ClusterStatus{
			Conditions: []resource.ResourceCondition{
				{
					Type: "Available", Status: "True", Reason: "Ready",
					LastTransitionTime: "2024-01-01T00:00:00Z", ObservedGeneration: 3,
				},
			},
		},
	}
	statuses := []resource.AdapterStatus{
		{
			Adapter: "sentinel",
			Conditions: []resource.AdapterCondition{
				{Type: "Health", Status: "True", Reason: "ok", Message: "all good"},
			},
		},
	}
	out := RenderDetail(&cl, nil, statuses, DetailOverview, false)
	if !strings.Contains(out, "abc") || !strings.Contains(out, "Generation:  3") {
		t.Errorf("unexpected overview summary: %s", out)
	}
	if !strings.Contains(out, "Resource conditions:") || !strings.Contains(out, "Adapter statuses:") {
		t.Errorf("expected condition sections: %s", out)
	}
	if !strings.Contains(out, "OBS_GEN") || !strings.Contains(out, "MESSAGE") {
		t.Errorf("expected full resource condition columns: %s", out)
	}
	if !strings.Contains(out, "sentinel") || !strings.Contains(out, "all good") {
		t.Errorf("expected adapter condition properties: %s", out)
	}
}

func TestSecsUntil(t *testing.T) {
	if secsUntil(time.Now().Add(-time.Second)) != 0 {
		t.Error("expected 0 for past time")
	}
	if got := secsUntil(time.Now().Add(1500 * time.Millisecond)); got != 2 {
		t.Errorf("secsUntil = %d, want 2", got)
	}
}
