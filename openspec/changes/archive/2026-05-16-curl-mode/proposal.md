# Proposal: curl-mode

## Why

When debugging API interactions or building automation around `hf`, users have no way to see the exact HTTP request being sent. Adding a `--curl` flag that prints the equivalent `curl` command to stderr for every API call closes this gap — users can copy-paste the command, replay it, or adapt it for scripts without needing to inspect source code.

## What Changes

- Add `--curl` persistent flag to the root command (available on every `hf` subcommand).
- When `--curl` is set, print a formatted `curl` invocation to stderr before each HTTP call to the HyperFleet REST API (`internal/api/client.go`) and the Maestro HTTP API (`internal/maestro/maestro.go`).
- Output goes to stderr so it does not pollute `--output json|table|yaml` stdout streams.
- Format: `[CURL] curl -s -X <METHOD> "<URL>" \` followed by header lines (`-H '...'`) and optionally `-d '<body>'` for requests with a JSON body.

## Capabilities

### New Capabilities

_(none — this is a cross-cutting debugging aid, not a new domain)_

### Modified Capabilities

- **`api-client`** — `internal/api/Client` gains a `curlMode` field; `NewClient` signature gains a `curlMode bool` parameter; the `do()` method prints the curl command when enabled.
- **`maestro`** — `maestro.Client` gains a `curlMode` field; `NewFromConfig` signature gains a `curlMode bool` parameter; the `get()` and `delete()` methods print curl commands when enabled.

## Impact

- `internal/api/client.go` — struct field + constructor param + print logic
- `internal/maestro/maestro.go` — struct field + constructor param + print logic in two methods
- `cmd/root.go` — new `curlMode bool` global flag
- `cmd/cluster.go` — pass `curlMode` to `newAPIClient`
- `cmd/maestro.go` — pass `curlMode` to both `NewFromConfig` calls
- All existing `api.NewClient` and `maestro.NewFromConfig` call sites in tests — add `false` as new final argument
- No new packages; no changes to `internal/output`, `internal/config`, or any command other than cluster and maestro
