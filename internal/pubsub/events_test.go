package pubsub

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildGenericReconcileEvent_RootResource(t *testing.T) {
	data, err := BuildGenericReconcileEvent("clusters", "cluster-123", nil, "clusters", "http://api.example.com", "v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var ev map[string]json.RawMessage
	if err := json.Unmarshal(data, &ev); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	checkString(t, ev, "specversion", "1.0")
	checkString(t, ev, "type", "com.redhat.hyperfleet.clusters.reconcile.v1")
	checkString(t, ev, "source", "/hyperfleet/service/sentinel")
	checkString(t, ev, "id", "cluster-123")
	checkString(t, ev, "datacontenttype", "application/json")

	var d map[string]json.RawMessage
	if err := json.Unmarshal(ev["data"], &d); err != nil {
		t.Fatalf("invalid data JSON: %v", err)
	}
	checkRawString(t, d, "id", "cluster-123")
	checkRawString(t, d, "kind", "clusters")
	checkRawString(t, d, "href", "http://api.example.com/api/hyperfleet/v1/clusters/cluster-123")

	if _, ok := d["owner_references"]; ok {
		t.Error("root resource must not have owner_references")
	}
}

func TestBuildGenericReconcileEvent_ChildResource(t *testing.T) {
	ancestors := []AncestorID{
		{TypeName: "clusters", ID: "cluster-123", Path: "clusters"},
	}
	data, err := BuildGenericReconcileEvent("nodepools", "np-456", ancestors, "clusters/cluster-123/nodepools", "http://api.example.com", "v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var ev map[string]json.RawMessage
	if err := json.Unmarshal(data, &ev); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	checkString(t, ev, "type", "com.redhat.hyperfleet.nodepools.reconcile.v1")
	checkString(t, ev, "id", "np-456")

	var d map[string]json.RawMessage
	if err := json.Unmarshal(ev["data"], &d); err != nil {
		t.Fatalf("invalid data JSON: %v", err)
	}
	checkRawString(t, d, "id", "np-456")
	checkRawString(t, d, "kind", "nodepools")
	checkRawString(t, d, "href", "http://api.example.com/api/hyperfleet/v1/clusters/cluster-123/nodepools/np-456")

	var owner map[string]string
	if err := json.Unmarshal(d["owner_references"], &owner); err != nil {
		t.Fatalf("invalid owner_references JSON: %v", err)
	}
	if owner["id"] != "cluster-123" {
		t.Errorf("owner.id = %q, want %q", owner["id"], "cluster-123")
	}
	if owner["kind"] != "clusters" {
		t.Errorf("owner.kind = %q, want %q", owner["kind"], "clusters")
	}
	if owner["href"] != "http://api.example.com/api/hyperfleet/v1/clusters/cluster-123" {
		t.Errorf("owner.href = %q", owner["href"])
	}
}

func TestBuildGenericReconcileEvent_ImmediateParentOnly(t *testing.T) {
	// Deeply nested: versions under channels — owner_references must be channels only, not any grandparent.
	ancestors := []AncestorID{
		{TypeName: "channels", ID: "ch-1", Path: "channels"},
	}
	data, err := BuildGenericReconcileEvent("versions", "v-2", ancestors, "channels/ch-1/versions", "http://api.example.com", "v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var ev map[string]json.RawMessage
	_ = json.Unmarshal(data, &ev)
	var d map[string]json.RawMessage
	_ = json.Unmarshal(ev["data"], &d)

	var owner map[string]string
	if err := json.Unmarshal(d["owner_references"], &owner); err != nil {
		t.Fatalf("invalid owner_references JSON: %v", err)
	}
	if owner["kind"] != "channels" {
		t.Errorf("owner.kind = %q, want %q", owner["kind"], "channels")
	}
}

func TestBuildGenericReconcileEvent_MissingID(t *testing.T) {
	_, err := BuildGenericReconcileEvent("clusters", "", nil, "clusters", "http://api.example.com", "v1")
	if err == nil {
		t.Fatal("expected error for empty resourceID")
	}
}

func TestBuildGenericReconcileEvent_TrailingSlashAPIURL(t *testing.T) {
	data, err := BuildGenericReconcileEvent("clusters", "c1", nil, "clusters", "http://api.example.com/", "v2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var ev map[string]json.RawMessage
	_ = json.Unmarshal(data, &ev)
	var d map[string]string
	_ = json.Unmarshal(ev["data"], &d)
	pathPart := strings.SplitN(d["href"], "://", 2)
	if len(pathPart) == 2 && strings.Contains(pathPart[1], "//") {
		t.Errorf("href path contains double slash: %s", d["href"])
	}
}

func checkString(t *testing.T, m map[string]json.RawMessage, key, want string) {
	t.Helper()
	raw, ok := m[key]
	if !ok {
		t.Errorf("missing key %q", key)
		return
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Errorf("key %q not a string: %v", key, err)
		return
	}
	if s != want {
		t.Errorf("%q = %q, want %q", key, s, want)
	}
}

func checkRawString(t *testing.T, m map[string]json.RawMessage, key, want string) {
	t.Helper()
	raw, ok := m[key]
	if !ok {
		t.Errorf("missing key %q", key)
		return
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Errorf("key %q not a string: %v", key, err)
		return
	}
	if s != want {
		t.Errorf("%q = %q, want %q", key, s, want)
	}
}
