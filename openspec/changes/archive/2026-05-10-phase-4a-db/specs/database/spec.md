# Database Operations — Phase 4a Delta

## ADDED Requirements

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
