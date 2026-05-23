## MODIFIED Requirements

### Requirement: Show Configuration

`hf config` and `hf config show` MUST display the active environment's configuration sections followed by a `state:` block at the bottom listing all non-empty runtime state variables. When writing to an interactive terminal, section header keys MUST be rendered in bold cyan and a separator line MUST appear between the last configuration section and the `state:` block. Color output is suppressed when `--no-color` is set, the `NO_COLOR` environment variable is set, or stdout is not a TTY.

When `hf config` is invoked with **no arguments**, the command MUST print the full Cobra help block (usage, subcommand list, flags) to stdout **before** the configuration output. `hf config show` continues to show only the configuration output without a preceding help block.

#### Scenario: Active environment file path displayed

- **WHEN** the user runs `hf config` or `hf config show` (no argument)
- **THEN** the CLI MUST print the absolute path of the active environment file
- **AND** the path MUST appear before the YAML sections
- **AND** the path MUST be the resolved absolute path of `~/.config/hf/environments/<name>.yaml`

#### Scenario: hf config bare — help precedes config output

- **WHEN** the user runs `hf config` with no arguments
- **AND** an active environment is configured
- **THEN** the CLI MUST print the command help block (including "Usage:" and available subcommands) first
- **AND** MUST then print the full active configuration output (same as `hf config show`)
- **AND** the help block MUST appear before the environment file path line

#### Scenario: Section headers colorized on TTY

- **WHEN** the user runs `hf config` and stdout is an interactive terminal
- **AND** `--no-color` is not set and `NO_COLOR` is not in the environment
- **THEN** top-level YAML section keys (e.g. `hyperfleet:`, `database:`, `state:`) MUST be rendered in bold cyan
- **AND** value lines MUST remain uncolored

#### Scenario: Separator between config and state sections

- **WHEN** the user runs `hf config` and the output contains both configuration sections and a `state:` block
- **THEN** a separator line (`────────────────────────────────────────`) MUST be printed between the last configuration section and the `state:` block

#### Scenario: Color suppressed — --no-color

- **WHEN** the user runs `hf config --no-color`
- **THEN** the output MUST contain no ANSI escape codes
- **AND** the separator line MUST still appear

#### Scenario: Color suppressed — NO_COLOR env var

- **WHEN** `NO_COLOR` is set in the environment
- **AND** the user runs `hf config`
- **THEN** the output MUST contain no ANSI escape codes
- **AND** the separator line MUST still appear

#### Scenario: Color suppressed — non-TTY stdout

- **WHEN** stdout is not a TTY (e.g. output is piped)
- **WHEN** the user runs `hf config`
- **THEN** the output MUST contain no ANSI escape codes
- **AND** the separator line MUST still appear

### Requirement: Set Configuration Value

The CLI SHALL allow setting individual configuration properties using dotted section.key notation. When called with no arguments the command SHALL launch an interactive fuzzy-finder to select the parameter, then prompt for a value. After any successful set the full active configuration MUST be displayed.

#### Scenario: Set a config property — non-interactive

- **GIVEN** an active environment is configured
- **WHEN** the user runs `hf config set <section.key> <value>`
- **THEN** the value MUST be written into the active environment file
- **AND** subsequent reads MUST return the new value
- **AND** the CLI MUST display the full active configuration (same output as `hf config show`) after the successful write

#### Scenario: Interactive set — fuzzy-find parameter selection

- **GIVEN** an active environment is configured
- **WHEN** the user runs `hf config set` with no arguments
- **THEN** the CLI MUST open an interactive fuzzy-finder listing all known `section.key` pairs with their current values (secrets shown as `<set>` or `<not set>`)
- **AND** the user can type to filter and press Enter to select a parameter
- **AND** after selection the CLI MUST prompt for the new value on stdin
- **AND** after the user enters a value the CLI MUST write it to the active environment file
- **AND** the CLI MUST display the full active configuration after the successful write

#### Scenario: Interactive set — user aborts selection

- **GIVEN** an active environment is configured
- **WHEN** the user runs `hf config set` with no arguments
- **AND** presses Escape or Ctrl+C in the fuzzy-finder
- **THEN** the CLI MUST exit with code 0 without modifying any configuration

#### Scenario: Invalid key format

- **GIVEN** the user runs `hf config set` without a dotted key
- **WHEN** the key does not contain a `.` separator
- **THEN** the command MUST exit with code 1
- **AND** MUST print `[ERROR] key must be in section.key format (e.g. hyperfleet.api-url)`

#### Scenario: Unknown section

- **GIVEN** the user runs `hf config set` with an unknown section prefix
- **WHEN** the section is not one of: hyperfleet, kubernetes, maestro, port-forward, database, rabbitmq, registry
- **THEN** the command MUST exit with code 1
- **AND** MUST print `[ERROR] unknown config section '<section>'`

#### Scenario: No active environment

- **GIVEN** no active environment is configured
- **WHEN** the user runs `hf config set <section.key> <value>`
- **THEN** the command MUST exit with code 1
- **AND** MUST print the no-active-environment error with guidance
