## Context

Builds on `hf-ui-server` and `ui-adapter-status-form`. The server already has GET proxy routes and POST proxy routes for status endpoints. The adapter status POST was incorrectly using `api.Post` instead of `api.Put` (the upstream API requires PUT for status updates). The UI lacked all write operations (create, patch, delete, force-delete) and had incorrect dot color logic.

## API Method Correction — PUT for Status Endpoints

The upstream HyperFleet API defines:
```
PUT /clusters/{id}/statuses
PUT /clusters/{id}/nodepools/{npid}/statuses
```
(verified in `plugins/clusters/plugin.go` in the API repository)

The previous implementation used `api.Post`, causing HTTP 405 from the upstream. Additionally, the API returns 204 No Content when the status is unchanged (no-op), which caused `EOF` errors in `decode[T]` since it tried to JSON-decode an empty body.

**Fixes:**
- Added `api.Put[T]` function in `internal/api/client.go` (mirrors `Get`/`Post` pattern).
- Added 204 early-return guard in `decode[T]`: returns zero value without error.
- `cmd/cluster.go` and `cmd/nodepool.go` adapter post-status commands use `api.Put`.
- No-op 204 detected via empty-adapter response: prints `(no-op: status unchanged)`.
- `internal/server/handlers.go`: renamed `proxyPOST` (for status) to `proxyPUT`; re-added `proxyPOST` for actual POST endpoints (create, force-delete).

## Force-Delete NodePool

The upstream provides `POST /clusters/{id}/nodepools/{npid}/force-delete` with body `{"reason":"..."}`. No cluster-level force-delete exists.

**CLI (`cmd/nodepool.go`):**
```
hf nodepool force-delete [id] --reason <reason>
```
Resolves nodepool ID interactively (same pattern as other nodepool commands). Requires `--reason`. Calls `api.Post[resource.NodePool]` to the force-delete path. Prints `NodePool '<id>' force-deleted` on success.

**Server route:**
```
POST /api/clusters/{id}/nodepools/{npid}/force-delete → proxyPOST → POST upstream
```

**UI:** Force Delete button appears in the actions bar when a nodepool row is selected. Clicking it reveals a confirm panel with a reason text input. Confirms by calling `api.forceDeleteNP`. Cancellable.

## Create Cluster / NodePool

**Server routes:**
```
POST /api/clusters                       → proxyPOST → POST upstream
POST /api/clusters/{id}/nodepools        → proxyPOST → POST upstream
```

**UI:**
- `+ Cluster` button in the page header (always visible).
- `+ NodePool` button in the actions bar (visible only when a cluster is selected, not a nodepool).
- Both open the form panel in `createCluster` or `createNodepool` mode.
- Form body is a `<textarea>` pre-filled with a JSON template (hardcoded from `cmd/assets/cluster-template.json` / `nodepool-template.json`).
- On submit: parse JSON, call `api.createCluster` or `api.createNodepool(clusterID, body)`, show success banner, close after 1200 ms.

## PATCH / DELETE Proxies

**Server routes (all previously missing):**
```
PATCH  /api/clusters/{id}                     → proxyPATCH  → PATCH  upstream
DELETE /api/clusters/{id}                     → proxyDELETE → DELETE upstream
PATCH  /api/clusters/{id}/nodepools/{npid}    → proxyPATCH  → PATCH  upstream
DELETE /api/clusters/{id}/nodepools/{npid}    → proxyDELETE → DELETE upstream
```

**Helpers added to `handlers.go`:**
- `proxyPATCH`: reads body, calls `api.Patch[json.RawMessage]`, writes raw response.
- `proxyDELETE`: calls `api.Delete[json.RawMessage]`, writes raw response.

These restore the PATCH/DELETE functionality that was in the frontend but had no backend route.

## Deleted Markers

Clusters and nodepools with `deleted_time != ""` display `❌` after their name in the table and in the detail panel title. This was implemented in the UI but was lost during a rebase; re-applied from `git reflog`.

## Dot Color Correctness

### Problem

`Finalized=False` is the normal operating state for all active resources — it means "not currently being deleted". Treating any `False` condition as a failure signal incorrectly turns all adapter dots red.

