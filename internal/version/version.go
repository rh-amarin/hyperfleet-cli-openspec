// Package version holds the CLI version string.
// The Version variable is overridden at build time via -ldflags:
//
//	go build -ldflags "-X github.com/rh-amarin/hyperfleet-cli/internal/version.Version=$(git describe --tags --always --dirty)"
package version

// Version is the canonical version string for the hf CLI.
// It defaults to "dev" and is replaced by the CI/CD build pipeline.
var Version = "dev"

// BuildTime is the RFC3339 timestamp of when the binary was compiled.
// It defaults to "unknown" and is injected at link time via -ldflags.
var BuildTime = "unknown"

// String returns the version string including build time.
func String() string {
	return Version + " (built " + BuildTime + ")"
}
