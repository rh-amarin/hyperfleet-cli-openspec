## ADDED Requirements

### Requirement: Watch Mode for Table Commands

The CLI SHALL support a `--watch` flag on `hf cluster list`, `hf nodepool list`, `hf table`, and `hf resources` that causes the table to refresh continuously at a configurable interval.

When `--watch` is active the CLI MUST:
1. Clear the terminal screen using ANSI escape sequences before each render.
2. Re-fetch all data from the API on every tick.
3. Re-render the full table after each fetch.
4. Continue until the user interrupts with SIGINT (Ctrl+C) or SIGTERM.
5. Exit cleanly on interrupt with no partial-line output.

A `-s <seconds>` flag (default `5`) controls the refresh interval. The minimum accepted value is `1`.

#### Scenario: Cluster list watch mode — basic refresh

- **WHEN** the user runs `hf cluster list --output table --watch`
- **THEN** the CLI MUST render the cluster table immediately, then re-render every 5 seconds
- **AND** each render MUST be preceded by a terminal clear

#### Scenario: Cluster list watch mode — custom frequency

- **WHEN** the user runs `hf cluster list --output table --watch -s 10`
- **THEN** the CLI MUST refresh every 10 seconds

#### Scenario: Nodepool list watch mode

- **WHEN** the user runs `hf nodepool list --output table --watch`
- **THEN** the CLI MUST render the nodepool table immediately, then re-render every 5 seconds

#### Scenario: Combined table watch mode

- **WHEN** the user runs `hf table --watch`
- **THEN** the CLI MUST render the combined cluster+nodepool table immediately, then re-render every 5 seconds
- **AND** `--output table` MUST NOT be required (table is the default output for `hf table` / `hf resources`)

#### Scenario: Watch mode — graceful exit

- **WHEN** the user sends SIGINT (Ctrl+C) while `--watch` is active
- **THEN** the CLI MUST exit with code 0 and leave the terminal in a clean state

#### Scenario: Watch mode — API error during refresh

- **WHEN** an API call fails during a watch refresh cycle
- **THEN** the CLI MUST exit with a non-zero code and print the error message

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

## MODIFIED Requirements

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
