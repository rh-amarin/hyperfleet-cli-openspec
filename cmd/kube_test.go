package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// ---- K8s API response fixtures ----

// nsListJSON builds a NamespaceList JSON for the given names/labels pairs.
// Each pair is (name, comma-separated "k=v" labels).
func nsListJSON(namespaces [][2]string) string {
	items := make([]string, 0, len(namespaces))
	for _, ns := range namespaces {
		labelPairs := strings.Split(ns[1], ",")
		labelsJSON := ""
		for i, kv := range labelPairs {
			if kv == "" {
				continue
			}
			parts := strings.SplitN(kv, "=", 2)
			if i > 0 {
				labelsJSON += ","
			}
			labelsJSON += fmt.Sprintf("%q:%q", parts[0], parts[1])
		}
		items = append(items, fmt.Sprintf(
			`{"apiVersion":"v1","kind":"Namespace","metadata":{"name":%q,"labels":{%s}}}`,
			ns[0], labelsJSON,
		))
	}
	return fmt.Sprintf(
		`{"apiVersion":"v1","kind":"NamespaceList","metadata":{"resourceVersion":"1"},"items":[%s]}`,
		strings.Join(items, ","),
	)
}

const kubeStatusOK = `{"apiVersion":"v1","kind":"Status","status":"Success","code":200}`
const kubeStatusNotFound = `{"apiVersion":"v1","kind":"Status","status":"Failure","reason":"NotFound","code":404,"message":"namespaces %q not found"}`

// ---- test helpers ----

