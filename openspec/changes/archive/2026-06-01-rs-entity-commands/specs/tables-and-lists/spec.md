## MODIFIED Requirements

### Requirement: Combined Resources Overview

The CLI SHALL display a combined table of all clusters and their nested nodepools via `hf rs` (no subcommand). The former `hf resources` and `hf table` root commands are deprecated and removed.

`hf rs` defaults to table output when `--output` is not set for the overview command. Pass `--output json` or `--output yaml` for structured output.

#### Scenario: Display combined resources JSON

- **GIVEN** clusters exist
- **WHEN** the user runs `hf rs --output json`
- **THEN** the CLI MUST output the cluster list as JSON (or equivalent structured overview payload)
- AND JSON output MUST NOT require per-nodepool adapter fetches unless documented otherwise

#### Scenario: Display combined resources table

- **GIVEN** clusters and nodepools exist
- **WHEN** the user runs `hf rs` or `hf rs --output table`
- **THEN** the CLI MUST:
  1. Fetch all clusters
  2. For each cluster, fetch its nodepools and adapter statuses
  3. For each nodepool, fetch its adapter statuses
- **AND** output a table with:
  - Fixed columns: `ID`, `NAME`, `GEN`
  - Dynamic condition columns: union of all `status.conditions[].type` across clusters and nodepools, excluding types ending in `Successful`
  - Dynamic adapter columns: union of all adapter names across all cluster and nodepool statuses
- **AND** cluster rows MUST appear with their full `id` and `name`
- **AND** nodepool rows MUST use hierarchical tree prefixes on `id` and `name` (ASCII tree, not only space indent)
- **AND** each cluster's nodepools MUST appear immediately after their parent cluster row
- **AND** dot rendering and the deletion marker MUST follow the rules in "Table Column Architecture"

#### Scenario: hf table and hf resources removed

- **GIVEN** this change is complete
- **WHEN** the user runs `hf table` or `hf resources`
- **THEN** the CLI MUST NOT register those commands
- **AND** operators MUST use `hf rs` for the combined view

#### Scenario: JSON output skips per-resource fetching

- **GIVEN** the user runs `hf rs --output json` for overview
- **THEN** the CLI MAY output the raw clusters list JSON without nodepool/adapter fetches
- AND table/watch modes MUST fetch nested data as required for columns

### Requirement: Watch Mode for Table Commands

The CLI SHALL support a `--watch` flag on `hf rs` (overview), `hf rs <entity> list`, and former-equivalent entity list commands that causes the table to refresh continuously at a configurable interval.

When `--watch` is active the CLI MUST:
1. Clear the terminal screen using ANSI escape sequences before each render.
2. Re-fetch all data from the API on every tick.
3. Re-render the full table after each fetch.
4. Continue until the user interrupts with SIGINT (Ctrl+C) or SIGTERM.
5. Exit cleanly on interrupt with no partial-line output.

A `-s <seconds>` flag (default `5`) controls the refresh interval. The minimum accepted value is `1`.

#### Scenario: Cluster list watch mode — basic refresh

- **WHEN** the user runs `hf rs clusters list --output table --watch`
- **THEN** the CLI MUST render the cluster table immediately, then re-render every 5 seconds
- **AND** each render MUST be preceded by a terminal clear

#### Scenario: Cluster list watch mode — custom frequency

- **WHEN** the user runs `hf rs clusters list --output table --watch -s 10`
- **THEN** the CLI MUST refresh every 10 seconds

#### Scenario: Nodepool list watch mode

- **WHEN** the user runs `hf rs nodepools list --output table --watch`
- **THEN** the CLI MUST render the nodepool table immediately, then re-render every 5 seconds

#### Scenario: Combined table watch mode

- **WHEN** the user runs `hf rs --watch`
- **THEN** the CLI MUST render the combined cluster+nodepool table immediately, then re-render every 5 seconds

#### Scenario: Watch mode — graceful exit

- **WHEN** the user sends SIGINT (Ctrl+C) while `--watch` is active
- **THEN** the CLI MUST exit with code 0 and leave the terminal in a clean state

#### Scenario: Watch mode — API error during refresh

- **WHEN** an API call fails during a watch refresh cycle on entity list commands
- **THEN** the CLI MUST exit with a non-zero code and print the error message
- **AND** overview refresh MUST tolerate partial failures per overview requirements (warnings, not hard fail)

## ADDED Requirements

### Requirement: RS Entity Table Views

The CLI SHALL support rich table output for `hf rs clusters` and `hf rs nodepools` via `list --output table` and `table`, using the same Table Column Architecture as former cluster and nodepool list tables.

#### Scenario: RS cluster table matches former cluster list table

- GIVEN the same cluster data
- WHEN the user runs `hf rs clusters table`
- THEN column sets (fixed, condition, adapter) and dot rendering MUST match the former `hf cluster list --output table` behavior

#### Scenario: RS nodepool table matches former nodepool list table

- GIVEN the same nodepool data under the same cluster
- WHEN the user runs `hf rs nodepools table`
- THEN column sets and dot rendering MUST match the former `hf nodepool list --output table` behavior
