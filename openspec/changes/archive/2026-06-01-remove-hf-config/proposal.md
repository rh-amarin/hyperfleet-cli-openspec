## Why

The `hf config` command group duplicates what users can do more safely by editing YAML files directly. `hf config set` rewrites environment files as flat maps (stripping structured sections like `resource-types`), while `hf config get/set/show` add maintenance burden for little value now that environments are self-contained files and `hf env` handles lifecycle.

## What Changes

- **BREAKING** — Remove the entire `hf config` command group (`show`, `get`, `set`, and bare `hf config`).
- **BREAKING** — Configuration changes are made by editing `~/.config/hf/environments/<name>.yaml` and `~/.config/hf/state.yaml` directly.
- Enhance `hf env show [name]` to display the active environment when `name` is omitted (replacing `hf config show`).
- After configuration output, print absolute paths to the environment file and `state.yaml`, plus a message directing the user to edit those files.
- Move display helpers from deleted `cmd/config.go` to `cmd/env_display.go`.
- Update error messages and docs that referenced `hf config` to use `hf env`.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `config`: Remove `hf config` commands; consolidate read-only display into `hf env show [name]` with file-path footer.
- `config-model`: Remove CLI `Set` requirement; add `StateFilePath()` helper.
- `command-hierarchy`: Remove `config.go` from command tree; document `hf env` in Cobra tree.
- `config-template`: Template seeding referenced via `hf env create` (not `cmd/config.go`).

## Impact

- `cmd/config.go` — deleted.
- `cmd/env_display.go` — new; YAML display, secret redaction, state block, file-path footer.
- `cmd/env.go` — `env show` accepts optional name; bare `hf env` picker calls `showEnvProfile`.
- `cmd/config_test.go` — deleted; helpers moved to `cmd/helpers_test.go`; env show tests in `cmd/env_test.go`.
- `internal/config/config.go` — `StateFilePath()`; error messages use `hf env create`.
- `internal/api/errors.go` — HTML hint uses `hf env show`.
- `README.md` — quickstart and command reference updated.

## Testing Scope

- `cmd/env_test.go`: `hf env show` with/without name, state block, secret redaction, file paths after config, edit message, no-active-env error, not-found error.
- `cmd/env_test.go`: bare `hf env` picker still activates and displays config.
- `cmd/resource_test.go`: resource-types visible in `hf env show` output.
- `cmd/helpers_test.go`: shared `runCmd` / `makeEnv` helpers (no new logic).
- `internal/config/config_test.go`: unchanged (Store API still tested).

Live verification: run `hf env show` against an existing active environment and confirm colorized output, state block, and file-path footer.
