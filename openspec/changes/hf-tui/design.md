## Context

`hf table --watch` renders a combined cluster+nodepool table in the alternate screen buffer with periodic API refresh. Data fetching and row building live in `cmd/resources.go` (`fetchResourceEntries`, `buildClusterRow`, `buildNodePoolRow`). JSON colorization exists in `internal/output` (`colorizeJSON`, `ColorizeYAMLSections`). Status filter semantics for adapter vs condition search already exist in `cmd/status_filter.go`.

The TUI must reuse this data layer rather than duplicating API paths or table column logic.

## Goals / Non-Goals

**Goals:**
- New `hf tui` command launches a full-screen interactive dashboard.
- Main view matches `hf table --watch` columns, spinners, countdown, and refresh interval (`-s`, default 5).
- Keyboard-driven selection, detail panel, format cycling, status view, and smart search.
- Reuse `fetchResourceEntries` (exported from cmd or moved to shared package) and existing output helpers for dots/colorization.
- Unit-testable model logic separate from bubbletea event loop.

**Non-Goals:**
- Write operations (create/delete/patch) from the TUI.
- Mouse support.
- Multiple simultaneous detail panels.
- Replacing or modifying `hf table --watch` behavior.

## Decisions

### D1 — bubbletea + bubbles + lipgloss

Use Charm's bubbletea ecosystem for the full-screen TUI. It handles alternate screen, resize, and key events cleanly.

*Alternative considered:* tview/tcell (already indirect via go-fuzzyfinder). Rejected: bubbletea composes better with overlay search and mode state machines.

### D2 — Export table data helpers from cmd to internal/tui

Move shared row-building types/functions used by both `resources.go` and TUI into `internal/tui/data.go` (or keep in cmd and call from tui via exported functions). Prefer extracting `fetchResourceEntries` and row builders to `internal/tui/data.go` to avoid cmd→tui circular imports.

Actually: keep fetch in cmd and pass a `DataFetcher` interface to `internal/tui` so tui doesn't import cmd. The cmd layer wires the real fetcher; tests inject httptest-backed fetcher.

### D3 — Detail panel content modes

Three resource detail formats cycle with `V`:
1. **json** — `json.MarshalIndent` + `output.ColorizeJSONBytes` (export colorize helper)
2. **yaml** — `yaml.Marshal` + `output.ColorizeYAMLSections`
3. **overview** — human-readable summary (id, name, generation, key conditions, nodepool count for clusters)

Pressing `S` toggles **statuses** view showing adapter status list (JSON/YAML of `ListResponse[AdapterStatus]` or formatted overview reusing status preview helpers from `status_filter.go` moved to `internal/tui/statuses.go`).

### D4 — Smart search (`/`)

When detail panel is in statuses view and user presses `/`:
- Show `bubbles/textinput` overlay at bottom.
- First character case determines filter mode:
  - **lowercase start** → filter adapters whose name partially matches (case-insensitive); display matching adapter statuses only.
  - **uppercase start** → filter conditions whose `type` partially matches; display cross-adapter condition lines with adapter name (reuse `statusConditionPreview` logic).
- `Esc` cancels search and restores unfiltered view.
- `Enter` applies filter and closes input.

### D5 — Layout

Split horizontal when detail open:
- Left: table (flex ~55%)
- Right: scrollable viewport (~45%)
When detail closed: table uses full width.

Footer shows key hints: `↑↓ select · Enter detail · V format · S statuses · / search · Esc close · q quit`

### D6 — Refresh loop

Mirror `runWatchFast`: 500 ms spinner tick, data fetch every `-s` seconds. Model holds cached entries; spinner tick re-renders without API call.

## Risks / Trade-offs

- **[Risk] TTY required** → Document ttyd testing in verification; unit tests mock tea program.
- **[Risk] Wide tables on narrow terminals** → Horizontal scroll on table panel; truncate detail if needed.
- **[Risk] New dependencies** → Acceptable for dedicated TUI; three charm packages only.

## Migration Plan

No migration — new command only. Ship behind `hf tui` subcommand.

## Open Questions

_(none — requirements are fully specified by user)_
