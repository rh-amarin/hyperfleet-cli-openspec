## Why

The `hf config show` output omits the runtime state variables (cluster-id, cluster-name, nodepool-id) that are the most commonly needed operational values, forcing users to check state.yaml manually. The `hf config get` command forces users to know and type both a section and a key as two separate args, despite the spec already specifying `section.key` dot notation for `config set`. Finally, commands with required positional args emit a bare Cobra error ("accepts N arg(s), received 0") instead of useful guidance, contradicting the spirit of the existing usage-message requirements in `errors-and-usage/spec.md`.

## What Changes

- `hf config show` — adds a `state:` block at the bottom of the resolved config output, listing all non-empty state variables (active-environment, cluster-id, cluster-name, nodepool-id).
- `hf config get` — reduced from two args (`<section> <key>`) to one (`<key>`), using `section.key` dot notation for config values (e.g. `hyperfleet.api-url`) or a plain key for state values (e.g. `cluster-id`). **BREAKING** for existing scripts using the two-arg form.
- All commands with required positional args — when called with zero arguments, show the command's full Cobra help text and exit 1, instead of the bare "accepts N arg(s), received 0" message.

## Capabilities

### New Capabilities

_(none — all changes are modifications to existing behaviour)_

### Modified Capabilities

- `config`: `hf config show` now includes a `state:` section; `hf config get` changes to single `section.key` argument.
- `errors-and-usage`: extends the "show usage when arguments are missing" requirement to cover all commands with required positional args, using the full Cobra help text rather than the raw Cobra arg-count error.

## Impact

- `cmd/config.go` — `configShowCmd` and `configGetCmd`
- `cmd/root.go` — new `helpOnNoArgs(n)` helper shared by all commands
- `main.go` — skip printing blank error strings (used by `helpOnNoArgs` to suppress duplicate output)
- `cmd/cluster.go`, `cmd/nodepool.go`, `cmd/pubsub.go`, `cmd/kube.go`, `cmd/db.go` — replace `cobra.ExactArgs(n)` with `helpOnNoArgs(n)` on user-facing commands
- `cmd/config_test.go` — update callers of the old two-arg `config get` form, add tests for state display and state key lookup
- No changes to `internal/` packages; no new dependencies

## Testing Scope

| Package | New test cases |
|---|---|
| `cmd` | `TestConfigShow_StateVariables` — state block appears in show output |
| `cmd` | `TestConfigGet_StateKey` — plain key looks up state value |
| `cmd` | `TestConfigGet_NoArgs_ShowsHelp` — zero args shows help, exits non-zero |
| `cmd` | Updated: `TestActiveEnvGuard_BlocksConfigGet`, `TestConfigGet_Found`, `TestConfigGet_NotFound`, `TestConfigSet_Valid` — new single-arg format |

Live cluster verification required: no (all changes are local config / UX only).
