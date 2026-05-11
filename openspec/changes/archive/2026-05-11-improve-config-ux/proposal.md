## Why

The current `hf config` surface has several UX problems: the configuration model uses a cascade of `config.yaml` + sparse environment overrides (making it hard to reason about what's actually active), `set` and `get` use inconsistent argument formats, and the command group lacks a clear first-run flow. The goal is a simpler, more predictable configuration experience where each environment is fully self-contained.

## What Changes

- **BREAKING** — Remove `config.yaml`. Configuration no longer cascades from a base file. Each environment profile is complete and self-contained.
- **BREAKING** — `hf config set` syntax changes from `set <section> <key> <value>` to `set <section.key> <value>` (symmetric with `get`).
- `hf config` with no subcommand now runs `show` instead of displaying help.
- `hf config show` errors with actionable guidance when no active environment is set.
- `hf config env create <name>` replaces the old `create`/`new` command: no flags, no prompts — copies a repo-bundled template, activates the environment, and prints the file path for the user to edit.
- Remove `hf config doctor`.
- Add `cmd/assets/config-template.yaml` as the canonical default environment file embedded in the binary.

## Capabilities

### New Capabilities

- `config-template`: A static YAML template file (`cmd/assets/config-template.yaml`) embedded in the binary, containing all config sections with their default values. Used by `env create` to seed new environment files.

### Modified Capabilities

- `config-model`: The storage model changes from a three-layer cascade (defaults → config.yaml → env profile) to a two-layer model (defaults → active env file). `config.yaml` is eliminated. `Set()` writes to the active env file.
- `config`: Command surface changes — `set` syntax, `env create` behavior, `hf config` default action, `show` no-env error, removal of `doctor`.

## Impact

- `internal/config/config.go`: rewrite `Load()`, `Get()`, `Set()`; remove config.yaml logic; update `ActivateEnvironment` to also write state immediately after creating file.
- `cmd/config.go`: update `configSetCmd`, `configCmd`, `configShowCmd`, `configEnvCreateCmd`; remove `configDoctorCmd` and all env create flags.
- `cmd/assets/config-template.yaml`: new file (embedded via `//go:embed`).
- All commands that call `s.Get()` continue to work unchanged — the interface is the same, only the backing storage changes.
- Tests in `cmd/config_test.go` and `internal/config/config_test.go` must be updated.
