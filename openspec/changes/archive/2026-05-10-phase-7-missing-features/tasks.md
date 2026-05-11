# Tasks: Phase 7 — Missing Features

## cmd/cluster.go

- [x] Add `clusterSearchCmd` with no-args (get from state) and name-arg (query API) behaviors
- [x] Add `clusterPatchCmd` for `spec` and `labels` counter increment
- [x] Update `clusterDeleteCmd`: change `ExactArgs(1)` to `MaximumNArgs(1)`, resolve ID from state, output deleted cluster JSON
- [x] Update `clusterCreateCmd`: change `Use` and `Args` to support positional `[name] [region] [version]`; positional args take precedence over `--name` flag and defaults
- [x] Update `clusterStatusesCmd` table: add `FINALIZED` column
- [x] Register `clusterSearchCmd` and `clusterPatchCmd` in `init()`
- [x] Update `clusterCmd.Long` to list new subcommands

## cmd/nodepool.go

- [x] Add `nodepoolSearchCmd` mirroring cluster search
- [x] Add `nodepoolPatchCmd` mirroring cluster patch
- [x] Update `nodepoolStatusesCmd` table: add `FINALIZED` column
- [x] Register `nodepoolSearchCmd` and `nodepoolPatchCmd` in `init()`
- [x] Update `nodepoolCmd.Long` to list new subcommands

## cmd/config.go

- [x] Update `configShowCmd`: add `cobra.MaximumNArgs(1)`; when env-name arg provided, display that profile with path, secrets redacted, `[active]` prefix if applicable
- [x] Add `Aliases: []string{"new"}` to `configEnvCreateCmd`

## Tests

- [x] `TestClusterDelete`: update assertion to expect deleted cluster JSON in output
- [x] `TestClusterDelete_FromState`: new test — delete using state cluster-id
- [x] `TestClusterSearch_ByName_Found`: name search finds cluster, persists ID
- [x] `TestClusterSearch_ByName_NotFound`: outputs `[]`, exits 0
- [x] `TestClusterSearch_ByName_Multiple`: multiple matches, uses first
- [x] `TestClusterSearch_NoArgs_WithState`: behaves like get
- [x] `TestClusterSearch_NoArgs_WithoutState`: specific error message
- [x] `TestClusterPatch_Spec`: increments spec.counter, sends correct PATCH body
- [x] `TestClusterPatch_Labels`: increments labels.counter, sends correct PATCH body
- [x] `TestClusterPatch_NoArgs`: prints usage to stdout, exits 1
- [x] `TestClusterCreate_PositionalArgs`: positional args used for name, region, version
- [x] `TestNodepoolSearch_ByName_Found`: name search finds nodepool, persists ID
- [x] `TestNodepoolSearch_ByName_NotFound`: outputs `[]`, exits 0
- [x] `TestNodepoolSearch_NoArgs_WithState`: behaves like get
- [x] `TestNodepoolSearch_NoArgs_WithoutState`: specific error message
- [x] `TestNodepoolPatch_Spec`: increments spec.counter
- [x] `TestNodepoolPatch_Labels`: increments labels.counter
- [x] `TestNodepoolPatch_NoArgs`: prints usage, exits 1

## Verification

- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./...` passes (zero failures across all 11 packages)
- [x] Save test output to `verification_proof/`
