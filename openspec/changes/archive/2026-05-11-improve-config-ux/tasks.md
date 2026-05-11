## 1. Template Asset

- [x] 1.1 Create `cmd/assets/config-template.yaml` with all sections and default values
- [x] 1.2 Add `//go:embed assets/config-template.yaml` to `cmd/config.go` (or a dedicated embed file in `cmd/`)

## 2. Rewrite `internal/config/config.go`

- [x] 2.1 Remove `s.config` field, `deepMergeConfig`, and all `config.yaml` loading/writing logic from `Load()`
- [x] 2.2 Update `Load()` to only create `state.yaml` and `environments/` dir; do NOT create `config.yaml`
- [x] 2.3 Update `Get()` precedence: HF_* env vars → active env file (`s.profile`) → built-in `defaults`
- [x] 2.4 Update `Set(section, key, value)` to resolve the active env file path from state and write there; return error if no active environment
- [x] 2.5 Remove `CountOverrides()` (no longer applicable to complete env files)

## 3. Update `cmd/config.go` — command surface

- [x] 3.1 Add `RunE` to `configCmd` that delegates to `configShowCmd` logic when no subcommand is given
- [x] 3.2 Update `configShowCmd`: when no active environment, print the no-active-env error with guidance (`→ run 'hf config env create <name>'`) and exit 1
- [x] 3.3 Update `configSetCmd`: change args from `<section> <key> <value>` (3 args) to `<section.key> <value>` (2 args, dotted notation); validate key contains `.`; validate section is known
- [x] 3.4 Remove `configDoctorCmd` and its registration in `init()`
- [x] 3.5 Rewrite `configEnvCreateCmd`: remove all flags (`--api-url`, `--api-token`, `--cluster-id`, `--nodepool-id`), require exactly one arg (`<name>`), copy embedded template bytes to env file, call `ActivateEnvironment`, print file path
- [x] 3.6 Remove the `new` alias from `configEnvCreateCmd` (command is `create` only)
- [x] 3.7 Remove package-level flag vars `envCreateAPIURL`, `envCreateAPIToken`, `envCreateClusterID`, `envCreateNPID`
- [x] 3.8 Update `configEnvListCmd` no-environments message to reference `hf config env create <name>`

## 4. Update tests

- [x] 4.1 Update `internal/config/config_test.go`: remove all tests that reference `config.yaml`; add tests for new `Load()` (no config.yaml created), `Get()` precedence, `Set()` writing to env file, `Set()` error when no active env
- [x] 4.2 Add a test asserting that every key in `cmd/assets/config-template.yaml` matches the corresponding value in the `defaults` map
- [x] 4.3 Update `cmd/config_test.go`: remove doctor tests; update `set` tests for dotted notation; add tests for `hf config` (no args) → show behavior; add tests for `env create` (creates file, activates, prints path, errors on duplicate)

## 5. Verification

- [x] 5.1 Run `go build ./...` and `go vet ./...` — zero errors
- [x] 5.2 Run `go test ./...` — zero failures; save output to `verification_proof/go-test.txt`
- [x] 5.3 Live verification: run `hf config env create local2`, confirm file created and activated, run `hf config show`, run `hf config set hyperfleet.api-url http://localhost:9000` and verify with `hf config get hyperfleet.api-url`; save output to `verification_proof/live-verification.txt`
