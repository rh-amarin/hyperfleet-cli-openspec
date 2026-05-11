# Design: Phase 7 — Missing Features

## Packages Modified

| File | Change |
|---|---|
| `cmd/cluster.go` | Add `clusterSearchCmd`, `clusterPatchCmd`; update `clusterDeleteCmd`, `clusterCreateCmd`, `clusterStatusesCmd` |
| `cmd/nodepool.go` | Add `nodepoolSearchCmd`, `nodepoolPatchCmd`; update `nodepoolStatusesCmd` |
| `cmd/config.go` | Update `configShowCmd` (optional env-name arg); add `new` alias to `configEnvCreateCmd` |
| `cmd/cluster_test.go` | Add tests for all new cluster behaviors |
| `cmd/nodepool_test.go` | Add tests for all new nodepool behaviors |

## Key Decisions

### cluster/nodepool search
- No-args case: reads `cluster-id`/`nodepool-id` from state directly (via `GetState`) and behaves like `get`. Uses a specific error message matching the spec rather than the generic `ClusterID()` error.
- With-name case: queries `?search=name='<name>'`, outputs `list.Items` (a JSON array, not the ListResponse wrapper), persists first result's ID. Warns on not-found (outputs `[]`) and multiple matches.

### cluster/nodepool patch
- Fetches the current resource, reads `spec.counter` or `labels.counter` as an integer (absent treated as 0), increments by 1, PATCHes with the new value as a string.
- No-args or invalid section: prints usage to stdout, returns a non-nil error (exit 1).

### cluster delete optional ID
- Changed `cobra.ExactArgs(1)` to `cobra.MaximumNArgs(1)`. ID resolution delegates to `s.ClusterID(explicit)` — same pattern as `get` and `conditions`.
- Also fixed: now outputs the deleted cluster JSON per spec (previously silent).

### cluster create positional args
- Added `cobra.MaximumNArgs(3)` and reads `args[0]`/`[1]`/`[2]` as name/region/version, taking precedence over the `--name` flag and built-in defaults. Preserves full backward compatibility with flag-based usage.

### config show [env-name]
- `cobra.MaximumNArgs(1)`. With a name: reads the profile file from disk, redacts secrets, prints the file path (prefixed `[active]` if it matches the active env). Without a name: original behavior unchanged.

### config env new alias
- Added `Aliases: []string{"new"}` to `configEnvCreateCmd`. No behavioral change.

### statuses table FINALIZED column
- Added `"FINALIZED"` header and extracts `Finalized` condition from `as.Conditions` using a switch (same pattern as `Available`).
