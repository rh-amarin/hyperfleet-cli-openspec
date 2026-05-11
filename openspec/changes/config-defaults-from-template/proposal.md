## Why

`internal/config/config.go` contains a hardcoded `defaults` map whose values duplicate those already defined in `cmd/assets/config-template.yaml`. Any addition or change to a config key must be made in two places, and a separate test (`TestConfigTemplateMatchesDefaults`) exists solely to detect when they drift apart. This change makes `config-template.yaml` the single authoritative source of default values and removes the duplication.

## What Changes

- Remove the hardcoded `defaults` map from `internal/config/config.go`
- Move `config-template.yaml` from `cmd/assets/` into `internal/config/assets/` so it can be embedded directly by the config package
- Embed the template in `internal/config` and parse it at `init()` time to populate defaults
- Export `ConfigTemplateYAML []byte` from `internal/config` so `cmd` can continue using it to seed new environment files
- Remove `cmd/assets.go` (embed moves to `internal/config`)
- Remove `cmd/assets_test.go` (`TestConfigTemplateMatchesDefaults` is no longer needed — drift is structurally impossible)

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `config-template`: the file location changes from `cmd/assets/config-template.yaml` to `internal/config/assets/config-template.yaml`; the embed is now in `internal/config` and the bytes are exported; the drift-detection test requirement is removed

## Testing Scope

- `internal/config`: add/update unit tests verifying that `Store.Get` returns template-defined defaults when no active environment is loaded, and that `ConfigTemplateYAML` is non-empty and parses cleanly
- No live cluster access required; all verification is build and unit-test only

## Impact

| File | Change |
|---|---|
| `internal/config/config.go` | Remove `defaults` map; add `//go:embed`; add `init()` to parse template |
| `internal/config/assets/config-template.yaml` | New location (moved from `cmd/assets/`) |
| `internal/config/config_test.go` | Add/update tests for template-driven defaults |
| `cmd/assets.go` | Remove (embed moves to `internal/config`) |
| `cmd/assets_test.go` | Remove (`TestConfigTemplateMatchesDefaults` is obsolete) |
| `cmd/config.go` | Update env-file seeding to use `config.ConfigTemplateYAML` |
