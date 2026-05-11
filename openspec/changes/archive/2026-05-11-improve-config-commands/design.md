## Context

`hf config show` currently prints only `active-environment` from state, then YAML sections from config.yaml. The state.yaml also holds `cluster-id`, `cluster-name`, and `nodepool-id` — the most commonly inspected operational values — but they are invisible in the output.

`hf config get` takes `<section> <key>` as two positional args. The existing `config/spec.md` already specifies dot notation (`section.key`) for `config set`, so `config get` is inconsistent with the spec and forces the user to know the section hierarchy.

When commands with required args are called with zero args, Cobra emits a raw error string ("accepts N arg(s), received 0"). The `rootCmd` sets `SilenceUsage: true`, so no help is printed — the user gets no actionable guidance.

## Goals / Non-Goals

**Goals:**
- `hf config show` includes a `state:` block showing all non-empty state keys.
- `hf config get <key>` accepts `section.key` (config) or plain `key` (state), matching the `config set` spec convention.
- All user-facing commands with required positional args show Cobra's full help on zero args.

**Non-Goals:**
- Changing the interactive `env create` prompting behaviour.
- Modifying any `internal/` package.
- Adding shell-completion support for the new `get` key format.

## Decisions

### D1 — `helpOnNoArgs(n int)` helper in `cmd/root.go`

A single `cobra.PositionalArgs`-returning helper replaces `cobra.ExactArgs(n)` on all user-facing commands. When `len(args) == 0` it calls `cmd.Help()` (prints full help to stdout) and returns `fmt.Errorf("")`. Because `rootCmd` has `SilenceErrors: true`, the blank error is not printed by Cobra. `main.go` is updated to skip printing blank error strings, so the only output is the help text, and the exit code is 1.

**Alternative considered:** Override `PersistentPreRunE` on `rootCmd` to detect zero args — rejected because `Args` validation fires before `PersistentPreRunE`; by the time `PersistentPreRunE` runs the wrong-arg count has already been accepted.

**Alternative considered:** `os.Exit(0)` inside the helper — rejected because it makes unit testing impossible.

### D2 — `hf config get <key>` with dot-notation split

In `configGetCmd.RunE`, split the single argument on the first `.`:
- If a dot is present → `section = key[:idx]`, `field = key[idx+1:]`, call `s.Get(section, field)`.
- If no dot → call `s.GetState(key)`.

This is unambiguous (state keys never contain dots; config keys always need a section) and consistent with the existing `config set` spec.

**Alternative considered:** Search all sections when no dot — rejected because keys like `host`, `port`, `user`, `password` appear in multiple sections, making the result order-dependent and surprising.

### D3 — `state:` block in `hf config show`

Collect known state keys (`active-environment`, `cluster-id`, `cluster-name`, `nodepool-id`) via `s.GetState(k)`, build a `map[string]string` of non-empty values, and pass it as the first section to the existing `marshalYAMLOrdered` function. No new helpers needed.

The existing `active-environment: <name>` header line is replaced by this block, keeping the output as valid YAML throughout.

## Risks / Trade-offs

- **BREAKING change for `config get`**: scripts using `hf config get hyperfleet api-url` must be updated to `hf config get hyperfleet.api-url`. Mitigated by clear release notes; the two-arg form now returns a Cobra arg-count error (still non-zero exit), not silent data.
- **Hidden daemon command**: `kube.go`'s `_daemon` command uses `cobra.ExactArgs(5)` and is an internal subprocess, not user-facing. It is deliberately left unchanged to avoid unintended help output during programmatic invocation.
- **`completion.go`**: uses `cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)` — not converted to `helpOnNoArgs` because `OnlyValidArgs` must be preserved; the completion command always receives exactly one arg from the shell.

## Migration Plan

1. Implement changes on a feature branch.
2. Run `go test ./...` and `go vet ./...` — must pass.
3. Save test output to `verification_proof/`.
4. Open PR; no live-cluster verification required (no API calls involved).

## Open Questions

_(none)_
