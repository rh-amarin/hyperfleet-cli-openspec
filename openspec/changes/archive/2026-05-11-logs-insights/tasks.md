# Tasks: logs-insights

- [x] 1. Export `ParseLogfmt` in `internal/kube/kube.go` (rename private `parseLogfmt` → `ParseLogfmt`; update the two internal call sites in `streamPodLogsFiltered`)
- [x] 2. Add `CollectLogs(ctx, cs, namespace, podPattern string, sinceSeconds int64) ([]string, error)` to `internal/kube/kube.go`
- [x] 3. Add `TestCollectLogs` in `internal/kube/kube_test.go` using `httptest.NewServer`
- [x] 4. Create `internal/insights/insights.go` with `ParseAPILogs`, `ParseSentinelLogs`, `ParseAdapterLogs` and their result types
- [x] 5. Create `internal/insights/insights_test.go` with unit tests for all three parsers covering normal, empty, and error cases
- [x] 6. Add `logsInsightsCmd` to `cmd/logs.go` (flag `--since`/`-s`, parallel log collection, call parsers, print formatted output)
- [x] 7. Run `go build ./...` and `go vet ./...` — fix any errors
- [x] 8. Run `go test ./...` — fix any failures; save output to `verification_proof/unit_tests.txt`
- [x] 9. Run `hf logs insights` against the live cluster; save output to `verification_proof/live_cluster.txt`
