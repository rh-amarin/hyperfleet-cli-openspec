# Tasks: interactive-context-selection

## Dependency

- [x] 1. Previous change `interactive-id-selection` must be merged (adds `internal/selector`, `clusterIDSel`, `nodepoolIDSel`)

## `cmd/cluster.go`

- [x] 2. Add `var clusterInteractive bool`
- [x] 3. Add `pickClusterInteractive(cmd *cobra.Command, s *config.Store) (string, error)` — fetch clusters via `clusterIDSel`, save selection to state, print `[INFO]` to stderr, return selected ID; return `("", nil)` on abort
- [x] 4. Add `-i`/`--interactive` flag on `clusterGetCmd`, `clusterPatchCmd`, `clusterDeleteCmd`, `clusterConditionsCmd`, `clusterStatusesCmd`, `clusterAdapterPostStatusCmd` in `init()`
- [x] 5. Add RunE guard to `clusterGetCmd`: `if clusterInteractive && explicit == "" { explicit, err = pickClusterInteractive(...) }`
- [x] 6. Add RunE guard to `clusterPatchCmd` (uses `args[1]` as explicit)
- [x] 7. Add RunE guard to `clusterDeleteCmd`
- [x] 8. Add RunE guard to `clusterConditionsCmd`
- [x] 9. Add RunE guard to `clusterStatusesCmd`
- [x] 10. Add RunE guard to `clusterAdapterPostStatusCmd` (no explicit arg; replaces `s.ClusterID("")` inline)

## `cmd/nodepool.go`

- [x] 11. Add `var nodepoolInteractive bool`
- [x] 12. Add `pickNodepoolInteractive(cmd *cobra.Command, s *config.Store, clusterID string) (string, error)` — fetch nodepools via `nodepoolIDSel`, save selection to state, print `[INFO]` to stderr, return selected ID; return `("", nil)` on abort
- [x] 13. Add `-i`/`--interactive` flag on `nodepoolGetCmd`, `nodepoolPatchCmd`, `nodepoolDeleteCmd`, `nodepoolConditionsCmd`, `nodepoolStatusesCmd`, `nodepoolAdapterPostStatusCmd` in `init()`
- [x] 14. Change `nodepoolDeleteCmd` `Args` from `helpOnNoArgs(1)` to `cobra.MaximumNArgs(1)`; add RunE guard with `if id == "" && !nodepoolInteractive { return fmt.Errorf(...) }`
- [x] 15. Add RunE guard to `nodepoolGetCmd`: call `requireClusterID` FIRST, then picker if `nodepoolInteractive && explicit == ""`
- [x] 16. Add RunE guard to `nodepoolPatchCmd`
- [x] 17. Add RunE guard to `nodepoolConditionsCmd`
- [x] 18. Add RunE guard to `nodepoolStatusesCmd`
- [x] 19. Add RunE guard to `nodepoolAdapterPostStatusCmd` (nodepool explicit only; cluster stays `s.ClusterID("")`)

## Reset functions

- [x] 20. Add `clusterInteractive = false` to `resetClusterFlags()` in `cmd/cluster_test.go`
- [x] 21. Add `nodepoolInteractive = false` to `resetNodepoolFlags()` in `cmd/nodepool_test.go`

## Unit Tests

- [x] 22. Add `TestPickClusterInteractive_Select` — mock selector returns index 1; verify `cluster-id` written to state
- [x] 23. Add `TestPickClusterInteractive_Abort` — mock returns -1; verify no state write and no further API call
- [x] 24. Add `TestClusterGetInteractive` — GET called with picked ID; verify cluster JSON returned
- [x] 25. Add `TestClusterDeleteInteractive` — DELETE called with picked ID; verify no error
- [x] 26. Add `TestPickNodepoolInteractive_Select` — mock selector returns index 1; verify `nodepool-id` written to state
- [x] 27. Add `TestPickNodepoolInteractive_Abort` — mock returns -1; verify no state write and no further API call
- [x] 28. Add `TestNodepoolGetInteractive` — GET called with picked ID; verify nodepool JSON returned
- [x] 29. Add `TestNodepoolDeleteInteractive` — DELETE called with picked ID; verify no error

## Verify

- [x] 30. (a) `go build ./...` — must pass with zero errors
- [x] 31. (b) `go vet ./...` — must report no issues
- [x] 32. (c) `go test ./...` — must pass with zero failures; capture full output and save to `verification_proof/tests.txt`
- [x] 33. (d) Live verification against real cluster — run at least two `-i` commands against `http://34.175.55.239:8000`; capture session output and save to `verification_proof/live.txt`