The CLI (`hf table`) already handled this correctly: it shows the `Available` condition dot for non-deleted resources and the `Finalized` condition dot only when `deleted_time` is set.

### Fix — UI (`buildRow`)

```js
const relevant = (s.conditions || []).filter(c =>
  isDeleted ? c.type === 'Finalized' : c.type !== 'Finalized'
);
let overall = relevant.length ? 'True' : '';
for (const c of relevant) {
  if (c.status === 'False') { overall = 'False'; break; }
  if (c.status === 'Unknown') overall = 'Unknown';
}
```

Non-deleted: all conditions except `Finalized` contribute to overall color (False > Unknown > True).  
Deleted: only `Finalized` contributes (reflects deletion progress).  
Empty relevant conditions: gray dot (was incorrectly green before).

### Fix — CLI (`cmd/resources.go`)

No change needed — `adapterDot` already picks the correct `condKey` (`Available` vs `Finalized`) based on `isDeleted`.

## Table Column Layout

### Previous

Dynamic condition columns collected from all clusters' `status.conditions` (all types, sorted Available-first/Reconciled-last), then adapter columns.

### New

Fixed columns: `Reconciled`, `LastKnownReconciled`, then adapter columns. This matches the two most useful aggregate convergence signals and removes per-adapter condition type noise.

**CLI:** `fixedCondCols = []string{"Reconciled", "LastKnownReconciled"}` — `condDot` picks the condition by type.  
**UI:** `FIXED_COND_COLS = ['Reconciled', 'LastKnownReconciled']` — `condForType` lookup.

`collectConditionCols` / `collectCondCols` are removed entirely.

## Decisions

### D1 — `proxyPUT` vs `proxyPOST` naming

`proxyPUT` is used only for the two status endpoints (upstream requires PUT). `proxyPOST` is used for create and force-delete (upstream requires POST). This preserves the semantic distinction even though both are triggered by browser POST.

### D2 — JSON editor for create forms (textarea, not a structured form)

Create payloads have nested spec objects that vary by environment. A structured form would be brittle. A JSON textarea with a template default allows engineers to edit any field without UI changes.

### D3 — Force-delete confirm panel (not a modal)

Consistent with the existing delete confirm pattern (inline confirm row in the actions bar). A reason string is mandatory; the submit button is disabled until the field is non-empty.

### D4 — Fixed condition columns instead of dynamic

Dynamic columns caused the table to reflow when a new condition type appeared, disrupting the layout in watch mode. Fixed columns are stable and the two chosen types (`Reconciled`, `LastKnownReconciled`) capture the authoritative convergence signal that the HyperFleet control plane computes.

## Package Layout Changes

```
internal/api/
  client.go          + Put[T] function; decode[T] 204 guard

cmd/
  cluster.go         adapter post-status: api.Post → api.Put; no-op detection
  nodepool.go        adapter post-status: api.Post → api.Put; no-op detection
                     + nodepoolForceDeleteCmd with --reason flag
  resources.go       - collectConditionCols; + fixedCondCols; + condDot
                     buildClusterRow, buildNodePoolRow: fixed cond cols before adapter cols

internal/server/
  handlers.go        proxyPOST → proxyPUT for statuses
                     + proxyPOST (real POST: create, force-delete)
                     + proxyPATCH, proxyDELETE
                     + handleCreateCluster, handleCreateNodePool
                     + handleForceDeleteNodePool
                     + handlePatchCluster, handlePatchNodePool
                     + handleDeleteCluster, handleDeleteNodePool
  server.go          route table: all CRUD methods + force-delete
  server_test.go     updated: POST→PUT expectation; + new route tests

internal/ui/static/
  index.html         - collectCondCols, condCols state
                     + FIXED_COND_COLS constant
                     + Finalized-aware buildRow dot logic
                     + ❌ deleted markers
                     + CLUSTER_TEMPLATE, NODEPOOL_TEMPLATE constants
                     + api.createCluster, api.createNodepool, api.forceDeleteNP
                     + openCreateForm, submitCreateForm
                     + startForceDelete, cancelForceDelete, confirmForceDelete
                     + #create-cluster-btn in header
                     + createNPBtn and forceDeleteBtn in renderActionsBar
```
