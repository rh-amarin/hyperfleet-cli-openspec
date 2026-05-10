# Output Formatting — Delta Spec

This spec is a delta against `openspec/specs/output-formatting/spec.md`.

## MODIFIED Requirements

### Requirement: Multi-Format Output Dispatch

#### Scenario: Table output (MODIFIED — header wrapping added)

- GIVEN `--output table` is set
- WHEN a command calls `Printer.PrintTable(headers, rows)`
- THEN the output MUST be rendered using aligned columns (tab-separated via tabwriter)
- AND headers MUST be displayed in uppercase
- AND any header whose uppercased form exceeds 10 characters MUST be split across multiple stacked header lines
- AND splitting MUST prefer underscore word boundaries; single tokens exceeding 10 characters MUST be hard-broken at 10 characters
- AND all header lines MUST be emitted before the data rows
- AND column alignment MUST be preserved across all header lines and data rows
- AND headers of 10 characters or fewer MUST be unaffected (single header row, as before)

## ADDED Requirements

### Requirement: Table Header Word Wrapping

The table renderer SHALL wrap column headers that exceed 10 characters into multiple stacked lines.

#### Scenario: Short header unchanged

- GIVEN a column header of 10 characters or fewer after uppercasing
- WHEN `PrintTable` renders the header
- THEN the header MUST appear on a single line without modification

#### Scenario: Underscore-split header

- GIVEN a column header longer than 10 characters that contains underscores (e.g., `CLUSTER_NAME`)
- WHEN `PrintTable` renders the header
- THEN each segment between underscores MUST be packed greedily into lines of ≤ 10 characters
- AND segments MUST be joined with `_` when they fit together within the limit
- AND the result MUST be multiple header rows, one per wrapped line

#### Scenario: Single-token long header hard-break

- GIVEN a column header longer than 10 characters with no underscores (e.g., `PROVISIONING`)
- WHEN `PrintTable` renders the header
- THEN the header MUST be hard-broken at every 10 characters into successive header lines

#### Scenario: Multi-column wrapping alignment

- GIVEN a table with a mix of short and long column headers
- WHEN `PrintTable` renders the table
- THEN all columns MUST align across the multi-line header block
- AND short-header columns MUST be padded with empty strings for their missing header lines
