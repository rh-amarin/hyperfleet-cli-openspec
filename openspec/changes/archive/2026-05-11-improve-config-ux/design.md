## Context

The current config system has three layers: built-in defaults → `config.yaml` → active env profile. Environment files only contain overrides. This cascade is hard to reason about — users cannot tell what their active configuration actually is without understanding the merge rules. The `config set` command writes to `config.yaml`, while env profiles are sparse, creating an implicit shared base that leaks between environments.

## Goals / Non-Goals

**Goals:**
- Each environment file is fully self-contained — no cascade from `config.yaml`
- `hf config` with no args shows the active configuration
- `hf config set section.key value` is symmetric with `get`
- `hf config env create <name>` creates a complete, ready-to-edit env file from a bundled template and activates it immediately
- Remove `doctor` command

**Non-Goals:**
- Interactive prompting during env creation
- Migration of existing `config.yaml` data
- Changes to state management (`state.yaml`, cluster-id, nodepool-id)

## Decisions

### 1. Eliminate `config.yaml` entirely

**Decision:** Remove `config.yaml` from the storage model. The precedence chain becomes: CLI flags → HF_* env vars → active env file → built-in defaults.

**Why:** Sparse env overrides on top of a shared `config.yaml` mean two environments can silently share state. A self-contained file per environment is easier to inspect, copy, and reason about.

**Alternative considered:** Keep `config.yaml` as a global base, make env files opt-in complete. Rejected — it preserves the confusion we're trying to eliminate.

### 2. `config set` writes to the active env file

**Decision:** `Set(section, key, value)` resolves the active env file path from `state.yaml` and writes there. Errors if no active environment.

**Why:** Consistent with the new model — there is no `config.yaml` to write to.

### 3. Template file embedded in the binary via `//go:embed`

**Decision:** `cmd/assets/config-template.yaml` is a regular file in the repo, embedded at compile time using `//go:embed`. `env create` copies its bytes directly to the new env file path.

**Why:** Keeps the defaults in one place, is human-readable in the repo, and requires no runtime file system access. Alternative (hardcoding defaults in Go) makes the template harder to read and update.

### 4. `env create` activates the new environment immediately

**Decision:** After writing the env file, `env create` calls `SetState("active-environment", name)` automatically.

**Why:** The next logical action after creating an environment is always to use it. Making the user run `activate` separately is unnecessary ceremony for the common path.

### 5. `configCmd.RunE` delegates to `configShowCmd`

**Decision:** Add a `RunE` to `configCmd` that calls the same logic as `configShowCmd`.

**Why:** Cobra's default for a command group with no args is to print help. Running `show` instead matches the expectation that `hf config` means "show me my config."

## Risks / Trade-offs

- **Breaking change for existing users** — `config.yaml` is silently ignored after this change. Users with existing `config.yaml` settings must create an env file. There is no migration. → Accepted; this is an internal developer tool with a small user base.
- **`config set` requires active env** — Scripts that run `hf config set` without an active environment will now fail. → Users must call `hf config env create` first, which is the correct precondition.
- **Template drift** — If built-in `defaults` in `config.go` diverge from `config-template.yaml`, users will see different behavior between a fresh env and the defaults used by `Get()`. → Mitigated by keeping `defaults` and the template in sync; a test will assert this.

## Open Questions

None. Scope is fully defined.
