# Design: cluster-nodepool-table-alias

## Pattern reference

`hf table` (alias for `hf resources`) is implemented in `cmd/resources.go` by pointing two
`*cobra.Command` vars at the same `RunE` function. This change uses a simpler variant: a dedicated
`RunE` closure that sets `outputFmt = "table"` then calls the existing render helper.

## `cmd/cluster.go`

```go
var clusterTableCmd = &cobra.Command{
    Use:   "table",
    Short: "List all clusters in table format (alias for: cluster list --output table)",
    RunE: func(cmd *cobra.Command, args []string) error {
        outputFmt = "table"
        return fetchAndRenderClusterList(cmd, 0, 0)
    },
}
```

Registered in `init()`:
```go
clusterCmd.AddCommand(clusterTableCmd)
```

## `cmd/nodepool.go`

```go
var clusterTableCmd = &cobra.Command{
    Use:   "table",
    Short: "List all nodepools in table format (alias for: nodepool list --output table)",
    RunE: func(cmd *cobra.Command, args []string) error {
        outputFmt = "table"
        return fetchAndRenderNodepoolList(cmd, 0, 0)
    },
}
```

Registered in `init()`:
```go
nodepoolCmd.AddCommand(nodepoolTableCmd)
```

## Rationale

- Setting `outputFmt = "table"` unconditionally is intentional: the command name carries the
  semantic intent. A user who runs `hf cluster table --output json` is contradicting themselves
  and should use `hf cluster list --output json` instead.
- No `--watch` / `--seconds` flags: keeping the alias minimal. If watch is needed later it can be
  added as a separate change.
- Both `fetchAndRenderClusterList` and `fetchAndRenderNodepoolList` read `outputFmt` from the
  package-level var — setting it before the call is the correct integration point.

## Impact

| File | Change |
|---|---|
| `cmd/cluster.go` | Add `clusterTableCmd`; register in `init()` |
| `cmd/nodepool.go` | Add `nodepoolTableCmd`; register in `init()` |
| `cmd/cluster_test.go` | Add `TestClusterTable` |
| `cmd/nodepool_test.go` | Add `TestNodepoolTable` |
