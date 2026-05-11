# Tasks: table-ansi-alignment

## 1. Implementation

- [x] 1.1 Add `stripANSI(s string) string` to `internal/output/printer.go` — rune-by-rune state machine that discards ANSI escape sequences
- [x] 1.2 Add `displayWidth(s string) int` — returns `utf8.RuneCountInString(stripANSI(s))`
- [x] 1.3 Rewrite `PrintTable` in `internal/output/printer.go` — remove `tabwriter`; compute `colWidths` from display widths across all rows; render with explicit per-cell padding; swap `"text/tabwriter"` import for `"unicode/utf8"`

## 2. Tests

- [x] 2.1 Add `TestPrintTable_ANSIAlignment` — inject raw ANSI codes in a data cell; assert STATUS header starts at the same byte offset as the ANSI prefix in the data row

## 3. Verification

- [x] 3.1 `go build ./...` passes — save to `verification_proof/build.txt`
- [x] 3.2 `go vet ./...` passes — save to `verification_proof/vet.txt`
- [x] 3.3 `go test ./...` passes — save to `verification_proof/test.txt`
- [x] 3.4 Commit `verification_proof/` files
