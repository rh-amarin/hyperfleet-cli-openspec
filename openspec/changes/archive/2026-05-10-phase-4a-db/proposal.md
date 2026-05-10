## Why

The HyperFleet CLI needs native PostgreSQL access so operators can query and manage
database state directly without needing `psql`, `kubectl exec`, or any external tools.
This is Phase 4a of the CLI build-out, adding `hf db` as the database operations domain.

## What Changes

- **New package** `internal/db`: pgxpool wrapper with `Config`, `Pool`, `Query`, `Exec`,
  and a `Querier` interface for testability.
- **New command group** `hf db` with three subcommands:
  - `hf db query <sql>` — execute arbitrary SQL, output as table/JSON/YAML
  - `hf db delete <target>` — delete records from clusters/nodepools/adapter_statuses/ALL
    with confirmation prompt and dependency-order execution
  - `hf db config` — display resolved DB connection parameters (password masked)
- **New dependency**: `github.com/jackc/pgx/v5`

## Capabilities

### New Capabilities

- `database`: Direct PostgreSQL operations via pgxpool — query, delete with confirmation,
  and config display.

### Modified Capabilities

_(none — no existing spec-level behavior changes)_

## Impact

- `go.mod` / `go.sum`: adds `github.com/jackc/pgx/v5` and its transitive dependencies
- New files: `internal/db/db.go`, `internal/db/db_test.go`, `cmd/db.go`, `cmd/db_test.go`
- No existing commands modified; `hf db` registers itself via `init()` on `rootCmd`

## Testing Scope

| Package | Test cases |
|---|---|
| `internal/db` | DSN construction from Config; Query column/row mapping; NULL rendering; truncation at 80 chars; Exec rowsAffected |
| `cmd/db` | `query` table output, JSON output; 0-row info message; file-read error; `delete` unknown target error; confirmation denied (aborted); `config` password masking |

Live cluster verification is required for:
- `hf db query "SELECT table_name FROM information_schema.tables WHERE table_schema='public'"` — must list expected tables
- `hf db config` — must show resolved host/port/name/user
