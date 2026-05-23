# Tasks: hf-env-top-level

## Implementation

- [x] 1. `internal/selector/selector.go`: Add `PreviewSelector` interface + `FuzzyPreviewSelector` struct with `SelectWithPreview`
- [x] 2. New `cmd/env.go`: `var envSel selector.PreviewSelector = selector.FuzzyPreviewSelector{}`; `envCmd` with picker `RunE`; 5 subcommands (`envListCmd`, `envCreateCmd`, `envActivateCmd`, `envDeleteCmd`, `envShowCmd`); `showEnvProfile` helper; `init()` wiring
- [x] 3. `cmd/config.go`: Delete `configEnvCmd`, `configEnvListCmd`, `configEnvCreateCmd`, `configEnvActivateCmd`, `configEnvDeleteCmd`, `configEnvShowCmd` vars and their `init()` entries; delete `showEnvProfile`; remove unused import (`path/filepath`)
- [x] 4. `cmd/root.go`: `isBypassCommand` — change `strings.Contains(path, "config env")` to `strings.HasPrefix(path, "hf env")`; update stale error message
- [x] 5. `cmd/config_test.go`: All `runCmd(t, dir, "config", "env", ...)` → `runCmd(t, dir, "env", ...)`; update bypass test lines; update `"hf config env create"` string → `"hf env create"`
- [x] 6. New `cmd/env_test.go`: `TestEnvPickerNoArgs_NoEnvironments`, `TestEnvPickerNoArgs_ActivatesAndShowsConfig`, `TestEnvPickerNoArgs_Abort` using `mockPreviewSel`; subcommand smoke tests

## Verification

- [x] 7. `go build ./...` — no errors
- [x] 8. `go vet ./...` — no warnings
- [x] 9. `go test ./cmd/... ./internal/selector/...` — all tests pass; save to `verification_proof/`
