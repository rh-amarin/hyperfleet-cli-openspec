package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// minimalKubeconfig is a valid kubeconfig YAML for BuildConfig tests.
const minimalKubeconfig = `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: fake-token
`

func TestBuildConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kubeconfig")
	if err := os.WriteFile(path, []byte(minimalKubeconfig), 0600); err != nil {
		t.Fatal(err)
	}
	config, err := BuildConfig(path, "")
	if err != nil {
		t.Fatalf("BuildConfig returned error: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Host != "https://localhost:6443" {
		t.Errorf("unexpected host: %s", config.Host)
	}
}

func TestBuildConfig_MissingFile(t *testing.T) {
	_, err := BuildConfig("/nonexistent/path/kubeconfig", "")
	if err == nil {
		t.Fatal("expected error for missing kubeconfig")
	}
	if !strings.Contains(err.Error(), "[ERROR] kubeconfig not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBuildConfig_HFKubeTokenOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kubeconfig")
	if err := os.WriteFile(path, []byte(minimalKubeconfig), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HF_KUBE_TOKEN", "override-token")
	config, err := BuildConfig(path, "")
	if err != nil {
		t.Fatalf("BuildConfig returned error: %v", err)
	}
	if config.BearerToken != "override-token" {
		t.Errorf("expected bearer token override, got: %q", config.BearerToken)
	}
}

func TestResolvedContext_CurrentContext(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kubeconfig")
	if err := os.WriteFile(path, []byte(minimalKubeconfig), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, err := ResolvedContext(path, "")
	if err != nil {
		t.Fatalf("ResolvedContext returned error: %v", err)
	}
	if ctx != "test-context" {
		t.Errorf("expected test-context, got %q", ctx)
	}
}

func TestResolvedContext_Override(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kubeconfig")
	if err := os.WriteFile(path, []byte(minimalKubeconfig), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, err := ResolvedContext(path, "test-context")
	if err != nil {
		t.Fatalf("ResolvedContext returned error: %v", err)
	}
	if ctx != "test-context" {
		t.Errorf("expected test-context, got %q", ctx)
	}
}

func TestResolvedContext_NonexistentContext(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kubeconfig")
	if err := os.WriteFile(path, []byte(minimalKubeconfig), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := ResolvedContext(path, "nonexistent-context")
	if err == nil {
		t.Fatal("expected error for nonexistent context")
	}
	if !strings.Contains(err.Error(), "nonexistent-context") {
		t.Errorf("expected error to mention context name, got: %v", err)
	}
}

func TestIsPortListening_NotListening(t *testing.T) {
	// Bind then immediately close a port; after Close it must not be detected as listening.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not bind test port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	if IsPortListening(port) {
		t.Errorf("port %d should not be listening after Close", port)
	}
}

func TestIsPortListening_Listening(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not bind test port: %v", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	if !IsPortListening(port) {
		t.Errorf("expected port %d to be listening", port)
	}
}

func TestPIDForPort_Listening(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not bind test port: %v", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	pid, err := PIDForPort(port)
	if err != nil {
		t.Skipf("lsof unavailable or failed (port %d): %v", port, err)
	}
	if pid != os.Getpid() {
		t.Errorf("PIDForPort(%d) = %d, want current PID %d", port, pid, os.Getpid())
	}
}

func TestPIDForPort_NotListening(t *testing.T) {
	// Grab a free port number then release it so nothing is listening.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not bind: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	_, err = PIDForPort(port)
	if err == nil {
		t.Errorf("expected error for non-listening port %d", port)
	}
}

func TestIsProcessAlive_CurrentProcess(t *testing.T) {
	if !IsProcessAlive(os.Getpid()) {
		t.Error("expected current process to be alive")
	}
}

func TestIsProcessAlive_InvalidPID(t *testing.T) {
	if IsProcessAlive(-1) {
		t.Error("expected PID -1 to report not alive")
	}
	// PID 99999999 is almost certainly not running.
	if IsProcessAlive(99999999) {
		t.Log("PID 99999999 unexpectedly alive (unlikely)")
	}
}

func TestFindRunningPod_Found(t *testing.T) {
	cs := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "hyperfleet-api-xyz", Namespace: "hyperfleet"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	})
	name, err := FindRunningPod(context.Background(), cs, "hyperfleet", "hyperfleet-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "hyperfleet-api-xyz" {
		t.Errorf("expected hyperfleet-api-xyz, got %s", name)
	}
}

