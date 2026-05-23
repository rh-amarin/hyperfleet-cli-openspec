## Why

Several gaps were found in the `hf ui` dashboard and the `hf table` / `hf cluster adapter post-status` CLI commands:

1. **PUT vs POST bug**: The upstream API expects `PUT /clusters/{id}/statuses` but the CLI and UI server were sending POST, causing `EOF` errors on 204 No Content responses.
2. **No write operations in the UI**: PATCH (update) and DELETE for clusters and nodepools existed on the CLI but were missing from the server proxy layer and the browser UI.
3. **No force-delete on stuck nodepools**: There was no CLI command or UI action to call `POST /clusters/{id}/nodepools/{npid}/force-delete` for nodepools stuck in Finalizing state.
4. **No create from the UI**: Engineers had to switch to the CLI to create clusters or nodepools; the dashboard had no creation form.
5. **Dot color incorrectness**: `Finalized=False` (the normal operating state — "not being deleted") was treated as a failure condition, turning all adapter dots red regardless of actual health.
6. **Table column noise**: The table showed every condition type dynamically, making it hard to read. The two most useful aggregate signals (`Reconciled`, `LastKnownReconciled`) were buried among many per-adapter condition columns.
7. **Missing deleted markers**: Clusters and nodepools with `deleted_time` set were not visually distinguished from live ones.

## What Changes

- **`internal/api/client.go`**: New `Put[T]` generic function; `decode[T]` returns zero value on 204 No Content instead of EOF error.
- **`cmd/cluster.go`** and **`cmd/nodepool.go`**: `adapter post-status` commands switch to `api.Put`; no-op 204 responses print `(no-op: status unchanged)`.
- **`cmd/nodepool.go`**: New `hf nodepool force-delete [id] --reason <reason>` command.
- **`internal/server/handlers.go`**: New `proxyPUT`, `proxyPATCH`, `proxyDELETE` helpers; `handlePostClusterStatuses` and `handlePostNodePoolStatuses` use `proxyPUT`; new handlers for PATCH/DELETE cluster and nodepool, create cluster/nodepool, and nodepool force-delete.
- **`internal/server/server.go`**: Route table extended to support POST (create), PATCH, DELETE for clusters and nodepools, and `POST /force-delete` for nodepools.
- **`cmd/resources.go`** and **`internal/ui/static/index.html`**: Table columns changed from dynamic condition types to fixed `Reconciled` + `LastKnownReconciled` before adapter columns. Adapter dot ignores `Finalized` for non-deleted resources; uses only `Finalized` for deleted resources.
- **`internal/ui/static/index.html`**: `+ Cluster` button in header, `+ NodePool` button in actions bar (JSON editor panel); force-delete confirm flow for nodepools; PATCH/DELETE wired to existing toolbar buttons; ❌ marker on deleted cluster/nodepool names.

## Capabilities

### New Capabilities

- `nodepool-force-delete`: `hf nodepool force-delete [id] --reason <reason>` permanently removes a nodepool stuck in Finalizing state.
- `ui-create-cluster`: `+ Cluster` button opens a JSON editor pre-filled with the cluster template; POST on submit.
- `ui-create-nodepool`: `+ NodePool` button (in actions bar, only when a cluster is selected) opens a JSON editor pre-filled with the nodepool template; POST on submit.
- `ui-force-delete-nodepool`: Force Delete button appears in actions bar when a nodepool is selected; requires a reason string; POSTs to the force-delete endpoint.
- `ui-patch-delete`: PATCH and DELETE buttons in the cluster/nodepool actions bar submit to the server proxy.

### Modified Capabilities

- `adapter-status-post`: CLI uses PUT (not POST); handles 204 No Content as a no-op.
- `ui-server`: Proxy layer now covers all CRUD methods plus force-delete and create. POST to status endpoints uses PUT upstream.
- `ui-table`: Fixed condition columns (Reconciled, LastKnownReconciled) replace dynamic per-type columns; adapter dot color is corrected for Finalized semantics.

## Impact

- **Modified files:** `internal/api/client.go`, `cmd/cluster.go`, `cmd/nodepool.go`, `cmd/resources.go`, `internal/server/handlers.go`, `internal/server/server.go`, `internal/server/server_test.go`, `internal/ui/static/index.html`
- **No new Go module dependencies**
- **Breaking change in `ui-adapter-status-form`**: `proxyPOST` is now split into `proxyPUT` (for status endpoints) and `proxyPOST` (for create/force-delete); server tests updated accordingly.

## Testing Scope

- `internal/server` — updated and new unit tests covering all new routes (PATCH, DELETE, POST create, force-delete) and PUT upstream method for status proxying.
- `cmd` — existing `TestClusterAdapterPostStatus` and `TestNodePoolAdapterPostStatus` updated to expect PUT upstream.
- Live verification: `hf ui` dashboard tested for dot correctness, create form submission, force-delete flow, and PATCH/DELETE actions.
