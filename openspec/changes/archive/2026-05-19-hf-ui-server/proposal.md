## Why

The CLI's `hf table --watch` provides real-time cluster/nodepool monitoring in the terminal, but the terminal is poorly suited for inspecting dense condition data — column widths are constrained, tooltips are impossible, and navigating between a cluster overview and its detailed adapter statuses requires multiple commands. A browser-based dashboard makes the same data far more explorable without requiring users to learn new API patterns.

## What Changes

- New `hf ui` command starts an embedded HTTP server (default port 8088) that serves a single-page HTML dashboard.
- New `internal/server` package handles HTTP routing and proxies read requests to the HyperFleet API using the existing `api.Client`.
- New `internal/ui/static/index.html` — a self-contained HTML file (CSS + JS inline, no external dependencies) embedded via `go:embed`.
- The dashboard auto-polls every 5 seconds, displays a cluster table with colored condition dots, and opens a side panel on row click showing full conditions, adapter statuses, and expandable nodepools.
- The server reuses config loading and `api.Client` initialization from existing commands — no new auth or URL logic.

## Capabilities

### New Capabilities

- `ui-server`: HTTP server mode for `hf`; serves the HTML dashboard and proxies HyperFleet API reads. Includes all backend routes and the embedded frontend asset.

### Modified Capabilities

_(none — existing CLI commands and specs are unchanged)_

## Impact

- **New files:** `cmd/ui.go`, `internal/server/server.go`, `internal/server/handlers.go`, `internal/ui/static/index.html`
- **Modified files:** `cmd/root.go` (add `AddCommand`)
- **Dependencies:** no new Go module dependencies; uses stdlib `net/http` and `embed`
- **No breaking changes** to existing commands or config

## Testing Scope

- `internal/server` — unit tests with `httptest.NewServer` covering:
  - Route dispatch (correct handler per path pattern)
  - Proxy correctness (upstream URL construction, response passthrough)
  - Error handling (upstream 4xx/5xx forwarded as JSON)
- Live verification: `hf ui` started against real cluster; browser confirms cluster rows and dots match `hf table` output

Verification steps requiring live cluster access:
- Confirming cluster/nodepool data populates the dashboard
- Confirming detail panel loads adapter statuses and conditions accurately
