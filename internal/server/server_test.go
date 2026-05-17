package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

// newTestServer creates a Server backed by a fake upstream httptest.Server.
// The upstream handler is provided by the caller.
func newTestServer(t *testing.T, upstream http.Handler) (*Server, *httptest.Server) {
	t.Helper()
	upstreamSrv := httptest.NewServer(upstream)
	t.Cleanup(upstreamSrv.Close)

	client := api.NewClient(upstreamSrv.URL+"/", "", false, false)
	srv := New(client, 0, []byte("<html>test</html>"))
	return srv, upstreamSrv
}

// do makes a GET request to the Server's route handler and returns the recorder.
func do(srv *Server, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	srv.route(rr, req)
	return rr
}

// doPost makes a POST request with a JSON body to the Server's route handler.
func doPost(srv *Server, path string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.route(rr, req)
	return rr
}

// TestHandleIndex serves the embedded HTML at GET /.
func TestHandleIndex(t *testing.T) {
	srv := New(nil, 0, []byte("<html>dashboard</html>"))
	rr := do(srv, "/")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html, got %q", ct)
	}
	if body := rr.Body.String(); body != "<html>dashboard</html>" {
		t.Errorf("unexpected body: %q", body)
	}
}

// TestRouteNotFound returns 404 for unknown paths.
func TestRouteNotFound(t *testing.T) {
	srv := New(nil, 0, nil)
	paths := []string{"/api/unknown", "/api/clusters/id/unknown", "/foo/bar"}
	for _, p := range paths {
		rr := do(srv, p)
		if rr.Code != http.StatusNotFound {
			t.Errorf("%s: expected 404, got %d", p, rr.Code)
		}
	}
}

// TestRouteMethodNotAllowed returns 405 for non-GET requests.
func TestRouteMethodNotAllowed(t *testing.T) {
	srv := New(nil, 0, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/clusters", nil)
	rr := httptest.NewRecorder()
	srv.route(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

// TestHandleClusterProxy verifies that a single-cluster request is proxied correctly.
func TestHandleClusterProxy(t *testing.T) {
	clusterJSON := `{"id":"abc","name":"test-cluster","kind":"Cluster"}`
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clusters/abc" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, clusterJSON)
	})

	srv, _ := newTestServer(t, upstream)
	rr := do(srv, "/api/clusters/abc")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var m map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &m); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if m["id"] != "abc" {
		t.Errorf("unexpected id: %v", m["id"])
	}
}

// TestHandleClustersWithStatuses verifies that GET /api/clusters returns items
// merged with adapter_statuses from per-cluster status calls.
func TestHandleClustersWithStatuses(t *testing.T) {
	clusters := resource.ListResponse[resource.Cluster]{
		Items: []resource.Cluster{
			{ID: "c1", Name: "cluster-one", Kind: "Cluster"},
			{ID: "c2", Name: "cluster-two", Kind: "Cluster"},
		},
		Kind:  "ClusterList",
		Total: 2,
	}

	statuses := resource.ListResponse[resource.AdapterStatus]{
		Items: []resource.AdapterStatus{
			{Adapter: "provisioner"},
		},
		Kind: "AdapterStatusList",
	}

	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/clusters":
			_ = json.NewEncoder(w).Encode(clusters)
		case "/clusters/c1/statuses", "/clusters/c2/statuses":
			_ = json.NewEncoder(w).Encode(statuses)
		default:
			http.NotFound(w, r)
		}
	})

	srv, _ := newTestServer(t, upstream)
	rr := do(srv, "/api/clusters")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp ClustersResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}
	for _, item := range resp.Items {
		if len(item.AdapterStatuses) != 1 {
			t.Errorf("cluster %s: expected 1 adapter status, got %d", item.ID, len(item.AdapterStatuses))
		}
		if item.AdapterStatuses[0].Adapter != "provisioner" {
			t.Errorf("cluster %s: unexpected adapter %q", item.ID, item.AdapterStatuses[0].Adapter)
		}
	}
}

