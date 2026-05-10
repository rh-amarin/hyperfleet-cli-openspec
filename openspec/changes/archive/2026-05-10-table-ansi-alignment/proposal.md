## Why

`PrintTable` uses `text/tabwriter` to align columns. `tabwriter` measures column widths in bytes. ANSI escape sequences inflate the byte length of colored cells without adding visible characters — a red dot `\033[31m●\033[0m 1` is 14 bytes but only 3 visible characters. This causes data columns to appear shifted right relative to their headers in any terminal that renders colors.

The bug was exposed when the `table-header-wrap` change made table output common enough that the misalignment became obvious on a real cluster.

## What Changes

`PrintTable` replaces `tabwriter` with manual column-width calculation using display widths (ANSI-stripped rune count). Two private helpers are added: `stripANSI` and `displayWidth`. Column widths are computed across all rows (header and data) using display widths, then each row is rendered with explicit per-cell padding.

No behavior change for plain (no-color) output. Colored table output is now visually aligned.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `output-formatting`: Table output requirement gains ANSI-aware column alignment.
