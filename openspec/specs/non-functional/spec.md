# Non-Functional Requirements Specification

## Purpose

Define non-functional requirements for the HyperFleet CLI covering shell completions, multi-format output, cross-platform builds, CI/CD pipelines, testing strategy, and distribution.

## Requirements

### Requirement: Shell Completions

The CLI SHALL generate shell completion scripts for all major shells.

#### Scenario: Generate bash completion

- GIVEN the user has installed `hf`
- WHEN the user runs `hf completion bash`
- THEN the CLI MUST output a bash completion script to stdout
- AND the output MUST be compatible with `source <(hf completion bash)`

#### Scenario: Generate zsh completion

- GIVEN the user has installed `hf`
- WHEN the user runs `hf completion zsh`
- THEN the CLI MUST output a zsh completion script to stdout
- AND the output MUST begin with `#compdef hf`

#### Scenario: Generate fish completion

- GIVEN the user has installed `hf`
- WHEN the user runs `hf completion fish`
- THEN the CLI MUST output a non-empty fish completion script to stdout

#### Scenario: Generate powershell completion

- GIVEN the user has installed `hf`
- WHEN the user runs `hf completion powershell`
- THEN the CLI MUST output a non-empty PowerShell completion script to stdout

#### Scenario: Reject invalid shell argument

- GIVEN the user runs `hf completion invalid`
- THEN the CLI MUST exit with a non-zero exit code
- AND MUST NOT output a completion script

#### Scenario: Dynamic completions

- GIVEN the user is tab-completing arguments
- WHEN a command accepts known values (e.g., `spec|labels` for patch, `start|stop|status` for port-forward)
- THEN the CLI MUST provide those values as completion suggestions
- AND for resource IDs (cluster-id, nodepool-id), the CLI SHOULD offer live completions by querying the API

#### Scenario: Completion installation instructions

- GIVEN the user runs `hf completion bash`
- WHEN the script is generated
- THEN the CLI MUST print instructions for installing the completion (e.g., `source <(hf completion bash)` or adding to `.bashrc`)

### Requirement: Output Flag Tab Completion

The CLI SHALL provide tab completions for the `--output` persistent flag.

#### Scenario: --output completions

- GIVEN a shell with tab completion enabled
- WHEN the user types `hf <command> --output <TAB>`
- THEN the shell MUST offer `json`, `table`, and `yaml` as completions
- AND MUST NOT offer file paths

### Requirement: Multi-Format Output

The CLI SHALL support a global `--output` flag for controlling output format on every command that produces data.

#### Scenario: JSON output

- GIVEN `--output json` is specified (or `-o json`)
- WHEN any command produces output
- THEN the output MUST be well-formed, pretty-printed JSON with 2-space indentation
- AND the output MUST be suitable for piping to `jq` or other JSON tools

#### Scenario: Table output

- GIVEN `--output table` is specified
- WHEN the command produces output
- THEN the output MUST be a human-readable formatted table with:
  - Column headers in uppercase
  - A separator line (`---`) under headers
  - Aligned columns using tab-width or padding
  - Colored status indicators (respecting `--no-color`)

#### Scenario: YAML output

- GIVEN `--output yaml` is specified
- WHEN any command produces output
- THEN the output MUST be valid YAML

#### Scenario: Default output format per command type

- GIVEN no `--output` flag is specified
- WHEN a command produces output
- THEN the default MUST be:
  - `json` for resource commands: `cluster list`, `cluster get`, `nodepool list`, `nodepool get`, `cluster conditions`, `nodepool conditions`, `cluster statuses`, `nodepool statuses`, `resources`, `table`
  - `table` for `repos` (always renders a table)
  - `text` for config commands (`hf config show`, `hf env *`), port-forward status, and log output

### Requirement: Cross-Platform Build and Distribution

The CLI SHALL be built and distributed for multiple platforms using GoReleaser.

#### Scenario: Supported platforms

