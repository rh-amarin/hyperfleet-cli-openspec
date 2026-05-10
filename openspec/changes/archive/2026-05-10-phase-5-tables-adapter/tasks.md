# Tasks: phase-5-tables-adapter

## Track A — Adapter Post-Status Commands

- [x] A1. Add `clusterAdapterCmd` group and `clusterAdapterPostStatusCmd` to `cmd/cluster.go`; register under `clusterCmd`
- [x] A2. Add `nodepoolAdapterCmd` group and `nodepoolAdapterPostStatusCmd` to `cmd/nodepool.go`; register under `nodepoolCmd`
- [x] A3. Write `TestClusterAdapterPostStatus` in `cmd/cluster_test.go` (POST True, invalid status)
- [x] A4. Write `TestNodePoolAdapterPostStatus` in `cmd/nodepool_test.go` (POST True)

## Track B — hf resources / hf table

- [x] B1. Replace stub `RunE` in `cmd/resources.go` with full table implementation; register `tableCmd` alias
- [x] B2. Write `TestResourcesTable` in `cmd/resources_test.go`
- [x] B3. Write `TestResourcesJSON` in `cmd/resources_test.go`

## Verification

- [x] V1. `go build ./...` passes — save output to `verification_proof/build.txt`
- [x] V2. `go vet ./...` passes — save output to `verification_proof/vet.txt`
- [x] V3. `go test ./...` passes — save output to `verification_proof/test.txt`
- [x] V4. Add `verification_proof/live_verification_note.txt`
