## Context

`technical-architecture-init` left the module with a working root command but no subcommands and no `internal/` packages. Before any feature branch can implement real behaviour, every command must be discoverable (so `hf --help` shows the full tree) and the module must compile and vet cleanly.

## Goals / Non-Goals

**Goals:**
- Register every domain command group with `rootCmd` so the full command tree is visible
- Provide `internal/version` so the version command has a well-typed source of truth
- Produce a `Makefile` that standardises the three developer build commands
- Achieve zero errors from `go build ./...`, `go vet ./...`, and `go test ./...`

**Non-Goals:**
- Implementing any real business logic (that belongs in subsequent per-domain changes)
- Adding new external Go module dependencies beyond cobra (already present)
- Integrating with the real HyperFleet API

## Decisions

### 1 — One file per command group in `cmd/`

Each domain stub (`cluster.go`, `nodepool.go`, etc.) defines a top-level cobra.Command and registers it in its `init()` function. This matches the existing `root.go` pattern and keeps future diffs scoped to a single file per domain.

**Alternative considered**: a single `cmd/commands.go` file. Rejected because it would create merge conflicts when multiple feature branches touch different domains simultaneously.

### 2 — Stubs print help; no `RunE`

Stubs set neither `Run` nor `RunE` so that invoking them directly (`hf cluster`) prints cobra's built-in help. This is the standard cobra pattern for group commands that only contain subcommands.

### 3 — `internal/version` is the single source of version truth

`var Version = "dev"` is declared in `internal/version/version.go`. The `Makefile` injects the git tag at link time via `-ldflags "-X github.com/rh-amarin/hyperfleet-cli/internal/version.Version=$(git describe ...)"`. This avoids duplicating the version string in `main.go` or any command file.

### 4 — Makefile targets mirror CLAUDE.md conventions

`make build` → `go build -o bin/hf .`  
`make test` → `go test ./...`  
`make vet` → `go vet ./...`

No additional tools (goreleaser, mage, etc.) are introduced at this stage.

## Risks / Trade-offs

- [Risk] Stub commands without subcommands registered look empty in `--help`. → Mitigation: acceptable for scaffold; real subcommands will be registered in subsequent changes.
- [Risk] `go.sum` may need an update if any indirect dependency hash changed. → Mitigation: `go mod tidy` is run during verification.

## Migration Plan

1. Write all artifact files in this change.
2. Run `go build ./...`, `go vet ./...`, `go test ./...` — all must pass.
3. Save verification proof.
4. Archive change.

## Open Questions

None.
