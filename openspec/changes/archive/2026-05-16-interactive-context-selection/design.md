# Design: interactive-context-selection

## Shared pickers

Two new package-level functions in `cmd/`:

```go
// pickClusterInteractive shows a fuzzy picker over all clusters, saves the selection to
// state, and returns the selected ID. Returns ("", nil) when the user aborts.
func pickClusterInteractive(cmd *cobra.Command, s *config.Store) (string, error)

// pickNodepoolInteractive shows a fuzzy picker over nodepools for clusterID, saves the
// selection to state, and returns the selected nodepool ID. Returns ("", nil) on abort.
func pickNodepoolInteractive(cmd *cobra.Command, s *config.Store, clusterID string) (string, error)
```

Both functions reuse the existing package-level selector vars (`clusterIDSel` and `nodepoolIDSel`) so that tests can swap in a `mockSel` without requiring a real terminal.

On selection, both print `[INFO] <kind> context set to: <name> (<id>)` to stderr (via `p.Info()`) and call `s.SetState(...)` to persist the chosen ID.

## Shared flag vars

```go
var clusterInteractive bool   // shared by all 6 cluster commands
var nodepoolInteractive bool  // shared by all 6 nodepool commands
```

Each command registers its own local `-i`/`--interactive` flag pointing to the shared var, identical to how `outputFmt` and `noColor` are shared via persistent flags on the root command. Tests reset these in `resetClusterFlags()` and `resetNodepoolFlags()`.

## Per-command RunE changes

### Cluster commands: `get`, `conditions`, `statuses`, `delete`

These already extract `explicit` from args and call `s.ClusterID(explicit)`. The change is a guard inserted between `explicit` resolution and `s.ClusterID`:

```go
if clusterInteractive && explicit == "" {
    explicit, err = pickClusterInteractive(cmd, s)
    if err != nil || explicit == "" {
        return err
    }
}
id, err := s.ClusterID(explicit)
```

### Cluster `patch`

Same pattern, but `explicit` is `args[1]` (second argument, optional):

```go
if clusterInteractive && explicit == "" {
    explicit, err = pickClusterInteractive(cmd, s)
    if err != nil || explicit == "" {
        return err
    }
}
id, err := s.ClusterID(explicit)
```

### Cluster `adapter post-status`

Uses `s.ClusterID("")` with no explicit arg. Replaced with:

```go
var clusterID string
if clusterInteractive {
    clusterID, err = pickClusterInteractive(cmd, s)
    if err != nil || clusterID == "" {
        return err
    }
} else {
    clusterID, err = s.ClusterID("")
    if err != nil {
        return err
    }
}
```

### Nodepool commands: `get`, `conditions`, `statuses`

These call `s.NodePoolID(explicit)` then `requireClusterID(s)`. The change adds a picker between them:

```go
clusterID, err := requireClusterID(s)
if err != nil {
    return err
}
if nodepoolInteractive && explicit == "" {
    explicit, err = pickNodepoolInteractive(cmd, s, clusterID)
    if err != nil || explicit == "" {
        return err
    }
}
id, err := s.NodePoolID(explicit)
```

Note: `requireClusterID` is called FIRST (before the picker) so that nodepool pickers always have a cluster to fetch from.

### Nodepool `patch`

Same as above — `explicit` comes from `args[1]`.

### Nodepool `delete`

Currently: `Args: helpOnNoArgs(1)`, `id := args[0]` (required positional).
After: `Args: cobra.MaximumNArgs(1)`, with an explicit guard:

```go
id := ""
if len(args) > 0 {
    id = args[0]
}
if id == "" && !nodepoolInteractive {
    return fmt.Errorf("[ERROR] nodepool ID required. Pass an explicit ID or use -i to select interactively.")
}
clusterID, err := requireClusterID(s)
...
if nodepoolInteractive && id == "" {
    id, err = pickNodepoolInteractive(cmd, s, clusterID)
    if err != nil || id == "" {
        return err
    }
}
```

### Nodepool `adapter post-status`

Uses `s.ClusterID("")` (no explicit cluster arg, must be set in state). Nodepool has optional `args[3]`. Both stay unchanged from state — only the nodepool-id gains `-i`:

```go
if nodepoolInteractive && explicit == "" {
    explicit, err = pickNodepoolInteractive(cmd, s, clusterID)
    if err != nil || explicit == "" {
        return err
    }
}
nodepoolID, err := s.NodePoolID(explicit)
```

## Flag registration in `init()`

Six flags added to `cluster.go` init():
```go
for _, cmd := range []*cobra.Command{
    clusterGetCmd, clusterPatchCmd, clusterDeleteCmd,
    clusterConditionsCmd, clusterStatusesCmd, clusterAdapterPostStatusCmd,
} {
    cmd.Flags().BoolVarP(&clusterInteractive, "interactive", "i", false,
        "interactively select the active cluster before running this command")
}
```
(Written as individual `Flags()` calls since the loop above is illustrative.)

Six flags added to `nodepool.go` init() for the same nodepool commands.

## Impact

| File | Change |
|---|---|
| `cmd/cluster.go` | `var clusterInteractive bool`; `pickClusterInteractive()`; `-i` flag on 6 commands; RunE guards in 6 commands |
| `cmd/nodepool.go` | `var nodepoolInteractive bool`; `pickNodepoolInteractive()`; `-i` flag on 6 commands; RunE guards in 6 commands; `nodepoolDeleteCmd` Args relaxed |
| `cmd/cluster_test.go` | `resetClusterFlags` reset; 4 new tests |
| `cmd/nodepool_test.go` | `resetNodepoolFlags` reset; 4 new tests |
