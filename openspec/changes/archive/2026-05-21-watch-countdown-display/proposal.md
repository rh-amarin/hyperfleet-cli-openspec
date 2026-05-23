## Why

`hf table --watch` and `hf resources --watch` silently loop — the screen refreshes every N seconds but gives no feedback between refreshes. Users cannot tell whether the CLI is alive, how long until the next update, or whether the data is stale. Other watch-mode CLIs (`kubectl`, `watch(1)`) always show a refresh timer.

## What Changes

In watch mode (`--watch` flag with table output), a single status line is printed above the table on every 500 ms spinner tick:

```
↻ 3s  ⠙
```

The number counts down from `N` (the `--seconds` value) to `1`, then resets to `N` after each data fetch. The braille spinner advances every tick so the user can see the process is alive even when the countdown does not change.

The line is flush-left, directly above the table header row — spatially near the `ID` column, which is always the first column.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `tables-and-lists`: The watch-mode scenario for `hf table --watch` / `hf resources --watch` MUST show a live countdown line above the table between data refreshes.

## Impact

- `cmd/resources.go`: `renderResourcesTable` gains a `secsLeft int` parameter; prints the countdown line when in watch mode. `runResources` tracks `nextRefresh time.Time` in the `runWatchFast` closure.
- No API, flag, config schema, or output-format changes.
- Non-watch (`hf table` without `--watch`) is unaffected — countdown line is suppressed when `frequencySecs == 0`.

## Testing Scope

| File | Changes |
|---|---|
| `cmd/resources_test.go` | Add `TestRenderResourcesTable_CountdownLine` and `TestRenderResourcesTable_NoCountdownInNonWatchMode` |
