## Context

`--curl` was introduced in change `2026-05-16-curl-mode` as a logging aid: print curl to stderr **before** `http.Do`. Specs in `api-client` and `maestro` describe "before every HTTP request" but do not explicitly require execution to continue — users now expect dry-run semantics.

Recent create-command work added `WithoutCurl()` for duplicate-check GETs and manual `PrintCurl()` on skip — a partial fix that still executes the GET when the cluster does not exist.

## Decisions

### D1 — Sentinel error: `api.ErrDryRun`

When `curlMode=true`, `Client.do()` prints the curl command then returns `(nil, ErrDryRun)` without calling `http.Do`. Typed helpers (`Get`, `Post`, …) propagate `ErrDryRun` unchanged.

Maestro client defines an equivalent `maestro.ErrDryRun` (or shares a small `internal/dryrun` package — prefer per-package sentinel to avoid new package).

### D2 — Command-layer handling

Add `isDryRun(err error) bool` helper in `cmd` (or use `errors.Is` inline). Extend `handleAPIError` to return `nil` when `errors.Is(err, api.ErrDryRun)` — this covers most cluster/nodepool commands in one place.

Commands that do not use `handleAPIError` must check `ErrDryRun` explicitly.

### D3 — No side effects on dry-run

When `ErrDryRun` is returned:
- MUST NOT write to stdout (no JSON/table/yaml API payload)
- MUST NOT call `SetState` / `Set`
- MUST exit 0

### D4 — Create commands: POST curl only

`hf cluster create --curl` and `hf nodepool create --curl` skip the duplicate-check GET entirely. The client prints one POST curl with the resolved template body and returns `ErrDryRun`. Remove `WithoutCurl()` / manual `PrintCurl()` branches.

Rationale: duplicate detection requires a live GET; dry-run users want the create command, not internal preflight.

### D5 — Watch + curl

When both `--watch` and `--curl` are set, print curl for the first fetch path and exit 0 without entering the watch loop. Rationale: watch without HTTP would spam identical curl lines every N seconds with no value.

### D6 — Interactive + curl

When both `--curl` and `-i` / `--interactive` are set on the same command, print `[ERROR] --curl cannot be used with interactive mode` and exit 1. Checked in command `RunE` before any API or fuzzy-finder call.

### D7 — Flag description

Update persistent flag help from "print equivalent curl command for each API request" to "print equivalent curl command and skip API requests (dry-run)".

### D8 — Multi-request commands

Commands that would perform multiple HTTP calls (e.g. `cluster delete` after a pre-fetch) print curl for **each** would-be request in order, stopping at the first `ErrDryRun` without executing any. For delete, if a confirmation prompt would appear, dry-run prints the DELETE curl and exits without prompting (no mutation).

### D9 — Remove `WithoutCurl` and `PrintCurl` if redundant

After dry-run is unified in `do()`, delete `WithoutCurl()` and `PrintCurl()` unless another caller needs them. Tests updated accordingly.

## Risks

| Risk | Mitigation |
|------|------------|
| Scripts relied on `--curl` + live output | Breaking change documented; users who want both should run without `--curl` |
| TUI / `hf ui` with `--curl` | TUI commands treat `ErrDryRun` as empty data or reject `--curl` at entry |
| Pubsub publish with `--curl` | Print event POST curl only, skip publish |

## Out of Scope

- `hf kube curl` subcommand (in-cluster curl exec) — unrelated to API `--curl` flag
- Auto-port-forward / kube operations — not HTTP API clients
