## Why

Phase 5 closes the two remaining CLI gaps: adapter status posting (which lets operators manually drive convergence during testing) and the `hf resources` / `hf table` combined overview that was stubbed but never implemented. Both are needed for day-to-day cluster operations.

## What Changes

- Add `hf cluster adapter post-status <adapter> <status> <gen>` — POST adapter status to `/clusters/{id}/statuses`
- Add `hf nodepool adapter post-status <adapter> <status> <gen>` — POST adapter status to `/clusters/{id}/nodepools/{id}/statuses`
- Replace `cmd/resources.go` stub with full implementation of `hf resources` / `hf table`
- `hf table` is an alias that calls the same `RunE` as `hf resources`
- Combined table shows clusters + their nodepools with dynamic condition and adapter columns

## Capabilities

### New Capabilities

- `adapter-status-commands`: CLI commands for posting adapter status to clusters and nodepools
- `resources-table`: Combined cluster+nodepool table view with dynamic condition and adapter columns

### Modified Capabilities

- `adapter-status`: Extends existing spec with Go command interface (adds `hf cluster adapter post-status` and `hf nodepool adapter post-status` shapes already documented in the spec)
- `tables-and-lists`: Implements the Combined Resources Overview requirement that was previously unimplemented

## Impact

- `cmd/cluster.go` — add `clusterAdapterCmd` and `clusterAdapterPostStatusCmd`
- `cmd/nodepool.go` — add `nodepoolAdapterCmd` and `nodepoolAdapterPostStatusCmd`
- `cmd/resources.go` — replace stub with full `RunE`; register `tableCmd` alias
- `internal/resource/resource.go` — all required types already present (`AdapterStatus`, `AdapterCondition`, `ConditionRequest`, `AdapterStatusCreateRequest`)
- Tests: `cmd/cluster_test.go`, `cmd/nodepool_test.go`, `cmd/resources_test.go`
- No new Go module dependencies required

## Testing Scope

| Package | Test cases |
|---------|-----------|
| `cmd` (cluster) | `TestClusterAdapterPostStatus` — POST True, invalid status |
| `cmd` (nodepool) | `TestNodePoolAdapterPostStatus` — POST True |
| `cmd` (resources) | `TestResourcesTable` — combined table output; `TestResourcesJSON` — raw JSON output |

Live cluster verification required for: `hf cluster adapter post-status`, `hf nodepool adapter post-status`, `hf resources`.
