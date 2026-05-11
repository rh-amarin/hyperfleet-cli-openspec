## Context

`hf config show` (active environment path) already exposes the file path when called with a named argument (`hf config show <env-name>`). The same information is absent from the default no-argument output. The path is computed as `filepath.Join(s.dir, "environments", active+".yaml")` in several places inside `internal/config/config.go` but is not exposed as a public method.

## Goals / Non-Goals

**Goals:**
- Print the active environment file path in `hf config` / `hf config show` (no-argument form).
- Keep the change minimal: one new method, one new output line.

**Non-Goals:**
- Changing `hf config show <env-name>` (already shows the path).
- Adding color or special formatting to the path line.
- Verifying that the file exists at runtime (it must exist if an environment is active).

## Decisions

**Add `EnvFilePath(name string) string` to `internal/config`.**

The path formula (`filepath.Join(s.dir, "environments", name+".yaml")`) is already duplicated four times in `config.go`. Exposing it as a method removes duplication and gives `cmd/config.go` a clean way to obtain the path without re-implementing the formula.

Alternative: inline the path formula in `cmd/config.go`. Rejected — it would create a fifth copy of the same formula and couple the command layer to the config directory layout.

**Print as a comment-style line below the environment name header.**

Format: `# ~/.config/hf/environments/e2e.yaml`  
Placed immediately after the `environment: <name>` header line (if any) and before the YAML sections. This matches the pattern already used by `hf config show <env-name>` which prints the path before the values.

Alternative: emit it as part of the YAML output under `state:`. Rejected — the path is metadata about the file, not a runtime state value.

## Risks / Trade-offs

- [Risk] Path printed but file deleted externally → path appears valid but is stale. Mitigation: none needed; `hf config set` and other writers will surface the real error at write time.
