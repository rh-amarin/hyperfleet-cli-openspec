# Delta Spec: cluster-nodepool-table-alias

## New subcommands

### `hf cluster table`

- Use: `table`
- Short: `List all clusters in table format (alias for: cluster list --output table)`
- Args: none
- Behavior: identical to `hf cluster list --output table`
- Registered on `clusterCmd`

### `hf nodepool table`

- Use: `table`
- Short: `List all nodepools in table format (alias for: nodepool list --output table)`
- Args: none
- Behavior: identical to `hf nodepool list --output table`
- Registered on `nodepoolCmd`

## Implementation

Both commands set the package-level `outputFmt = "table"` before delegating to the existing
list render helper (`fetchAndRenderClusterList` / `fetchAndRenderNodepoolList`).

No new flags, no watch mode, no changes to existing commands.
