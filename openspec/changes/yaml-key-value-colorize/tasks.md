## 1. Rewrite `ColorizeYAMLSections`

- [x] 1.1 Add ANSI constants `ansiBoldWhite`, `ansiGreen`, `ansiDim` to `internal/output/config_display.go` (keep existing `ansiBoldCyan` and `ansiResetBold`/`ansiReset`)
- [x] 1.2 Rewrite `ColorizeYAMLSections` with a line classifier:
  - Top-level section header (`^[a-zA-Z][\w-]*:\s*$`) → bold cyan (unchanged)
  - Indented `key: value` line → key in bold white + dim `: ` + value in green (or dim if sentinel `<…>`)
  - Indented `key:` line with no value → key in bold white + dim `:`
  - All other lines → unchanged
- [x] 1.3 Remove the now-internal `isSectionHeader` helper if it is inlined, or keep it; ensure `isWordChar` is retained if still needed

## 2. Unit Tests

- [x] 2.1 Update `internal/output/config_display_test.go`: existing `noColor=true` test still passes (no ANSI for any line type)
- [x] 2.2 Add test: indented `  api-url: http://example.com` with `noColor=false` → key segment contains `\033[1;37m`, value segment contains `\033[32m`
- [x] 2.3 Add test: indented `  token: <not set>` with `noColor=false` → value segment contains `\033[2m` (dim), not `\033[32m`
- [x] 2.4 Add test: indented `  context:` (no value) with `noColor=false` → key segment contains `\033[1;37m`, no green code present
- [x] 2.5 Existing `SectionSeparator` tests continue to pass unchanged

## 3. Verify

- [x] 3.1 `go build ./...` succeeds
- [x] 3.2 `go vet ./...` passes
- [x] 3.3 `go test ./... 2>&1 | tee openspec/changes/yaml-key-value-colorize/verification_proof/tests.txt`
- [x] 3.4 Run `hf config` against http://34.175.170.44:8000/ and save raw output (with ANSI) to `openspec/changes/yaml-key-value-colorize/verification_proof/live.txt`
- [ ] 3.5 Commit everything and push to main
