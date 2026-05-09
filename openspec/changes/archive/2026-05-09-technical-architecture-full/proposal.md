## Why

The initial change (`technical-architecture-init`) established `go.mod`, `main.go`, and `cmd/root.go` (tasks 1.1–2.3). The remaining architecture scaffold — all domain command stubs, `internal/version`, a `Makefile`, and unit tests — is needed to make the module compile cleanly and give every contributor a runnable starting point before full feature implementation begins.

## What Changes

- Add stub Cobra sub-command files for every domain: `cmd/cluster.go`, `cmd/nodepool.go`, `cmd/config.go`, `cmd/db.go`, `cmd/maestro.go`, `cmd/pubsub.go`, `cmd/rabbitmq.go`, `cmd/kube.go`, `cmd/logs.go`, `cmd/repos.go`, `cmd/resources.go`
- Add `cmd/version.go` that prints the version from `internal/version`
- Add `cmd/completion.go` registering Cobra's built-in completion command
- Add `internal/version/version.go` with `var Version = "dev"` and `String()` helper
- Add `Makefile` with `build`, `test`, and `vet` targets
- Add unit tests: `internal/version/version_test.go` and `cmd/root_test.go`
- Save verification proof files for `go build`, `go vet`, `go test`, `bin/hf --help`, and `bin/hf version`

## Capabilities

### New Capabilities

- `technical-architecture-full`: Complete Go module scaffold — all cmd stubs registered with rootCmd, internal/version package, Makefile, and unit tests. No business logic beyond routing and version printing.

### Modified Capabilities

_(none — no existing requirement-level behavior is changing)_

## Impact

- **Code**: adds ~15 new files; all in the existing module
- **Dependencies**: no new external dependencies; uses only stdlib + cobra (already in `go.mod`)
- **Build**: after this change `go build ./...`, `go vet ./...`, and `go test ./...` all pass
- **API**: no API calls; stubs print help text only

## Testing Scope

| Package | Test cases |
|---|---|
| `internal/version` | `Version` is non-empty; `String()` returns `Version` |
| `cmd` | `Execute()` returns nil for `--help`; all six global flags present on rootCmd |

## Verification Steps Requiring Live Cluster

- None — all verification steps are local (`go build`, `go vet`, `go test`, binary smoke-test).
