# Tasks: ui-management-actions

## 1. API Client — PUT + 204 fix

- [x] 1.1 Add `Put[T]` generic function to `internal/api/client.go` mirroring `Post[T]`
- [x] 1.2 Add 204 No Content early-return guard in `decode[T]` (returns zero value, nil error)

## 2. CLI — adapter post-status PUT fix

- [x] 2.1 `cmd/cluster.go`: switch `clusterAdapterPostStatusCmd` from `api.Post` to `api.Put`; detect empty-adapter no-op response and print `(no-op: status unchanged)`
- [x] 2.2 `cmd/nodepool.go`: same fix for `nodepoolAdapterPostStatusCmd`
- [x] 2.3 Update `TestClusterAdapterPostStatus` in `cmd/cluster_test.go` to expect upstream method PUT
- [x] 2.4 Update `TestNodePoolAdapterPostStatus` in `cmd/nodepool_test.go` to expect upstream method PUT

## 3. CLI — force-delete nodepool

- [x] 3.1 Add `nodepoolForceDeleteCmd` to `cmd/nodepool.go` with `--reason` required flag; resolves ID interactively; calls `api.Post` to `nodepools/{id}/force-delete`
- [x] 3.2 Register `nodepoolForceDeleteCmd` in `init()` on `nodepoolCmd`

## 4. Server — proxy helpers

- [x] 4.1 Rename `proxyPOST` (status variant) → `proxyPUT` in `internal/server/handlers.go`; use `api.Put[json.RawMessage]`; add 204 no-content forwarding
- [x] 4.2 Re-add `proxyPOST` for real POST endpoints (create, force-delete); uses `api.Post[json.RawMessage]`; add 204 no-content forwarding
- [x] 4.3 Add `proxyPATCH` helper: reads body, calls `api.Patch[json.RawMessage]`, writes raw response
- [x] 4.4 Add `proxyDELETE` helper: calls `api.Delete[json.RawMessage]`, writes raw response

## 5. Server — new route handlers

- [x] 5.1 Add `handleCreateCluster` → `proxyPOST(w, r, "clusters")`
- [x] 5.2 Add `handleCreateNodePool` → `proxyPOST(w, r, "clusters/{id}/nodepools")`
- [x] 5.3 Add `handleForceDeleteNodePool` → `proxyPOST(w, r, "clusters/{id}/nodepools/{npid}/force-delete")`
- [x] 5.4 Add `handlePatchCluster` → `proxyPATCH(w, r, "clusters/{id}")`
- [x] 5.5 Add `handleDeleteCluster` → `proxyDELETE(w, r, "clusters/{id}")`
- [x] 5.6 Add `handlePatchNodePool` → `proxyPATCH(w, r, "clusters/{id}/nodepools/{npid}")`
- [x] 5.7 Add `handleDeleteNodePool` → `proxyDELETE(w, r, "clusters/{id}/nodepools/{npid}")`

## 6. Server — route table updates

- [x] 6.1 `/api/clusters`: add POST → `handleCreateCluster`
- [x] 6.2 `/api/clusters/{id}`: add PATCH → `handlePatchCluster`, DELETE → `handleDeleteCluster`
- [x] 6.3 `/api/clusters/{id}/nodepools`: add POST → `handleCreateNodePool`
- [x] 6.4 `/api/clusters/{id}/nodepools/{npid}`: add PATCH → `handlePatchNodePool`, DELETE → `handleDeleteNodePool`
- [x] 6.5 `/api/clusters/{id}/nodepools/{npid}/force-delete`: new route, POST only → `handleForceDeleteNodePool`
- [x] 6.6 `/api/clusters/{id}/statuses` POST: use `handlePostClusterStatuses` (now calls `proxyPUT`)
- [x] 6.7 `/api/clusters/{id}/nodepools/{npid}/statuses` POST: use `handlePostNodePoolStatuses` (now calls `proxyPUT`)

## 7. Server tests

- [x] 7.1 Update `TestPostClusterStatusesProxy`: expect upstream method PUT (was POST)
- [x] 7.2 Update `TestRouteMethodNotAllowed`: use `newTestServer`; only assert methods that are still disallowed
- [x] 7.3 Update `TestPostToNonStatusRouteReturns405`: remove routes that now accept POST (`/api/clusters`, `/api/clusters/{id}/nodepools`)
- [x] 7.4 Add `TestPostNodePoolStatusesProxy`: verify path forwarded correctly
- [x] 7.5 Add `TestPostStatusesUpstreamError`: verify 422 forwarded with correct content-type

## 8. UI — dot color correctness

- [x] 8.1 `buildRow` in `index.html`: filter adapter conditions by `isDeleted` — non-deleted excludes `Finalized`, deleted uses only `Finalized`
- [x] 8.2 Empty relevant conditions → gray dot (was incorrectly green)
- [x] 8.3 `buildNPRow`: same `isDeleted` guard (NP adapter columns show `·` placeholder; no NP adapter data fetched in table view)

## 9. UI + CLI — table column layout

- [x] 9.1 `cmd/resources.go`: remove `collectConditionCols`; add `fixedCondCols = []string{"Reconciled", "LastKnownReconciled"}`; add `condDot` helper; `buildClusterRow` and `buildNodePoolRow` emit fixed cond cells before adapter cells
- [x] 9.2 `index.html`: remove `collectCondCols`; add `FIXED_COND_COLS = ['Reconciled', 'LastKnownReconciled']`; `renderHeader` emits fixed cond `<th>` elements; `buildRow` and `buildNPRow` emit fixed cond `<td>` cells

## 10. UI — deleted markers

- [x] 10.1 `buildRow`: append ` ❌` to cluster name when `cluster.deleted_time` is set
- [x] 10.2 `buildNPRow`: append ` ❌` to nodepool name when `np.deleted_time` is set

## 11. UI — create cluster/nodepool forms

- [x] 11.1 Add `CLUSTER_TEMPLATE` and `NODEPOOL_TEMPLATE` JS constants (JSON from `cmd/assets/*.json`)
- [x] 11.2 Add `api.createCluster` and `api.createNodepool` to the `api` object
- [x] 11.3 Add `openCreateForm(mode, clusterID)`: sets `state.formMode`, renders `<textarea>` pre-filled with template JSON, opens form panel
- [x] 11.4 Add `submitCreateForm()`: parses JSON from textarea, calls correct API, shows success banner, closes after 1200 ms
- [x] 11.5 Update `closeForm()` to reset `formMode` to `'status'` and clear `formCreateClusterID`
- [x] 11.6 Update `submitForm()` dispatcher to call `submitCreateForm()` when `formMode` is `createCluster` or `createNodepool`
- [x] 11.7 Add `#create-cluster-btn` in page header HTML with CSS
- [x] 11.8 `renderActionsBar`: add `createNPBtn` (when no nodepool selected) and `forceDeleteBtn` (when nodepool selected)

## 12. UI — force-delete nodepool

- [x] 12.1 Add `api.forceDeleteNP` to the `api` object
- [x] 12.2 Add `startForceDelete(clusterID, npID)`, `cancelForceDelete()`, `confirmForceDelete()` JS functions
- [x] 12.3 Add `#force-delete-confirm-wrap` panel HTML with reason input and Confirm/Cancel buttons
- [x] 12.4 Update `_restoreDeleteConfirm()` to also restore force-delete panel state on detail panel refresh

## 13. Verification

- [x] 13.1 `go build ./...` — zero errors
- [x] 13.2 `go vet ./...` — zero warnings
- [x] 13.3 `go test ./...` — all tests pass
