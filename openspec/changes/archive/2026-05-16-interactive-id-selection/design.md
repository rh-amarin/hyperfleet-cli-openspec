# Design: interactive-id-selection

## New package: `internal/selector`

Provides a thin abstraction over the fuzzy-finder so commands can be tested without a real terminal.

```go
package selector

// Item is a selectable row: shown as "<Name>  <ID>" in the picker.
type Item struct {
    ID   string
    Name string
}

// Selector picks one item from a list.
// Returns index -1 and nil error when the user aborts (Esc / Ctrl+C).
type Selector interface {
    Select(items []Item) (int, error)
}

// FuzzySelector is the live implementation backed by go-fuzzyfinder.
type FuzzySelector struct{}

func (FuzzySelector) Select(items []Item) (int, error) {
    idx, err := fuzzyfinder.Find(items, func(i int) string {
        return fmt.Sprintf("%-40s  %s", items[i].Name, items[i].ID)
    })
    if err == fuzzyfinder.ErrAbort {
        return -1, nil
    }
    return idx, err
}
```

`FuzzySelector` is the zero-value default; test code supplies a `mockSelector` that returns a hard-coded index.

## New dependency

`github.com/ktr0731/go-fuzzyfinder` — fzf-style interactive picker for Go. Handles raw-mode terminal I/O, fuzzy filtering, and arrow-key navigation. No transitive dependencies beyond `go-runewidth` (already indirect via kubernetes client-go). Added to `go.mod` and `go.sum`.

## `cmd/cluster.go` changes

**Flag** (package-level var, registered in `init()`):
```go
var clusterIDInteractive bool
// clusterIDCmd.Flags().BoolVarP(&clusterIDInteractive, "interactive", "i", false,
//     "Interactively select and set the active cluster")
```

**`RunE` branch**:
```go
if clusterIDInteractive {
    return runClusterIDInteractive(cmd, s, selector.FuzzySelector{})
}
// ... existing non-interactive path unchanged
```

**`runClusterIDInteractive(cmd, s, sel)`**:
1. `newAPIClient(s)` → build client
2. `api.Get[resource.ListResponse[resource.Cluster]](ctx, client, "clusters")` → fetch list
3. If `len(items) == 0` → `return fmt.Errorf("no clusters available")`
4. Build `[]selector.Item` from clusters
5. `sel.Select(items)` → idx
6. If idx == -1 (abort) → return nil
7. `s.SetState("cluster-id", items[idx].ID)` → persists to state.yaml
8. `fmt.Fprintf(cmd.OutOrStdout(), "Active cluster set to: %s (%s)\n", items[idx].Name, items[idx].ID)`

## `cmd/nodepool.go` changes

Same pattern. One additional prerequisite step:

**`runNodepoolIDInteractive(cmd, s, sel)`**:
1. `requireClusterID(s)` → clusterID (exits with error if not set)
2. `api.Get[resource.ListResponse[resource.NodePool]](ctx, client, npBase(clusterID))` → fetch list
3. If `len(items) == 0` → `return fmt.Errorf("no nodepools available for cluster %s", clusterID)`
4. Build `[]selector.Item` from nodepools
5. `sel.Select(items)` → idx
6. If idx == -1 → return nil
7. `s.SetState("nodepool-id", items[idx].ID)`
8. Print confirmation to stdout

## Test approach

Commands accept `sel selector.Selector` as a parameter, called by `RunE` with `selector.FuzzySelector{}`. Tests pass a `mockSelector`:

```go
type mockSelector struct{ idx int }
func (m mockSelector) Select(_ []selector.Item) (int, error) { return m.idx, nil }
```

HTTP is stubbed via `httptest.NewServer` (project standard). State writes are verified by reading the temp state file created by the test's config setup.

## Impact

| File | Change |
|---|---|
| `go.mod`, `go.sum` | Add `github.com/ktr0731/go-fuzzyfinder` |
| `internal/selector/selector.go` | New — `Item`, `Selector` interface, `FuzzySelector` |
| `internal/selector/selector_test.go` | New — interface compliance check |
| `cmd/cluster.go` | Add `-i` flag, `runClusterIDInteractive` |
| `cmd/nodepool.go` | Add `-i` flag, `runNodepoolIDInteractive` |
| `cmd/cluster_test.go` | Add `TestClusterIDInteractive_*` |
| `cmd/nodepool_test.go` | Add `TestNodepoolIDInteractive_*` |
