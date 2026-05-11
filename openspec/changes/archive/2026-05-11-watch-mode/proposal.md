## Why

Operators monitoring HyperFleet clusters must manually re-run table commands to see current state, creating friction during convergence events. A continuous watch mode eliminates this loop and adds a live activity signal so operators can immediately see which adapters are actively reporting without interpreting timestamps.

## What Changes

- `hf cluster list --output table` gains a `--watch` flag and a `-s <seconds>` frequency flag (default 5 s).
- `hf nodepool list --output table` gains the same `--watch` and `-s` flags.
- `hf table` / `hf resources` gains `--watch` and `-s` flags AND defaults to table output format (no `--output table` required; JSON/YAML still selectable via `--output`).
- When `--watch` is active, the command clears the screen and re-renders the table on every tick.
- Each adapter column gains an **activity character**: a braille spinner frame (⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏) prepended to the cell when the adapter's `last_report_time` is within `2 × frequency` seconds of now; otherwise the cell renders unchanged.
- The spinner frame advances on every refresh cycle (global frame counter mod 10).

## Capabilities

### New Capabilities

_(none — all changes extend existing commands)_

### Modified Capabilities

- `tables-and-lists`: cluster list table, nodepool list table, and combined table gain `--watch` / `-s` flags, a continuous refresh loop, and per-adapter activity indicator rendering.
- `output-formatting`: new spinner/activity character rendering rule added to the dot-rendering model for adapter columns.

## Impact

- **`cmd/cluster.go`** — add `--watch` and `-s` flags to `clusterListCmd`; add watch loop calling the existing table render path.
- **`cmd/nodepool.go`** — add `--watch` and `-s` flags to `nodepoolListCmd`; add watch loop.
- **`cmd/resources.go`** — add `--watch` and `-s` flags to `resourcesCmd` / `tableCmd`; default output format to `table`; add watch loop.
- **`internal/output/`** — new `SpinnerFrame(tick int) string` helper; `adapterDot` (or equivalent) accepts an `active bool` parameter that prepends the spinner character.
- **`cmd/cluster_test.go`**, **`cmd/nodepool_test.go`**, **`cmd/resources_test.go`** — tests for new flags; activity indicator unit tests (no live cluster required for unit tests).
- No new dependencies required; `time.Sleep` / `time.Ticker` from stdlib suffice.
- No API changes; activity is computed client-side from `last_report_time`.

## Testing Scope

| Package | Test cases needed |
|---------|-------------------|
| `internal/output` | `SpinnerFrame` cycles correctly; `adapterDot` with `active=true` prepends spinner; `active=false` unchanged |
| `cmd` (cluster) | `--watch` flag registered; `-s` flag registered; table rendered with activity indicator when `last_report_time` is recent |
| `cmd` (nodepool) | Same as cluster |
| `cmd` (resources) | Default output is `table` without `--output table`; `--watch` registered; activity indicator present |

Live cluster access required for: verifying the watch loop refreshes correctly in a real terminal, and confirming `last_report_time` values from the API match expected activity thresholds.
