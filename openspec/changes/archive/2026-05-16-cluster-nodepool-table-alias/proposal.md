## Why

`hf cluster list --output table` and `hf nodepool list --output table` are the most frequent
read operations during a debugging session. Requiring `--output table` every time is tedious.
The existing `hf table` (alias for `hf resources`) established the precedent that short
table-specific subcommands are welcome. This change follows that pattern at the domain level:
`hf cluster table` and `hf nodepool table`.

## What Changes

Two new subcommands, each backed by the existing list render function with `outputFmt` forced to `"table"`:

- **`hf cluster table`** — equivalent to `hf cluster list --output table`
  - Registered on `clusterCmd` alongside the existing `list`, `get`, `search`, etc. subcommands.
  - Shares the same `fetchAndRenderClusterList` implementation as `clusterListCmd`.
  - Forces `outputFmt = "table"` before delegating; `--watch` / `--seconds` flags are NOT carried over (out of scope for a simple alias).

- **`hf nodepool table`** — equivalent to `hf nodepool list --output table`
  - Registered on `nodepoolCmd`.
  - Shares the same `fetchAndRenderNodepoolList` implementation as `nodepoolListCmd`.
  - Same constraints as cluster table.

No new packages. No changes to existing commands or flags.

## Testing Scope

| Package | Test cases |
|---|---|
| `cmd` (`cluster_test.go`) | `TestClusterTable` — GET /clusters called; response rendered as table (headers ID/NAME/GEN/STATUS present) |
| `cmd` (`nodepool_test.go`) | `TestNodepoolTable` — GET /clusters/{id}/nodepools called; response rendered as table (headers ID/NAME/TYPE/GEN/REPLICAS/STATUS present) |

Live cluster access required for step (d) only.
