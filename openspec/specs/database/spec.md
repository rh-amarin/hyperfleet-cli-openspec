# Database Operations Specification

## Purpose

Provide CLI commands for direct PostgreSQL database operations against the HyperFleet
database, using native Go (`pgxpool`) instead of `psql` subprocesses or `kubectl exec`.

## Config Keys

| Key | Default |
|---|---|
| `database.host` | `localhost` |
| `database.port` | `5432` |
| `database.name` | `hyperfleet` |
| `database.user` | `hyperfleet` |
| `database.password` | `foobar-bizz-buzz` |

DSN format: `postgres://<user>:<password>@<host>:<port>/<name>`
## Requirements
### Requirement: Execute SQL Query

The CLI SHALL execute arbitrary SQL queries against the HyperFleet PostgreSQL database.

#### Scenario: Query with inline SQL

- GIVEN database connection config is set (database.host/port/name/user/password)
- WHEN the user runs `hf db query "<SQL>"`
- THEN the CLI MUST connect via pgxpool using the resolved DSN
- AND execute the query natively (no subprocess)
- AND output results as a formatted table

#### Scenario: Query from file

- GIVEN database connection config is set
- WHEN the user runs `hf db query -f <filepath>`
- THEN the CLI MUST read the SQL from the specified file
- AND execute it natively against the database
- AND output results as a formatted table
- AND exit with code 1 and an `[ERROR]` message if the file cannot be read

#### Scenario: Query output format

- GIVEN `hf db query` returns rows
- THEN the default output format MUST be table (rendered with tabwriter, columns in query order)
- AND `--output json` MUST output rows as a JSON array of objects keyed by column name
- AND `--output yaml` MUST output the same data as YAML

#### Scenario: Query result rendering

- WHEN the query returns rows
- THEN the CLI MUST render results as a formatted table with:
  - Column headers in uppercase, in the order returned by the query
  - `NULL` values displayed as the literal string `NULL` (not blank)
  - Fields exceeding 80 characters truncated with a trailing `…`
  - Aligned columns using tabwriter

#### Scenario: Query returns no rows

- WHEN the query returns 0 rows
- THEN the CLI MUST print `[INFO] No rows returned.` and exit 0

#### Scenario: List database tables

- GIVEN the database is accessible
- WHEN the user runs `hf db query "SELECT table_name FROM information_schema.tables WHERE table_schema='public'"`
- THEN the output MUST list the tables: `migrations`, `adapter_statuses`, `clusters`, `node_pools`

### Requirement: Delete Records

The CLI SHALL delete all records from a specified table, or from all tables, with a confirmation prompt.

The `<target>` argument is required and MUST be one of: `clusters`, `nodepools`, `adapter_statuses`, or `ALL`.
The argument values MUST be offered as shell completions.

#### Scenario: Unknown delete target

- GIVEN the user provides an unrecognized target name
- WHEN the user runs `hf db delete <unknown>`
- THEN the CLI MUST display `[ERROR] Unknown target '<unknown>'. Valid targets are: clusters, nodepools, adapter_statuses, ALL.`
- AND exit with code 1

#### Scenario: Delete all records from a single table

- GIVEN database connection is configured
- WHEN the user runs `hf db delete clusters`
- THEN the CLI MUST show the total row count for that table
- AND prompt the user for confirmation (requiring `yes`)
- AND run `DELETE FROM clusters` only after confirmation
- AND print the count of deleted rows

- WHEN the user runs `hf db delete nodepools`
- THEN the same behavior applies for the `node_pools` table

- WHEN the user runs `hf db delete adapter_statuses`
- THEN the same behavior applies for the `adapter_statuses` table

#### Scenario: Delete all records from all tables

- GIVEN database connection is configured
- WHEN the user runs `hf db delete ALL`
- THEN the CLI MUST show the row count for each table
- AND prompt the user for confirmation (requiring `yes`)
- AND delete in dependency order: `adapter_statuses` first, then `node_pools`, then `clusters`
- AND print the count of deleted rows per table
- AND if a DELETE fails for one table, the CLI MUST display `[ERROR] Failed to delete from <table>: <error>`, continue to the next table, and report the final row counts for each table that succeeded

#### Scenario: Confirmation denied

- WHEN the user does not confirm (any input other than `yes`)
- THEN the CLI MUST print "Aborted" and exit 0 without deleting anything

### Requirement: Database Configuration Display

The CLI SHALL display the resolved database connection parameters.

#### Scenario: Show DB config

- GIVEN the CLI is running
- WHEN the user runs `hf db config`
- THEN the CLI MUST print host, port, name, user as plain values
- AND mask the password as `<set>` or `<not set>`
- AND require no database connection to run
- AND output format is always plain text; `--output` flag does not apply to this command

#### Scenario: Delete output format

- GIVEN `hf db delete` completes
- THEN the output MUST be plain text (row count lines per table)
- AND `--output` flag does not apply to this command

### Requirement: CLI Database Query Command

The CLI SHALL implement `hf db query` to execute arbitrary SQL against the HyperFleet PostgreSQL database.

#### Scenario: Query with inline SQL

- GIVEN database connection config is set (database.host/port/name/user/password)
- WHEN the user runs `hf db query "<SQL>"`
- THEN the CLI MUST connect via pgxpool using the resolved DSN
- AND execute the query natively (no subprocess)
- AND output results as a formatted table by default

#### Scenario: Query from file

- GIVEN database connection config is set
- WHEN the user runs `hf db query -f <filepath>`
- THEN the CLI MUST read the SQL from the specified file
- AND execute it natively against the database
- AND exit with code 1 and an `[ERROR]` message if the file cannot be read

#### Scenario: Query output format

- GIVEN `hf db query` returns rows
- THEN the default output format MUST be table (rendered with tabwriter, columns in query order)
- AND `--output json` MUST output rows as a JSON array of objects keyed by column name
- AND `--output yaml` MUST output the same data as YAML

#### Scenario: Query returns no rows

- WHEN the query returns 0 rows
- THEN the CLI MUST print `[INFO] No rows returned.` and exit 0

### Requirement: CLI Database Exec Command

The CLI SHALL implement `hf db exec` to execute DML SQL statements.

#### Scenario: Execute DML SQL

- GIVEN database connection config is set
- WHEN the user runs `hf db exec "<SQL>"`
- THEN the CLI MUST execute the SQL natively
- AND print `Rows affected: <n>` on success

### Requirement: CLI Database Delete Command

The CLI SHALL implement `hf db delete` to delete records with confirmation.

#### Scenario: Unknown delete target

- GIVEN the user provides an unrecognized target name
- WHEN the user runs `hf db delete <unknown>`
- THEN the CLI MUST display `[ERROR] Unknown target '<unknown>'. Valid targets are: clusters, nodepools, adapter_statuses, ALL.`
- AND exit with code 1

#### Scenario: Delete with confirmation

- GIVEN database connection is configured
- WHEN the user runs `hf db delete clusters` (or nodepools, adapter_statuses, ALL)
- THEN the CLI MUST show the row count and prompt for confirmation
- AND delete only after the user types `yes`
- AND print "Aborted" if confirmation denied

#### Scenario: Delete ALL in dependency order

- WHEN the user runs `hf db delete ALL`
- THEN the CLI MUST delete in order: adapter_statuses, node_pools, clusters

### Requirement: CLI Database Config Display

The CLI SHALL implement `hf db config` to display resolved connection parameters.

#### Scenario: Show DB config

- WHEN the user runs `hf db config`
- THEN the CLI MUST print host, port, name, user
- AND mask the password as `<set>` or `<not set>`
- AND require no database connection to run

