## 1. Remove hf config command group

- [x] 1.1 Delete `cmd/config.go` (configCmd, configShowCmd, configGetCmd, configSetCmd, interactive set)
- [x] 1.2 Remove `configCmd` registration from root command tree
- [x] 1.3 Move display helpers to `cmd/env_display.go` (`showEnvProfile`, redaction, state block, file-path footer)

## 2. Enhance hf env show

- [x] 2.1 Change `env show` to accept optional `[name]` (`cobra.MaximumNArgs(1)`)
- [x] 2.2 Default to active environment when name omitted; error with guidance if none active
- [x] 2.3 Print environment file path and state file path after all YAML output
- [x] 2.4 Print edit message: "Edit these files to change configuration and runtime state."
- [x] 2.5 Mark active env file path with `[active]` suffix when applicable
- [x] 2.6 Update bare `hf env` picker to call `showEnvProfile` after activation

## 3. Internal package updates

- [x] 3.1 Add `Store.StateFilePath()` in `internal/config/config.go`
- [x] 3.2 Update no-active-environment error messages to reference `hf env create`
- [x] 3.3 Update `internal/api/errors.go` HTML hint to reference `hf env show`
- [x] 3.4 Update `cmd/ui.go` error messages to reference `hf env`

## 4. Tests

- [x] 4.1 Create `cmd/helpers_test.go` with shared `runCmd`, `makeEnv`, `setActiveEnv` helpers
- [x] 4.2 Delete `cmd/config_test.go`; remove config show/get/set tests
- [x] 4.3 Add/update `cmd/env_test.go` tests for new show behavior (paths after config, edit message, optional name, state block)
- [x] 4.4 Update `cmd/resource_test.go` to use `hf env show` instead of `hf config show`

## 5. Documentation

- [x] 5.1 Update `README.md` quickstart and command reference (remove hf config, document hf env show)
- [x] 5.2 Update delta specs under `openspec/changes/remove-hf-config/specs/`

## 6. Verification

- [x] 6.1 Run `go build ./...` and `go vet ./...` — zero errors
- [x] 6.2 Run `go test ./...` — zero failures; save output to `verification_proof/go-test.txt`
- [x] 6.3 Save `go vet` output to `verification_proof/go-vet.txt`
- [x] 6.4 Live verification: run `hf env show` and save output to `verification_proof/live.txt`
