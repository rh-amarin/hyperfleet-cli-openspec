## 1. Extend `internal/kube`

- [x] 1.1 Add `StartResult` struct to `internal/kube/kube.go` with fields `Name string`, `PID int`, `LocalPort int`, `RemotePort int`, `Namespace string`, `PodName string`
- [x] 1.2 Change `StartPortForward` signature from `(...) (int, int, int, error)` to `(...) (StartResult, error)` and populate all fields (capture namespace + pod name from `FindRunningPod`; set `PodName = ""` when pod lookup fails)
- [x] 1.3 Update `internal/kube/kube_test.go` to use `StartResult` return type in all relevant test assertions

## 2. Update `cmd/kube.go`

- [x] 2.1 Extract `printPortForwardStatus(w io.Writer, noColor bool)` helper in `cmd/kube.go` that calls `kube.ListPortForwards()` and renders the bullet table (same logic as current `pfStatusCmd` inline code)
- [x] 2.2 Update `pfStatusCmd` to call `printPortForwardStatus` instead of the inlined logic
- [x] 2.3 Update `pfStartCmd` to use the new `StartResult` return value and print the enriched `[INFO] Started <name> (<namespace>/<podName>): …` line (omit the `(…)` token when `PodName` is empty)
- [x] 2.4 Add call to `printPortForwardStatus` at the end of `pfStartCmd`, after all services have been started

## 3. Verify

- [x] 3.1 `go build ./...` succeeds
- [x] 3.2 `go vet ./...` passes
- [x] 3.3 `go test ./... 2>&1 | tee openspec/changes/port-forward-start-feedback/verification_proof/tests.txt`
- [x] 3.4 Run `hf kube port-forward start` against the live cluster and save output to `openspec/changes/port-forward-start-feedback/verification_proof/live.txt`; confirm namespace/pod appear in start lines and status table is shown
- [x] 3.5 Commit all changed files (implementation + tasks.md + verification_proof/) and push to main
