# Tasks: cluster-nodepool-id-commands

- [x] 1. Add `hf cluster id` — `clusterIDCmd` in `cmd/cluster.go`, registered in `init()`
- [x] 2. Add `hf nodepool id` — `nodepoolIDCmd` in `cmd/nodepool.go`, registered in `init()`
- [x] 3. Add `npBase` and `requireClusterID` helpers to `cmd/nodepool.go`; update all nodepool API paths to use cluster-scoped route
- [x] 4. Add `TestClusterID` to `cmd/cluster_test.go`
- [x] 5. Add `TestNodepoolID` to `cmd/nodepool_test.go`
- [x] 6. Run `go build ./...` — must pass with zero errors
- [x] 7. Run `go vet ./...` — must pass with zero warnings
- [x] 8. Run `go test ./...` — must pass with zero failures; save output to `verification_proof/tests.txt`
- [x] 9. Commit and push all changes
