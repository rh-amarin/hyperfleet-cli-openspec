# Delta for config

## MODIFIED Requirements

### Requirement: Show Configuration

`hf config` and `hf config show` MUST display the active environment's configuration sections followed by a `state:` block at the bottom listing all non-empty runtime state variables. When writing to an interactive terminal, section header keys MUST be rendered in bold cyan and a separator line MUST appear between the last configuration section and the `state:` block. Color output is suppressed when `--no-color` is set, the `NO_COLOR` environment variable is set, or stdout is not a TTY.

#### Scenario: Active environment file path displayed

- **WHEN** the user runs `hf config` or `hf config show` (no argument)
- **THEN** the CLI MUST print the absolute path of the active environment file
- **AND** the path MUST appear before the YAML sections
- **AND** the path MUST be the resolved absolute path of `~/.config/hf/environments/<name>.yaml`

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
