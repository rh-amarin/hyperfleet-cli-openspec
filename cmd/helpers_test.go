// Package cmd shared test helpers for command integration tests.
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// runCmd executes the root command with the given args, setting HF_CONFIG_DIR to
// a temp dir. Returns stdout content and the command error.
func runCmd(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	t.Setenv("HF_CONFIG_DIR", dir)
	if err := preloadResourceCommands(args); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)

	return buf.String(), err
}

// makeEnv creates a named environment file in dir/environments/<name>.yaml.
func makeEnv(t *testing.T, dir, name, apiURL string) {
	t.Helper()
	makeEnvRaw(t, dir, name, "hyperfleet:\n  api-url: "+apiURL+"\n")
}

// makeReconciledClusterEnv adds clusters and nodepools resource-types for hf rs tests.
func makeReconciledClusterEnv(t *testing.T, dir, name, apiURL string) {
	t.Helper()
	makeEnvRaw(t, dir, name, fmt.Sprintf(`hyperfleet:
  api-url: %s
  api-version: v1
resource-types:
  clusters:
    path: clusters
    state-key: cluster-id
    create-template: clusters.json
  nodepools:
    parent: clusters
    path: "clusters/{cluster_id}/nodepools"
    state-key: nodepool-id
    create-template: nodepools.json
`, apiURL))
}

// makeEnvRaw creates a named environment file with arbitrary content.
func makeEnvRaw(t *testing.T, dir, name, content string) {
	t.Helper()
	envDir := filepath.Join(dir, "environments")
	if err := os.MkdirAll(envDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envDir, name+".yaml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

// setActiveEnv writes state.yaml in dir to set the active environment.
func setActiveEnv(t *testing.T, dir, name string) {
	t.Helper()
	statePath := filepath.Join(dir, "state.yaml")
	if err := os.WriteFile(statePath, []byte("active-environment: "+name+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
}
