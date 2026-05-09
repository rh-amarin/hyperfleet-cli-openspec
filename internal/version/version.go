// Package version holds the CLI version string.
// The Version variable is overridden at build time via -ldflags:
//
//	go build -ldflags "-X github.com/rh-amarin/hyperfleet-cli/internal/version.Version=$(git describe --tags --always --dirty)"
package version

// Version is the canonical version string for the hf CLI.
// It defaults to "dev" and is replaced by the CI/CD build pipeline.
var Version = "dev"

// String returns the version string.
func String() string {
	return Version
}
