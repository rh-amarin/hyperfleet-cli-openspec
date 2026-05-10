## Why

The HyperFleet CLI has its `internal/config` package fully built but no `hf config` commands to expose it. Users cannot manage environments, read or write config values, or verify API connectivity from the CLI — forcing them to edit YAML files manually.

## What Changes

- `cmd/root.go`: Add `RequireActiveEnvironment()` guard to `PersistentPreRunE`, bypassed for `config env *`, `version`, `completion`, and `help` commands.
- `cmd/config.go`: Implement the full `hf config` command tree — `show`, `get`, `set`, and the `env` subcommand family (`list`, `create`, `activate`, `delete`, `show`) plus `doctor`.
- `cmd/config_test.go`: Unit tests for all config commands (file I/O via `t.TempDir()`, HTTP via `httptest.NewServer`).

## Capabilities

### New Capabilities

- `config-commands`: Full `hf config` command tree: show/get/set, env CRUD, and connectivity doctor.
- `active-env-guard`: Root `PersistentPreRunE` guard that fails with a clear error when no active environment is set, bypassed for always-available commands.

### Modified Capabilities

- `config`: The existing config spec documents the `show` and `set` behaviors plus the env subcommands and the active-env guard scenarios.

## Impact

- `cmd/config.go` — rewrite (currently a stub)
- `cmd/root.go` — add guard logic to `PersistentPreRunE`
- No new external dependencies; all implementation uses `internal/config`, `internal/api`, `internal/output`, and `net/http`

## Testing Scope

- `cmd/config_test.go`: white-box tests using package `cmd` (internal)
  - `hf config show` (requires active env, renders YAML)
  - `hf config get` / `hf config set` (key lookup, unknown key error)
  - `hf config env list` (table output)
  - `hf config env create` / `activate` / `delete` / `show`
  - `hf config doctor` (uses `httptest.NewServer` — reachable and unreachable cases)
  - Active-env guard: bypass for `config env *`, `version`, `completion`; fires for `config show`, `config set`
