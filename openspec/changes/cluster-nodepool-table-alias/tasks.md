# Tasks: cluster-nodepool-table-alias

## `cmd/cluster.go`

- [x] 1. Add `clusterTableCmd` — sets `outputFmt = "table"`, calls `fetchAndRenderClusterList(cmd, 0, 0)`
- [x] 2. Register `clusterTableCmd` on `clusterCmd` in `init()`

## `cmd/nodepool.go`

- [x] 3. Add `nodepoolTableCmd` — sets `outputFmt = "table"`, calls `fetchAndRenderNodepoolList(cmd, 0, 0)`
- [x] 4. Register `nodepoolTableCmd` on `nodepoolCmd` in `init()`

## Unit Tests

- [x] 5. Add `TestClusterTable` to `cmd/cluster_test.go` — GET /clusters returns list; output contains table headers (ID, NAME, GEN, STATUS)
- [x] 6. Add `TestNodepoolTable` to `cmd/nodepool_test.go` — GET /clusters/{id}/nodepools returns list; output contains table headers (ID, NAME, TYPE, GEN, REPLICAS, STATUS)

## Verify

- [x] 7. (a) `go build ./...` — must pass with zero errors
- [x] 8. (b) `go vet ./...` — must report no issues
- [x] 9. (c) `go test ./...` — must pass with zero failures; capture full output and save to `verification_proof/tests.txt`
- [x] 10. (d) Live verification against real cluster — run `hf cluster table` and `hf nodepool table` against `http://34.175.55.239:8000`; capture output and save to `verification_proof/live.txt`
