## Context

Three commands render data as tables: `hf cluster list --output table` (`cmd/cluster.go`), `hf nodepool list --output table` (`cmd/nodepool.go`), and `hf table` / `hf resources` (`cmd/resources.go`). All three call `p.PrintTable(headers, rows)` as their terminal step.

Currently all three share a package-level `outputFmt` global set by the persistent `--output` flag. The watch loop must be grafted onto each command's `RunE` without duplicating the data-fetch and table-build logic that already exists.

`AdapterStatus.LastReportTime` is an RFC 3339 string already present in `resource.AdapterStatus`; no API changes are needed.

## Goals / Non-Goals

**Goals:**
- Add `--watch` boolean flag and `-s int` (seconds, default 5) to `clusterListCmd`, `nodepoolListCmd`, `resourcesCmd`, and `tableCmd`.
- On each tick: clear the terminal, re-fetch all required resources, re-render the table.
- Prepend a braille spinner frame to adapter cell values when `last_report_time` is within `2 × frequency` seconds of now.
- `hf table` / `hf resources` default to `table` output format when no `--output` flag is given.
- Graceful exit on SIGINT (Ctrl+C) with no partial-line artifacts.

**Non-Goals:**
- Differential/incremental rendering (redraw only changed rows).
- Watch mode for non-table output formats (`--output json` / `--output yaml`).
- Changing the spinner frame rate independently of the fetch frequency.
- Any persistent spinner library dependency.

## Decisions

### D1 — Screen clearing

Use the ANSI escape sequence `\033[H\033[2J` written directly to `cmd.OutOrStdout()` before each render. This is the portable approach already implied by the project's colored-dot rendering. Alternative considered: `os.Stdout.Write([]byte("\033c"))` (full terminal reset) — rejected because it clears scroll-back and is too aggressive for a monitoring use-case.

### D2 — Watch loop mechanics

Use a package-local helper `runWatch(ctx context.Context, s int, fn func() error)` that:
1. Calls `fn()` immediately on first tick.
2. Uses `time.NewTicker(time.Duration(s) * time.Second)` for subsequent ticks.
3. Returns when context is cancelled or `fn()` returns an error.

Each command's `RunE` checks `watchMode` and, if set, installs a `context.WithCancel` wired to `os.Signal` (SIGINT/SIGTERM) before entering `runWatch`. This is placed in `cmd/watch.go` (new file) to avoid duplicating the signal-handling boilerplate across three command files.

### D3 — Activity indicator placement

Add two new helpers in `internal/output/`:
- `SpinnerFrame(tick int) string` — returns `spinnerFrames[tick%10]` where `spinnerFrames` is the braille array.
- `IsActive(lastReportTime string, frequencySecs int) bool` — parses the RFC 3339 string and returns true if `time.Since(t) < 2 × frequency`.

`adapterDot` (in `cmd/resources.go`) and the equivalent in `cmd/cluster.go` / `cmd/nodepool.go` accept two new parameters: `tick int` and `frequencySecs int`. They call `output.IsActive(as.LastReportTime, frequencySecs)` and prepend `output.SpinnerFrame(tick) + " "` when active.

These are pure functions with no I/O, easy to unit-test without `httptest`.

### D4 — Default output format for `hf table` / `hf resources`

Introduce a command-local `init()` override: set `outputFmt = "table"` as the default value for the `--output` flag on `resourcesCmd` and `tableCmd` only. Cobra respects per-command flag defaults, so this does not affect the global `outputFmt` used by other commands.

Alternatively we could special-case `if cmd.Name() == "table"` — rejected, more brittle.

### D5 — watch flag scope

`--watch` and `-s` are local flags (not persistent) on each command, so they do not accidentally propagate to subcommands.

## Risks / Trade-offs

- **Flicker on slow terminals** — full clear + redraw causes visible flicker. Mitigation: acceptable for an operator-focused CLI; differential rendering is explicitly a non-goal.
- **RFC 3339 parse errors for `last_report_time`** — if the field is empty or malformed, `IsActive` returns `false` (adapter shows no spinner). This is the safe default.
- **Signal handling in tests** — the signal goroutine must not leak in unit tests. Mitigation: `runWatch` accepts a `context.Context`; tests cancel the context directly rather than sending signals.
- **`time.Sleep` vs `time.Ticker` for the first iteration** — using `Ticker` means the first render happens at time 0 (immediate), subsequent renders at `s`-second intervals. This is the correct UX.
