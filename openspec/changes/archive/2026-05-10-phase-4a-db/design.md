## Overview

Phase 4a adds native PostgreSQL access to the HyperFleet CLI. The implementation is split
into two layers: `internal/db` (the pgxpool wrapper) and `cmd/db.go` (the Cobra command tree).

## Package: internal/db

**File:** `internal/db/db.go`

### Config

```go
type Config struct {
    Host     string
    Port     string
    Name     string
    User     string
    Password string
}
```

### NewFromConfig

Reads `database.*` keys from the config store:

```go
func NewFromConfig(s *config.Store) *Config
```

### DSN

`postgres://<user>:<password>@<host>:<port>/<name>`

### Pool

Opens a pgxpool connection pool. Returns `(*pgxpool.Pool, error)`.

```go
func (c *Config) Pool(ctx context.Context) (*pgxpool.Pool, error)
```

### Querier Interface

Defined for testability; allows cmd/db_test.go to inject a mock:

```go
type Querier interface {
    Query(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) ([]string, [][]string, error)
    Exec(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (int64, error)
}
```

Since `Query` and `Exec` are standalone functions (not methods on Config), the interface is
satisfied by a thin adapter type `defaultQuerier` that delegates to the package-level functions.

### Query function

```go
func Query(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (headers []string, rows [][]string, error)
```

- Executes `pool.Query(ctx, sql, args...)`
- Extracts column names from `rows.FieldDescriptions()` as headers (in query order)
- For each row, scans all columns as `interface{}` and converts to string:
  - `nil` → `"NULL"`
  - `[]byte` → string conversion
  - everything else → `fmt.Sprintf("%v", v)`
- Truncates any field value exceeding 80 chars to 79 chars + `…`

### Exec function

```go
func Exec(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (int64, error)
```

- Calls `pool.Exec(ctx, sql, args...)`
- Returns `commandTag.RowsAffected()`

## Package: cmd

**File:** `cmd/db.go`

All subcommands load the DB config via `db.NewFromConfig(cfgStore)` (where `cfgStore` is
set by root's `PersistentPreRunE`). The root's active-env guard handles the environment check.

### hf db query <sql>

- Accepts inline SQL as the first positional argument OR `-f <filepath>` flag
- Connects via `dbCfg.Pool(ctx)` then calls `db.Query(ctx, pool, sql)`
- 0 rows: prints `[INFO] No rows returned.` and exits 0
- Otherwise renders via `output.Printer`:
  - `table`: renders with tabwriter; headers uppercase; columns in query order
  - `json`: JSON array of objects keyed by column name
  - `yaml`: YAML of the same structure
- Errors: `fmt.Errorf("[ERROR] %v", err)`

### hf db exec <sql>

- Accepts inline SQL as the first positional argument
- Connects via `dbCfg.Pool(ctx)` then calls `db.Exec(ctx, pool, sql)`
- On success: prints `Rows affected: <n>`
- Errors: `fmt.Errorf("[ERROR] %v", err)`

### hf db delete <target>

- Target is required: one of `clusters`, `nodepools`, `adapter_statuses`, `ALL`
- Shell completions provided via `ValidArgs`
- Unknown target: `[ERROR] Unknown target '<t>'. Valid targets are: clusters, nodepools, adapter_statuses, ALL.`
- Shows row count per table, prompts for `yes` confirmation
- Confirmation denied: prints `Aborted` and exits 0
- Execution order for ALL: `adapter_statuses` → `node_pools` → `clusters`
- Per-table error: prints `[ERROR] Failed to delete from <table>: <err>`, continues
- Output is always plain text; `--output` flag does not apply

### hf db config

- No DB connection needed
- Prints host, port, name, user as plain values
- Password: `<set>` if non-empty, `<not set>` if empty
- Output is always plain text; `--output` flag does not apply

## Testing Strategy

**internal/db/db_test.go** (no real DB needed):
- DSN construction correctness
- NULL rendering in Query result
- Truncation at 80 chars

**cmd/db_test.go** (mock Querier, httptest-style in-process):
- `query`: table output headers/rows, JSON output, 0-row info message, file-read error (`-f`)
- `exec`: rows-affected output
- `delete`: unknown target error, confirmation denied
- `config`: password masking (`<set>` vs `<not set>`)

Integration tests (tagged `//go:build integration`) for real DB operations.

## Key Decisions

- **Interface for testability**: The `Querier` interface + `defaultQuerier` adapter allows
  command tests to swap in a mock without touching the pgxpool code.
- **No subprocess**: All DB operations use pgxpool natively, no `psql` shells.
- **Truncation in Query, not output**: Truncation at 80 chars happens in `internal/db.Query`
  so the table renderer doesn't need special casing.
