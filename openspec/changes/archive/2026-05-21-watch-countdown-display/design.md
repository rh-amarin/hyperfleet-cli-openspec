## Context

`runWatchFast` fires two independent tickers:
- **spinner ticker** — 500 ms — calls `fn(tick, refresh=false)` for live animation without API calls
- **data ticker** — N seconds — calls `fn(tick, refresh=true)` to fetch fresh data

`renderResourcesTable(cmd, entries, adCols, tick, frequencySecs int)` is called on every tick (both kinds). It has no countdown information today.

## Countdown tracking

Add `nextRefresh time.Time` as a closure variable in `runResources`. Set it to `time.Now().Add(duration)` immediately after each successful data fetch (`refresh=true`). This is always set before the first render because `runWatchFast` calls `fn(0, true)` as its first action.

Compute `secsLeft` using ceiling integer arithmetic — no `math` import:

```go
func secsUntil(t time.Time) int {
    d := time.Until(t)
    if d <= 0 {
        return 0
    }
    return int((d + time.Second - 1) / time.Second)
}
```

Examples: 3.7 s → 4, 0.2 s → 1, 5.0 s → 5, 0 s → 0.

## Signature change

```go
// Before
func renderResourcesTable(cmd *cobra.Command, entries []clusterEntry, adapterCols []string, tick, frequencySecs int) error

// After
func renderResourcesTable(cmd *cobra.Command, entries []clusterEntry, adapterCols []string, tick, frequencySecs, secsLeft int) error
```

Non-watch call site (`fetchAndRenderResources`) passes `secsLeft=0`.

## Countdown line format

```go
if frequencySecs > 0 {
    fmt.Fprintf(cmd.OutOrStdout(), "↻ %ds  %s\n", secsLeft, output.SpinnerFrame(tick))
}
```

Printed before `p.PrintTable(...)` so it appears at the top of the screen — one line above the column headers, flush left near the `ID` column.

No extra blank line between the countdown and the table. The line is always one line tall so it does not cause column-width instability.

## Why `frequencySecs > 0` and not a separate bool

`frequencySecs > 0` already distinguishes watch mode from non-watch mode at this call site. Using a separate bool would require more parameters without clarity benefit.

## Non-destructive for non-watch

`fetchAndRenderResources` passes `secsLeft=0` and `frequencySecs=0`, so the guard `if frequencySecs > 0` suppresses the line entirely. `hf table` without `--watch` is byte-for-byte unchanged.

## Import additions

`"fmt"` is already imported in `cmd/resources.go`; no new imports required.
