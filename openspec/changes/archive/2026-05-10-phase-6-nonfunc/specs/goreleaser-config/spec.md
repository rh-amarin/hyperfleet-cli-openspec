# GoReleaser Config Spec

## ADDED Requirements

### Requirement: Cross-platform release builds

The project SHALL provide a GoReleaser configuration that produces release-ready binaries for all supported platforms.

#### Scenario: Linux amd64 build

- GIVEN a tagged release `v*`
- WHEN GoReleaser runs
- THEN it MUST produce a statically linked `hf` binary for `linux/amd64`
- AND the archive MUST be named `hf_<version>_linux_amd64.tar.gz`

#### Scenario: macOS arm64 build

- GIVEN a tagged release `v*`
- WHEN GoReleaser runs
- THEN it MUST produce a statically linked `hf` binary for `darwin/arm64`
- AND the archive MUST be named `hf_<version>_darwin_arm64.tar.gz`

#### Scenario: Windows build uses zip format

- GIVEN a tagged release `v*`
- WHEN GoReleaser runs
- THEN the Windows archives MUST use `.zip` format instead of `.tar.gz`

#### Scenario: Version is injected at build time

- GIVEN a tagged release `v1.2.3`
- WHEN GoReleaser builds the binary
- THEN `hf version` MUST output `1.2.3`
- AND the binary MUST have been built with `-X github.com/rh-amarin/hyperfleet-cli/internal/version.Version=1.2.3`

#### Scenario: Checksums file is generated

- GIVEN a tagged release
- WHEN GoReleaser runs
- THEN it MUST produce a `checksums.txt` file listing SHA256 hashes of all archives
