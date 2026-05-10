## Packages Touched

- `internal/output/printer.go` — replace `tabwriter` in `PrintTable`; add `stripANSI`, `displayWidth`

No new packages. No new external dependencies. `"text/tabwriter"` import removed; `"unicode/utf8"` added.

## Key Decisions

### Why tabwriter fails with ANSI codes

`tabwriter` pads each cell to a column-width measured in bytes. ANSI SGR sequences like `\033[31m` (5 bytes) and `\033[0m` (4 bytes) are invisible in the terminal but are counted. A colored dot cell `\033[31m●\033[0m 1` is 14 bytes — tabwriter allocates 14+2=16 bytes for that column, but only 3 characters are visible (`● 1`). The header `READY` (5 bytes) also gets padded to 16 bytes, rendering as 16 visible chars. Every column after a colored cell is offset by the hidden ANSI bytes.

### Replacement approach

Pre-compute column widths using display widths across all rows, then render each cell followed by explicit padding:

```
colWidth[c] = max(displayWidth(cell)) for all cells in column c
padding     = colWidth[c] - displayWidth(cell) + 2  (2-space separator, matches old tabwriter padding)
```

`displayWidth(s)` = `utf8.RuneCountInString(stripANSI(s))`, which strips escape sequences then counts Unicode code points. This correctly handles:
- ANSI SGR sequences (`\033[31m`, `\033[0m`, etc.)
- Multi-byte Unicode characters like `●` (3 bytes, 1 rune, display width 1)

The last column gets no trailing padding (same as tabwriter behavior).

### stripANSI state machine

Simple rune-by-rune walk: on `\033` set `inEsc = true`; while `inEsc` discard runes until a letter terminates the sequence; otherwise copy rune to output. Handles all CSI sequences used in this codebase (`\033[Nm` and `\033[0m`).

### No API change

`WrapHeader` remains exported. `stripANSI` and `displayWidth` are private — tested indirectly through `PrintTable`.

## Test Plan

- `TestPrintTable_ANSIAlignment` — inject raw ANSI codes in a data cell; verify that STATUS column header starts at the same byte offset as the ANSI prefix in the data row (proving the two-column separator is consistent despite ANSI inflation)
- All existing `TestPrintTable_*` and `TestWrapHeader_*` tests must continue to pass
