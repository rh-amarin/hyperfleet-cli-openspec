## Why

Running `hf config` or `hf kube port-forward` with no arguments gives the user either raw YAML output or nothing — no orientation, no hint of available subcommands. Improving these bare invocations makes the CLI self-documenting and removes the friction of needing a separate `--help` invocation to understand what's available.

## What Changes

- `hf config` (no args) prints the Cobra help block first, then shows the active configuration — the same YAML output as `hf config show`, preceded by usage and subcommand list.
- `hf config set` (no args) launches an interactive fuzzy-finder over all known `section.key` pairs with their current values, prompts for the new value on selection, sets it, then displays the full config. The two-argument non-interactive form (`hf config set <section.key> <value>`) continues to work and also displays the full config after a successful set.
- `hf kube port-forward` (no args) prints the Cobra help block, then runs the equivalent of `hf kube port-forward status`. If any port-forward is not connected, the user is prompted with `[y/N]` to run `hf kube port-forward start`.

## Capabilities

### New Capabilities

_(none — all changes are delta modifications to existing capability specs)_

### Modified Capabilities

- `config`: New scenarios for `hf config` bare invocation (help before config) and `hf config set` interactive mode (fuzzy-pick → prompt → set → show full config). Existing non-interactive set and show scenarios are unchanged.
- `kubernetes`: New scenario for `hf kube port-forward` bare invocation (help + status + conditional start prompt).

## Impact

- `cmd/config.go`: `configCmd.RunE` updated; `configSetSel selector.Selector` package-level var added for testability; `configSetCmd` gets updated `Args` validator (0 or 2 args), interactive path in `RunE`, and displays full config after every successful set.
- `cmd/kube.go`: `portForwardCmd` gains a `RunE`.
- `cmd/config_test.go`: New tests for help-before-config, config-after-set, and interactive set (using the `mockSel` test double already defined in `cluster_test.go`).
- `cmd/kube_test.go`: New test for bare `port-forward` invocation with no tracked port-forwards.
- Dependencies: `github.com/ktr0731/go-fuzzyfinder` already a direct dependency via `internal/selector` — no new deps.
- No API, wire format, or config-file schema changes.

## Testing Scope

| File | New Test Cases |
|---|---|
| `cmd/config_test.go` | `TestConfigNoArgs_ShowsHelpBeforeConfig`, `TestConfigSet_ShowsConfigAfterSet`, `TestConfigSet_Interactive` |
| `cmd/kube_test.go` | `TestPortForwardNoArgs_NoForwards` |

Live cluster verification: not required — all changes are local UX only (no new API calls; port-forward start path is integration-only and covered by existing manual verification).
