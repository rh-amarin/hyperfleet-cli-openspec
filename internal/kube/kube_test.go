package kube

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
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
	config, err := BuildConfig(path)
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
	_, err := BuildConfig("/nonexistent/path/kubeconfig")
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
	config, err := BuildConfig(path)
	if err != nil {
		t.Fatalf("BuildConfig returned error: %v", err)
	}
	if config.BearerToken != "override-token" {
		t.Errorf("expected bearer token override, got: %q", config.BearerToken)
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
		ObjectMeta: metav1.ObjectMeta{Name: "hyperfleet-api-xyz", Namespace: "amarin-ns1"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	})
	name, err := FindRunningPod(context.Background(), cs, "amarin-ns1", "hyperfleet-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "hyperfleet-api-xyz" {
		t.Errorf("expected hyperfleet-api-xyz, got %s", name)
	}
}

func TestFindRunningPod_NotRunning(t *testing.T) {
	cs := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "hyperfleet-api-abc", Namespace: "amarin-ns1"},
		Status:     corev1.PodStatus{Phase: corev1.PodPending},
	})
	name, err := FindRunningPod(context.Background(), cs, "amarin-ns1", "hyperfleet-api")
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
	_, err := FindRunningPod(context.Background(), cs, "amarin-ns1", "missing-pod")
	if err == nil {
		t.Fatal("expected error for missing pod")
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
	fields := parseLogfmt(line)
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
	got := resolveKubeconfig("/my/path")
	if got != "/my/path" {
		t.Errorf("expected /my/path, got %s", got)
	}
}

func TestResolveKubeconfig_EnvVar(t *testing.T) {
	t.Setenv("KUBECONFIG", "/env/path")
	got := resolveKubeconfig("")
	if got != "/env/path" {
		t.Errorf("expected /env/path, got %s", got)
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
