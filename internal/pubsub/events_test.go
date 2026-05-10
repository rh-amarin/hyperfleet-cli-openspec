package pubsub

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildClusterEvent(t *testing.T) {
	data, err := BuildClusterEvent("cluster-123", "http://api.example.com", "v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var ev map[string]json.RawMessage
	if err := json.Unmarshal(data, &ev); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	checkString(t, ev, "specversion", "1.0")
	checkString(t, ev, "type", "com.redhat.hyperfleet.cluster.reconcile.v1")
	checkString(t, ev, "source", "/hyperfleet/service/sentinel")
	checkString(t, ev, "id", "cluster-123")
	checkString(t, ev, "datacontenttype", "application/json")

	var d map[string]string
	if err := json.Unmarshal(ev["data"], &d); err != nil {
		t.Fatalf("invalid data JSON: %v", err)
	}
	if d["id"] != "cluster-123" {
		t.Errorf("data.id = %q, want %q", d["id"], "cluster-123")
	}
	if d["kind"] != "Cluster" {
		t.Errorf("data.kind = %q, want %q", d["kind"], "Cluster")
	}
	wantHref := "http://api.example.com/api/hyperfleet/v1/clusters/cluster-123"
	if d["href"] != wantHref {
		t.Errorf("data.href = %q, want %q", d["href"], wantHref)
	}
}

func TestBuildNodePoolEvent(t *testing.T) {
	data, err := BuildNodePoolEvent("cluster-123", "np-456", "http://api.example.com", "v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var ev map[string]json.RawMessage
	if err := json.Unmarshal(data, &ev); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	checkString(t, ev, "specversion", "1.0")
	checkString(t, ev, "type", "com.redhat.hyperfleet.nodepool.reconcile.v1")
	checkString(t, ev, "id", "np-456")

	var d map[string]json.RawMessage
	if err := json.Unmarshal(ev["data"], &d); err != nil {
		t.Fatalf("invalid data JSON: %v", err)
	}

	checkRawString(t, d, "id", "np-456")
	checkRawString(t, d, "kind", "NodePool")
	wantHref := "http://api.example.com/api/hyperfleet/v1/clusters/cluster-123/nodepools/np-456"
	checkRawString(t, d, "href", wantHref)

	var owner map[string]string
	if err := json.Unmarshal(d["owner_references"], &owner); err != nil {
		t.Fatalf("invalid owner_references JSON: %v", err)
	}
	if owner["id"] != "cluster-123" {
		t.Errorf("owner.id = %q, want %q", owner["id"], "cluster-123")
	}
	if owner["kind"] != "Cluster" {
		t.Errorf("owner.kind = %q, want %q", owner["kind"], "Cluster")
	}
	wantClusterHref := "http://api.example.com/api/hyperfleet/v1/clusters/cluster-123"
	if owner["href"] != wantClusterHref {
		t.Errorf("owner.href = %q, want %q", owner["href"], wantClusterHref)
	}
}

func TestBuildClusterEvent_TrailingSlashAPIURL(t *testing.T) {
	data, err := BuildClusterEvent("c1", "http://api.example.com/", "v2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The href path must not contain consecutive slashes (excluding the scheme "://").
	var ev map[string]json.RawMessage
	_ = json.Unmarshal(data, &ev)
	var d map[string]string
	_ = json.Unmarshal(ev["data"], &d)
	href := d["href"]
	// Strip scheme (http://) before checking for double slashes in the path.
	pathPart := strings.SplitN(href, "://", 2)
	if len(pathPart) == 2 && strings.Contains(pathPart[1], "//") {
		t.Errorf("href path contains double slash: %s", href)
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
