# Database Operations Specification (delta)

This delta spec implements all requirements from `openspec/specs/database/spec.md` unchanged.
No requirement modifications. Implementation adds:

- `internal/db` package with `Config`, `Pool`, `Query`, `Exec`, `Querier` interface
- `cmd/db.go` with `hf db query`, `hf db delete`, `hf db config` subcommands

See parent spec at `openspec/specs/database/spec.md` for full requirements.
