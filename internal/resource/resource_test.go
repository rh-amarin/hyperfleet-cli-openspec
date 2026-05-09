package resource_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"gopkg.in/yaml.v3"
)

func TestClusterRoundTrip(t *testing.T) {
	raw := `{
		"id": "abc-123",
		"kind": "Cluster",
		"name": "test-cluster",
		"generation": 2,
		"labels": {"env": "test", "region": "us-east"},
		"spec": {"replicas": 3, "platform": {"type": "aws"}},
		"status": {"conditions": [
			{"type": "Available", "status": "True", "last_transition_time": "2024-01-01T00:00:00Z",
			 "observed_generation": 2, "created_time": "2024-01-01T00:00:00Z", "last_updated_time": "2024-01-01T00:00:00Z"}
		]},
		"created_by": "user1",
		"created_time": "2024-01-01T00:00:00Z",
		"updated_by": "user1",
		"updated_time": "2024-01-01T00:00:00Z",
		"href": "/api/hyperfleet/v1/clusters/abc-123"
	}`

	var c resource.Cluster
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if c.ID != "abc-123" {
		t.Errorf("ID: got %q, want %q", c.ID, "abc-123")
	}
	if c.Labels["env"] != "test" {
		t.Errorf("Labels[env]: got %q, want %q", c.Labels["env"], "test")
	}
	if c.Spec["replicas"].(float64) != 3 {
		t.Errorf("Spec[replicas]: got %v, want 3", c.Spec["replicas"])
	}
	if len(c.Status.Conditions) != 1 || c.Status.Conditions[0].Type != "Available" {
		t.Errorf("Status.Conditions: got %v", c.Status.Conditions)
	}

	// Re-marshal and check no data loss
	out, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var c2 resource.Cluster
	if err := json.Unmarshal(out, &c2); err != nil {
		t.Fatalf("re-unmarshal error: %v", err)
	}
	if c2.ID != c.ID || c2.Name != c.Name {
		t.Errorf("round-trip mismatch: %+v vs %+v", c, c2)
	}
}

func TestNodePoolRoundTrip(t *testing.T) {
	raw := `{
		"id": "np-456",
		"kind": "NodePool",
		"name": "pool-1",
		"generation": 1,
		"labels": {"pool": "default"},
		"spec": {"replicas": 2, "platform": {"type": "gcp"}},
		"status": {"conditions": []},
		"owner_references": {"id": "abc-123", "kind": "Cluster", "href": "/api/hyperfleet/v1/clusters/abc-123"},
		"created_by": "user1",
		"created_time": "2024-01-01T00:00:00Z",
		"updated_by": "user1",
		"updated_time": "2024-01-01T00:00:00Z",
		"href": "/api/hyperfleet/v1/clusters/abc-123/nodepools/np-456"
	}`

	var np resource.NodePool
	if err := json.Unmarshal([]byte(raw), &np); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if np.ID != "np-456" {
		t.Errorf("ID: got %q, want %q", np.ID, "np-456")
	}
	if np.OwnerReferences.ID != "abc-123" {
		t.Errorf("OwnerReferences.ID: got %q, want %q", np.OwnerReferences.ID, "abc-123")
	}
	if np.Spec["replicas"].(float64) != 2 {
		t.Errorf("Spec[replicas]: got %v, want 2", np.Spec["replicas"])
	}
}

func TestListResponseRoundTrip(t *testing.T) {
	raw := `{
		"items": [{"id": "c1", "kind": "Cluster", "name": "c1", "generation": 1, "labels": {}, "spec": {}, "status": {"conditions": []}, "created_by": "", "created_time": "", "updated_by": "", "updated_time": "", "href": ""}],
		"kind": "ClusterList",
		"page": 1,
		"size": 1,
		"total": 1
	}`

	var lr resource.ListResponse[resource.Cluster]
	if err := json.Unmarshal([]byte(raw), &lr); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(lr.Items) != 1 {
		t.Errorf("Items: got %d, want 1", len(lr.Items))
	}
	if lr.Kind != "ClusterList" {
		t.Errorf("Kind: got %q, want %q", lr.Kind, "ClusterList")
	}
	if lr.Total != 1 {
		t.Errorf("Total: got %d, want 1", lr.Total)
	}
}

