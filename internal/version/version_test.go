package version_test

import (
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/version"
)

func TestVersionNonEmpty(t *testing.T) {
	if version.Version == "" {
		t.Fatal("Version must not be empty")
	}
}

func TestStringReturnsVersion(t *testing.T) {
	if got := version.String(); got != version.Version {
		t.Fatalf("String() = %q; want %q", got, version.Version)
	}
}
