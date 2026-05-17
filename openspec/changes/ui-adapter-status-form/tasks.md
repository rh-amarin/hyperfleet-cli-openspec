## 1. Backend — POST proxy routes

- [x] 1.1 Add `proxyPOST` helper in `internal/server/handlers.go` that reads `r.Body`, wraps it in `json.RawMessage`, calls `api.Post[json.RawMessage]`, and writes the raw response
- [x] 1.2 Add `handlePostClusterStatuses(w, r, clusterID)` that calls `proxyPOST` with path `clusters/{id}/statuses`
- [x] 1.3 Add `handlePostNodePoolStatuses(w, r, clusterID, npID)` that calls `proxyPOST` with path `clusters/{id}/nodepools/{npid}/statuses`
- [x] 1.4 Update `route()` in `server.go`: remove the top-level GET-only guard; instead, for the two status paths check `r.Method` and dispatch to GET or POST handler; all other paths return 405 for non-GET
- [x] 1.5 Add unit tests in `server_test.go`: POST cluster statuses passthrough, POST nodepool statuses passthrough, upstream 422 forwarded, POST to non-status route → 405

## 2. Frontend — Form panel

- [x] 2.1 Add form panel HTML to `index.html` after `#detail-panel`: a `<div id="form-panel">` with header (title + close button) and a `<div id="form-body">` for dynamic content; start with `width: 0` (hidden)
- [x] 2.2 Add CSS for `#form-panel` matching the detail panel slide-in pattern; add styles for form inputs, the numeric stepper (−/display/+ layout), radio groups, condition rows, add/remove buttons, and success/error banners
- [x] 2.3 Add `openStatusForm(clusterID, adapter, npID=null)` JS function: sets `state.formClusterID` / `state.formNpID`, renders the form pre-filled with `adapter` data (or blank if null), opens the panel
- [x] 2.4 Add `renderForm(adapter)` that builds the form HTML: adapter name input, generation stepper bound to `state.formGen`, conditions list from adapter conditions, "+ Add condition" button, Cancel / Submit buttons
- [x] 2.5 Add `renderConditionRow(idx, cond)` that renders one condition row: type text input, True/False/Unknown radio group, reason text, message text, × remove button
- [x] 2.6 Wire the − / + stepper buttons to increment/decrement `state.formGen` and re-render the stepper display; disable − when value is 0
- [x] 2.7 Wire the "+ Add condition" button to push an empty condition to `state.formConditions` and re-render the conditions list
- [x] 2.8 Wire condition row × buttons to remove that condition from `state.formConditions` and re-render
- [x] 2.9 Wire condition field inputs (type, reason, message) and radio buttons to update `state.formConditions[idx]` on change
- [x] 2.10 Add `submitStatusForm()`: build `AdapterStatusCreateRequest` from form state, POST to `/api/clusters/{id}/statuses` (or nodepool variant), disable Submit during flight, show success banner and refresh detail on 2xx, show error banner on non-2xx
- [x] 2.11 Add `closeForm()` function that slides out `#form-panel` and clears form state
- [x] 2.12 Make each `adapter-block` in `renderAdapterBlock()` clickable: add `onclick="openStatusForm(...)"` with the cluster ID and adapter object; add a cursor-pointer style
- [x] 2.13 Add a "+ Report" button in the adapter statuses section header of `renderDetail()` that calls `openStatusForm(clusterID, null)`
- [x] 2.14 Wire the close button and Cancel button in the form panel to `closeForm()`

## 3. Verification

- [x] 3.1 Run `go build ./...` — zero errors
- [x] 3.2 Run `go vet ./...` — zero warnings
- [x] 3.3 Run `go test ./internal/server/...` and save output to `verification_proof/server_tests_form.txt`
- [x] 3.4 Run `go test ./...` and save output to `verification_proof/all_tests_form.txt`
- [ ] 3.5 Live: open `hf ui`, click an adapter status block, confirm form opens pre-filled with correct adapter/generation/conditions
- [ ] 3.6 Live: edit generation stepper, toggle a condition status radio, submit; confirm 2xx and detail panel refreshes
- [ ] 3.7 Live: open blank form via "+ Report", fill manually, submit successfully
