## Context

Phase 1 laid down `internal/api`, `internal/config`, `internal/resource`, and `internal/output`. Phase 2 delivered `hf config` (show/get/set, env management, doctor). Phase 3a now completes the primary user-facing cluster commands. `cmd/cluster.go` currently exists as a stub with only the parent command registered.

## Goals / Non-Goals

**Goals**
- Implement all subcommands specified in `openspec/specs/cluster-lifecycle/spec.md`
- Use only existing internal packages — no new internal packages needed
- Full httptest-based unit test coverage
- Behave exactly like the original bash scripts (API errors → exit 0, output as JSON)

**Non-Goals**
- `hf cluster search` (search-by-name flow) — deferred to Phase 3b
- `hf cluster patch` (counter increment) — deferred to Phase 3b
- `hf cluster id` — deferred to Phase 3b
- `hf cluster adapter post-status` — deferred to adapter phase

## Decisions

### D1: cfgStore access pattern

`cfgStore` is a package-level `*config.Store` in `cmd/root.go`, initialized in `PersistentPreRunE`. All subcommands in the `cmd` package can access it directly — no need to pass it via context or closure.

However, `cfgStore` is `nil` until `PersistentPreRunE` runs. In tests, we drive `rootCmd` through `Execute()` which fires the hook, so tests get a properly initialized store.

### D2: API client construction

Each command's `RunE` constructs its own `*api.Client`:
```go
baseURL := cfgStore.Get("hyperfleet", "api-url") + "/api/hyperfleet/" + cfgStore.Get("hyperfleet", "api-version") + "/"
token := cfgStore.Get("hyperfleet", "token")
client := api.NewClient(baseURL, token, verbose)
```
This keeps commands stateless and testable — tests point cfgStore at a temp dir with an env profile whose `api-url` is the httptest server URL.

### D3: API error handling (RFC 7807 → exit 0)

Per spec, API errors (4xx/5xx) must be printed as JSON and exit 0. The `api.Get[T]` family returns `(*APIError, error)`. When the error is `*api.APIError`, commands call `printer.Print(err)` and return `nil` (not the error). Non-API errors (network failure, config missing) are returned as errors → Cobra prints them → exit 1.

### D4: Output format

Default output for `list` is `json` (JSON array). Default for `get`/`create`/`update`/`delete` is `json`. Table output is available via `--output table` on all commands. The `--output` flag is global (from root.go).

### D5: Duplicate-create guard

Per spec, `hf cluster create` checks for an existing cluster by name before POSTing:
- GET `/clusters?search=name='<name>'`
- If items > 0: print `[WARN] Cluster '<name>' already exists, skipping creation` → return nil
- Otherwise: POST

### D6: Conditions and statuses endpoints

`hf cluster conditions <id>` fetches the cluster object (GET /clusters/{id}) and extracts `status.conditions`. The table shows: TYPE STATUS LAST TRANSITION REASON MESSAGE.

`hf cluster statuses <id>` fetches GET /clusters/{id}/statuses → `ListResponse[AdapterStatus]`. The table shows: ADAPTER GEN AVAILABLE (Available condition status as dot).

### D7: ID resolution

Commands that accept an optional `[id]` arg (get, delete, conditions, statuses) resolve via:
```go
clusterID := ""
if len(args) > 0 { clusterID = args[0] }
id, err := cfgStore.ClusterID(clusterID)
```
`config.ClusterID()` handles the "explicit arg > state > error" precedence.

## Risks / Trade-offs

- [Risk] `PersistentPreRunE` sets up `cfgStore` but test helpers call `rootCmd.Execute()` so the hook fires correctly. → Verified in Phase 2 tests; same pattern applies here.
- [Risk] The conditions endpoint is actually a field on the cluster object, not a separate endpoint. → Implementation fetches the cluster and extracts conditions; the statuses endpoint is `/clusters/{id}/statuses`.

## Migration Plan

Replace `cmd/cluster.go` stub in-place. No config migration needed.

## Open Questions

_(none)_
