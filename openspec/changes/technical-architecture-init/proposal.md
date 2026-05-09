## Why

The HyperFleet CLI (`hf`) needs a well-defined Go module structure to replace a suite of fragile bash scripts with a single self-contained binary. Establishing the architectural skeleton — module layout, Cobra command tree, shared internal packages — is the prerequisite for every other feature change.

## What Changes

- Initialize `go.mod` with module `github.com/rh-amarin/hyperfleet-cli` at Go 1.22
- Create `main.go` as the entry point that delegates to the Cobra root command
- Create `cmd/root.go` defining the root `hf` command with all global persistent flags (`--config`, `--output`, `--no-color`, `--verbose`/`-v`, `--api-url`, `--api-token`)
- Stub `cmd/` files for each command group: cluster, nodepool, adapter, config, db, maestro, pubsub, rabbitmq, kube, logs, repos, resources
- Create `internal/version/` package providing build-time version info
- Add a top-level `Makefile` with `build`, `test`, and `vet` targets
- Wire `hf version` and `hf completion` as built-in commands

## Capabilities

### New Capabilities

- `technical-architecture`: Go module scaffold, Cobra command tree, global flags, and internal package skeleton that all other capabilities build upon

### Modified Capabilities

(none — this is the foundational scaffold; no existing spec requirements are changing)

## Impact

- All subsequent `cmd/` and `internal/` packages depend on this scaffold existing
- `go.mod` determines dependency versions for cobra, client-go, pgx, pubsub, yaml.v3
- Global flags defined in `cmd/root.go` propagate to every subcommand via Cobra's persistent-flag mechanism
- The `Makefile` targets are used by CI and the Definition of Done verification steps

## Testing Scope

| Package | Test Cases Needed |
|---|---|
| `cmd/` (root) | `hf --help` exits 0; global flags are registered; unknown flag returns error |
| `internal/version` | `Version()` returns a non-empty string; build-time override via ldflags |

## Verification Steps Requiring Live Cluster Access

- None at the scaffold stage — all tests are unit tests that do not require network access.
