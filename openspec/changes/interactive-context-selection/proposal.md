## Why

The `--interactive` / `-i` flag was added to `hf cluster id` and `hf nodepool id` in the previous change. All other commands that fall back to an active cluster-id or nodepool-id from state still require the user to have set the active context in advance — or to pass an explicit ID argument.

Adding `-i` to those commands lets the user select the target cluster or nodepool on the fly, without a separate `hf cluster id -i` step first. This is especially useful for commands like `hf cluster delete -i` or `hf nodepool conditions -i`.

## What Changes

**Shared pickers in `cmd/cluster.go` and `cmd/nodepool.go`:**
- `pickClusterInteractive(cmd, s)` — fetches all clusters, shows fuzzy picker (via existing `clusterIDSel`), saves selection to state, prints `[INFO]` to stderr, returns selected ID. Returns `""` on abort.
- `pickNodepoolInteractive(cmd, s, clusterID)` — fetches nodepools for the given cluster, shows fuzzy picker (via existing `nodepoolIDSel`), saves selection to state, prints `[INFO]` to stderr, returns selected ID.

**New shared flag var per domain:**
- `var clusterInteractive bool` — registered as `-i`/`--interactive` on 6 cluster commands.
- `var nodepoolInteractive bool` — registered as `-i`/`--interactive` on 6 nodepool commands.

**Cluster commands receiving `-i`:** `get`, `patch`, `delete`, `conditions`, `statuses`, `adapter post-status`

**Nodepool commands receiving `-i`:** `get`, `patch`, `delete`, `conditions`, `statuses`, `adapter post-status`

`nodepoolDeleteCmd`'s `Args` changes from `helpOnNoArgs(1)` (required ID) to `cobra.MaximumNArgs(1)` so it can run with no positional arg when `-i` is set. A RunE guard preserves the existing error when neither arg nor `-i` is given.

The existing `clusterIDSel` and `nodepoolIDSel` selector vars (already injectable for tests) are reused by the new helpers — no new selector vars needed.

## Testing Scope

| Package | Test cases |
|---|---|
| `cmd` (`cluster_test.go`) | `TestPickClusterInteractive_Select` — helper saves state and returns ID; `TestPickClusterInteractive_Abort` — returns `""` without state write; `TestClusterGetInteractive` — GET called with picked ID; `TestClusterDeleteInteractive` — DELETE called with picked ID |
| `cmd` (`nodepool_test.go`) | `TestPickNodepoolInteractive_Select` — helper saves state and returns ID; `TestPickNodepoolInteractive_Abort` — returns `""` without state write; `TestNodepoolGetInteractive` — GET called with picked ID; `TestNodepoolDeleteInteractive` — DELETE called with picked ID |

Spot-checks for `get` and `delete` are sufficient — all other commands share the same helper injection point. Live cluster access is required for step (d) only.
