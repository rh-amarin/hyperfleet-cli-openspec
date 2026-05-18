# Tasks: interactive-id-selection

## Dependency

- [x] 1. Add `github.com/ktr0731/go-fuzzyfinder` to `go.mod` and `go.sum` (`go get github.com/ktr0731/go-fuzzyfinder`)

## New package: `internal/selector`

- [x] 2. Create `internal/selector/selector.go` — `Item` struct, `Selector` interface, `FuzzySelector` implementation

## `cmd/cluster.go`

- [x] 3. Add `var clusterIDInteractive bool` and register `--interactive` / `-i` flag on `clusterIDCmd` in `init()`
- [x] 4. Add `runClusterIDInteractive(cmd, s, sel selector.Selector) error` — fetch clusters, open picker, save selection via `s.SetState("cluster-id", ...)`, print confirmation
- [x] 5. Branch in `clusterIDCmd.RunE`: when `clusterIDInteractive` is true, call `runClusterIDInteractive(cmd, s, selector.FuzzySelector{})`

## `cmd/nodepool.go`

- [x] 6. Add `var nodepoolIDInteractive bool` and register `--interactive` / `-i` flag on `nodepoolIDCmd` in `init()`
- [x] 7. Add `runNodepoolIDInteractive(cmd, s, sel selector.Selector) error` — require cluster-id, fetch nodepools, open picker, save selection via `s.SetState("nodepool-id", ...)`, print confirmation
- [x] 8. Branch in `nodepoolIDCmd.RunE`: when `nodepoolIDInteractive` is true, call `runNodepoolIDInteractive(cmd, s, selector.FuzzySelector{})`

## Unit Tests

- [x] 9. Create `internal/selector/selector_test.go` — compile-time `Selector` interface compliance check for `FuzzySelector`
- [x] 10. Add `TestClusterIDInteractive_Select` to `cmd/cluster_test.go` — mock selector returns index 1; verify correct `cluster-id` written to state and correct stdout message
- [x] 11. Add `TestClusterIDInteractive_Abort` to `cmd/cluster_test.go` — mock selector returns -1; verify no state write and no stdout output
- [x] 12. Add `TestClusterIDInteractive_Empty` to `cmd/cluster_test.go` — API returns empty list; verify `"no clusters available"` error
- [x] 13. Add `TestNodepoolIDInteractive_Select` to `cmd/nodepool_test.go` — mock selector returns index 0; verify correct `nodepool-id` written to state and correct stdout message
- [x] 14. Add `TestNodepoolIDInteractive_Abort` to `cmd/nodepool_test.go` — mock selector returns -1; verify no state write
- [x] 15. Add `TestNodepoolIDInteractive_NoCluster` to `cmd/nodepool_test.go` — no cluster-id in state; verify error before any API call

## Verify

- [x] 16. (a) `go build ./...` — must pass with zero errors
- [x] 17. (b) `go vet ./...` — must report no issues
- [x] 18. (c) `go test ./...` — must pass with zero failures; capture full output and save to `verification_proof/tests.txt`
- [x] 19. (d) Live verification against real cluster — run `hf cluster id -i` and `hf nodepool id -i` interactively; capture session output and save to `verification_proof/live.txt`
