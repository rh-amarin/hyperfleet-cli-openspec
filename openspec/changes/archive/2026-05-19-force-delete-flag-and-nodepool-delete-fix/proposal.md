## Why

`hf cluster delete` has no way to force-delete a stuck cluster, and the equivalent `hf nodepool force-delete` subcommand is a separate command that users must discover separately. Additionally, `hf nodepool delete` ignores the active nodepool-id from state when called with no arguments, forcing users to always pass an explicit ID — inconsistent with every other single-resource nodepool command.

## What Changes

- **Add `--force` flag to `hf cluster delete`**: when set, POST to `clusters/<id>/force-delete` instead of DELETE to `clusters/<id>`. An optional `--reason` flag passes a reason string in the request body.
- **Add `--force` flag to `hf nodepool delete`**: when set, POST to `clusters/<clusterID>/nodepools/<id>/force-delete`. Reuses the same endpoint already called by `hf nodepool force-delete`.
- **Fix `hf nodepool delete` active-nodepool fallback**: when no explicit ID is provided and not in interactive mode, resolve the nodepool-id from state via `s.NodePoolID("")` (matching the behavior of `get`, `patch`, `conditions`, and `statuses`). Remove the early error that prevented this fallback.

## Capabilities

### New Capabilities

- `cluster-force-delete`: `hf cluster delete --force [--reason <text>]` calls `POST clusters/<id>/force-delete` to remove a stuck cluster.

### Modified Capabilities

- `cluster-lifecycle`: `hf cluster delete` gains `--force` and `--reason` flags.
- `nodepool-lifecycle`: `hf nodepool delete` gains `--force` and `--reason` flags, and is fixed to fall back to the active nodepool-id from state.

## Impact

- `cmd/cluster.go`: `clusterDeleteCmd` — new flag vars, branching on `--force`.
- `cmd/nodepool.go`: `nodepoolDeleteCmd` — new flag vars, branching on `--force`, fix ID resolution.
- `cmd/cluster_test.go`: new test cases for `--force` on cluster delete.
- `cmd/nodepool_test.go`: new test cases for `--force` on nodepool delete and the state-fallback fix.
- No new dependencies. No breaking changes to existing flags or subcommands.

### Testing Scope

- `cmd` package — `cluster_test.go`:
  - `hf cluster delete --force` calls `POST clusters/<id>/force-delete`
  - `hf cluster delete --force --reason "stuck"` passes reason in body
  - `hf cluster delete` (no `--force`) still calls `DELETE clusters/<id>`
- `cmd` package — `nodepool_test.go`:
  - `hf nodepool delete` with no args and nodepool-id in state resolves from state
  - `hf nodepool delete --force` calls `POST .../force-delete`
  - `hf nodepool delete --force --reason "stuck"` passes reason in body
  - `hf nodepool delete` (no `--force`) still calls `DELETE .../nodepools/<id>`

Live cluster verification required for: `hf cluster delete --force`, `hf nodepool delete --force`, and `hf nodepool delete` with active state fallback.
