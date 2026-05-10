# Tasks: table-header-wrap

## 1. Implementation

- [ ] 1.1 Add `wrapHeader(s string, maxWidth int) []string` to `internal/output/printer.go` — split on underscores greedily, hard-break single tokens at maxWidth
- [ ] 1.2 Modify `PrintTable` in `internal/output/printer.go` — uppercase headers, apply `wrapHeader` with maxWidth=10, compute max line count N, emit N tabwriter rows (padding shorter headers with empty strings), then emit data rows as before

## 2. Tests

- [ ] 2.1 Add `TestWrapHeader_ShortString` — header ≤ 10 chars returns a single unchanged element
- [ ] 2.2 Add `TestWrapHeader_ExactlyTen` — header of exactly 10 chars returns a single unchanged element
- [ ] 2.3 Add `TestWrapHeader_UnderscoreSplit` — `"CLUSTER_NAME"` → `["CLUSTER", "NAME"]`
- [ ] 2.4 Add `TestWrapHeader_MultiSegment` — `"OBSERVED_GENERATION"` splits correctly across two lines
- [ ] 2.5 Add `TestWrapHeader_LongSingleToken` — `"PROVISIONING"` hard-breaks at 10 chars
- [ ] 2.6 Add `TestWrapHeader_GreedyPack` — short tokens are packed together before wrapping (e.g., `"A_B_TOOLONG_X"`)
- [ ] 2.7 Add `TestPrintTable_LongHeader` — full `PrintTable` output contains multi-line header rows; data rows remain aligned
- [ ] 2.8 Add `TestPrintTable_ShortHeaders` — existing short-header table output is unchanged

## 3. Verification

- [ ] 3.1 `go build ./...` passes — save output to `verification_proof/build.txt`
- [ ] 3.2 `go vet ./...` passes — save output to `verification_proof/vet.txt`
- [ ] 3.3 `go test ./...` passes — save output to `verification_proof/test.txt`
- [ ] 3.4 Commit `verification_proof/` files
