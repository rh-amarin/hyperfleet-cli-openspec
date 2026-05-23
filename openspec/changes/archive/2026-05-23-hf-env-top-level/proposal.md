## Why

`hf config env` is buried two levels deep. Every time a developer switches clusters or environments they type six tokens. There is also no quick way to see all environments and their full configuration without running multiple commands. `hf env` as a top-level command with an interactive fuzzy-picker addresses both problems.

## What Changes

### New: `hf env` top-level command

- **Bare `hf env`**:
  - If no environments exist: prints the command help and a hint to run `hf env create <name>`.
  - If one or more environments exist: launches a split-screen fuzzy picker — left panel shows the list of environment names (active one marked with ✓), right panel shows the full YAML of the highlighted environment with syntax coloring. Selecting an environment activates it and then shows the full active config (same as `hf config show`).
- **Subcommands**: `hf env create`, `hf env activate`, `hf env list` (alias `ls`), `hf env delete` (alias `rm`), `hf env show` — identical behavior to the current `hf config env` equivalents.

### Removed: `hf config env`

`hf config env` and all its subcommands are removed. The config group retains: `show`, `get`, `set`.

## Capabilities

### New Capabilities

_(none — functionality already existed under `hf config env`)_

### Modified Capabilities

- `config`: `hf config env` group is removed; `hf config` subcommands are now `show`, `get`, `set` only.
- `environment-management`: env commands are now at the top level (`hf env *`); bare `hf env` is a new interactive picker.

## Impact

- `internal/selector/selector.go`: new `PreviewSelector` interface + `FuzzyPreviewSelector` implementation using `go-fuzzyfinder`'s `WithPreviewWindow`.
- New `cmd/env.go`: `envCmd`, injectable `envSel selector.PreviewSelector`, 5 subcommands, picker `RunE`, `showEnvProfile` helper.
- `cmd/config.go`: remove `configEnvCmd` and all 5 subcommand vars; remove from `init()`; move `showEnvProfile` to `env.go`.
- `cmd/root.go`: update `isBypassCommand` to bypass `hf env *` instead of `config env *`.
- `cmd/config_test.go`: update all `config env` invocations to `env`; update bypass test.
- New `cmd/env_test.go`: picker tests (no envs, activate+show, abort).
