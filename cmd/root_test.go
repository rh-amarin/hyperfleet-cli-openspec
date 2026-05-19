// Package cmd contains tests for the root command and global flags.
// Using the internal (white-box) test package so we can access package-level vars.
package cmd

import (
	"bytes"
	"os"
	"testing"
)

// TestExecuteHelp verifies that Execute() returns nil when --help is passed.
func TestExecuteHelp(t *testing.T) {
	// Redirect output to a buffer so help text doesn't pollute test output.
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--help"})

	err := Execute()
	if err != nil {
		t.Fatalf("Execute() returned error for --help: %v", err)
	}

	// Reset args and output so other tests are unaffected.
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestGlobalFlagsRegistered verifies that every global flag is present on the root command.
func TestGlobalFlagsRegistered(t *testing.T) {
	flags := []string{
		"config",
		"output",
		"no-color",
		"verbose",
		"api-url",
		"api-token",
	}

	pf := rootCmd.PersistentFlags()
	for _, name := range flags {
		if pf.Lookup(name) == nil {
			t.Errorf("persistent flag %q is not registered on rootCmd", name)
		}
	}
}

// TestAutoPortForward_DisabledByDefault verifies that when auto-port-forward is not
// enabled, PersistentPreRunE does not set HF_API_URL to a localhost forwarded address.
func TestAutoPortForward_DisabledByDefault(t *testing.T) {
	t.Setenv("HF_API_URL", "")

	dir := t.TempDir()
	makeEnv(t, dir, "test", "http://api.example.com:8000")
	setActiveEnv(t, dir, "test")

	// Exercise PersistentPreRunE; actual K8s call failing is fine.
	_, _ = runCmd(t, dir, "kube", "namespace-clean")

	apiURL := os.Getenv("HF_API_URL")
	if len(apiURL) >= 16 && apiURL[:16] == "http://127.0.0.1" {
		t.Errorf("auto-port-forward should be disabled by default, but HF_API_URL was set to %q", apiURL)
	}
}
