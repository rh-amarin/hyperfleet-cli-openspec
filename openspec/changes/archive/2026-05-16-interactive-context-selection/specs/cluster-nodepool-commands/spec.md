# Delta Spec: interactive-context-selection

## Scope

`cmd/cluster.go`, `cmd/nodepool.go`, `cmd/cluster_test.go`, `cmd/nodepool_test.go`

## New shared helpers

```go
// cmd/cluster.go
func pickClusterInteractive(cmd *cobra.Command, s *config.Store) (string, error)
// Uses clusterIDSel (injectable). Fetches clusters, shows picker.
// On select: s.SetState("cluster-id", id), p.Info("[INFO] cluster context set to: <name> (<id>)"), returns id.
// On abort: returns ("", nil).

// cmd/nodepool.go
func pickNodepoolInteractive(cmd *cobra.Command, s *config.Store, clusterID string) (string, error)
// Uses nodepoolIDSel (injectable). Fetches nodepools for clusterID, shows picker.
// On select: s.SetState("nodepool-id", id), p.Info("[INFO] nodepool context set to: <name> (<id>)"), returns id.
// On abort: returns ("", nil).
```

## New flag vars

```go
var clusterInteractive bool   // cmd/cluster.go — shared by 6 cluster commands
var nodepoolInteractive bool  // cmd/nodepool.go — shared by 6 nodepool commands
```

## Cluster commands: `-i` flag and RunE guard

Commands: `clusterGetCmd`, `clusterPatchCmd`, `clusterDeleteCmd`, `clusterConditionsCmd`, `clusterStatusesCmd`

Guard pattern (between `explicit` resolution and `s.ClusterID(explicit)`):
```go
if clusterInteractive && explicit == "" {
    explicit, err = pickClusterInteractive(cmd, s)
    if err != nil || explicit == "" {
        return err
    }
}
```

`clusterAdapterPostStatusCmd` — no `explicit` arg; replaces `s.ClusterID("")`:
```go
var clusterID string
if clusterInteractive {
    clusterID, err = pickClusterInteractive(cmd, s)
    if err != nil || clusterID == "" { return err }
} else {
    clusterID, err = s.ClusterID("")
    if err != nil { return err }
}
```

## Nodepool commands: `-i` flag and RunE guard

Commands: `nodepoolGetCmd`, `nodepoolPatchCmd`, `nodepoolConditionsCmd`, `nodepoolStatusesCmd`

Guard pattern (`requireClusterID` called FIRST, then picker):
```go
clusterID, err := requireClusterID(s)
if err != nil { return err }
if nodepoolInteractive && explicit == "" {
    explicit, err = pickNodepoolInteractive(cmd, s, clusterID)
    if err != nil || explicit == "" { return err }
}
id, err := s.NodePoolID(explicit)
```

`nodepoolDeleteCmd`:
- `Args` changes from `helpOnNoArgs(1)` → `cobra.MaximumNArgs(1)`
- RunE guard:
```go
id := ""
if len(args) > 0 { id = args[0] }
if id == "" && !nodepoolInteractive {
    return fmt.Errorf("[ERROR] nodepool ID required. Pass an explicit ID or use -i to select interactively.")
}
clusterID, err := requireClusterID(s)
...
if nodepoolInteractive && id == "" {
    id, err = pickNodepoolInteractive(cmd, s, clusterID)
    if err != nil || id == "" { return err }
}
```

`nodepoolAdapterPostStatusCmd`:
```go
if nodepoolInteractive && explicit == "" {
    explicit, err = pickNodepoolInteractive(cmd, s, clusterID)
    if err != nil || explicit == "" { return err }
}
nodepoolID, err := s.NodePoolID(explicit)
```

## Flag registration in `init()`

`cmd/cluster.go`:
```go
for _, c := range []*cobra.Command{
    clusterGetCmd, clusterPatchCmd, clusterDeleteCmd,
    clusterConditionsCmd, clusterStatusesCmd, clusterAdapterPostStatusCmd,
} {
    c.Flags().BoolVarP(&clusterInteractive, "interactive", "i", false,
        "interactively select the active cluster before running this command")
}
```

`cmd/nodepool.go`:
```go
for _, c := range []*cobra.Command{
    nodepoolGetCmd, nodepoolPatchCmd, nodepoolDeleteCmd,
    nodepoolConditionsCmd, nodepoolStatusesCmd, nodepoolAdapterPostStatusCmd,
} {
    c.Flags().BoolVarP(&nodepoolInteractive, "interactive", "i", false,
        "interactively select the active nodepool before running this command")
}
```

## Test additions

`cmd/cluster_test.go`:
- `clusterInteractive = false` added to `resetClusterFlags()`
- `TestPickClusterInteractive_Select` — mock returns idx 0; verify state + stderr INFO
- `TestPickClusterInteractive_Abort` — mock returns -1; verify no state write
- `TestClusterGetInteractive` — GET called with picked cluster ID
- `TestClusterDeleteInteractive` — DELETE called with picked cluster ID

`cmd/nodepool_test.go`:
- `nodepoolInteractive = false` added to `resetNodepoolFlags()`
- `TestPickNodepoolInteractive_Select` — mock returns idx 0; verify state + stderr INFO
- `TestPickNodepoolInteractive_Abort` — mock returns -1; verify no state write
- `TestNodepoolGetInteractive` — GET called with picked nodepool ID
- `TestNodepoolDeleteInteractive` — DELETE called with picked nodepool ID
