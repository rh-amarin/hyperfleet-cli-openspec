## Why

Cluster and nodepool operations are implemented twice: legacy `hf cluster` / `hf nodepool` / `hf table` and config-driven `hf rs`. That splits tests, docs, and operator habits. The API is already modeled under `resource-types` (`clusters`, `nodepools`); the CLI should expose **one** command surface — `hf rs` — and deprecate the legacy groups.

## What Changes

- Add normative spec **rs-entity-commands** for the full `hf rs <entity>` tree
- Implement missing `hf rs` parity (conditions, statuses, rich tables, force-delete, counter-patch, create flags, `hf rs` overview with adapter columns)
- **Extract** shared logic from `cmd/cluster.go`, `cmd/nodepool.go`, `cmd/resources.go` into code paths used by `hf rs`
- **Deprecate** `hf cluster`, `hf nodepool`, `hf table`, and `hf resources` — warn and delegate to `hf rs` during transition
- **Remove** legacy command registration once parity tests pass (same change, Phase 3)
- Update README, completions, and specs to list `hf rs` as canonical

## Capabilities

### New Capabilities

- `rs-entity-commands`: Canonical `hf rs <entity>` command contract and deprecation of legacy cluster/nodepool/table entrypoints

### Modified Capabilities

- `command-hierarchy`: Remove `hf cluster` / `hf nodepool` from long-term tree; document `hf rs` as replacement; deprecate `hf resources` / `hf table`
- `generic-resource-lifecycle`: Overview with adapters; patch/search/create/delete parity
- `adapter-status`: `hf rs <entity> adapter-report` replaces `hf cluster|nodepool adapter post-status`
- `config-model`: Template defaults for `clusters` / `nodepools`
- `tables-and-lists`: Combined and per-entity tables served via `hf rs` only after migration
- `cluster-lifecycle` / `nodepool-lifecycle`: Requirements **REMOVED** or redirected to `rs-entity-commands` on archive (legacy command names no longer normative)

## Impact

- **Delete or gut:** `cmd/cluster.go`, `cmd/nodepool.go` — after extract, only deprecation shim or removal
- **Redirect:** `cmd/resources.go` — overview logic moves to `hf rs`; `resourcesCmd` / `tableCmd` removed
- **Grow:** `cmd/resource*.go` — all cluster/nodepool behavior
- **Tests:** Migrate `cluster_test.go` / `nodepool_test.go` / `resources_test.go` assertions to `resource_*_test.go`; remove legacy command tests last
- **Docs:** README, command reference, shell completions

## Testing Scope

- Parity: each legacy scenario has an `hf rs` equivalent test (httptest)
- Deprecation: legacy invocations still exit 0 but stderr contains `deprecated` and stdout matches `hf rs`
- After removal: no tests reference `hf cluster` / `hf nodepool` / `hf table` as primary API

## Live Verification

- Verify **only** `hf rs` commands for: overview, clusters list/table/watch, nodepools CRUD, conditions, statuses, adapter-report, force-delete
- Save to `verification_proof/live.txt`
- Optional appendix: confirm deprecated alias prints warning once before removal
