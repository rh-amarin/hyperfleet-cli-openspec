## Context

`hf` is a pure CLI tool built with Cobra. It has no HTTP server logic today. All outbound API calls go through `internal/api`'s generic `Get[T]`/`Post[T]` functions using a configured `*api.Client`. Config and auth resolution live in `internal/config`. The existing `hf table --watch` command polls the HyperFleet API every N seconds and re-renders a terminal table; we want to replicate this pattern in a browser.

The binary must remain self-contained: no `kubectl`, `npm`, or external assets at runtime.

## Goals / Non-Goals

**Goals:**
- New `hf ui` subcommand starts an HTTP server and serves a browser dashboard.
- Dashboard shows live cluster/nodepool status matching `hf table` data and dot semantics.
- Detail side panel shows full conditions and adapter statuses for a selected cluster/nodepool.
- Zero new Go module dependencies (stdlib `net/http` + `embed` only).
- All backend routes are read-only proxies — no writes through the UI server.

**Non-Goals:**
- Write operations (create/delete cluster, scale nodepool) via the UI.
- Authentication on the UI server itself (the server is localhost-only by default).
- WebSocket or SSE live push (polling is sufficient and simpler).
- Multi-environment switching from the UI (environment is fixed to what's active at startup).

## Decisions

### D1 — stdlib `net/http` with manual path routing (no gorilla/mux)

The route set is small and stable (7 paths). A hand-written `switch` on `strings.Split(path, "/")` avoids pulling in a router dependency. Path variable extraction (`{id}`, `{npid}`) is done by index into the split slice.

*Alternative considered:* gorilla/mux (already in go.mod). Rejected because the added import surface is not worth it for 7 fixed routes, and the instructions say to avoid unnecessary dependencies.

### D2 — Single embedded `index.html` (inline CSS + JS via `go:embed`)

One file keeps the asset pipeline trivial: no build step, no asset bundler, binary size increase is bounded. CSS and JS are written directly into `<style>` and `<script>` tags.

*Alternative considered:* Separate `static/` directory with multiple files via `embed.FS`. Rejected: complicates navigation and editing for what is a single-page app with no framework.

### D3 — Browser polling every 5 s (not SSE or WebSocket)

The dashboard is a read-only observer, same as `hf table --watch`. SSE would require keeping a persistent HTTP connection open per tab. Polling with a `setInterval` + countdown timer gives the same user experience with simpler server code and no goroutine leaks.

### D4 — Proxy pattern: handlers fetch from upstream API using `api.Get[json.RawMessage]`

Rather than deserializing into typed structs and re-serializing, handlers proxy the raw JSON bytes from the upstream API directly to the browser. This avoids any data loss from partial struct mapping and means the frontend always sees the full API response shape.

*Exception:* The `/api/clusters` list endpoint also fetches per-cluster statuses and attaches them so the table view has condition data in a single request. This does deserialize into `resource.ListResponse[resource.Cluster]` and `resource.ListResponse[resource.AdapterStatus]` to assemble the combined payload.

### D5 — Frontend: vanilla JS, no framework

The UI is static HTML + vanilla `fetch()` + DOM manipulation. No React/Vue/Svelte. Keeps the embedded file small, loads instantly, and has no build-time dependencies. Chart-like rendering (dots, tooltips, side panel) is all CSS + DOM.

### D6 — `hf ui` bypasses the active-environment check via `_` prefix naming convention

Looking at `cmd/root.go`, daemon-style commands prefixed with `_` are bypass candidates. Instead, `hf ui` will explicitly handle the case: it calls `config.Store.RequireActiveEnvironment()` itself and prints a friendly error if none is set, rather than relying on PersistentPreRunE. This keeps the error message contextual.

## Package Layout

```
cmd/
  ui.go                  # Cobra command: flags, config load, server.Start()

internal/
  server/
    server.go            # Server struct, Listen(), route table
    handlers.go          # per-route handlers; proxy logic
  ui/
    static/
      index.html         # single-file embedded dashboard
    embed.go             # //go:embed directive
```

## API Routes (server → upstream)

| Browser request | Upstream API path |
|---|---|
| `GET /api/clusters` | `clusters` (+ per-cluster statuses fetched and merged) |
| `GET /api/clusters/{id}` | `clusters/{id}` |
| `GET /api/clusters/{id}/statuses` | `clusters/{id}/statuses` |
| `GET /api/clusters/{id}/nodepools` | `clusters/{id}/nodepools` |
| `GET /api/clusters/{id}/nodepools/{npid}` | `clusters/{id}/nodepools/{npid}` |
| `GET /api/clusters/{id}/nodepools/{npid}/statuses` | `clusters/{id}/nodepools/{npid}/statuses` |

## Frontend Architecture

**Two-panel layout:**
- Left: cluster table (full height, scrollable). Columns: Name, Age, then dynamic condition columns (Available first, Reconciled last, others alpha).
- Right: detail panel, hidden until a row is clicked. Slides in with CSS transform. Shows cluster metadata, full conditions, adapter statuses with sub-conditions, and an expandable nodepool list.

**Condition dot rendering** (matches `output.StatusDot()` semantics):
- `True` → `#3fb950` (green)
- `False` → `#f85149` (red)
- `Unknown` → `#d29922` (yellow)
- missing/empty → `#484f58` (gray)

**Dynamic column detection:** Collect all condition types from fetched clusters; sort with Available first, Reconciled last, others alphabetical between. Column headers are the condition type abbreviated to 4 chars with full name in `title` attribute (tooltip).

**Hover tooltips:** Custom CSS tooltip on each dot showing condition type, status, reason, message, and last transition time.

**Polling:** `setInterval(poll, 5000)`. Header shows countdown. On each poll, re-fetch `/api/clusters`, diff rows by cluster ID, update changed cells in place (no full re-render to avoid scroll position reset). If a cluster is selected, also re-fetch its statuses.

## Risks / Trade-offs

**[Risk] Combined clusters+statuses request on `/api/clusters` is N+1 upstream calls**
→ Mitigation: statuses are fetched concurrently per cluster using goroutines with a semaphore (max 5 concurrent). Acceptable for dashboards with O(10s) of clusters; revisit if cluster counts grow to hundreds.

**[Risk] Raw JSON proxy leaks full upstream API shape to the browser**
→ Acceptable: the server is localhost-only and the user already has CLI access to the same data.

**[Risk] Single HTML file becomes hard to maintain at scale**
→ Mitigation: keep JS under ~500 lines. If the file grows beyond that, split into a proper `embed.FS` with separate files. The design explicitly notes this escape hatch.

**[Risk] Port 8088 conflicts with other local services**
→ Mitigation: `--port` flag allows override. Error message on bind failure is clear.
