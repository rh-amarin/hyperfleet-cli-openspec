## 1. Output Helpers — Spinner and Activity

- [ ] 1.1 Add `SpinnerFrame(tick int) string` to `internal/output/` returning `spinnerFrames[tick%10]` for the braille sequence
- [ ] 1.2 Add `IsActive(lastReportTime string, frequencySecs int) bool` to `internal/output/` using RFC 3339 parse + `time.Since` comparison
- [ ] 1.3 Write unit tests for `SpinnerFrame`: verify all 10 frames in order and that frame 10 == frame 0
- [ ] 1.4 Write unit tests for `IsActive`: recent timestamp → true; stale timestamp → false; empty/malformed → false

## 2. Watch Loop Helper

- [ ] 2.1 Create `cmd/watch.go` with `runWatch(ctx context.Context, out io.Writer, s int, fn func(tick int) error)` — calls `fn(0)` immediately, then on each ticker tick; clears screen via ANSI before each call
- [ ] 2.2 Add signal wiring helper in `cmd/watch.go`: `watchContext(parent context.Context) (context.Context, context.CancelFunc)` that cancels on SIGINT/SIGTERM
- [ ] 2.3 Write unit tests for `runWatch`: verify `fn` is called on first tick; verify it stops when context is cancelled; verify it stops and returns error when `fn` returns an error

## 3. Adapter Activity in Table Rendering

- [ ] 3.1 Update `adapterDot` in `cmd/resources.go` to accept `tick int` and `frequencySecs int`; prepend `output.SpinnerFrame(tick) + " "` when `output.IsActive(as.LastReportTime, frequencySecs)` is true
- [ ] 3.2 Update all callers of `adapterDot` in `cmd/resources.go` (`buildClusterRow`, `buildNodePoolRow`) to pass `tick` and `frequencySecs`
- [ ] 3.3 Confirm equivalent adapter-dot logic in `cmd/cluster.go` and `cmd/nodepool.go` (if duplicated there) is updated the same way

## 4. `hf cluster list` — Watch Flag

- [ ] 4.1 Add package-level `watchMode bool` and `watchSecs int` vars in `cmd/cluster.go`
- [ ] 4.2 Register `--watch` (bool, default false) and `-s` (int, default 5) on `clusterListCmd` in `init()`
- [ ] 4.3 In `runClusterList` (or equivalent), when `--output table` and `--watch` are both set: obtain a watch context, call `runWatch` with the fetch-and-render logic passing `tick` through to `adapterDot`
- [ ] 4.4 Write unit tests for `clusterListCmd`: `--watch` flag is registered; `-s` flag is registered; table output with active adapter prepends spinner character

## 5. `hf nodepool list` — Watch Flag

- [ ] 5.1 Add `watchMode bool` and `watchSecs int` vars in `cmd/nodepool.go`
- [ ] 5.2 Register `--watch` and `-s` on `nodepoolListCmd` in `init()`
- [ ] 5.3 In `runNodePoolList`, when `--output table` and `--watch` are both set: obtain a watch context, call `runWatch` with fetch-and-render logic passing `tick` to `adapterDot`
- [ ] 5.4 Write unit tests for `nodepoolListCmd`: `--watch` flag registered; `-s` flag registered; table output with active adapter prepends spinner

## 6. `hf table` / `hf resources` — Watch Flag + Default Table Format

- [ ] 6.1 Change the default value of `--output` on `resourcesCmd` and `tableCmd` to `"table"` in `init()` (per-command flag default, not the global)
- [ ] 6.2 Add `watchMode bool` and `watchSecs int` vars in `cmd/resources.go`
- [ ] 6.3 Register `--watch` and `-s` on `resourcesCmd` and `tableCmd` in `init()`
- [ ] 6.4 In `runResources`, when `--watch` is set: obtain a watch context, call `runWatch` with the full fetch-and-render logic passing `tick` and `watchSecs` to `adapterDot`
- [ ] 6.5 Write unit tests for `resourcesCmd`: no `--output` flag defaults to table; `--watch` flag registered; `-s` flag registered; `--output json` still works

## 7. Build and Test Verification

- [ ] 7.1 Run `go vet ./...` — fix any issues
- [ ] 7.2 Run `go build ./...` — fix any compilation errors
- [ ] 7.3 Run `go test ./...` — all tests pass; save output to `verification_proof/unit_tests.txt`
- [ ] 7.4 Live verification: run `hf cluster list --output table --watch -s 3` against real cluster; save terminal output to `verification_proof/live_cluster_watch.txt`
- [ ] 7.5 Live verification: run `hf nodepool list --output table --watch -s 3`; save output to `verification_proof/live_nodepool_watch.txt`
- [ ] 7.6 Live verification: run `hf table --watch -s 3`; save output to `verification_proof/live_table_watch.txt`
