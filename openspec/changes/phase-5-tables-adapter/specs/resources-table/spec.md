# Resources Table — Delta Spec

## Delta against: openspec/specs/tables-and-lists/spec.md

This delta implements the Combined Resources Overview requirement (`hf resources` / `hf table`) that was previously a stub.

## Implementation

### hf resources / hf table

Both commands produce the same output. `hf table` is registered as a cobra alias.

Default output format: table (no `--output table` flag required).

### Data fetched

1. GET `/clusters` → all clusters
2. For each cluster: GET `/clusters/{id}/statuses` → adapter statuses
3. For each cluster: GET `/clusters/{id}/nodepools` → nodepools
4. For each nodepool: GET `/clusters/{id}/nodepools/{npid}/statuses` → adapter statuses

### Column layout

- Fixed: `ID`, `NAME`, `GEN`
- Condition columns: unique `status.conditions[].type` across all clusters and nodepools, insertion order, excluding types ending in `Successful`
- Adapter columns: unique adapter names across all adapter statuses, insertion order

### Dot rendering with generation suffix

`StatusDot(status, noColor) + " " + strconv.Itoa(generation)` — e.g., `● 3`

### Deletion marker

When `deleted_time != ""`: append ` ❌` to GEN cell value. Adapter columns on deleted resources use `Finalized` condition instead of `Available`.

### Nodepool row indentation

Nodepool `ID` and `NAME` values are prefixed with `"  "` (two spaces).

### JSON / YAML output

When `--output json` or `--output yaml` is used, emit the raw clusters list response only (nodepools and adapter statuses are not fetched in non-table mode).
