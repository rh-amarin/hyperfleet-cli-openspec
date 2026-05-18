## Context

The existing `hf ui` server (`internal/server/`) handles only GET requests. The dashboard (`internal/ui/static/index.html`) is a single-page vanilla-JS app with a two-panel layout: cluster table (left) and detail panel (right). The detail panel renders adapter status blocks as `<div class="adapter-block">` elements. No write path exists yet in the server or the frontend.

The upstream API accepts adapter status reports via:
- `POST clusters/{clusterID}/statuses`
- `POST clusters/{clusterID}/nodepools/{npID}/statuses`

Request body shape (`resource.AdapterStatusCreateRequest`):
```json
{
  "adapter": "cl-job",
  "observed_generation": 1,
  "observed_time": "2026-05-17T10:00:00Z",
  "conditions": [
    { "type": "Ready", "status": "True", "reason": "...", "message": "..." }
  ]
}
```

## Goals / Non-Goals

**Goals:**
- Form panel that slides in to the right of the detail panel, same visual language.
- Clicking an adapter block pre-fills the form; a blank "+ Report" button opens an empty form.
- Generation numeric stepper (integer, min 0).
- Status radio buttons True / False / Unknown per condition row.
- Dynamic condition list: add empty row, remove any row.
- POST via two new server proxy routes; forward raw request body to upstream.
- Success banner + detail panel refresh on 2xx; error banner on non-2xx.

**Non-Goals:**
- Authentication or CSRF protection (server is localhost-only).
- Editing metadata fields (job_name, attempt, etc.) — conditions only.
- NodePool-level status form triggered from nodepool expand (cluster-level only in this change; the route is wired, the UI trigger is cluster adapter statuses only).

## Decisions

### D1 — Third sliding panel (form) sits to the right of detail panel

The layout becomes `[table] [detail] [form]`. The form panel uses the same CSS pattern as the detail panel (`width: 0` → `width: 420px` transition). When open, the detail panel stays visible; clicking a different adapter on the detail panel switches the form prefill without closing.

*Alternative considered:* Modal overlay. Rejected — it obscures the detail panel, making it impossible to compare the current state while editing.

### D2 — POST proxy handlers read `r.Body` and forward raw bytes

The server's `proxyPOST` helper reads `r.Body`, sets `Content-Type: application/json`, and forwards to the upstream path using `api.Client.do` — but `do` is unexported. Instead, use `api.Post[json.RawMessage]` with a `json.RawMessage` body (which marshals as-is). This keeps the proxy pattern consistent with GET.

*Note:* `json.Marshal(json.RawMessage(body))` is a no-op — RawMessage marshals its bytes verbatim.

### D3 — Observed time auto-set in the browser to `time.Now().toISOString()`

No date-picker. The form sets `observed_time` to the current UTC instant at submit time. This mirrors what the CLI adapters do.

### D4 — Form state lives entirely in JS (no server session)

The form is a plain HTML `<form>` element managed by JS event listeners. No fetch on every keystroke — only on submit.

### D5 — Route dispatcher extended to handle POST for status paths only

The existing `route()` switch checks `r.Method == http.MethodGet` at the top and returns 405 for anything else. This guard is moved inside each case so that status-create paths accept both GET and POST, while all other paths remain GET-only.

## Package Layout Changes

```
internal/server/
  handlers.go      + proxyPOST() helper
                   + handlePostClusterStatuses()
                   + handlePostNodePoolStatuses()
  server.go        route() updated: status paths accept GET and POST
  server_test.go   + POST passthrough tests, 405 for non-status paths

internal/ui/static/
  index.html       + form panel HTML, CSS, JS
```

## Form UI Layout

```
┌─ Form panel (420px) ──────────────────────┐
│ Report Adapter Status          [×]         │
├───────────────────────────────────────────┤
│ Adapter   [cl-job              ]           │
│ Generation  [−] [ 1 ] [+]                 │
│ Time      2026-05-17 10:00:00 UTC (auto)  │
├─ CONDITIONS ──────────────────────────────┤
│ Type     [Ready         ]                  │
│ Status   ○ True  ○ False  ○ Unknown       │
│ Reason   [ReconciledAll  ]   [×]          │
│ Message  [All adapters…  ]               │
│                                            │
│ [+ Add condition]                          │
├───────────────────────────────────────────┤
│            [Cancel]  [Submit Report]       │
└───────────────────────────────────────────┘
```

## Risks / Trade-offs

**[Risk] `json.RawMessage` round-trip through `api.Post` may double-encode**  
→ Mitigation: `json.RawMessage` implements `json.Marshaler` and marshals its bytes verbatim. Verified behaviour — no double-encoding.

**[Risk] Form panel pushes the table panel very narrow on small screens**  
→ Mitigation: form panel has `min-width: 0` and hides when closed; table panel has `overflow-x: auto`. Acceptable for a developer tool.

**[Risk] Submitting to a nodepool status endpoint from the cluster detail panel**  
→ Scoped out. The cluster detail panel only shows cluster-level adapter statuses; the POST route for nodepools is wired in the backend but not yet triggered from the UI. Nodepool form trigger is deferred.
