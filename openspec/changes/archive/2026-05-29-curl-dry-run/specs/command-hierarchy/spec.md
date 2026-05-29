## ADDED Requirements

### Requirement: Global Persistent Flags

The root command SHALL expose global flags available on all subcommands.

#### Scenario: Persistent flags registered

- WHEN global flags are registered
- THEN the following persistent flags MUST be available on every command:
  - `--output <format>` / `-o`: output format (`json`, `table`, `yaml`); default varies per command
  - `--no-color`: disable colored output
  - `--verbose` / `-v`: enable verbose/debug logging
  - `--api-url <url>`: override API URL for this invocation
  - `--api-token <token>`: override API token for this invocation
  - `--curl`: print equivalent curl command to stderr and skip API requests (dry-run)
- NOTE: There is no `--force-color` flag. Color is enabled when stdout is a TTY, disabled otherwise.

### Requirement: Curl Dry-Run Command Behavior

When `--curl` is set, commands that call the HyperFleet or Maestro HTTP clients MUST behave as dry-runs.

#### Scenario: List command dry-run

- GIVEN the active environment is configured
- WHEN the user runs `hf cluster list --curl`
- THEN the CLI MUST print the GET curl for the list endpoint to stderr
- AND MUST NOT print cluster data to stdout
- AND MUST exit with code 0

#### Scenario: Watch with curl

- GIVEN the user runs a watch command with both `--watch` and `--curl` (e.g. `hf cluster list --watch --curl`)
- THEN the CLI MUST print the curl for the first fetch to stderr
- AND MUST NOT enter the watch refresh loop
- AND MUST exit with code 0

#### Scenario: Interactive mode incompatible with curl

- GIVEN a command supports `-i` / `--interactive`
- WHEN the user passes both `--curl` and the interactive flag
- THEN the CLI MUST print `[ERROR] --curl cannot be used with interactive mode`
- AND exit with code 1

#### Scenario: No state mutations on dry-run

- GIVEN `--curl` is set
- WHEN a command would normally persist state (e.g. `SetState` after create)
- THEN the CLI MUST NOT mutate `state.yaml` or environment files
- AND MUST exit with code 0 after printing curl (when the command's primary API call is reached)
