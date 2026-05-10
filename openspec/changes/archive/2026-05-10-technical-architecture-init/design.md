## Context

The HyperFleet CLI currently exists as a collection of bash scripts. This change establishes the Go module skeleton that every subsequent implementation change will build upon. There is no existing Go code; we are starting from zero.

Key constraints from the spec:
- Module name: `github.com/rh-amarin/hyperfleet-cli`
- Binary name: `hf`
- Go 1.22+ required (for type parameters in `internal/api`)
- CLI framework: `github.com/spf13/cobra`
- No external tool dependencies for core operations

## Goals / Non-Goals

**Goals:**
- Produce a compilable `go.mod` and `main.go` that builds a working `hf` binary
- Define `cmd/root.go` with all global persistent flags specified in the spec
- Create stub `cmd/<domain>.go` files that register empty command groups with the root
- Create `internal/version/` with a build-time overridable version string
- Provide a `Makefile` with `build`, `test`, and `vet` targets
- Ensure `go build ./...` and `go vet ./...` pass with zero errors

**Non-Goals:**
- Implementing any actual command logic (cluster, nodepool, db, etc.) — those are separate changes
- Implementing `internal/api`, `internal/config`, `internal/output`, or any other internal package — those are separate changes
- Writing integration or live-cluster tests — not required at scaffold stage

## Decisions

### Decision 1: Single `main.go` → `cmd.Execute()` pattern

**Choice:** `main.go` contains only `func main() { cmd.Execute() }` (or equivalent), delegating entirely to the Cobra tree defined in `cmd/`.

**Rationale:** Keeps `main.go` trivially small; all testable logic lives in `cmd/` or `internal/`. Standard pattern for Cobra-based CLIs.

**Alternative considered:** Putting flag parsing in `main.go`. Rejected — makes `main.go` harder to test and contradicts the Cobra pattern.

### Decision 2: `PersistentPreRunE` for global flag wiring

**Choice:** Global flags (`--output`, `--verbose`, `--no-color`, etc.) are read into package-level variables in `cmd/root.go` via `PersistentPreRunE`, not via `init()` calls.

**Rationale:** `PersistentPreRunE` runs after flag parsing and before each command's `RunE`, giving a single consistent injection point. Package-level vars (`var outputFormat string`) are set once per invocation and readable by all `cmd/` files without passing contexts.

**Alternative considered:** Using Viper for flag-to-config binding. Rejected — adds dependency complexity; the spec calls for a custom config package (`internal/config`), not Viper.

### Decision 3: Stub `cmd/<domain>.go` files with no-op subcommands

**Choice:** Every domain file (`cluster.go`, `nodepool.go`, etc.) registers its top-level command with `rootCmd.AddCommand(...)` in an `init()` function, with a placeholder `RunE` that returns `nil`.

**Rationale:** Keeps the binary compilable and shows the full command tree in `--help` output from day one. Subsequent changes fill in the real `RunE` bodies without touching the wiring.

**Alternative considered:** Creating all domain files in one PR with full implementations. Rejected — violates the incremental change model; each domain is its own change.

### Decision 4: `internal/version` with ldflags override

**Choice:** `internal/version/version.go` exports `var Version = "dev"` which is overridden at build time via `-ldflags "-X github.com/rh-amarin/hyperfleet-cli/internal/version.Version=$(git describe --tags)"`.

**Rationale:** Standard Go pattern; keeps version info out of `main.go` and testable. The `Makefile` `build` target sets this automatically.

### Decision 5: Minimal `go.mod` with only cobra as an initial dependency

**Choice:** `go.mod` starts with only `github.com/spf13/cobra` as a direct dependency. All other dependencies (client-go, pgx, pubsub, etc.) are added by the changes that implement those packages.

**Rationale:** Avoids pulling in thousands of transitive dependencies before any code actually uses them. Keeps the initial `go.sum` small and the build fast.

## Risks / Trade-offs

- [Risk] Stub `RunE` returning nil means `hf cluster` prints no output rather than an error → Mitigation: stubs should call `cmd.Help()` so the user sees usage information
- [Risk] Adding all domain `init()` calls at once creates a wide surface area → Mitigation: each domain file is self-contained; adding or removing a file has no effect on other files
- [Risk] `go.mod` minimum Go version set too low → Mitigation: pin to 1.22 explicitly to guarantee type-parameter support for `internal/api`

## Migration Plan

1. Create `go.mod` and `main.go` (Task 1-2)
2. Create `cmd/root.go` with global flags (Task 3)
3. Create stub `cmd/<domain>.go` files (Task 4)
4. Create `internal/version/version.go` (Task 5)
5. Add `Makefile` (Task 6)
6. Verify compilation and tests pass

No rollback needed — this is a net-new module; there is no prior Go code to revert.

## Open Questions

- None — the spec fully defines the required package structure and command tree.