func TestAdapterStatusRoundTrip(t *testing.T) {
	raw := `{
		"adapter": "openshift",
		"observed_generation": 3,
		"conditions": [
			{"type": "Synced", "status": "True", "last_transition_time": "2024-01-01T00:00:00Z"}
		],
		"created_time": "2024-01-01T00:00:00Z",
		"last_report_time": "2024-01-02T00:00:00Z",
		"data": {"extra_key": "extra_val"}
	}`

	var as resource.AdapterStatus
	if err := json.Unmarshal([]byte(raw), &as); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if as.Adapter != "openshift" {
		t.Errorf("Adapter: got %q, want %q", as.Adapter, "openshift")
	}
	if len(as.Conditions) != 1 || as.Conditions[0].Status != "True" {
		t.Errorf("Conditions: got %v", as.Conditions)
	}
	if as.Data["extra_key"] != "extra_val" {
		t.Errorf("Data[extra_key]: got %v", as.Data["extra_key"])
	}
}

// TestClusterYAMLRoundTrip verifies that yaml.v3 uses the explicit yaml tags
// (snake_case) rather than lowercasing Go field names directly.
func TestClusterYAMLRoundTrip(t *testing.T) {
	c := resource.Cluster{
		ID:          "abc-123",
		Kind:        "Cluster",
		Name:        "test-cluster",
		Generation:  2,
		Labels:      map[string]string{"env": "test"},
		Spec:        map[string]any{"replicas": 3},
		CreatedBy:   "user1",
		CreatedTime: "2024-01-01T00:00:00Z",
		UpdatedBy:   "user1",
		UpdatedTime: "2024-01-02T00:00:00Z",
		Href:        "/api/hyperfleet/v1/clusters/abc-123",
		Status: resource.ClusterStatus{
			Conditions: []resource.ResourceCondition{
				{
					Type:               "Available",
					Status:             "True",
					LastTransitionTime: "2024-01-01T00:00:00Z",
					ObservedGeneration: 2,
					CreatedTime:        "2024-01-01T00:00:00Z",
					LastUpdatedTime:    "2024-01-01T00:00:00Z",
				},
			},
		},
	}

	out, err := yaml.Marshal(c)
	if err != nil {
		t.Fatalf("yaml.Marshal error: %v", err)
	}
	yamlStr := string(out)

	// Verify snake_case keys are present (not camelCase lowercased)
	for _, want := range []string{"created_time", "updated_time", "created_by", "updated_by", "last_transition_time", "observed_generation", "last_updated_time"} {
		if !strings.Contains(yamlStr, want) {
			t.Errorf("YAML missing key %q:\n%s", want, yamlStr)
		}
	}

	// Verify camelCase-collapsed keys are NOT present
	for _, bad := range []string{"createdtime", "updatedtime", "createdby", "updatedby", "lasttransitiontime", "observedgeneration"} {
		if strings.Contains(yamlStr, bad) {
			t.Errorf("YAML contains incorrectly lowercased key %q:\n%s", bad, yamlStr)
		}
	}

	// Round-trip: unmarshal back and verify field values
	var c2 resource.Cluster
	if err := yaml.Unmarshal(out, &c2); err != nil {
		t.Fatalf("yaml.Unmarshal error: %v", err)
	}
	if c2.CreatedTime != c.CreatedTime {
		t.Errorf("CreatedTime round-trip: got %q, want %q", c2.CreatedTime, c.CreatedTime)
	}
	if c2.UpdatedBy != c.UpdatedBy {
		t.Errorf("UpdatedBy round-trip: got %q, want %q", c2.UpdatedBy, c.UpdatedBy)
	}
	if len(c2.Status.Conditions) != 1 || c2.Status.Conditions[0].LastTransitionTime != "2024-01-01T00:00:00Z" {
		t.Errorf("Conditions round-trip: got %+v", c2.Status.Conditions)
	}
}