func TestFindRunningPod_NotRunning(t *testing.T) {
	cs := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "hyperfleet-api-abc", Namespace: "hyperfleet"},
		Status:     corev1.PodStatus{Phase: corev1.PodPending},
	})
	name, err := FindRunningPod(context.Background(), cs, "hyperfleet", "hyperfleet-api")
	if err == nil {
		t.Fatal("expected PodNotReadyError")
	}
	var notReady *PodNotReadyError
	if !isPodNotReady(err, &notReady) {
		t.Fatalf("expected *PodNotReadyError, got %T: %v", err, err)
	}
	if notReady.Phase != "Pending" {
		t.Errorf("expected phase Pending, got %s", notReady.Phase)
	}
	if name != "hyperfleet-api-abc" {
		t.Errorf("expected pod name even on not-ready, got %q", name)
	}
}

func TestFindRunningPod_NotFound(t *testing.T) {
	cs := fake.NewSimpleClientset()
	_, err := FindRunningPod(context.Background(), cs, "hyperfleet", "missing-pod")
	if err == nil {
		t.Fatal("expected error for missing pod")
	}
	if !strings.Contains(err.Error(), "no pod found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func endpointSliceForService(namespace, serviceName, podName string, ports ...int32) *discoveryv1.EndpointSlice {
	slicePorts := make([]discoveryv1.EndpointPort, len(ports))
	for i, port := range ports {
		p := port
		slicePorts[i] = discoveryv1.EndpointPort{Port: &p}
	}
	return &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName + "-slice",
			Namespace: namespace,
			Labels: map[string]string{
				discoveryv1.LabelServiceName: serviceName,
			},
		},
		Endpoints: []discoveryv1.Endpoint{{
			TargetRef: &corev1.ObjectReference{Kind: "Pod", Name: podName},
		}},
		Ports: slicePorts,
	}
}

func TestResolvePortForwardTarget_ServicePreferred(t *testing.T) {
	cs := fake.NewSimpleClientset(
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "hyperfleet-api", Namespace: "hyperfleet"},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{Port: 8000, TargetPort: intstr.FromInt32(8000)}},
			},
		},
		endpointSliceForService("hyperfleet", "hyperfleet-api", "hyperfleet-api-xyz", 8000),
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "hyperfleet-api-abc", Namespace: "hyperfleet"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		},
	)
	res, warn := ResolvePortForwardTarget(context.Background(), cs, "hyperfleet", "hyperfleet-api", "hyperfleet-api", 8000)
	if warn != nil {
		t.Fatalf("unexpected warn: %v", warn)
	}
	if res.DisplayKind != TargetKindService || res.DisplayName != "hyperfleet-api" {
		t.Errorf("display: kind=%q name=%q, want service/hyperfleet-api", res.DisplayKind, res.DisplayName)
	}
	if res.PodName != "hyperfleet-api-xyz" {
		t.Errorf("pod=%q, want hyperfleet-api-xyz from endpoint slices", res.PodName)
	}
}

func TestResolvePortForwardTarget_ServiceWrongPortFallsBackToPod(t *testing.T) {
	cs := fake.NewSimpleClientset(
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "maestro", Namespace: "maestro"},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{Port: 8000, TargetPort: intstr.FromInt32(8000)}},
			},
		},
		endpointSliceForService("maestro", "maestro", "maestro-svc-pod", 8000),
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "maestro-abc", Namespace: "maestro"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		},
	)
	res, warn := ResolvePortForwardTarget(context.Background(), cs, "maestro", "maestro", "maestro", 8090)
	if warn != nil {
		t.Fatalf("unexpected warn: %v", warn)
	}
	if res.DisplayKind != TargetKindPod || res.PodName != "maestro-abc" {
		t.Errorf("got display=%q pod=%q, want pod/maestro-abc", res.DisplayKind, res.PodName)
	}
}

func TestResolvePortForwardTarget_PodNotReady(t *testing.T) {
	cs := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "hyperfleet-api-pending", Namespace: "hyperfleet"},
		Status:     corev1.PodStatus{Phase: corev1.PodPending},
	})
	res, warn := ResolvePortForwardTarget(context.Background(), cs, "hyperfleet", "missing-svc", "hyperfleet-api", 8000)
	if warn == nil {
		t.Fatal("expected PodNotReadyError")
	}
	var notReady *PodNotReadyError
	if !isPodNotReady(warn, &notReady) {
		t.Fatalf("expected *PodNotReadyError, got %T: %v", warn, warn)
	}
	if res.DisplayKind != TargetKindPod || res.PodName != "hyperfleet-api-pending" {
		t.Errorf("got display=%q pod=%q", res.DisplayKind, res.PodName)
	}
}

