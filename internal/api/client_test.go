package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
)

type testResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func newClient(t *testing.T, baseURL string) *api.Client {
	t.Helper()
	return api.NewClient(baseURL, "test-token", false, false, "", "")
}

func TestGetHappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: got %s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testResource{ID: "1", Name: "test"})
	}))
	defer srv.Close()

	c := newClient(t, srv.URL+"/api/hyperfleet/v1/")
	res, err := api.Get[testResource](context.Background(), c, "resources/1")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if res.ID != "1" || res.Name != "test" {
		t.Errorf("unexpected result: %+v", res)
	}
}

func TestPostHappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %s, want POST", r.Method)
		}
		var body testResource
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		body.ID = "new-id"
		json.NewEncoder(w).Encode(body)
	}))
	defer srv.Close()

	c := newClient(t, srv.URL+"/api/hyperfleet/v1/")
	res, err := api.Post[testResource](context.Background(), c, "resources", testResource{Name: "new"})
	if err != nil {
		t.Fatalf("Post error: %v", err)
	}
	if res.ID != "new-id" {
		t.Errorf("unexpected ID: %q", res.ID)
	}
}

func TestPatchHappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method: got %s, want PATCH", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testResource{ID: "1", Name: "patched"})
	}))
	defer srv.Close()

	c := newClient(t, srv.URL+"/api/hyperfleet/v1/")
	res, err := api.Patch[testResource](context.Background(), c, "resources/1", map[string]string{"name": "patched"})
	if err != nil {
		t.Fatalf("Patch error: %v", err)
	}
	if res.Name != "patched" {
		t.Errorf("unexpected Name: %q", res.Name)
	}
}

func TestDeleteHappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method: got %s, want DELETE", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testResource{ID: "1", Name: "deleted"})
	}))
	defer srv.Close()

	c := newClient(t, srv.URL+"/api/hyperfleet/v1/")
	res, err := api.Delete[testResource](context.Background(), c, "resources/1")
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	if res.ID != "1" {
		t.Errorf("unexpected ID: %q", res.ID)
	}
}

func TestRFC7807ErrorParsing404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"type":   "about:blank",
			"title":  "Not Found",
			"status": 404,
			"detail": "Cluster not found",
		})
	}))
	defer srv.Close()

	c := newClient(t, srv.URL+"/api/hyperfleet/v1/")
	_, err := api.Get[testResource](context.Background(), c, "clusters/bad-id")
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *api.APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 404 {
		t.Errorf("Status: got %d, want 404", apiErr.Status)
	}
	if apiErr.Title != "Not Found" {
		t.Errorf("Title: got %q", apiErr.Title)
	}
	if apiErr.Detail != "Cluster not found" {
		t.Errorf("Detail: got %q", apiErr.Detail)
	}
	if !strings.Contains(apiErr.Error(), "[404]") {
		t.Errorf("Error() format: got %q", apiErr.Error())
	}
}

func TestRFC7807ValidationErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"type":   "about:blank",
			"title":  "Bad Request",
			"status": 400,
			"detail": "Validation failed",
			"errors": []map[string]any{
				{"field": "name", "message": "required", "constraint": "required"},
			},
		})
	}))
	defer srv.Close()

	c := newClient(t, srv.URL+"/api/hyperfleet/v1/")
	_, err := api.Post[testResource](context.Background(), c, "resources", nil)
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *api.APIError, got %T", err)
	}
	if len(apiErr.Errors) != 1 {
		t.Fatalf("Errors len: got %d, want 1", len(apiErr.Errors))
	}
	if apiErr.Errors[0].Field != "name" {
		t.Errorf("Field: got %q", apiErr.Errors[0].Field)
	}
}

func TestNonJSONErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer srv.Close()

	c := newClient(t, srv.URL+"/api/hyperfleet/v1/")
	_, err := api.Get[testResource](context.Background(), c, "resources/1")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *api.APIError, got %T", err)
	}
	if apiErr.Status != 500 {
		t.Errorf("Status: got %d", apiErr.Status)
	}
	if apiErr.Detail != "internal server error" {
		t.Errorf("Detail: got %q", apiErr.Detail)
	}
}

func TestHTMLErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("<!DOCTYPE html><html><body>Bad Gateway</body></html>"))
	}))
	defer srv.Close()

	c := newClient(t, srv.URL+"/api/hyperfleet/v1/")
	_, err := api.Get[testResource](context.Background(), c, "resources/1")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *api.APIError, got %T", err)
	}
	if !strings.Contains(apiErr.Detail, "Received HTML response") {
		t.Errorf("HTML detail prefix missing: got %q", apiErr.Detail)
	}
}

func TestBearerTokenHeader(t *testing.T) {
	var capturedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testResource{ID: "1"})
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL+"/api/hyperfleet/v1/", "secret-token", false, false, "", "")
	api.Get[testResource](context.Background(), c, "resources/1")

	if capturedAuth != "Bearer secret-token" {
		t.Errorf("Authorization header: got %q", capturedAuth)
	}
}

func TestNoTokenNoAuthHeader(t *testing.T) {
	var capturedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testResource{ID: "1"})
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL+"/api/hyperfleet/v1/", "", false, false, "", "")
	api.Get[testResource](context.Background(), c, "resources/1")

	if capturedAuth != "" {
		t.Errorf("expected no Authorization header, got %q", capturedAuth)
	}
}

