# Design: cluster-nodepool-id-commands

## New subcommands

### `hf cluster id`

Registered under `clusterCmd`. Reads `cluster-id` from state via `s.GetState("cluster-id")`.

- If set: prints the ID to stdout and exits 0.
- If empty: returns `[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.` and exits 1.

No flags. Output is a plain string (the raw ID), not JSON — consistent with how `hf config get` works for scalar values.

### `hf nodepool id`

Registered under `nodepoolCmd`. Reads `nodepool-id` from state via `s.GetState("nodepool-id")`.

- If set: prints the ID to stdout and exits 0.
- If empty: returns `[ERROR] No nodepool-id set in state. Run 'hf nodepool create' or 'hf nodepool search <name>' first.` and exits 1.

No flags. Same plain-string output convention as `hf cluster id`.

## Bug fix: nodepool cluster-scoped API paths

All `hf nodepool` subcommands were constructing paths as `nodepools/...` (flat). The nodepool lifecycle spec requires `clusters/{cluster_id}/nodepools/...`.

Two helpers added to `cmd/nodepool.go`:

```go
// npBase returns the cluster-scoped nodepool collection path.
func npBase(clusterID string) string {
    return "clusters/" + clusterID + "/nodepools"
}

// requireClusterID reads cluster-id from state with the spec-mandated error message.
func requireClusterID(s interface{ GetState(string) string }) (string, error) {
    id := s.GetState("cluster-id")
    if id == "" {
        return "", fmt.Errorf("[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.")
    }
    return id, nil
}
```

Every nodepool API call now uses `npBase(clusterID)` as the path prefix.

## Impact

| File | Change |
|---|---|
| `cmd/cluster.go` | Add `clusterIDCmd`, register in `init()` |
| `cmd/nodepool.go` | Add `nodepoolIDCmd`, `npBase`, `requireClusterID`; update all API paths; register in `init()` |
| `cmd/cluster_test.go` | Add `TestClusterID` |
| `cmd/nodepool_test.go` | Add `TestNodepoolID` |
