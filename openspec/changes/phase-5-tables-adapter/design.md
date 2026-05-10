## Overview

Two implementation tracks sharing a single change:

1. **Adapter post-status commands** — new cobra subcommand trees under `clusterCmd` and `nodepoolCmd`
2. **`hf resources` / `hf table`** — replace stub `RunE` with full combined table implementation

## Packages Affected

| Package | Change |
|---------|--------|
| `cmd/cluster.go` | Add `clusterAdapterCmd` group + `clusterAdapterPostStatusCmd` leaf |
| `cmd/nodepool.go` | Add `nodepoolAdapterCmd` group + `nodepoolAdapterPostStatusCmd` leaf |
| `cmd/resources.go` | Replace stub with full `RunE`; add `tableCmd` alias registered to `rootCmd` |
| `internal/resource/resource.go` | No change — all types already present |

## Adapter Post-Status Design

### Command trees

```
hf cluster adapter post-status <adapter_name> <status> <generation>
hf nodepool adapter post-status <adapter_name> <status> <generation> [nodepool_id]
```

Both use `cobra.ExactArgs(3)` (nodepool uses `cobra.RangeArgs(3,4)` to allow optional 4th arg).

### Validation

- `<status>` must be `True`, `False`, or `Unknown`; reject with `[ERROR] Invalid status value '<v>'. Must be one of: True, False, Unknown.` and exit 1
- `<generation>` parsed with `strconv.Atoi`; reject non-integer with error

### Request construction

```go
resource.AdapterStatusCreateRequest{
    Adapter:            adapterName,
    ObservedGeneration: int32(gen),
    ObservedTime:       time.Now().UTC().Format(time.RFC3339),
    Conditions: []resource.ConditionRequest{
        {Type: "Available", Status: status, Reason: "ManualStatusPost", Message: "Status posted via hf adapter post-status"},
        {Type: "Applied",   Status: status, Reason: "ManualStatusPost", Message: "Status posted via hf adapter post-status"},
        {Type: "Health",    Status: status, Reason: "ManualStatusPost", Message: "Status posted via hf adapter post-status"},
        {Type: "Finalized", Status: status, Reason: "ManualStatusPost", Message: "Status posted via hf adapter post-status"},
    },
}
```

### Endpoints

- Cluster: `POST clusters/{cluster_id}/statuses`
- NodePool: `POST clusters/{cluster_id}/nodepools/{nodepool_id}/statuses`

Note: the existing `newAPIClient` base URL ends in `clusters/` for cluster scope; for nodepools it's routed via `nodepools/{id}/statuses`. The API client's relative path handling means we call `api.Post[resource.AdapterStatus](..., "clusters/"+clusterID+"/statuses", body)` from the cluster-level base URL.

### Output

Default format is JSON (printer default). On HTTP 200, print the returned `AdapterStatus`. On HTTP 204, `api.Post` returns the zero value; print an empty `{}`. On success, also print `[INFO] Posted adapter status for <adapter> on cluster <cluster_id>` to stderr.

## Resources / Table Design

### Data flow

```
GET /clusters → []Cluster
For each cluster:
  GET /clusters/{id}/statuses → []AdapterStatus
  GET /clusters/{id}/nodepools → []NodePool (to get nodepool count and rows)
  For each nodepool:
    GET /clusters/{id}/nodepools/{npid}/statuses → []AdapterStatus
```

### Column building algorithm

1. Collect all unique condition types from cluster and nodepool `status.conditions`, excluding types ending in `Successful`. Preserve insertion order.
2. Collect all unique adapter names from all `AdapterStatus` items. Preserve insertion order.
3. Fixed headers: `ID`, `NAME`, `GEN` — then condition columns — then adapter columns.

### Row building

For each cluster:
- `ID` = cluster ID
- `NAME` = cluster name
- `GEN` = `strconv.Itoa(gen)` + ` ❌` if `deleted_time != ""`
- Condition columns: find matching condition in `cluster.Status.Conditions`; render `StatusDot(cond.Status, noColor) + " " + strconv.Itoa(int(cond.ObservedGeneration))`; `-` if absent
- Adapter columns: find matching `AdapterStatus`; use `Available` condition (or `Finalized` if cluster is being deleted); render `StatusDot(status, noColor) + " " + strconv.Itoa(int(as.ObservedGeneration))`; `-` if absent

For each nodepool (indented under its cluster):
- `ID` = `"  " + nodepool.ID`
- `NAME` = `"  " + nodepool.Name`
- `GEN` = same deletion marker logic
- Condition columns from nodepool conditions
- Adapter columns from nodepool adapter statuses

### Output format

- `--output table` (default): use `p.PrintTable(headers, rows)`
- `--output json`: print the raw clusters list response as JSON
- `--output yaml`: print the clusters list as YAML

### `hf table` alias

Register a second cobra command `tableCmd` pointing at the same `RunE` as `resourcesCmd`:

```go
var tableCmd = &cobra.Command{
    Use:   "table",
    Short: "Alias for hf resources",
    RunE:  resourcesCmd.RunE,
}
```

Both registered with `rootCmd.AddCommand`.

## Error handling

Same pattern as other commands: `handleAPIError(p, err)` for HTTP errors, plain `fmt.Errorf` for argument validation.

## Key decisions

- No new dependencies required
- Nodepool rows are fetched per-cluster (existing `clusters/{id}/nodepools` endpoint)
- Adapter status for nodepools: `clusters/{cluster_id}/nodepools/{id}/statuses`
- Column ordering is insertion-order (first seen across all resources)
- `hf resources` always outputs table by default — the `--output` flag controls format
