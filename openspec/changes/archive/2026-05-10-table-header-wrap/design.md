## Packages Touched

- `internal/output/printer.go` — modify `PrintTable`; add `wrapHeader` helper

No new packages. No new external dependencies.

## Key Decisions

### Where to split

A new private function `wrapHeader(s string, maxWidth int) []string` handles the logic:

1. If `len(s) <= maxWidth`, return `[]string{s}` — no wrapping needed.
2. Split `s` by `_`. Greedily accumulate tokens (joined back with `_`) into the current line as long as the result stays ≤ `maxWidth`. When adding the next token would exceed the limit, flush the current line and start a new one.
3. If a single token (no underscores) exceeds `maxWidth`, hard-break it at `maxWidth` character intervals.

This handles the common uppercase snake_case column names (`CLUSTER_ID`, `OBSERVED_GENERATION`, `PROVISIONING_STATE`) cleanly. CamelCase or single-word long headers fall back to hard-breaking.

### How multi-line headers are emitted via tabwriter

`tabwriter` aligns columns by tab stops — it operates on complete rows. A multi-line header is achieved by emitting N separate rows before the data rows, where N = max number of lines across all headers. For headers shorter than N lines, empty strings pad the missing lines. Example for `["ID", "CLUSTER_NAME", "STATUS"]` (maxWidth=10):

```
wrapHeader("ID")           → ["ID"]
wrapHeader("CLUSTER_NAME") → ["CLUSTER", "NAME"]
wrapHeader("STATUS")       → ["STATUS"]
```

Row 1: `"ID\tCLUSTER\tSTATUS\n"`
Row 2: `"\tNAME\t\n"`

tabwriter aligns both rows to the same column widths, resulting in a stacked two-line header over the `CLUSTER_NAME` column.

### maxWidth constant

Hard-coded as 10 at the call site in `PrintTable`. Not exposed as a parameter — the requirement is specific to 10 characters and adding a parameter creates unnecessary complexity.

### No separator line between header and rows

Consistent with existing table rendering (no horizontal rule is drawn today). Adding one is out of scope for this change.

### Backward compatibility

Headers of ≤ 10 characters behave identically to before (single-row header). The only observable change is additional header rows for long column names.

## Test Plan

- Unit tests in `printer_test.go` (package `output_test`):
  - `TestWrapHeader_ShortString` — ≤10 chars, returns single element unchanged
  - `TestWrapHeader_ExactlyTen` — exactly 10 chars, returns single element
  - `TestWrapHeader_UnderscoreSplit` — e.g. `"CLUSTER_NAME"` → `["CLUSTER", "NAME"]`
  - `TestWrapHeader_MultiSegment` — e.g. `"OBSERVED_GENERATION"` → `["OBSERVED", "GENERATION"]`
  - `TestWrapHeader_LongSingleToken` — e.g. `"PROVISIONING"` (12 chars) → `["PROVISIONI", "NG"]`
  - `TestWrapHeader_GreedyPack` — e.g. `"A_B_C"` packs greedily before wrapping
  - `TestPrintTable_LongHeader` — full `PrintTable` call verifying multi-line header rows appear in output
  - `TestPrintTable_ShortHeaders` — confirms short-header tables are unchanged
