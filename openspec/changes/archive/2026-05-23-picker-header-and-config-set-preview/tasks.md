# Tasks: picker-header-and-config-set-preview

## Implementation

- [x] 1. `internal/selector/selector.go`: add `header string` to `PreviewSelector.SelectWithPreview`; update `FuzzyPreviewSelector` to use `WithHeader(header)` when non-empty
- [x] 2. `internal/selector/selector_test.go`: add compile-time check `var _ PreviewSelector = FuzzyPreviewSelector{}`
- [x] 3. `cmd/env.go`: add header constant; pass it to `envSel.SelectWithPreview`
- [x] 4. `cmd/config.go`: change `configSetSel` type to `selector.PreviewSelector`; add `renderConfigPreview`; update `configSetInteractive` to use `SelectWithPreview` with preview + header
- [x] 5. `cmd/env_test.go`: update `mockPreviewSel.SelectWithPreview` signature (add `_ string`)
- [x] 6. `cmd/config_test.go`: replace `mockSel{idx:0}` with `mockPreviewSel{idx:0}` in `TestConfigSet_Interactive`

## Verification

- [x] 7. `go build ./...` — no errors
- [x] 8. `go vet ./...` — no warnings
- [x] 9. `go test ./cmd/... ./internal/selector/...` — all tests pass; save to `verification_proof/`
