## Context

Configuration is already stored in self-contained environment YAML files plus `state.yaml`. The `hf config` group provided show/get/set on top of those files, but `set` is unsafe for structured config (`resource-types`) and the interactive picker duplicates editor workflows. `hf env` already handles create/activate/list/delete/show for named environments.

## Goals / Non-Goals

**Goals:**
- Remove `hf config` entirely.
- Make `hf env show [name]` the single read-only config inspection command.
- Show file paths and an edit hint **after** configuration output (not before).
- Preserve colorized YAML, secret redaction, and active-env state block behavior.

**Non-Goals:**
- Changing the config storage model or precedence chain.
- Adding a new `hf config validate` command.
- Fixing `Store.Set()` round-trip (no CLI setter remains).

## Decisions

### 1. Delete `cmd/config.go` rather than deprecate

**Decision:** Remove the command group in one release. No deprecation period.

**Why:** Small internal user base; `hf env show` is a direct replacement for `show`; file editing replaces `get`/`set`.

### 2. `hf env show [name]` replaces `hf config show`

**Decision:** Optional `name` argument defaults to active environment. Output order: env YAML → state block (if active) → separator → file paths → edit message.

**Why:** Matches user request; paths at the bottom reinforce that files are the source of truth.

### 3. Display helpers in `cmd/env_display.go`

**Decision:** Extract `showEnvProfile`, `formatEnvFileForDisplay`, `redactEnvFileYAML`, and state marshaling from deleted `config.go`.

**Why:** Keeps `env.go` focused on Cobra wiring; helpers are shared by `env show` and the bare `hf env` picker preview.

### 4. Test helpers in `cmd/helpers_test.go`

**Decision:** Move `runCmd`, `makeEnv`, `setActiveEnv`, etc. out of deleted `config_test.go`.

**Why:** Many packages' tests depend on these helpers; they outlive any single command file.

## Risks / Trade-offs

- **Breaking scripts using `hf config set/get`** — Users must edit YAML or use `HF_*` env vars. → Document in README and proposal.
- **Main specs updated during implementation** — Delta specs in the change folder capture the same requirements for archive merge. → Archive will reconcile.

## Open Questions

None.
