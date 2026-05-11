package cmd

import (
	"strings"
	"testing"
)

func TestCompletionBash(t *testing.T) {
	out, err := runCmd(t, "", "completion", "bash")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "bash") {
		t.Error("expected bash completion output")
	}
}

func TestCompletionZsh(t *testing.T) {
	out, err := runCmd(t, "", "completion", "zsh")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "#compdef") {
		t.Error("expected zsh completion output")
	}
}

func TestCompletionFish(t *testing.T) {
	out, err := runCmd(t, "", "completion", "fish")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Error("expected non-empty fish completion output")
	}
}

func TestCompletionPowerShell(t *testing.T) {
	out, err := runCmd(t, "", "completion", "powershell")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Error("expected non-empty powershell completion output")
	}
}
