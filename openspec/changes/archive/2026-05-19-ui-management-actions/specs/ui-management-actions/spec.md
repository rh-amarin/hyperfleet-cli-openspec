# Delta Spec: UI Management Actions

Extends `hf-ui-server` and `ui-adapter-status-form`.

---

## 1. API Client

### REQ-API-PUT-1
`internal/api` must export a `Put[T any]` function with the same signature as `Post[T]`.

### REQ-API-204-1
`decode[T]` must return a zero `T` and `nil` error when the upstream response status is 204 No Content, instead of attempting to JSON-decode an empty body.

---

## 2. Adapter Status — PUT Semantics

### REQ-ADAPTER-PUT-1
`hf cluster adapter post-status` and `hf nodepool adapter post-status` must use `api.Put` (not `api.Post`) when calling the upstream status endpoint, because the upstream API defines the route as `PUT /clusters/{id}/statuses`.

### REQ-ADAPTER-NOOP-1
When the upstream returns 204 No Content (status unchanged), the CLI must print `(no-op: status unchanged)` and exit 0.

### REQ-UI-STATUS-PUT-1
`POST /api/clusters/{id}/statuses` and `POST /api/clusters/{id}/nodepools/{npid}/statuses` on the UI server must forward the request body to the upstream via `PUT` (not POST).

---

## 3. Force-Delete NodePool

### REQ-NODEPOOL-FORCE-DELETE-1
`hf nodepool force-delete [id] --reason <reason>` permanently removes a nodepool stuck in Finalizing state by calling `POST /clusters/{clusterID}/nodepools/{id}/force-delete` with body `{"reason":"..."}`.

**Behaviour:**
- `--reason` is required; exit 1 with `[ERROR] --reason is required` if omitted.
- ID resolution follows the same interactive pattern as other nodepool commands.
- Prints `NodePool '<id>' force-deleted` on success.

### REQ-SERVER-FORCE-DELETE-1
`POST /api/clusters/{id}/nodepools/{npid}/force-delete` must be proxied to `POST clusters/{id}/nodepools/{npid}/force-delete` upstream via `proxyPOST`.

### REQ-UI-FORCE-DELETE-1
When a nodepool row is selected in the dashboard, the actions bar must show a **Force Delete** button.

### REQ-UI-FORCE-DELETE-2
Clicking Force Delete must reveal an inline confirm panel with a reason text input and Confirm / Cancel buttons. The Confirm button must be disabled until the reason field is non-empty.

### REQ-UI-FORCE-DELETE-3
On confirmation the UI must call `POST /api/clusters/{clusterID}/nodepools/{npID}/force-delete` with body `{"reason":"..."}` and refresh the table on success.

---

## 4. Create Cluster / NodePool from UI

### REQ-SERVER-CREATE-CLUSTER-1
`POST /api/clusters` must be proxied to `POST clusters` upstream via `proxyPOST`.

### REQ-SERVER-CREATE-NODEPOOL-1
`POST /api/clusters/{id}/nodepools` must be proxied to `POST clusters/{id}/nodepools` upstream via `proxyPOST`.

### REQ-UI-CREATE-CLUSTER-1
The page header must contain a **+ Cluster** button always visible regardless of selection state.

### REQ-UI-CREATE-NODEPOOL-1
The actions bar must contain a **+ NodePool** button visible only when a cluster (not a nodepool) is selected.

### REQ-UI-CREATE-FORM-1
Clicking either create button must open the form panel in create mode with a `<textarea>` pre-filled with the appropriate JSON template (cluster template or nodepool template).

### REQ-UI-CREATE-FORM-2
On submit the UI must parse the textarea content as JSON, call the appropriate create API, show a success banner, and close the form after 1200 ms.

---

## 5. PATCH / DELETE Proxy

### REQ-SERVER-PATCH-CLUSTER-1
`PATCH /api/clusters/{id}` must be proxied to `PATCH clusters/{id}` upstream (reads body, calls `api.Patch[json.RawMessage]`).

### REQ-SERVER-DELETE-CLUSTER-1
`DELETE /api/clusters/{id}` must be proxied to `DELETE clusters/{id}` upstream (calls `api.Delete[json.RawMessage]`).

