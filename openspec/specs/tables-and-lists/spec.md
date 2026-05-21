# Tables and Lists Specification

## Purpose

Provide aggregated table and list views that display multiple resources in formatted output, including cluster lists, cluster tables with dynamic condition and adapter columns, nodepool tables, and a combined overview table.

## Table Column Architecture

All table views (cluster, nodepool, combined) follow a unified dual-column architecture:

**Fixed columns** vary per resource type (see individual requirements below).

**Condition columns** — one column per unique `status.conditions[].type` found across all displayed resources, with `*Successful` types excluded. Column ordering is determined by `unique` deduplication across the dataset.

**Adapter columns** — one column per unique adapter name found across all adapter status responses. These appear after condition columns.

### Dot rendering

Status values in condition and adapter columns are rendered as colored dots with an inline generation suffix:

- `True` → green `●` + `<observed_generation>` (e.g., `● 2`)
- `False` → red `●` + `<observed_generation>` (e.g., `● 1`)
- `Unknown` → yellow `●` + `<observed_generation>` (e.g., `● 3`)
- absent → `-`

For condition columns the generation comes from `condition.observed_generation`. For adapter columns it comes from `adapterStatus.observed_generation`.

### Deletion marker

When a resource has `deleted_time` set, the generation cell MUST append a red `❌` icon (e.g., `3 ❌`). Additionally, adapter columns for that resource MUST use the `Finalized` condition (instead of `Available`) when reading the adapter's status, reflecting that the adapter is reporting on finalization rather than availability.
## Requirements
### Requirement: List Clusters (JSON)

The CLI SHALL list all clusters as raw JSON.

#### Scenario: Empty cluster list

- GIVEN no clusters exist in the API
- WHEN the user runs `hf cluster list`
- THEN the CLI MUST output `{"items": [], "kind": "ClusterList", "page": 1, "size": 0, "total": 0}`

**Example**:
```json
{"items": [], "kind": "ClusterList", "page": 1, "size": 0, "total": 0}
```

#### Scenario: Populated cluster list