- GIVEN GoReleaser is configured
- WHEN a release is built
- THEN binaries MUST be produced for:
  - `linux/amd64`
  - `linux/arm64`
  - `darwin/amd64`
  - `darwin/arm64`
  - `windows/amd64`
- AND each binary MUST be a statically linked, self-contained executable

#### Scenario: Linux amd64 archive naming

- GIVEN a tagged release `v*`
- WHEN GoReleaser runs
- THEN the Linux amd64 archive MUST be named `hf_<version>_linux_amd64.tar.gz`

#### Scenario: macOS arm64 archive naming

- GIVEN a tagged release `v*`
- WHEN GoReleaser runs
- THEN the macOS arm64 archive MUST be named `hf_<version>_darwin_arm64.tar.gz`

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

#### Scenario: Release artifacts

- GIVEN a new version is tagged
- WHEN GoReleaser runs
- THEN it MUST produce:
  - Compressed archives (`.tar.gz` for linux/mac, `.zip` for windows)
  - SHA256 checksums file
  - GitHub Release with changelog from conventional commits
- AND the binary name MUST be `hf`

#### Scenario: Version information

- GIVEN the binary is built with ldflags
- WHEN the user runs `hf version`
- THEN the CLI MUST display the following fields as plain text:
  ```
  Version:    v1.2.3
  Commit:     abc1234
  Built:      2026-04-28T12:00:00Z
  Go version: go1.22.0
  OS/Arch:    linux/amd64
  ```
- AND values MUST be injected at build time via `-ldflags`
- AND `--output json` MUST be supported, outputting the same fields as a JSON object

#### Scenario: Homebrew and package managers

- GIVEN the CLI is released
- WHEN distribution channels are configured
- THEN GoReleaser SHOULD generate a Homebrew formula
- AND the CLI SHOULD be installable via `brew install hf`

### Requirement: CI/CD Pipeline

The CLI SHALL have automated build, test, and release pipelines via GitHub Actions.

#### Scenario: CI on pull requests

- GIVEN a pull request is opened against `main` or a commit is pushed to `main`
- WHEN the CI workflow runs
- THEN it MUST execute in order: `go build ./...`, `go vet ./...`, `go test ./...`
- AND it MUST NOT run integration tests (no `-tags integration` flag)
- AND the workflow MUST fail if any step exits non-zero

#### Scenario: Go version is pinned via go.mod

- GIVEN the workflow runs
- WHEN `actions/setup-go` installs Go
- THEN it MUST read the Go version from `go.mod`
- AND MUST enable the Go module cache

#### Scenario: Release on tag push

- GIVEN a tag matching `v*` is pushed
- WHEN the release workflow runs
- THEN it MUST invoke GoReleaser with `--clean`
- AND produce cross-platform binaries, archives, checksums, and a GitHub Release automatically

#### Scenario: Full git history is available to GoReleaser

- GIVEN the release workflow runs
- WHEN the repository is checked out
- THEN `fetch-depth: 0` MUST be set so GoReleaser can generate a full changelog

#### Scenario: GITHUB_TOKEN is passed to GoReleaser

- GIVEN the release workflow runs
- WHEN GoReleaser attempts to create a GitHub release and upload assets
- THEN it MUST have access to `GITHUB_TOKEN` from secrets
- AND the token MUST be passed via the `env` block of the goreleaser-action step

### Requirement: Error Output Conventions

The CLI SHALL follow consistent conventions for all output to stdout and stderr.

| Situation | Destination | Format | Exit code |
|---|---|---|---|
| API error (RFC 9457) | stdout | rendered via `--output` format | 0 |
| CLI usage error | stderr | `Error: <msg>` (usage suppressed) | 1 |
| `[WARN]` message | stderr | `[WARN] <msg>` | 0 |
| `[INFO]` message | stderr | `[INFO] <msg>` | 0 |
| `[ERROR]` message | stderr | `[ERROR] <msg>` | 1 (unless otherwise specified per command) |

`SilenceUsage: true` on the root command suppresses the usage block for all subcommands on runtime errors.

