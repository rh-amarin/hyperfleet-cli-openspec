# Tasks: `--search` Flag for `cluster list` and `nodepool list`

## Checklist

- [x] 1. Add `clusterListSearch` flag var and wire `--search` flag to `clusterListCmd` in `cmd/cluster.go`
- [x] 2. Update `fetchAndRenderClusterList` to append `?search=<url-encoded>` to path when flag is set
- [x] 3. Add `nodepoolListSearch` flag var and wire `--search` flag to `nodepoolListCmd` in `cmd/nodepool.go`
- [x] 4. Update `fetchAndRenderNodepoolList` to append `?search=<url-encoded>` to path when flag is set
- [x] 5. Write unit tests in `cmd/cluster_test.go`: no-flag path, simple label filter, compound expression, API 400 error passthrough
- [x] 6. Write unit tests in `cmd/nodepool_test.go`: same four cases
- [x] 7. Run `go build ./...` — zero errors
- [x] 8. Run `go vet ./...` — zero errors
- [x] 9. Run `go test ./...` — zero failures; capture output to `verification_proof/go_test.txt`
- [x] 10. Verify against live API at `34.175.55.239:8000`; capture to `verification_proof/live.txt`
- [x] 11. Update spec delta files in `specs/cluster-lifecycle/spec.md` and `specs/nodepool-lifecycle/spec.md`