// writeKubeconfig writes a minimal HTTP kubeconfig file pointing to serverURL.
func writeKubeconfig(t *testing.T, dir, serverURL string) string {
	t.Helper()
	cfg := fmt.Sprintf(
		"apiVersion: v1\nkind: Config\nclusters:\n- cluster:\n    server: %s\n  name: test\n"+
			"contexts:\n- context:\n    cluster: test\n    user: test\n  name: test\n"+
			"current-context: test\nusers:\n- name: test\n  user:\n    token: test-token\n",
		serverURL,
	)
	path := filepath.Join(dir, "kubeconfig")
	if err := os.WriteFile(path, []byte(cfg), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func resetKubeFlags() {
	kubeConfigFlag = ""
}

// setupKubeEnv creates a temp dir with an active environment (no API URL needed for kube cmds).
func setupKubeEnv(t *testing.T, serverURL string) (dir, kubeconfigPath string) {
	t.Helper()
	dir = t.TempDir()
	makeEnv(t, dir, "test", "http://unused:8000")
	setActiveEnv(t, dir, "test")
	kubeconfigPath = writeKubeconfig(t, dir, serverURL)
	return
}

// runNamespaceClean runs `hf kube namespace-clean` with the given kubeconfig and stdin.
func runNamespaceClean(t *testing.T, dir, kubeconfigPath, stdin string) (string, error) {
	t.Helper()
	resetKubeFlags()
	rootCmd.SetIn(strings.NewReader(stdin))
	defer rootCmd.SetIn(nil)
	return runCmd(t, dir, "kube", "--kubeconfig", kubeconfigPath, "namespace-clean")
}

// ---- tests ----

func TestNamespaceClean_Empty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/namespaces" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nsListJSON(nil))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	dir, kubeconfig := setupKubeEnv(t, ts.URL)
	out, err := runNamespaceClean(t, dir, kubeconfig, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No namespaces with 'hyperfleet' labels found") {
		t.Errorf("expected empty-list message, got: %q", out)
	}
}

func TestNamespaceClean_Abort(t *testing.T) {
	var deleteCalled atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/namespaces":
			fmt.Fprint(w, nsListJSON([][2]string{
				{"hyperfleet-e2e-test", "managed-by=hyperfleet"},
				{"kube-system", "kubernetes.io/metadata.name=kube-system"},
			}))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/v1/namespaces/"):
			deleteCalled.Add(1)
			fmt.Fprint(w, kubeStatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	dir, kubeconfig := setupKubeEnv(t, ts.URL)
	out, err := runNamespaceClean(t, dir, kubeconfig, "n\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "hyperfleet-e2e-test") {
		t.Errorf("expected namespace list in output, got: %q", out)
	}
	if !strings.Contains(out, "Aborted") {
		t.Errorf("expected Aborted message, got: %q", out)
	}
	if n := deleteCalled.Load(); n != 0 {
		t.Errorf("expected 0 DELETE calls after abort, got %d", n)
	}
}

func TestNamespaceClean_Success(t *testing.T) {
	var deleteCalled atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/namespaces":
			fmt.Fprint(w, nsListJSON([][2]string{
				{"hyperfleet-e2e-test", "managed-by=hyperfleet"},
				{"hyperfleet-staging", "app=hyperfleet-worker"},
			}))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/v1/namespaces/"):
			deleteCalled.Add(1)
			fmt.Fprint(w, kubeStatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	dir, kubeconfig := setupKubeEnv(t, ts.URL)
	out, err := runNamespaceClean(t, dir, kubeconfig, "y\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := deleteCalled.Load(); n != 2 {
		t.Errorf("expected 2 DELETE calls, got %d", n)
	}
	if !strings.Contains(out, "Done. Deleted 2 namespace(s).") {
		t.Errorf("expected done summary, got: %q", out)
	}
	if !strings.Contains(out, "[DELETED]") {
		t.Errorf("expected [DELETED] lines in output, got: %q", out)
	}
}

func TestNamespaceClean_PartialFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/namespaces":
			fmt.Fprint(w, nsListJSON([][2]string{
				{"hyperfleet-good", "managed-by=hyperfleet"},
				{"hyperfleet-bad", "managed-by=hyperfleet"},
			}))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/namespaces/hyperfleet-good":
			fmt.Fprint(w, kubeStatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/namespaces/hyperfleet-bad":
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, kubeStatusNotFound, "hyperfleet-bad")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	dir, kubeconfig := setupKubeEnv(t, ts.URL)
	out, err := runNamespaceClean(t, dir, kubeconfig, "y\n")
	if err == nil {
		t.Fatal("expected non-nil error for partial failure")
	}
	if !strings.Contains(out, "[DELETED]") {
		t.Errorf("expected at least one [DELETED] in output, got: %q", out)
	}
	if !strings.Contains(err.Error(), "failed to delete") {
		t.Errorf("expected 'failed to delete' in error, got: %v", err)
	}
}

func TestNamespaceClean_LabelValueMatch(t *testing.T) {
	var deleteCalled atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/namespaces":
			// Label key does NOT contain "hyperfleet" but label value does.
			fmt.Fprint(w, nsListJSON([][2]string{
				{"my-tenant-ns", "owner=hyperfleet-team"},
			}))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/v1/namespaces/"):
			deleteCalled.Add(1)
			fmt.Fprint(w, kubeStatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	dir, kubeconfig := setupKubeEnv(t, ts.URL)
	out, err := runNamespaceClean(t, dir, kubeconfig, "y\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := deleteCalled.Load(); n != 1 {
		t.Errorf("expected 1 DELETE call for value-matched namespace, got %d", n)
	}
	_ = out
}

func TestNamespaceClean_NonHyperfleetNamespacesSkipped(t *testing.T) {
	var deleteCalled atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/namespaces":
			fmt.Fprint(w, nsListJSON([][2]string{
				{"kube-system", "kubernetes.io/metadata.name=kube-system"},
				{"default", "kubernetes.io/metadata.name=default"},
			}))
		case r.Method == http.MethodDelete:
			deleteCalled.Add(1)
			fmt.Fprint(w, kubeStatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	dir, kubeconfig := setupKubeEnv(t, ts.URL)
	out, err := runNamespaceClean(t, dir, kubeconfig, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := deleteCalled.Load(); n != 0 {
		t.Errorf("expected 0 DELETE calls for non-hyperfleet namespaces, got %d", n)
	}
	if !strings.Contains(out, "No namespaces with 'hyperfleet' labels found") {
		t.Errorf("expected empty-list message, got: %q", out)
	}
}