- GIVEN clusters exist
- WHEN the user runs `hf cluster list`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters`
- AND output the full JSON response with all cluster objects in the `items` array
- AND include pagination fields: `page`, `size`, `total`

**Example** (abbreviated):
```json
{
  "items": [
    {"id": "019dc049-5096-7f33-af06-8efe296e9e25", "kind": "Cluster", "name": "test-cluster-beta",  "generation": 1, ...},
    {"id": "019dc049-43a8-7a42-b44a-8d7f89e9e10f", "kind": "Cluster", "name": "test-cluster-alpha", "generation": 3, ...}
  ],
  "kind": "ClusterList",
  "page": 1,
  "size": 2,
  "total": 2
}
```

### Requirement: Cluster Table View

The CLI SHALL display clusters in a formatted table with dynamic condition and adapter columns when the `--table` flag is passed to `hf cluster list`.

#### Scenario: Empty table

- GIVEN no clusters exist
- WHEN the user runs `hf cluster list --table`
- THEN the CLI MUST output table headers only: `ID  NAME  GEN` with a separator line

**Example**:
```
ID   NAME  GEN
---  ---   ---
```

#### Scenario: Populated table

- GIVEN clusters exist
- WHEN the user runs `hf cluster list --table`
- THEN the CLI MUST:
  1. Fetch all clusters via GET `/api/hyperfleet/v1/clusters`
  2. For each cluster, fetch its adapter statuses via GET `/api/hyperfleet/v1/clusters/{id}/statuses`
- AND output a table with:
  - Fixed columns: `ID`, `NAME`, `GEN`
  - Dynamic condition columns: one per unique `status.conditions[].type` across all clusters, excluding types ending in `Successful`
  - Dynamic adapter columns: one per unique adapter name across all fetched statuses
- AND dot rendering MUST follow the rules in "Table Column Architecture" above
- AND the deletion marker MUST apply to any cluster with `deleted_time` set

**Example** — two clusters; `test-cluster-alpha` (gen 3) has three adapters reporting, `test-cluster-beta` (gen 1) has none. `● N` = colored dot + generation number (green=True, red=False):
```
ID                                    NAME                GEN  Available  Reconciled  cl-deployment  cl-job  cl-namespace
---                                   ---                 ---  ---        ---         ---            ---     ---
019dc049-5096-7f33-af06-8efe296e9e25  test-cluster-beta   1    ● 1(red)   ● 1(red)    -              -       -
019dc049-43a8-7a42-b44a-8d7f89e9e10f  test-cluster-alpha  3    ● 3(red)   ● 3(red)    ● 3(green)     ● 3(red) ● 3(green)
```

**Example** — cluster being deleted (`deleted_time` is set, gen 4):
```
ID                                    NAME                GEN    Available  Reconciled  cl-deployment  cl-namespace
---                                   ---                 ---    ---        ---         ---            ---
019dc049-43a8-7a42-b44a-8d7f89e9e10f  test-cluster-alpha  4 ❌  ● 4(red)   ● 4(red)    ● 4(green)     ● 4(green)
```
(Adapter columns for deleted resources show the `Finalized` condition instead of `Available`.)

### Requirement: NodePool Table View

The CLI SHALL display nodepools in the current cluster as a formatted table when the `--table` flag is passed to `hf nodepool list`.

#### Scenario: Populated nodepool table

- GIVEN nodepools exist in the current cluster
- WHEN the user runs `hf nodepool list --table`
- THEN the CLI MUST:
  1. Fetch all nodepools via GET `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools`
  2. For each nodepool, fetch its adapter statuses via GET `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{id}/statuses`
- AND output a table with:
  - Fixed columns: `ID`, `NAME`, `REPLICAS`, `TYPE`, `GEN`
  - Dynamic condition columns: one per unique `status.conditions[].type` across all nodepools, excluding types ending in `Successful`
  - Dynamic adapter columns: one per unique adapter name across all fetched statuses
  - `REPLICAS` MUST show `spec.replicas`
  - `TYPE` MUST show `spec.platform.type` (or `-` if absent)
- AND dot rendering and the deletion marker MUST follow the rules in "Table Column Architecture" above

**Example** — two nodepools; `workers-1` (gen 2, converged) and `workers-2` (gen 1, no adapters yet):
```
ID                                    NAME      REPLICAS  TYPE           GEN  Available  Reconciled  np-configmap
---                                   ---       ---       ---            ---  ---        ---         ---
019dc049-e79e-72a9-94f8-0056a11193cd  workers-2  1        n2-standard-4  1    ● 1(red)   ● 1(red)    -
019dc049-e76c-7be1-b201-0db50e2c8ecb  workers-1  1        n2-standard-4  2    ● 2(green) ● 2(green)  ● 2(green)
```

### Requirement: Combined Resources Overview

The CLI SHALL display a combined table of all clusters and their nested nodepools.

`hf table` and `hf resources` MUST default to table output format without requiring `--output table`. JSON and YAML remain available via `--output json` and `--output yaml` respectively.

#### Scenario: Display combined resources table (default table format)

- **GIVEN** clusters and nodepools exist
- **WHEN** the user runs `hf table` (no `--output` flag)
- **THEN** the CLI MUST render the full combined table (not JSON)

#### Scenario: Override to JSON

- **GIVEN** clusters exist
- **WHEN** the user runs `hf table --output json`
- **THEN** the CLI MUST output the cluster list as JSON (existing behavior preserved)

#### Scenario: Display combined resources table (existing behavior)

- **GIVEN** clusters and nodepools exist
- **WHEN** the user runs `hf table`
- **THEN** the CLI MUST:
  1. Fetch all clusters
  2. For each cluster, fetch its nodepools and adapter statuses
  3. For each nodepool, fetch its adapter statuses
- **AND** output a table with:
  - Fixed columns: `ID`, `NAME`, `GEN`
  - Dynamic condition columns: union of all `status.conditions[].type` across clusters and nodepools, excluding types ending in `Successful`
  - Dynamic adapter columns: union of all adapter names across all cluster and nodepool statuses
- **AND** cluster rows MUST appear with their full `id` and `name`
- **AND** nodepool rows MUST be indented with two spaces on `id` and `name` to show hierarchy
- **AND** each cluster's nodepools MUST appear immediately after their parent cluster row
- **AND** dot rendering and the deletion marker MUST follow the rules in "Table Column Architecture"

### Requirement: Watch Mode for Table Commands

`hf table --watch`, `hf resources --watch`, `hf cluster list --watch`, and `hf nodepool list --watch` MUST continuously refresh the table output at the configured interval. The refresh interval MUST default to 5 seconds and MAY be changed with `--seconds / -s`.

For commands that use fast-tick rendering (500 ms spinner interval decoupled from the data-fetch interval), the CLI MUST additionally display a live countdown line above the table headers on every render tick, showing the number of seconds remaining until the next data fetch and an animated braille spinner.

#### Scenario: Combined table watch mode — countdown line shown

- GIVEN `hf table --watch` is running with a refresh interval of N seconds
- WHEN the table is rendered (every 500 ms)
- THEN the CLI MUST print a line of the form `↻ Xs  <spinner>` above the table headers
- AND `X` MUST be the ceiling of the number of seconds remaining until the next data fetch (range: 1 to N)
- AND `<spinner>` MUST be the current braille spinner frame, advancing every 500 ms
- AND the line MUST appear flush left, directly above the `ID` column header

#### Scenario: Combined table watch mode — countdown resets after data refresh

- GIVEN `hf table --watch` is running with a refresh interval of N seconds
- WHEN a data fetch completes successfully
- THEN the countdown MUST reset to N on the next render tick

#### Scenario: No countdown line in non-watch mode

- GIVEN the user runs `hf table` without `--watch`
- WHEN the table is rendered once and exits
- THEN the CLI MUST NOT print any `↻` countdown line
- AND the output MUST be byte-for-byte identical to the output before this change

### Requirement: Adapter Activity Indicator

Each adapter column in table output SHALL display a braille spinner character prepended to the cell value when the adapter's `last_report_time` is within `2 × frequency` seconds of the current time, indicating the adapter is actively reporting.

The spinner character advances through the sequence `⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏` (cycling by refresh tick count modulo 10). When the adapter is not active (or `--watch` is not in use), no spinner character is prepended and the cell renders as before.

The activity check computes `time.Since(lastReportTime) < 2 × frequencySecs`. If `last_report_time` is absent or unparseable, the adapter is treated as inactive.

#### Scenario: Active adapter shows spinner

- **WHEN** an adapter's `last_report_time` is within `2 × frequency` seconds of now
- **AND** `--watch` is active
- **THEN** the adapter cell MUST be prefixed with the current spinner frame (e.g., `⠋ ● 3`)

#### Scenario: Inactive adapter shows no spinner

- **WHEN** an adapter's `last_report_time` is older than `2 × frequency` seconds
- **OR** `--watch` is not active
- **THEN** the adapter cell MUST render without a spinner prefix (e.g., `● 3`)

#### Scenario: Missing last_report_time

- **WHEN** an adapter's `last_report_time` is empty or unparseable
- **THEN** the adapter MUST be treated as inactive (no spinner)

