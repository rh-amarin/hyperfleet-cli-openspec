package maestro_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/rh-amarin/hyperfleet-cli/internal/maestro"
)

// newTestClient builds a Client pointed at the given test server URL.
func newTestClient(t *testing.T, serverURL string) *maestro.Client {
	t.Helper()
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfg, []byte("maestro:\n  http-endpoint: "+serverURL+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := config.New(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	return maestro.NewFromConfig(s)
}

func TestListResourceBundles(t *testing.T) {
	payload := maestro.ResourceBundleList{
		Kind:  "ResourceBundleList",
		Total: 1,
		Items: []maestro.ResourceBundle{
			{ID: "id-1", Name: "mw-cluster1", ConsumerName: "cluster1", Version: 2, ManifestCount: 1},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/maestro/v1/resource-bundles" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	list, err := c.ListResourceBundles(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list.Items))
	}
	if list.Items[0].Name != "mw-cluster1" {
		t.Errorf("unexpected name: %s", list.Items[0].Name)
	}
}

func TestListResourceBundlesByConsumer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/maestro/v1/resource-bundles" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query().Get("search")
		if q == "" {
			t.Errorf("expected search query param, got none")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(maestro.ResourceBundleList{
			Kind:  "ResourceBundleList",
			Total: 0,
			Items: []maestro.ResourceBundle{},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	list, err := c.ListResourceBundlesByConsumer(context.Background(), "cluster1")
	if err != nil {
		t.Fatal(err)
	}
	if list == nil {
		t.Fatal("expected non-nil list")
	}
}

func TestGetResourceBundle_found(t *testing.T) {
	payload := maestro.ResourceBundleList{
		Kind:  "ResourceBundleList",
		Total: 2,
		Items: []maestro.ResourceBundle{
			{ID: "id-1", Name: "mw-cluster1"},
			{ID: "id-2", Name: "mw-cluster2"},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	rb, err := c.GetResourceBundle(context.Background(), "mw-cluster2")
	if err != nil {
		t.Fatal(err)
	}
	if rb == nil {
		t.Fatal("expected non-nil bundle")
	}
	if rb.ID != "id-2" {
		t.Errorf("expected id-2, got %s", rb.ID)
	}
}

func TestGetResourceBundle_notFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(maestro.ResourceBundleList{Kind: "ResourceBundleList", Items: []maestro.ResourceBundle{}})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	rb, err := c.GetResourceBundle(context.Background(), "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if rb != nil {
		t.Errorf("expected nil, got %+v", rb)
	}
}

func TestDeleteResourceBundle(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.DeleteResourceBundle(context.Background(), "id-1"); err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if gotPath != "/api/maestro/v1/resource-bundles/id-1" {
		t.Errorf("unexpected path: %s", gotPath)
	}
}

func TestListConsumers(t *testing.T) {
	payload := maestro.ConsumerList{
		Kind:  "ConsumerList",
		Total: 1,
		Items: []maestro.Consumer{{ID: "c-1", Kind: "Consumer", Name: "cluster1"}},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/maestro/v1/consumers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	list, err := c.ListConsumers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 || list.Items[0].Name != "cluster1" {
		t.Errorf("unexpected consumer list: %+v", list)
	}
}

func TestHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.ListResourceBundles(context.Background())
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}
