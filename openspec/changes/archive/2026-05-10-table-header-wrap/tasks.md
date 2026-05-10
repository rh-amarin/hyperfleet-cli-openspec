# Tasks: table-header-wrap

## 1. Implementation

- [x] 1.1 Add `wrapHeader(s string, maxWidth int) []string` to `internal/output/printer.go` — split on underscores greedily, hard-break single tokens at maxWidth
- [x] 1.2 Modify `PrintTable` in `internal/output/printer.go` — uppercase headers, apply `wrapHeader` with maxWidth=10, compute max line count N, emit N tabwriter rows (padding shorter headers with empty strings), then emit data rows as before

## 2. Tests

- [x] 2.1 Add `TestWrapHeader_ShortString` — header ≤ 10 chars returns a single unchanged element
- [x] 2.2 Add `TestWrapHeader_ExactlyTen` — header of exactly 10 chars returns a single unchanged element
- [x] 2.3 Add `TestWrapHeader_UnderscoreSplit` — `"CLUSTER_NAME"` → `["CLUSTER", "NAME"]`
- [x] 2.4 Add `TestWrapHeader_MultiSegment` — `"OBSERVED_GENERATION"` splits correctly across two lines
- [x] 2.5 Add `TestWrapHeader_LongSingleToken` — `"PROVISIONING"` hard-breaks at 10 chars
- [x] 2.6 Add `TestWrapHeader_GreedyPack` — short tokens are packed together before wrapping (e.g., `"A_B_TOOLONG_X"`)
- [x] 2.7 Add `TestPrintTable_LongHeader` — full `PrintTable` output contains multi-line header rows; data rows remain aligned
- [x] 2.8 Add `TestPrintTable_ShortHeaders` — existing short-header table output is unchanged

## 3. Verification

- [x] 3.1 `go build ./...` passes — save output to `verification_proof/build.txt`
- [x] 3.2 `go vet ./...` passes — save output to `verification_proof/vet.txt`
- [x] 3.3 `go test ./...` passes — save output to `verification_proof/test.txt`
- [x] 3.4 Commit `verification_proof/` files
