## Context

`hf config`, `hf config set`, and `hf kube port-forward` are among the most frequently used commands. Currently, bare invocations (no subcommand or no arguments) either silently dump YAML or show nothing. The codebase already has `internal/selector` (backed by `go-fuzzyfinder`) used for interactive cluster and nodepool selection — the same infrastructure applies here.

Key existing patterns to reuse:
- `clusterIDSel selector.Selector = selector.FuzzySelector{}` in `cmd/cluster.go` — injectable selector for testability
- `mockSel` in `cmd/cluster_test.go` — shared test double, accessible across all `package cmd` tests
- `cmd.Help()` — Cobra helper that writes to `cmd.OutOrStdout()`, captured by `rootCmd.SetOut` in tests
- `configShowCmd.RunE(cmd, args)` — callable directly from other `RunE` functions without shelling out

## Goals / Non-Goals

**Goals:**
- `hf config` (bare) prints help then active config — one command does both orientation and state inspection
- `hf config set` (bare) enters interactive fuzzy-find mode; 2-arg form continues working and now shows config after set
- `hf kube port-forward` (bare) prints help then port-forward status; if any are down, prompts to start all

**Non-Goals:**
- No changes to the `hf config show`, `hf config get`, or any `hf config env *` commands
- No changes to `hf kube port-forward start/stop/status` subcommands
- No new external dependencies
- No changes to config file schema or API contracts

## Decisions

### 1. `hf config` — `cmd.Help()` then `configShowCmd.RunE`

`configCmd.RunE` calls `cmd.Help()` (writes to `OutOrStdout()`) followed by `configShowCmd.RunE(cmd, args)`. This reuses 100% of the existing show implementation and keeps both outputs going to the same writer — testable via `rootCmd.SetOut`.

Alternative considered: duplicate the show logic inline. Rejected — no benefit, adds maintenance burden.

### 2. `hf config set` interactive — package-level selector var

A package-level `configSetSel selector.Selector = selector.FuzzySelector{}` variable follows the exact pattern of `clusterIDSel` in `cluster.go`. Tests override it with `mockSel{idx: N}` from `cluster_test.go` (shared across the `cmd` package). No interface change needed — `selector.Selector` already exists.

Items are built from `knownKeysForSection` across all sections in canonical order, giving a deterministic first item (`hyperfleet.api-url`) — important for test predictability. Secret keys show `<set>` / `<not set>` as current value, matching the display convention in `resolvedSection`.

### 3. `hf config set` value prompt — `bufio.Scanner(cmd.InOrStdin())`

After fuzzy selection, the value is read from `cmd.InOrStdin()`. In tests, `rootCmd.SetIn(strings.NewReader("value\n"))` controls input without touching os.Stdin. This is the same stdin injection pattern already used by `kubeNamespaceCleanCmd`.

### 4. `hf config set` — show config after every successful set

Both the interactive path and the 2-arg non-interactive path call `configShowCmd.RunE(cmd, nil)` after `s.Set()` succeeds. `configShowCmd.RunE` creates a fresh `config.Store` and loads from disk, so it always reflects the just-written value.

### 5. `hf config set` Args validator — 0 or 2 args

A custom `cobra.PositionalArgs` func returns nil for 0 or 2 args, and calls `cmd.Help()` + blank error for any other count. This mirrors the `helpOnNoArgs` helper already in `root.go`.

### 6. `hf kube port-forward` bare — `pfStartCmd.RunE(cmd, nil)` for start

`portForwardCmd.RunE` calls `pfStartCmd.RunE(portForwardCmd, nil)` directly when the user confirms start. `nil` args causes `servicesForArgs` to return all 4 predefined services (existing behaviour). Passing `portForwardCmd` as `cmd` is safe because `OutOrStdout()` walks up to `rootCmd` — the same writer used throughout.

Alternative considered: extract a shared `startAllPortForwards(cmd, s)` helper. Rejected — the existing `pfStartCmd.RunE` already encapsulates exactly this logic; a helper abstraction would add indirection without value.

## Risks / Trade-offs

- **Fuzzy finder requires a TTY**: `go-fuzzyfinder` returns a non-abort error when no terminal is available. In non-interactive contexts (pipes, CI), `hf config set` with 0 args will error. Mitigation: users in non-interactive contexts should use the 2-arg form; the error message is self-explanatory.
- **`pfStartCmd.RunE` called with different `cmd`**: The start command is invoked with `portForwardCmd` as the cobra.Command receiver. All internal calls (`loadConfig`, `kube.StartPortForward`) are self-contained; only `cmd.OutOrStdout()` is affected, and it correctly resolves to rootCmd's writer. Tested via the existing pattern where port-forward tests set `rootCmd.SetOut`.
- **Double `loadConfig()` call in port-forward bare path**: `portForwardCmd.RunE` calls `loadConfig()` and then `pfStartCmd.RunE` calls it again. This is a minor redundancy; config loading is fast (disk read) and idempotent — not worth refactoring.

## Migration Plan

No migration required. The 2-arg `hf config set <section.key> <value>` form is preserved. The only user-visible behavior changes are additive: help before config, optional interactive mode, config shown after set, and the port-forward status+prompt.
