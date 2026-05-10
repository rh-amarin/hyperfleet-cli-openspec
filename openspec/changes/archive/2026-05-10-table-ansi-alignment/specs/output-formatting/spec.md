# Output Formatting — Delta Spec

This spec is a delta against `openspec/specs/output-formatting/spec.md`.

## MODIFIED Requirements

### Requirement: Multi-Format Output Dispatch

#### Scenario: Table output (MODIFIED — ANSI-aware column alignment)

- GIVEN `--output table` is set
- WHEN a command calls `Printer.PrintTable(headers, rows)`
- THEN the output MUST be rendered with aligned columns
- AND column widths MUST be computed from visible display widths (ANSI escape codes excluded)
- AND colored data cells MUST be visually aligned under their column headers in an ANSI terminal

## ADDED Requirements

### Requirement: ANSI-Aware Table Column Widths

The table renderer SHALL compute column widths from display widths, not byte lengths.

#### Scenario: Colored cell does not inflate column width

- GIVEN a data cell contains ANSI escape sequences (e.g., `\033[31m●\033[0m`)
- WHEN `PrintTable` computes the column width for that cell
- THEN the width MUST be the number of visible characters after stripping ANSI sequences
- AND subsequent columns MUST start at the same visual position in both header and data rows