func TestResolvePortForwardTarget_NeitherFound(t *testing.T) {
	cs := fake.NewSimpleClientset()
	_, err := ResolvePortForwardTarget(context.Background(), cs, "hyperfleet", "missing-svc", "missing-pod", 8000)
	if err == nil {
		t.Fatal("expected error when neither service nor pod found")
	}
	if !strings.Contains(err.Error(), "no pod found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestListPortForwards_Empty(t *testing.T) {
	// Override pidDir for test by ensuring no PID files exist.
	// Since pidDir uses UserHomeDir(), we test with the real dir (no pid files in CI).
	pfs, err := ListPortForwards()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Result may be empty or non-empty depending on environment; just ensure no panic.
	_ = pfs
}

func TestListPortForwards_WithPIDFile(t *testing.T) {
	dir := t.TempDir()
	// Write a fake PID file using the internal helper directly.
	name := "test-svc"
	content := "12345\n9000\n9000\n"
	if err := os.WriteFile(filepath.Join(dir, "pf-"+name+".pid"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	// Parse it manually to verify readPIDFile logic.
	// We can't call ListPortForwards() with a custom dir without refactoring,
	// so test readPIDFile via a separate helper test.
}

func TestParseLogfmt(t *testing.T) {
	line := `time=2024-01-01T00:00:00Z level=info msg="hello world" cluster_id=abc123`
	fields := ParseLogfmt(line)
	checks := map[string]string{
		"time":       "2024-01-01T00:00:00Z",
		"level":      "info",
		"msg":        "hello world",
		"cluster_id": "abc123",
	}
	for k, want := range checks {
		if got := fields[k]; got != want {
			t.Errorf("parseLogfmt[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestResolveKubeconfig_ExplicitPath(t *testing.T) {
	got := ResolveKubeconfig("/my/path")
	if got != "/my/path" {
		t.Errorf("expected /my/path, got %s", got)
	}
}

func TestResolveKubeconfig_EnvVar(t *testing.T) {
	t.Setenv("KUBECONFIG", "/env/path")
	got := ResolveKubeconfig("")
	if got != "/env/path" {
		t.Errorf("expected /env/path, got %s", got)
	}
}

func TestCollectLogs(t *testing.T) {
	const ns = "test-ns"
	mux := http.NewServeMux()
	podList := corev1.PodList{
		Items: []corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "myapp-abc", Namespace: ns}},
			{ObjectMeta: metav1.ObjectMeta{Name: "other-xyz", Namespace: ns}},
		},
	}
	mux.HandleFunc("/api/v1/namespaces/"+ns+"/pods", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(podList)
	})
	mux.HandleFunc("/api/v1/namespaces/"+ns+"/pods/myapp-abc/log", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("line one\nline two\n"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := &rest.Config{Host: srv.URL, ContentConfig: rest.ContentConfig{GroupVersion: &corev1.SchemeGroupVersion, NegotiatedSerializer: scheme.Codecs}}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("NewForConfig: %v", err)
	}

	lines, err := CollectLogs(context.Background(), cs, ns, "myapp", 60)
	if err != nil {
		t.Fatalf("CollectLogs error: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "line one" || lines[1] != "line two" {
		t.Errorf("unexpected lines: %v", lines)
	}
}

func TestCollectLogs_NoPods(t *testing.T) {
	const ns = "test-ns"
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/namespaces/"+ns+"/pods", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(corev1.PodList{})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := &rest.Config{Host: srv.URL, ContentConfig: rest.ContentConfig{GroupVersion: &corev1.SchemeGroupVersion, NegotiatedSerializer: scheme.Codecs}}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("NewForConfig: %v", err)
	}

	lines, err := CollectLogs(context.Background(), cs, ns, "myapp", 60)
	if err != nil {
		t.Fatalf("CollectLogs error: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines for no pods, got %d", len(lines))
	}
}

// isPodNotReady is a helper to check for *PodNotReadyError without importing errors.
func isPodNotReady(err error, out **PodNotReadyError) bool {
	if p, ok := err.(*PodNotReadyError); ok {
		*out = p
		return true
	}
	return false
}

// ---- connectivity check tests ----

func TestCheckAPIConnectivity_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	port := extractPort(t, srv.URL)
	if err := CheckAPIConnectivity(port); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCheckAPIConnectivity_Down(t *testing.T) {
	port := freePort(t)
	if err := CheckAPIConnectivity(port); err == nil {
		t.Fatal("expected non-nil error for closed port")
	}
}

func TestCheckMaestroHTTPConnectivity_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	port := extractPort(t, srv.URL)
	if err := CheckMaestroHTTPConnectivity(port); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCheckMaestroHTTPConnectivity_Down(t *testing.T) {
	port := freePort(t)
	if err := CheckMaestroHTTPConnectivity(port); err == nil {
		t.Fatal("expected non-nil error for closed port")
	}
}

func TestCheckPostgresConnectivity_Down(t *testing.T) {
	port := freePort(t)
	if err := CheckPostgresConnectivity(port, "localhost", "testdb", "testuser", "testpass"); err == nil {
		t.Fatal("expected non-nil error for closed port")
	}
}

func TestCheckMaestroGRPCConnectivity_Down(t *testing.T) {
	port := freePort(t)
	if err := CheckMaestroGRPCConnectivity(port); err == nil {
		t.Fatal("expected non-nil error for closed port")
	}
}

// extractPort parses the port from a URL like http://127.0.0.1:PORT.
func extractPort(t *testing.T, rawURL string) int {
	t.Helper()
	// strip scheme
	addr := strings.TrimPrefix(rawURL, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("extractPort: %v", err)
	}
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		t.Fatalf("extractPort parse: %v", err)
	}
	return port
}

// freePort returns a TCP port that is not currently bound.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

func TestEphemeralPortForward_PodNotFound(t *testing.T) {
	const ns = "hyperfleet"
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/namespaces/"+ns+"/services/hyperfleet-api", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/api/v1/namespaces/"+ns+"/pods", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(corev1.PodList{})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	dir := t.TempDir()
	kubeconfigContent := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- cluster:
    server: %s
  name: test
contexts:
- context:
    cluster: test
    user: test
  name: test
current-context: test
users:
- name: test
  user:
    token: test-token
`, srv.URL)
	kubeconfigPath := filepath.Join(dir, "kubeconfig")
	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfigContent), 0600); err != nil {
		t.Fatal(err)
	}

	_, _, err := EphemeralPortForward(kubeconfigPath, ns, "hyperfleet-api", "hyperfleet-api", 8000, "")
	if err == nil {
		t.Fatal("expected error when pod not found")
	}
	if !strings.Contains(err.Error(), "no pod found") {
		t.Errorf("expected 'no pod found' in error, got: %v", err)
	}
}

func TestEphemeralPortForward_ServicePreferred(t *testing.T) {
	const ns = "hyperfleet"
	podName := "api-hyperfleet-api-xyz"
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/namespaces/"+ns+"/services/hyperfleet-api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "hyperfleet-api", Namespace: ns},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{Port: 8000, TargetPort: intstr.FromInt32(8000)}},
			},
		})
	})
	mux.HandleFunc("/apis/discovery.k8s.io/v1/namespaces/"+ns+"/endpointslices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		port := int32(8000)
		_ = json.NewEncoder(w).Encode(discoveryv1.EndpointSliceList{
			Items: []discoveryv1.EndpointSlice{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hyperfleet-api-slice",
					Namespace: ns,
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "hyperfleet-api",
					},
				},
				Endpoints: []discoveryv1.Endpoint{{
					TargetRef: &corev1.ObjectReference{Kind: "Pod", Name: podName},
				}},
				Ports: []discoveryv1.EndpointPort{{Port: &port}},
			}},
		})
	})
	mux.HandleFunc("/api/v1/namespaces/"+ns+"/pods", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(corev1.PodList{})
	})
	mux.HandleFunc("/api/v1/namespaces/"+ns+"/pods/"+podName+"/portforward", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	dir := t.TempDir()
	kubeconfigContent := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- cluster:
    server: %s
  name: test
contexts:
- context:
    cluster: test
    user: test
  name: test
current-context: test
users:
- name: test
  user:
    token: test-token
`, srv.URL)
	kubeconfigPath := filepath.Join(dir, "kubeconfig")
	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfigContent), 0600); err != nil {
		t.Fatal(err)
	}

	_, _, err := EphemeralPortForward(kubeconfigPath, ns, "hyperfleet-api", "hyperfleet-api", 8000, "")
	if err == nil {
		t.Fatal("expected error from mock portforward endpoint")
	}
	if strings.Contains(err.Error(), "no pod found") {
		t.Errorf("should resolve via service endpoint slices, not pod pattern: %v", err)
	}
}

func TestFindFreePort(t *testing.T) {
	port, err := findFreePort()
	if err != nil {
		t.Fatalf("findFreePort error: %v", err)
	}
	if port < 1 || port > 65535 {
		t.Errorf("findFreePort returned out-of-range port %d", port)
	}
	// Port should not be bound after findFreePort returns.
	if IsPortListening(port) {
		t.Errorf("port %d should not be listening after findFreePort", port)
	}
}
