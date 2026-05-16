## Why

`hf kube port-forward start` currently prints only a terse `[INFO] Started …` line per service, giving no indication of which namespace or pod was selected and no confirmation that the tunnels came up healthy. Users are left guessing whether the right pod was targeted and must run a separate `hf kube port-forward status` command to verify.

## What Changes

- The `[INFO] Started …` line for each service will include the namespace and pod name that was resolved (e.g. `Started hyperfleet-api (amarin-ns1/hyperfleet-api-7d9f8b-xkj2p): localhost:8000 → 8000 (pid 12345)`).
- After all port-forwards are started, `port-forward start` will automatically print the same status table produced by `port-forward status` (colored process-alive bullets).
- `StartPortForward` return type changes from three scalars `(pid, localPort, remotePort, error)` to a `StartResult` struct that includes `Namespace` and `PodName`; callers are updated accordingly.
- A `printPortForwardStatus` helper is extracted so both `pfStartCmd` and `pfStatusCmd` share the same rendering logic without duplication.

## Capabilities

### New Capabilities

- `port-forward-start-feedback`: Enriched output for `port-forward start` — namespace/pod in the per-service start line, followed by an inline status table.

### Modified Capabilities

- `kubernetes`: The `port-forward start` output format changes; `StartPortForward` signature changes (internal API only, no CLI flag changes).

## Impact

- **`internal/kube/kube.go`**: New `StartResult` struct; updated `StartPortForward` signature and return.
- **`cmd/kube.go`**: `pfStartCmd` updated to use `StartResult` and print enriched output; `pfStatusCmd` updated to call extracted `printPortForwardStatus` helper.
- **`internal/kube/kube_test.go`**: Tests updated for the new return type.
- No changes to CLI flags, config keys, or external APIs.

## Testing Scope

- `internal/kube`: Unit tests for `StartPortForward` returning the correct `StartResult` fields.
- `cmd/kube.go` is exercised by live verification (daemon processes require a real cluster).

Live cluster verification is required to confirm the enriched start output and inline status table display correctly.
