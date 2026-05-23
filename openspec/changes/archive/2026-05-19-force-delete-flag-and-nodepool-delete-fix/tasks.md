## 1. Fix nodepool delete ID resolution

- [x] 1.1 In `nodepoolDeleteCmd`, replace the early-exit on empty ID with `s.NodePoolID(explicit)` (same pattern as `nodepoolGetCmd`)

## 2. Add --force to hf cluster delete

- [x] 2.1 Add `clusterDeleteForce bool` and `clusterDeleteReason string` package-level vars in `cmd/cluster.go`
- [x] 2.2 Modify `clusterDeleteCmd.RunE` to POST `clusters/<id>/force-delete` when `--force` is set, and print `[INFO] Cluster '<id>' force-deleted`
- [x] 2.3 Register `--force` and `--reason` flags on `clusterDeleteCmd` in `init()`

## 3. Add --force to hf nodepool delete

- [x] 3.1 Add `nodepoolDeleteForce bool` and `nodepoolDeleteReason string` package-level vars in `cmd/nodepool.go`
- [x] 3.2 Modify `nodepoolDeleteCmd.RunE` to POST `clusters/<clusterID>/nodepools/<id>/force-delete` when `--force` is set, and print `[INFO] NodePool '<id>' force-deleted`
- [x] 3.3 Register `--force` and `--reason` flags on `nodepoolDeleteCmd` in `init()`

## 4. Tests

- [x] 4.1 Add `cluster_test.go` tests: `hf cluster delete --force` calls POST force-delete; `--reason` passes body; no `--force` still calls DELETE
- [x] 4.2 Add `nodepool_test.go` tests: `hf nodepool delete` with no args resolves from state; `--force` calls POST force-delete; `--reason` passes body

## 5. Verification

- [x] 5.1 Run `go build ./...` and `go vet ./...` — must pass
- [x] 5.2 Run `go test ./cmd/...` and save output to `verification_proof/unit-tests-cmd.txt`
- [x] 5.3 Live: run `hf nodepool delete` with active state and verify it resolves the active nodepool-id; save output to `verification_proof/live-nodepool-delete-state-fallback.txt`
- [x] 5.4 Live: run `hf cluster delete --force` and `hf nodepool delete --force` against a disposable resource; save output to `verification_proof/live-force-delete.txt`