func TestVerboseLogging(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testResource{ID: "1"})
	}))
	defer srv.Close()

	// verbose=true should not panic or error
	c := api.NewClient(srv.URL+"/api/hyperfleet/v1/", "", true, false, "", "")
	_, err := api.Get[testResource](context.Background(), c, "resources/1")
	if err != nil {
		t.Fatalf("Get with verbose: %v", err)
	}
}

func TestResourceHref(t *testing.T) {
	c := api.NewClient("http://localhost:8000/api/hyperfleet/v1/", "", false, false, "", "")

	href := c.ResourceHref("clusters/abc-123")
	want := "http://localhost:8000/api/hyperfleet/v1/clusters/abc-123"
	if href != want {
		t.Errorf("ResourceHref: got %q, want %q", href, want)
	}
}

func TestTimeoutErrorFormat(t *testing.T) {
	// Server that blocks until its context is cancelled (simulates a hung server).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL+"/api/hyperfleet/v1/", "", false, false, "", "")

	// A context with a very short deadline causes the HTTP client to return a
	// timeout error (url.Error wrapping context.DeadlineExceeded, Timeout()==true).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	_, err := api.Get[testResource](ctx, c, "resources/1")
	if err == nil {
		t.Fatal("expected error from timeout")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "[ERROR] Request to") {
		t.Errorf("timeout error missing '[ERROR] Request to' prefix: got %q", errStr)
	}
	if !strings.Contains(errStr, "timed out after 30s") {
		t.Errorf("timeout error missing 'timed out after 30s': got %q", errStr)
	}
}

func TestCurlModeDryRun(t *testing.T) {
	var reqCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testResource{ID: "1"})
	}))
	defer srv.Close()

	var stderr strings.Builder
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&stderr, r)
		close(done)
	}()

	c := api.NewClient(srv.URL+"/api/hyperfleet/v1/", "tok", false, true, "", "")
	_, err = api.Get[testResource](context.Background(), c, "resources/1")

	w.Close()
	os.Stderr = old
	<-done

	if !errors.Is(err, api.ErrDryRun) {
		t.Fatalf("expected ErrDryRun, got %v", err)
	}
	if reqCount != 0 {
		t.Errorf("expected no HTTP requests in curl mode, got %d", reqCount)
	}
	if !strings.Contains(stderr.String(), "[CURL]") {
		t.Errorf("expected curl output on stderr, got: %q", stderr.String())
	}
}

func TestCurlModeDryRunPost(t *testing.T) {
	var reqCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount++
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL+"/api/hyperfleet/v1/", "", false, true, "", "")
	_, err := api.Post[testResource](context.Background(), c, "clusters", map[string]string{"name": "x"})
	if !errors.Is(err, api.ErrDryRun) {
		t.Fatalf("expected ErrDryRun, got %v", err)
	}
	if reqCount != 0 {
		t.Errorf("expected no HTTP requests, got %d", reqCount)
	}
}

func TestIdentityHeaderInjected(t *testing.T) {
	methods := []struct {
		name string
		fn   func(c *api.Client) error
	}{
		{"GET", func(c *api.Client) error {
			_, err := api.Get[testResource](context.Background(), c, "resources/1")
			return err
		}},
		{"POST", func(c *api.Client) error {
			_, err := api.Post[testResource](context.Background(), c, "resources", map[string]string{"name": "x"})
			return err
		}},
		{"PATCH", func(c *api.Client) error {
			_, err := api.Patch[testResource](context.Background(), c, "resources/1", map[string]string{"name": "y"})
			return err
		}},
		{"DELETE", func(c *api.Client) error {
			_, err := api.Delete[testResource](context.Background(), c, "resources/1")
			return err
		}},
	}

	for _, m := range methods {
		m := m
		t.Run(m.name, func(t *testing.T) {
			var gotHeader string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotHeader = r.Header.Get("X-HyperFleet-Identity")
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(testResource{ID: "1", Name: "test"})
			}))
			defer srv.Close()

			c := api.NewClient(srv.URL+"/", "", false, false, "X-HyperFleet-Identity", "openspec@test.com")
			_ = m.fn(c)

			if gotHeader != "openspec@test.com" {
				t.Errorf("expected identity header 'openspec@test.com', got %q", gotHeader)
			}
		})
	}
}

func TestIdentityHeaderAbsentWhenNotConfigured(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-HyperFleet-Identity")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testResource{ID: "1", Name: "test"})
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL+"/", "", false, false, "", "")
	_, _ = api.Get[testResource](context.Background(), c, "resources/1")

	if gotHeader != "" {
		t.Errorf("expected no identity header, got %q", gotHeader)
	}
}

func TestIdentityHeaderInCurlOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	old := os.Stderr
	r, w, _ := os.Pipe()
	var stderr strings.Builder
	os.Stderr = w
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&stderr, r)
		close(done)
	}()

	c := api.NewClient(srv.URL+"/", "", false, true, "X-HyperFleet-Identity", "openspec@test.com")
	_, _ = api.Get[testResource](context.Background(), c, "resources/1")

	w.Close()
	os.Stderr = old
	<-done

	if !strings.Contains(stderr.String(), "X-HyperFleet-Identity: openspec@test.com") {
		t.Errorf("expected identity header in curl output, got: %q", stderr.String())
	}
}
