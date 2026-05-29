## Why

The `--curl` flag currently prints a curl command **and still executes** the HTTP request. Users expect `--curl` to be a dry-run: show the exact command they can copy-paste without hitting the API, mutating state, or producing misleading list/create output. This mismatch was reported for `hf cluster list --curl` (shows curl plus real JSON) and `hf cluster create --curl` (shows the duplicate-check GET instead of the create POST).

## What Changes

- **BREAKING:** `--curl` becomes a dry-run mode — print curl to stderr, skip all HTTP requests, exit 0 with no API data on stdout
- Update flag help text to describe dry-run semantics
- `internal/api` and `internal/maestro` clients return a sentinel `ErrDryRun` after printing curl; callers treat it as success with no output
- `cluster create` / `nodepool create` with `--curl`: print POST curl only (skip duplicate-check GET)
- `--curl` with `--watch`: print curl for the first fetch only, do not enter the watch loop
- `--curl` with `-i` / `--interactive`: reject with a clear error (interactive flows require live list data)
- Remove the ad-hoc `WithoutCurl` / manual `PrintCurl` workaround on create commands (superseded by unified dry-run)
- No state mutations (`SetState`, etc.) when `--curl` is set

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `api-client`: curl mode is dry-run; no HTTP execution; `ErrDryRun` sentinel
- `maestro`: curl mode is dry-run; no HTTP execution
- `command-hierarchy`: update `--curl` global flag description and dry-run interaction rules
- `errors-and-usage`: duplicate-creation behavior under `--curl`
- `cluster-lifecycle`: create dry-run scenario
- `nodepool-lifecycle`: create dry-run scenario

## Impact

- `internal/api/client.go` — dry-run in `do()`, export `ErrDryRun`, simplify/remove `WithoutCurl` if unused
- `internal/maestro/maestro.go` — dry-run in `get()` / `delete()`
- `cmd/root.go` — flag help text
- `cmd/cluster.go`, `cmd/nodepool.go` — skip duplicate preflight when `curlMode`; handle `ErrDryRun`
- `cmd/*` — all commands using `handleAPIError` or direct API calls must treat `ErrDryRun` as success
- `cmd/cluster_test.go`, `internal/api/client_test.go`, maestro tests — update/add dry-run coverage
- User-facing docs: `README.md`, `hf-user-guide.md` (if they mention `--curl`)

## Testing Scope

- `internal/api`: curl mode prints curl, returns `ErrDryRun`, does not call HTTP transport
- `internal/maestro`: same for GET and DELETE
- `cmd`: `cluster list --curl` — stderr has curl, stdout empty (or no JSON body), exit 0
- `cmd`: `cluster create --curl` — stderr has POST curl only (no search GET), no POST to server, exit 0
- `cmd`: `cluster list --watch --curl` — single curl line, no watch loop
- `cmd`: `cluster get -i --curl` — error exit

## Live Verification

- `hf cluster list --curl` — curl on stderr, no cluster JSON on stdout
- `hf cluster create --curl` — POST curl on stderr, no cluster created
- `hf maestro list --curl` — GET curl on stderr, no bundle list on stdout
