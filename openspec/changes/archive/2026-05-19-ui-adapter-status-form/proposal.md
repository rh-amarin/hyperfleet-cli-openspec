## Why

The `hf ui` dashboard is read-only. When debugging adapter behaviour it is useful to manually inject an adapter status report — the same thing `hf cluster statuses create` does on the CLI — without leaving the browser. Clicking an existing adapter status pre-fills the form so engineers can resend a slightly modified report in seconds.

## What Changes

- New **form panel** slides in from the right of the detail panel when an adapter status block is clicked; same visual style as the existing detail panel.
- The form is pre-filled with the clicked adapter's current data (name, generation, conditions with status/reason/message).
- **Observed Generation** is a numeric stepper with `−` / `+` buttons.
- **Condition status** is a radio group: `True` / `False` / `Unknown`.
- Conditions can be added (empty row) or removed (× button per row).
- Submitting POSTs to the correct upstream endpoint via two new server routes.
- On success the form shows a brief confirmation and the detail panel refreshes.
- The form can also be opened blank (no prefill) via a **"+ Report"** button in the adapter statuses section header.

## Capabilities

### New Capabilities

- `ui-status-form`: Browser form for submitting adapter status reports from within the `hf ui` dashboard. Covers both cluster-level and nodepool-level status endpoints.

### Modified Capabilities

- `ui-server`: Two new POST proxy routes are added to `internal/server/handlers.go` and the route dispatcher in `server.go`.

## Impact

- **Modified files:** `internal/server/server.go`, `internal/server/handlers.go`, `internal/server/server_test.go`, `internal/ui/static/index.html`
- **No new Go module dependencies**
- **No breaking changes** to existing routes or frontend behaviour

## Testing Scope

- `internal/server` — new unit tests for POST proxy routes: success passthrough, upstream error forwarding, method-not-allowed guard for non-POST to status paths.
- Live verification: open `hf ui`, click an adapter status, confirm form pre-fill, edit generation, change a condition status radio, submit, confirm the upstream API receives the POST.

Verification steps requiring live cluster access:
- Confirming the submitted report appears when the detail panel refreshes after submission.
