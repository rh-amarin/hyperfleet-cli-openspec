## Context

- `hf rs` (`cmd/resource.go`) dynamically registers per-type commands from `resource-types`; `clusters` and `nodepools` are in the bundled template.
- Today, operators use **three parallel surfaces** for the same API resources:
  - `hf cluster` / `hf nodepool` (typed commands in `cmd/cluster.go`, `cmd/nodepool.go`)
  - `hf table` / `hf resources` (combined adapter overview in `cmd/resources.go`)
  - `hf rs` / `hf rs clusters|nodepools` (generic commands; partially implemented)
- Partial `hf rs` work exists: overview, `adapter-report`, tolerant fetch, deletion markers.
- **End state:** `hf rs` is the only supported interface for cluster/nodepool lifecycle and the combined operational overview. Legacy top-level groups are removed after a short deprecation window.

## Goals / Non-Goals

**Goals:**

- **`hf rs <entity>` is canonical** for all cluster and nodepool operations (and any future `resource-types` entries)
- **Full parity** with current `hf cluster` / `hf nodepool` behavior before legacy removal
- **`hf rs` replaces `hf table` / `hf resources`** for the combined cluster+nodepool adapter view (same columns, watch, spinners)
- **Consolidate implementation** — extract shared logic from `cluster.go` / `nodepool.go` / `resources.go` into handlers invoked only from `hf rs`; legacy commands become thin deprecated delegates, then are deleted
- **Deprecation UX** — legacy commands print `[WARN] deprecated: use hf rs …` and forward to the `hf rs` implementation during the transition release
- Normative spec in `rs-entity-commands` delta; update command-hierarchy and tables-and-lists accordingly

**Non-Goals:**

- Deprecating unrelated top-level groups (`hf kube`, `hf db`, `hf maestro`, etc.)
- Per-type YAML capability flags in v1 (hardcode `clusters` / `nodepools` profiles)
- Changing HyperFleet API paths or auth

## Decisions

### D1 — `hf rs` owns cluster/nodepool behavior

All lifecycle, conditions, statuses, adapter simulation, and rich tables live under `hf rs clusters` and `hf rs nodepools`. Implementation uses entity profiles (`clusters`, `nodepools`) in `cmd/resource_profiles.go` (or equivalent).

**Rejected:** Keeping `hf cluster` as a parallel long-term API.

### D2 — Extract shared internals, delete duplicate command trees

1. Move table row builders, conditions/statuses renderers, create/patch/delete/force-delete logic into shared functions (e.g. `cmd/resource_cluster.go`, `internal/resource/table.go`).
2. `hf rs` subcommands call those functions with resolved paths from `config.Store`.
3. `hf cluster` / `hf nodepool` / `hf table` / `hf resources` temporarily call the **same** functions with a deprecation printer.
4. Final step of this change (or immediate follow-up task in same PR): **unregister** `clusterCmd`, `nodepoolCmd`, `resourcesCmd`, `tableCmd` from `rootCmd`.

**Rejected:** Leaving duplicate HTTP/render paths in `cluster.go` indefinitely.

### D3 — Combined overview = `hf rs` (not `hf table`)

`hf rs` with no subcommand MUST render the **adapter-rich** combined cluster+nodepool table (equivalent to today's `hf table --output table --watch`), built by reusing `fetchResourceEntries` / `renderResourcesTable` logic keyed off `resource-types` parent/child (`clusters` → `nodepools`).

- `hf table` and `hf resources` become deprecated aliases for `hf rs` (same flags: `--watch`, `-s`).
- After deprecation window, only `hf rs` remains.

**Rejected:** Keeping `hf table` as a permanent second overview command.

### D4 — Subcommand naming on `hf rs`

| Legacy | Canonical `hf rs` |
|---|---|
| `hf cluster adapter post-status` | `hf rs clusters adapter-report` |
| `hf nodepool adapter post-status` | `hf rs nodepools adapter-report` |
| `hf cluster table` | `hf rs clusters table` |
| `hf resources` / `hf table` | `hf rs` (overview) |

### D5 — Patch, create, delete parity on `hf rs` only

- Patch: counter increment when `--file` omitted; file patch to `/spec` or `/labels` when set
- Create: templates under `templates/`, duplicate guard, entity-specific flags/positionals
- Delete / force-delete: same API paths as today; only exposed on `hf rs`

### D6 — Deprecation mechanics

During transition (one release minimum):

```go
// cluster.go — temporary
RunE: func(cmd, args) error {
    deprecate(cmd.ErrOrStderr(), "hf cluster", "hf rs clusters")
    return runRsEntity(cmd, "clusters", translateArgs(args))
}
```

- Help text on legacy groups: `Deprecated: use hf rs clusters …`
- README and spec index list `hf rs` only; legacy names in a "Removed / migrated" table
- Tests for legacy commands either removed or assert deprecation + identical output to `hf rs`

## Migration Plan

| Phase | Deliverable |
|---|---|
| **1. Parity** | Implement missing `hf rs` subcommands (table, conditions, statuses, force-delete, rich list tables, overview with adapters) |
| **2. Delegate** | Legacy `hf cluster`, `hf nodepool`, `hf table`, `hf resources` call shared code; emit deprecation warning |
| **3. Remove** | Unregister legacy commands from Cobra tree; update shell completions, README, hf-user-guide |
| **4. Archive** | Merge OpenSpec deltas; verification_proof includes only `hf rs` commands |

Rollback: keep deprecated delegate shims behind a build tag or one release branch if needed; do not maintain two implementations.

## Risks / Trade-offs

- **[Risk] Breaking scripts** using `hf cluster create` → Mitigate with deprecation release that still works but warns; document migration table in README
- **[Risk] Completion flags** reference `cluster` subcommand in tests → Rewire tests to `hf rs clusters`; delete `cluster_test.go` only after parity tests exist on `hf rs`
- **[Trade-off] Large PR** — acceptable; logical order is extract → wire `hf rs` → delegate legacy → delete legacy
- **[Trade-off] `hf table` default output** — `hf rs` already defaults to table; aligns with replacing `hf table`

## Open Questions

- **None blocking:** Deprecation is in scope; `hf table` delegates to `hf rs` in this change (Phase 2–3 above).
- Confirm whether Phase 3 (unregister legacy) ships in the same PR as parity or the immediately following PR (recommend same change once parity tests pass).
