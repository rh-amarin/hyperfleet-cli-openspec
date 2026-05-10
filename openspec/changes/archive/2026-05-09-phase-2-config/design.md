## Context

Phase 1 delivered `internal/config` (Store, Load, Get, Set, ListEnvironments, ActivateEnvironment, RequireActiveEnvironment). Phase 2 wires this into Cobra commands so users can manage their config from the CLI. The `internal/api` client, `internal/output` Printer, and `internal/resource` types are all available.

## Goals / Non-Goals

**Goals:**
- Expose `hf config show/get/set` for reading and writing `config.yaml`.
- Expose `hf config env list/create/activate/delete/show` for managing environment profiles.
- Expose `hf config doctor` for API connectivity validation.
- Add a `PersistentPreRunE` guard in `cmd/root.go` that gates all commands on an active environment (with well-defined bypasses).
- Full unit-test coverage via `cmd/config_test.go`.

**Non-Goals:**
- Interactive prompts for `hf config env create` (flags only — no terminal prompt library).
- `hf config env new` prompt flow described in the config spec (that is a separate interactive-UX capability).
- Implementing commands from other domains (cluster, nodepool, db, etc.).

## Decisions

### D1: Active-env guard placement

Guard lives in `rootCmd.PersistentPreRunE`. It inspects `cmd.CommandPath()` (the full path string like `"hf config env activate"`) and skips the check when the path contains `"config env"`, or the leaf command is `"version"`, `"completion"`, or `"help"`.

**Why not per-command RunE?** A single guard in PersistentPreRunE is DRY and ensures new commands are guarded by default.

### D2: Config Store initialisation

Each command that needs the store calls `config.NewFromEnv()` and `s.Load()` inline in its `RunE`. No global singleton — this keeps tests hermetic (each test sets `HF_CONFIG_DIR` via `t.TempDir()`).

### D3: `hf config show` output

Renders `config.yaml` sections as YAML (via `internal/output` Printer). Secrets (`token`, `database.password`, `rabbitmq.password`) are redacted to `<set>` or `<not set>`. The command requires an active environment (subject to the guard).

### D4: `hf config set` argument format

Accepts `<section> <key> <value>` as three positional args (matching the instruction spec). The valid sections are hardcoded as the canonical list. Unknown sections return `[ERROR] Unknown config section '<section>'`.

### D5: `hf config doctor` timeout

Uses a separate `http.Client` with a **5-second timeout** (not the 30s default). Hits `GET <api-url>/healthz` (or just the base URL) and reports reachable/unreachable.

### D6: `hf config env create` flags

Flags: `--api-url`, `--api-token`, `--cluster-id`, `--nodepool-id`. The environment file is created at `~/.config/hf/environments/<name>.yaml`. If the file exists, return `[ERROR] Environment '<name>' already exists`.

### D7: Test isolation

All cmd tests set `HF_CONFIG_DIR` to `t.TempDir()` so they never touch the real `~/.config/hf/`. Commands initialise the store from the env var via `config.NewFromEnv()`.

## Risks / Trade-offs

- [Risk] PersistentPreRunE bypass list could miss new bypass-worthy commands → Mitigation: the bypass list is explicit and documented; add new entries as new commands are added.
- [Risk] Doctor endpoint path unknown → Use `GET /healthz` with fallback to `GET /` and accept any 2xx response as "reachable".

## Migration Plan

No migrations required — this is additive. The stub `cmd/config.go` is replaced entirely.
