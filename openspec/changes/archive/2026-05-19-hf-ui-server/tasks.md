## 1. Backend — Server Package

- [x] 1.1 Create `internal/server/server.go` with `Server` struct holding `*api.Client` and port, and a `Listen()` method that registers routes and calls `http.ListenAndServe`
- [x] 1.2 Create `internal/server/handlers.go` with a `proxyJSON` helper that fetches a path from the upstream API as `json.RawMessage` and writes it to the response with `Content-Type: application/json`
- [x] 1.3 Implement `handleClusters` that fetches `clusters`, concurrently fetches per-cluster statuses (max 5 goroutines), merges them into a combined JSON object, and returns the result
- [x] 1.4 Implement `handleCluster` (single cluster by ID), `handleClusterStatuses`, `handleNodePools`, `handleNodePool`, `handleNodePoolStatuses` as thin `proxyJSON` wrappers
- [x] 1.5 Implement the route dispatcher in `server.go` using `strings.Split` on the URL path to match the 6 API routes plus `GET /`
- [x] 1.6 Write unit tests in `internal/server/server_test.go` using `httptest.NewServer` for the upstream and `httptest.NewRecorder` for the response; cover: cluster list merge, single cluster proxy, 404 forwarding, unknown route → 404

## 2. Frontend Asset

- [x] 2.1 Create `internal/ui/static/index.html` — dark-theme single-page app with inline CSS and JS; implement the two-panel layout (cluster table left, detail panel right hidden by default)
- [x] 2.2 Implement cluster table rendering in JS: fetch `/api/clusters`, compute dynamic condition columns (Available first, Reconciled last, others alpha), render rows with colored dots
- [x] 2.3 Implement `statusDot(status)` JS function returning a colored `●` span matching CLI semantics (True=green, False=red, Unknown=yellow, missing=gray)
- [x] 2.4 Implement CSS hover tooltips on each dot showing condition type, status, reason, message, and last transition time
- [x] 2.5 Implement 5-second polling with `setInterval`; show countdown in header; update changed rows in-place (diff by cluster ID) to preserve scroll position
- [x] 2.6 Implement row-click handler that opens the detail side panel (CSS transform slide-in); fetch `/api/clusters/{id}/statuses` and render full conditions and adapter statuses
- [x] 2.7 Implement expandable nodepools section in the detail panel: list nodepool names, on expand fetch `/api/clusters/{id}/nodepools/{npid}/statuses` and render inline
- [x] 2.8 Implement close button (×) that hides the detail panel and clears selection

## 3. Embed and Command Wiring

- [x] 3.1 Create `internal/ui/embed.go` with `//go:embed static/index.html` and export `FS embed.FS`
- [x] 3.2 Wire `GET /` in server to serve `index.html` from the embedded FS with `Content-Type: text/html`
- [x] 3.3 Create `cmd/ui.go` with `newUICmd()`: parse `--port` (default 8088) and `--open` flags, load config, call `RequireActiveEnvironment()`, instantiate `api.Client`, call `server.Listen()`
- [x] 3.4 Add `rootCmd.AddCommand(newUICmd())` in `cmd/root.go`

## 4. Verification

- [x] 4.1 Run `go build ./...` and confirm zero errors
- [x] 4.2 Run `go vet ./...` and confirm zero warnings
- [x] 4.3 Run `go test ./internal/server/...` and save output to `verification_proof/server_tests.txt`
- [x] 4.4 Run `go test ./...` and save full output to `verification_proof/all_tests.txt`
- [x] 4.5 Start `hf ui` against the real cluster; open `http://localhost:8088` in a browser and confirm cluster rows with condition dots appear and match `hf table` output; save a screenshot or terminal recording to `verification_proof/live_ui.txt`
- [x] 4.6 Click a cluster row and confirm the detail panel opens with correct conditions and adapter statuses
- [x] 4.7 Expand a nodepool in the detail panel and confirm its conditions load
