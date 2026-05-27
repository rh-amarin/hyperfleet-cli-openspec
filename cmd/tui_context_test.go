package cmd

import (
	"os"
	"testing"
)

func TestLocalPortFromEndpoint(t *testing.T) {
	tests := []struct {
		endpoint string
		want     int
		ok       bool
	}{
		{"http://127.0.0.1:8000", 8000, true},
		{"127.0.0.1:8090", 8090, true},
		{"", 0, false},
		{"http://localhost", 0, false},
	}
	for _, tc := range tests {
		got, ok := localPortFromEndpoint(tc.endpoint)
		if got != tc.want || ok != tc.ok {
			t.Errorf("localPortFromEndpoint(%q) = (%d, %v), want (%d, %v)", tc.endpoint, got, ok, tc.want, tc.ok)
		}
	}
}

func TestAutoPortForwardLinesUsesEnvEndpoints(t *testing.T) {
	t.Setenv("HF_API_URL", "http://127.0.0.1:49152")
	t.Setenv("HF_MAESTRO_HTTP", "http://127.0.0.1:49153")
	t.Setenv("HF_MAESTRO_GRPC", "127.0.0.1:49154")

	dir := t.TempDir()
	makeEnvRaw(t, dir, "dev", `hyperfleet:
  auto-port-forward: "true"
  api-url: http://old:8000
maestro:
  http-endpoint: http://old:8100
  grpc-endpoint: old:8090
`)
	setActiveEnv(t, dir, "dev")
	t.Setenv("HF_CONFIG_DIR", dir)

	s, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	lines := autoPortForwardLines(s)
	if len(lines) != 3 {
		t.Fatalf("expected 3 auto port-forward lines, got %d", len(lines))
	}
	if lines[0].Name != "hyperfleet-api" || lines[0].LocalPort != 49152 {
		t.Fatalf("hyperfleet-api line = %+v, want port 49152", lines[0])
	}
	if lines[1].LocalPort != 49153 {
		t.Fatalf("maestro-http line = %+v, want port 49153", lines[1])
	}
	if lines[2].LocalPort != 49154 {
		t.Fatalf("maestro-grpc line = %+v, want port 49154", lines[2])
	}
}

func TestTuiContextSnapshotAutoPortForwardFlag(t *testing.T) {
	dir := t.TempDir()
	makeEnvRaw(t, dir, "dev", `hyperfleet:
  auto-port-forward: "true"
`)
	setActiveEnv(t, dir, "dev")
	t.Setenv("HF_CONFIG_DIR", dir)
	os.Unsetenv("HF_API_URL")
	os.Unsetenv("HF_MAESTRO_HTTP")
	os.Unsetenv("HF_MAESTRO_GRPC")

	info := tuiContextSnapshot()
	if !info.AutoPortForward {
		t.Fatal("expected AutoPortForward=true")
	}
	if len(info.PortForwards) != 3 {
		t.Fatalf("expected 3 auto port-forward entries, got %d", len(info.PortForwards))
	}
}
