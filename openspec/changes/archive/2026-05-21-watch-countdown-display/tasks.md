# Tasks: watch-countdown-display

## Implementation

- [x] 1. `cmd/resources.go`: Add `secsUntil(t time.Time) int` helper — ceiling of `time.Until(t)` in seconds, clamped to 0
- [x] 2. `cmd/resources.go`: In `runResources`, add `var nextRefresh time.Time` to the `runWatchFast` closure; set `nextRefresh = time.Now().Add(...)` after each `refresh=true` data fetch; compute `secsLeft := secsUntil(nextRefresh)` and pass to `renderResourcesTable`
- [x] 3. `cmd/resources.go`: Update `renderResourcesTable` signature to `(..., tick, frequencySecs, secsLeft int) error`; before `p.PrintTable`, print `↻ Xs  <spinner>\n` when `frequencySecs > 0`
- [x] 4. `cmd/resources.go`: Update `fetchAndRenderResources` call to `renderResourcesTable` to pass `secsLeft=0`
- [x] 5. `cmd/resources_test.go`: Add `TestRenderResourcesTable_CountdownLine` — call `renderResourcesTable` with `frequencySecs=5, secsLeft=3`, assert output contains `↻ 3s`
- [x] 6. `cmd/resources_test.go`: Add `TestRenderResourcesTable_NoCountdownInNonWatchMode` — call with `frequencySecs=0, secsLeft=0`, assert output does NOT contain `↻`

## Verification

- [x] 7. `go build ./...` — no errors
- [x] 8. `go vet ./...` — no warnings
- [x] 9. `go test ./cmd/...` — all tests pass; save output to `verification_proof/`