// TestHandleClustersUpstreamError forwards upstream 4xx errors with correct status.
func TestHandleClustersUpstreamError(t *testing.T) {
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"title":"Not Found","status":404,"detail":"cluster not found"}`)
	})

	srv, _ := newTestServer(t, upstream)
	rr := do(srv, "/api/clusters/missing")

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

// TestRoutesMapToCorrectUpstreamPaths verifies each route hits the correct upstream path.
func TestRoutesMapToCorrectUpstreamPaths(t *testing.T) {
	cases := []struct {
		browserPath  string
		upstreamPath string
	}{
		{"/api/clusters/abc", "/clusters/abc"},
		{"/api/clusters/abc/statuses", "/clusters/abc/statuses"},
		{"/api/clusters/abc/nodepools", "/clusters/abc/nodepools"},
		{"/api/clusters/abc/nodepools/np1", "/clusters/abc/nodepools/np1"},
		{"/api/clusters/abc/nodepools/np1/statuses", "/clusters/abc/nodepools/np1/statuses"},
	}

	for _, tc := range cases {
		t.Run(tc.browserPath, func(t *testing.T) {
			var got string
			upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{}`)
			})
			srv, _ := newTestServer(t, upstream)
			do(srv, tc.browserPath)
			if got != tc.upstreamPath {
				t.Errorf("expected upstream path %q, got %q", tc.upstreamPath, got)
			}
		})
	}
}

// TestPostClusterStatusesProxy verifies that POST /api/clusters/{id}/statuses
// forwards the body to the upstream and returns the upstream response.
func TestPostClusterStatusesProxy(t *testing.T) {
	reqBody := `{"adapter":"cl-job","observed_generation":1,"observed_time":"2026-05-17T10:00:00Z","conditions":[]}`
	var receivedBody string
	var receivedMethod string

	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		b, _ := fmt.Fprintf(w, "") // capture below
		_ = b
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		receivedBody = buf.String()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"adapter":"cl-job"}`)
	})

	srv, _ := newTestServer(t, upstream)
	rr := doPost(srv, "/api/clusters/abc/statuses", []byte(reqBody))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if receivedMethod != http.MethodPost {
		t.Errorf("expected upstream method POST, got %q", receivedMethod)
	}
	var sent map[string]any
	if err := json.Unmarshal([]byte(receivedBody), &sent); err != nil {
		t.Fatalf("upstream received invalid JSON: %v", err)
	}
	if sent["adapter"] != "cl-job" {
		t.Errorf("upstream body missing adapter field, got %v", sent)
	}
}

// TestPostNodePoolStatusesProxy verifies POST /api/clusters/{id}/nodepools/{npid}/statuses.
func TestPostNodePoolStatusesProxy(t *testing.T) {
	var receivedPath string
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{}`)
	})

	srv, _ := newTestServer(t, upstream)
	body := []byte(`{"adapter":"np-provisioner","observed_generation":2,"observed_time":"2026-05-17T10:00:00Z","conditions":[]}`)
	rr := doPost(srv, "/api/clusters/c1/nodepools/np1/statuses", body)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if receivedPath != "/clusters/c1/nodepools/np1/statuses" {
		t.Errorf("expected upstream path /clusters/c1/nodepools/np1/statuses, got %q", receivedPath)
	}
}

// TestPostStatusesUpstreamError verifies upstream 422 is forwarded correctly.
func TestPostStatusesUpstreamError(t *testing.T) {
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"title":"Unprocessable Entity","status":422,"detail":"adapter name required"}`)
	})

	srv, _ := newTestServer(t, upstream)
	rr := doPost(srv, "/api/clusters/abc/statuses", []byte(`{}`))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json content-type, got %q", ct)
	}
}

// TestPostToNonStatusRouteReturns405 verifies GET-only routes reject POST.
func TestPostToNonStatusRouteReturns405(t *testing.T) {
	srv := New(nil, 0, nil)
	paths := []string{
		"/api/clusters",
		"/api/clusters/abc",
		"/api/clusters/abc/nodepools",
		"/api/clusters/abc/nodepools/np1",
	}
	for _, p := range paths {
		rr := doPost(srv, p, nil)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", p, rr.Code)
		}
	}
}
