// Package cmd contains tests for the root command and global flags.
// Using the internal (white-box) test package so we can access package-level vars.
package cmd

import (
	"bytes"
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
