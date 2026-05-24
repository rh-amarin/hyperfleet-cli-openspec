## Why

Operators monitoring HyperFleet clusters today rely on repeatedly running `hf table --watch` or piping JSON through external tools to inspect resource details. A k9s-style terminal UI provides live cluster/nodepool overview with inline detail inspection, status filtering, and format switching — all in one interactive session without leaving the terminal.

## What Changes

- New top-level command `hf tui` launches a full-screen terminal application.
- Main panel mirrors `hf table --watch`: combined cluster+nodepool table with auto-refresh, countdown, and adapter activity spinners.
- Cursor keys (↑/↓) select a cluster or nodepool row; selection opens a scrollable detail panel on the right.
- Detail panel defaults to syntax-highlighted JSON of the selected resource; `V` cycles JSON → YAML → overview; `S` switches to adapter statuses view.
- `Esc` closes the detail panel; `/` opens a smart search field with case-sensitive adapter vs condition-type filtering (same semantics as existing status filter UI).
- `-s / --seconds` flag controls refresh interval (default 5, minimum 1).

## Capabilities

### New Capabilities

- `tui`: Interactive terminal dashboard for HyperFleet cluster/nodepool monitoring and inspection.

### Modified Capabilities

_(none — `hf table --watch` behavior is unchanged; TUI is a new command)_

## Impact

- **`cmd/tui.go`** — new Cobra command registering `hf tui`.
- **`internal/tui/`** — new package: bubbletea model, table rendering, detail panel, search overlay, data fetcher.
- **`go.mod`** — adds `github.com/charmbracelet/bubbletea`, `bubbles`, and `lipgloss` dependencies.
- **`openspec/specs/`** — new `tui/spec.md` capability spec after archive.

## Testing Scope

| Package / area | Test cases |
|---|---|
| `internal/tui/` | Model init; row selection; detail panel open/close; format cycling (json/yaml/overview); statuses view toggle; search filter logic (lowercase adapter / uppercase condition); detail content rendering; scroll offset bounds |
| `internal/tui/` (HTTP) | Data fetcher against `httptest.NewServer` returning cluster/nodepool/status payloads |
| `cmd/tui.go` | Command registration; `-s` flag defaults and validation; requires active environment |

## Live Verification

- `hf tui` against real cluster: table renders, selection opens detail panel, key bindings work.
- Manual verification via **ttyd** (non-interactive shells cannot run bubbletea).