### REQ-SERVER-PATCH-NODEPOOL-1
`PATCH /api/clusters/{id}/nodepools/{npid}` must be proxied to `PATCH clusters/{id}/nodepools/{npid}` upstream.

### REQ-SERVER-DELETE-NODEPOOL-1
`DELETE /api/clusters/{id}/nodepools/{npid}` must be proxied to `DELETE clusters/{id}/nodepools/{npid}` upstream.

---

## 6. Deleted Resource Markers

### REQ-UI-DELETED-MARKER-1
Cluster rows with `deleted_time` set must display ` ❌` after the cluster name in the table.

### REQ-UI-DELETED-MARKER-2
NodePool rows with `deleted_time` set must display ` ❌` after the nodepool name in the table.

---

## 7. Adapter Dot Color — Finalized Semantics

### REQ-DOT-FINALIZED-1
For non-deleted resources (`deleted_time` is empty), the `Finalized` condition must be excluded from the adapter dot overall color calculation. `Finalized=False` means "not being deleted" and is a healthy state.

### REQ-DOT-FINALIZED-2
For deleted resources (`deleted_time` is set), only the `Finalized` condition must contribute to the adapter dot color. All other conditions are excluded.

### REQ-DOT-EMPTY-1
An adapter status entry with zero relevant conditions (after the Finalized filter) must render a gray dot, not green.

**Priority rule (unchanged):** False > Unknown > True > gray (empty).

---

## 8. Table Column Layout

### REQ-TABLE-FIXED-COND-1
The cluster/nodepool table (both `hf table` CLI and browser UI) must show exactly two fixed condition columns before the adapter columns: `Reconciled` and `LastKnownReconciled` (in that order).

### REQ-TABLE-FIXED-COND-2
No other condition types from `cluster.status.conditions` or `nodepool.status.conditions` must appear as table columns. The dynamic `collectConditionCols` / `collectCondCols` logic is removed.

### REQ-TABLE-COL-ORDER-1
Column order: **Name | Age | Gen | Reconciled | LastKnownReconciled | [adapter columns sorted by earliest created_time]**.

---

## 9. Method Routing — Updated Route Table

| Browser method + path | Upstream method + path | Handler |
|---|---|---|
| `GET /api/clusters` | `GET clusters` (+ per-cluster statuses merged) | `handleClusters` |
| `POST /api/clusters` | `POST clusters` | `handleCreateCluster` |
| `GET /api/clusters/{id}` | `GET clusters/{id}` | `handleCluster` |
| `PATCH /api/clusters/{id}` | `PATCH clusters/{id}` | `handlePatchCluster` |
| `DELETE /api/clusters/{id}` | `DELETE clusters/{id}` | `handleDeleteCluster` |
| `GET /api/clusters/{id}/statuses` | `GET clusters/{id}/statuses` | `handleClusterStatuses` |
| `POST /api/clusters/{id}/statuses` | `PUT clusters/{id}/statuses` | `handlePostClusterStatuses` |
| `GET /api/clusters/{id}/nodepools` | `GET clusters/{id}/nodepools` | `handleNodePools` |
| `POST /api/clusters/{id}/nodepools` | `POST clusters/{id}/nodepools` | `handleCreateNodePool` |
| `GET /api/clusters/{id}/nodepools/{npid}` | `GET clusters/{id}/nodepools/{npid}` | `handleNodePool` |
| `PATCH /api/clusters/{id}/nodepools/{npid}` | `PATCH clusters/{id}/nodepools/{npid}` | `handlePatchNodePool` |
| `DELETE /api/clusters/{id}/nodepools/{npid}` | `DELETE clusters/{id}/nodepools/{npid}` | `handleDeleteNodePool` |
| `GET /api/clusters/{id}/nodepools/{npid}/statuses` | `GET clusters/{id}/nodepools/{npid}/statuses` | `handleNodePoolStatuses` |
| `POST /api/clusters/{id}/nodepools/{npid}/statuses` | `PUT clusters/{id}/nodepools/{npid}/statuses` | `handlePostNodePoolStatuses` |
| `POST /api/clusters/{id}/nodepools/{npid}/force-delete` | `POST clusters/{id}/nodepools/{npid}/force-delete` | `handleForceDeleteNodePool` |
