## Why

Operators frequently need to switch their active cluster or nodepool context. Today the only path is `hf cluster search <name>` or `hf nodepool search <name>`, which requires knowing the exact name ahead of time. There is no way to browse what's available and pick interactively.

Adding `--interactive` / `-i` to `hf cluster id` and `hf nodepool id` lets the user open a fuzzy-searchable picker, scroll or type to filter, and confirm with Enter — the selected ID is immediately saved as the active context, identical to what `hf cluster search` does.

## What Changes

- `internal/selector/` (new package): `Selector` interface + `FuzzySelector` implementation wrapping `go-fuzzyfinder`. The interface allows test injection without spawning a real terminal.
- `go.mod` / `go.sum`: add `github.com/ktr0731/go-fuzzyfinder`.
- `cmd/cluster.go`: add `--interactive` / `-i` flag to `clusterIDCmd`. When set, fetches all clusters (`GET /clusters`), opens the fuzzy picker (name + id columns), saves the selection via `s.SetState("cluster-id", ...)`, and prints `Active cluster set to: <name> (<id>)`. Non-interactive path is unchanged.
- `cmd/nodepool.go`: same pattern for `nodepoolIDCmd`. Requires `cluster-id` to be set (uses existing `requireClusterID`), fetches nodepools for that cluster (`GET /clusters/{cluster_id}/nodepools`), and saves via `s.SetState("nodepool-id", ...)`.

## UX

```
$ hf cluster id -i
# opens full-screen fuzzy finder:
# > prod                 (typing filters)
#   prod-cluster-eu        019dc049-e79e-72a9-94f8-0056a11193cd
#   prod-cluster-us        019dc049-e76c-7be1-b201-0db50e2c8ecb
# ↑↓ to navigate, Enter to select, Esc to cancel
Active cluster set to: prod-cluster-eu (019dc049-e79e-72a9-94f8-0056a11193cd)

$ hf nodepool id -i
# opens fuzzy finder listing nodepools under the active cluster
Active nodepool set to: workers-1 (019dc049-abcd-1234-b201-0db50e2c8ecb)
```

Esc / Ctrl+C exits cleanly with no state change and exit code 0.

## Testing Scope

| Package | Test cases |
|---|---|
| `internal/selector` | Compile-time interface check; `Item` fields; `Select` returns abort (-1, nil) on `ErrAbort` — tested via a stub `fuzzyfinder` in a build-tag-isolated file |
| `cmd` (`cluster_test.go`) | `TestClusterIDInteractive_Select`: mock selector returns index 1 → correct cluster-id saved, correct stdout; `TestClusterIDInteractive_Abort`: mock returns -1 → no state write, no output, exit 0; `TestClusterIDInteractive_Empty`: API returns 0 clusters → error "no clusters available" |
| `cmd` (`nodepool_test.go`) | `TestNodepoolIDInteractive_Select`: mock selector returns index 0 → correct nodepool-id saved; `TestNodepoolIDInteractive_Abort`: no state write; `TestNodepoolIDInteractive_NoCluster`: missing cluster-id → error before API call |

Live cluster access is required for verification step (d) only — `hf cluster id -i` and `hf nodepool id -i` against the real LoadBalancer. Steps (a)–(c) are fully local.
