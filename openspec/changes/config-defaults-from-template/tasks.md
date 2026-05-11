## 1. Move Template File

- [x] 1.1 Create `internal/config/assets/` directory
- [x] 1.2 Copy `cmd/assets/config-template.yaml` to `internal/config/assets/config-template.yaml`
- [x] 1.3 Delete `cmd/assets/config-template.yaml` and `cmd/assets/` directory

## 2. Update `internal/config`

- [x] 2.1 Add `//go:embed assets/config-template.yaml` directive and `var ConfigTemplateYAML []byte` to `config.go`
- [x] 2.2 Replace the hardcoded `defaults` declaration with `var defaults map[string]map[string]string`
- [x] 2.3 Add `func init()` that unmarshals `ConfigTemplateYAML` into `defaults`, panicking on parse error

## 3. Update `cmd`

- [x] 3.1 Delete `cmd/assets.go` (embed directive moves to `internal/config`)
- [x] 3.2 Delete `cmd/assets_test.go` (`TestConfigTemplateMatchesDefaults` is obsolete)
- [x] 3.3 Update `cmd/config.go` line 249: replace `configTemplateYAML` with `config.ConfigTemplateYAML`

## 4. Unit Tests

- [x] 4.1 In `internal/config/config_test.go`: add test that `Store.Get` returns a template-defined value (e.g. `hyperfleet.api-url`) when no active environment is loaded
- [x] 4.2 In `internal/config/config_test.go`: add test that `ConfigTemplateYAML` is non-empty and parses into a non-empty map with no error

## 5. Verify

- [x] 5.1 `go build ./...` succeeds
- [x] 5.2 `go vet ./...` passes
- [x] 5.3 `go test ./... 2>&1 | tee verification_proof/tests.txt`
- [ ] 5.4 Commit `verification_proof/` to git