### Requirement: Testing Strategy

The CLI SHALL maintain a comprehensive test suite.

#### Scenario: Unit tests

- GIVEN all `internal/` packages contain business logic
- WHEN tests are run
- THEN each package MUST have unit tests covering:
  - Config loading, merging, and precedence
  - API client request construction and error parsing
  - Output formatting (JSON, table, YAML rendering)
  - Resource type marshaling/unmarshaling

#### Scenario: Integration tests

- GIVEN the CLI interacts with external services (API, database, Kubernetes)
- WHEN integration tests are run
- THEN tests MUST use:
  - `httptest.Server` for API client tests against a mock HyperFleet API
  - `testcontainers-go` or similar for PostgreSQL integration tests
  - `envtest` from controller-runtime for Kubernetes client tests
- AND integration tests MUST be tagged with `//go:build integration`

#### Scenario: Command tests

- GIVEN commands are defined in `cmd/`
- WHEN command tests are run
- THEN each command MUST have at least one test verifying:
  - Correct flag parsing
  - Expected API calls (using HTTP mock)
  - Expected output format
  - Exit code behavior

#### Scenario: End-to-end tests

- GIVEN a full test environment is available
- WHEN E2E tests are run
- THEN the CLI MUST be tested against a real HyperFleet API instance
- AND the test suite MUST exercise the full cluster lifecycle: create, get, list, patch, adapter-status, conditions, delete

### Requirement: Graceful Degradation

The CLI SHALL degrade gracefully when optional dependencies are unavailable.

#### Scenario: Missing GCP credentials

- GIVEN GCP credentials are not configured
- WHEN the user runs `hf pubsub list` or `hf pubsub publish`
- THEN the CLI MUST display a clear error: `[ERROR] GCP credentials not found. Run 'gcloud auth application-default login' or set GOOGLE_APPLICATION_CREDENTIALS`
- AND other commands MUST continue to work

#### Scenario: Unreachable API

- GIVEN the HyperFleet API is unreachable
- WHEN any API command is invoked
- THEN the CLI MUST display a clear error with the attempted URL
- AND suggest checking `hf config show` and `hf kube port-forward status`

#### Scenario: Database unreachable

- GIVEN the PostgreSQL database is unreachable
- WHEN a database command is invoked
- THEN the CLI MUST display the connection error with host:port
- AND suggest checking `hf config show` for database settings

### Requirement: Performance

The CLI SHALL respond promptly for common operations.

#### Scenario: Parallel data fetching

- GIVEN `hf repos` queries 7 repositories
- WHEN the data is fetched
- THEN requests MUST be made concurrently using goroutines
- AND the total execution time MUST be bounded by the slowest single request, not the sum of all requests

#### Scenario: Large list pagination

- GIVEN the API returns paginated results
- WHEN a list command encounters pagination
- THEN the CLI MUST handle pagination transparently, fetching all pages

### Requirement: Security

The CLI SHALL follow security best practices.

#### Scenario: Token storage

- GIVEN API tokens are stored in config
- WHEN the config file is created or updated
- THEN the file permissions MUST be set to `0600` (owner read/write only)

#### Scenario: No token in logs

- GIVEN verbose mode is enabled
- WHEN HTTP requests are logged
- THEN the `Authorization` header value MUST be masked (e.g., `Bearer <redacted>`)
- AND the `password` fields in config show MUST remain masked

#### Scenario: TLS verification

- GIVEN the API URL uses HTTPS
- WHEN the CLI connects
- THEN TLS certificate verification MUST be enabled by default
- AND `--insecure` flag MAY be provided to skip verification with a warning

#### Scenario: Non-TTY color disabling

- GIVEN the CLI writes output to a pipe or file (stdout is not a TTY)
- WHEN any command produces output with colored elements
- THEN ANSI color codes MUST be disabled automatically
- AND the output MUST be plain text suitable for piping to other tools
- AND there is no `--force-color` flag; use `--no-color` to explicitly disable color on a TTY
