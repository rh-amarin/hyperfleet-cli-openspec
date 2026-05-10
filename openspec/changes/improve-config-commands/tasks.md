# Tasks: improve-config-commands

## 1. Shared helper

- [x] 1.1 Add `helpOnNoArgs(n int) cobra.PositionalArgs` to `cmd/root.go` — prints help and returns blank error when `len(args) == 0`, falls back to `cobra.ExactArgs(n)` otherwise
- [x] 1.2 Update `main.go` to skip printing blank error strings (guard: `if msg := err.Error(); msg != ""`)

## 2. `hf config show` — state block

- [x] 2.1 In `configShowCmd.RunE` (resolved-config path), collect non-empty state keys (`active-environment`, `cluster-id`, `cluster-name`, `nodepool-id`) via `s.GetState`
- [x] 2.2 Build `out["state"]` map and prepend `"state"` to the `allSections` slice passed to `marshalYAMLOrdered`
- [x] 2.3 Remove the standalone `active-environment: <name>` header line (replaced by the state block)

## 3. `hf config get` — single-arg dot notation

- [x] 3.1 Change `configGetCmd.Use` to `"get <key>"` and `Args` to `helpOnNoArgs(1)`
- [x] 3.2 In `RunE`: split arg on first `.`; if dot found → `s.Get(section, field)`; if no dot → `s.GetState(key)`
- [x] 3.3 Update error message to reference the single key (e.g. `[ERROR] Config key 'hyperfleet.api-url' not found`)

## 4. Apply `helpOnNoArgs` to other commands

- [x] 4.1 `cmd/config.go` — `configSetCmd` (n=3), `configEnvCreateCmd` (n=1), `configEnvActivateCmd` (n=1), `configEnvDeleteCmd` (n=1), `configEnvShowCmd` (n=1)
- [x] 4.2 `cmd/cluster.go` — `clusterUpdateCmd` (n=1), `clusterAdapterPostStatusCmd` (n=3)
- [x] 4.3 `cmd/nodepool.go` — `nodepoolUpdateCmd` (n=1), `nodepoolDeleteCmd` (n=1)
- [x] 4.4 `cmd/pubsub.go` — `pubsubPublishClusterCmd` (n=1), `pubsubPublishNodePoolCmd` (n=1)
- [x] 4.5 `cmd/kube.go` — `kubeDebugCmd` (n=1); leave `pfDaemonCmd` unchanged (hidden internal command)
- [x] 4.6 `cmd/db.go` — `dbExecCmd` (n=1), `dbDeleteCmd` (n=1)

## 5. Tests

- [x] 5.1 Update `TestActiveEnvGuard_BlocksConfigGet` — use new single-arg form `"hyperfleet.api-url"`
- [x] 5.2 Update `TestConfigGet_Found` — use `"hyperfleet.api-url"`
- [x] 5.3 Update `TestConfigGet_NotFound` — use `"hyperfleet.nonexistent-key"`
- [x] 5.4 Update `TestConfigSet_Valid` read-back — use `"hyperfleet.api-version"`
- [x] 5.5 Add `TestConfigShow_StateVariables` — state.yaml with cluster-id/cluster-name; assert both appear in output under `state:`
- [x] 5.6 Update `TestConfigShow` — assert `state:` section present
- [x] 5.7 Add `TestConfigGet_StateKey` — state.yaml with cluster-id; assert plain-key lookup works
- [x] 5.8 Add `TestConfigGet_NoArgs_ShowsHelp` — zero args returns error and output contains help text

## 6. Verification

- [x] 6.1 Run `go build ./...` — must pass with no errors
- [x] 6.2 Run `go vet ./...` — must pass with no errors
- [x] 6.3 Run `go test ./cmd/... -v` — must pass; save output to `verification_proof/cmd_test.txt`
- [x] 6.4 Run `go test ./... -v` — must pass; save output to `verification_proof/all_test.txt`
