package version_test

import (
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/version"
)

func TestVersionNonEmpty(t *testing.T) {
	if version.Version == "" {
		t.Fatal("Version must not be empty")
	}
}

func TestBuildTimeNonEmpty(t *testing.T) {
	if version.BuildTime == "" {
		t.Fatal("BuildTime must not be empty")
	}
}

func TestStringContainsVersion(t *testing.T) {
	if got := version.String(); !strings.Contains(got, version.Version) {
		t.Fatalf("String() = %q; must contain Version %q", got, version.Version)
	}
}

func TestStringContainsBuildTime(t *testing.T) {
	if got := version.String(); !strings.Contains(got, version.BuildTime) {
		t.Fatalf("String() = %q; must contain BuildTime %q", got, version.BuildTime)
	}
}
