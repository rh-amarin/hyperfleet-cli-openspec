## 1. Spec and config (baseline)

- [x] 1.1 Add OpenSpec change `rs-entity-commands` with proposal, design, and delta specs
- [x] 1.2 Seed `resource-types` with `clusters` and `nodepools` in config template
- [x] 1.3 Add `templates/clusters.json` and `templates/nodepools.json` (copy from embedded cluster/nodepool templates)

## 2. Extract shared implementation (canonical path = hf rs)

- [x] 2.1 Extract table/row builders from `cmd/resources.go` into shared package used by `hf rs`
- [x] 2.2 Extract conditions/statuses/force-delete/create/patch/delete from `cluster.go` / `nodepool.go` into shared handlers parameterized by entity name + resolved path
- [x] 2.3 Wire all `hf rs clusters|nodepools` subcommands to shared handlers

## 3. hf rs subcommands (parity before deprecation)

- [x] 3.1 Register `table`, `conditions`, `statuses`, `force-delete`; `clusters delete --force --reason`
- [x] 3.2 Rich list/table for `clusters` and `nodepools` (condition + adapter columns, watch spinners)
- [x] 3.3 `hf rs` overview uses adapter-rich combined table (replaces `hf table` behavior)
- [x] 3.4 Create/patch/search/delete parity (duplicate guard, counter patch, flags, 404 messages)
- [x] 3.5 `adapter-report` (already implemented; verify against spec)

## 4. Tests on hf rs only

- [x] 4.1 Port critical scenarios from `cluster_test.go` / `nodepool_test.go` to `resource_*_test.go` using `hf rs clusters|nodepools`
- [ ] 4.2 Add parity test: `hf rs clusters table` equals former `hf cluster list -o table` output shape
- [ ] 4.3 Add overview test: `hf rs` matches former `hf table` column layout (when 3.3 done)

## 5. Deprecate legacy commands

- [ ] 5.1 Add `deprecate()` helper — print `[WARN] hf cluster is deprecated; use hf rs clusters …` on stderr
- [ ] 5.2 Delegate `hf cluster *`, `hf nodepool *`, `hf table`, `hf resources` to equivalent `hf rs` paths (warn once per invocation)
- [ ] 5.3 Update legacy command Short/Long help with deprecation notice

> Skipped in this pass: went straight to removal (§6).

## 6. Remove legacy commands

- [x] 6.1 Unregister `clusterCmd`, `nodepoolCmd`, `resourcesCmd`, `tableCmd` from `rootCmd`
- [ ] 6.2 Remove or slim `cmd/cluster.go`, `cmd/nodepool.go`, `cmd/resources.go` to moved shared code only
- [x] 6.3 Update shell completions and README command table to `hf rs` only
- [ ] 6.4 Delete redundant tests that only exercised legacy command paths

## 7. Spec archive prep

- [ ] 7.1 Add delta REMOVED requirements for `cluster-lifecycle` / `nodepool-lifecycle` command names (on archive)
- [ ] 7.2 Update `tables-and-lists` delta: combined view normative path is `hf rs`

## 8. Verification (Definition of Done)

- [x] 8.1 `go test ./...` — save to `verification_proof/go-test.txt`
- [x] 8.2 `go vet ./...` — save to `verification_proof/go-vet.txt`
- [x] 8.3 Live verification using **only** `hf rs` — save to `verification_proof/live.txt`
