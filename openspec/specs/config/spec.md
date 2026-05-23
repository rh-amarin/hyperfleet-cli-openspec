# Spec: config

## Purpose

Manage HyperFleet CLI configuration through environment profiles and runtime state. Provides commands to show and set configuration values, manage named environment profiles, and display active resource IDs.

## Requirements

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

The CLI SHALL allow setting individual configuration properties using dotted section.key notation. When called with no arguments the command SHALL launch an interactive fuzzy-finder to select the parameter. The picker MUST use a split-screen layout: the left panel lists filterable section.key pairs with their current values; the right panel MUST display the full active configuration as colorized YAML. A header MUST be shown above the item list explaining the operation and keyboard shortcuts. After any successful set the full active configuration MUST be displayed.

#### Scenario: Interactive set — split-screen preview

- GIVEN an active environment is configured
- WHEN the user runs `hf config set` with no arguments
- THEN the CLI MUST display a split-screen picker
- AND the left panel MUST list all known section.key pairs with their current values
- AND the right panel MUST display the full active configuration as colorized YAML
- AND a header MUST appear above the item list describing the operation and keyboard shortcuts

#### Scenario: Interactive set — fuzzy-find parameter selection

- GIVEN an active environment is configured
- WHEN the user runs `hf config set` with no arguments
- THEN the CLI MUST open an interactive fuzzy-finder listing all known `section.key` pairs with their current values (secrets shown as `<set>` or `<not set>`)
- AND the user can type to filter and press Enter to select a parameter
- AND after selection the CLI MUST prompt for the new value on stdin
- AND after the user enters a value the CLI MUST write it to the active environment file
- AND the CLI MUST display the full active configuration after the successful write

#### Scenario: Interactive set — user aborts selection

- GIVEN an active environment is configured
- WHEN the user runs `hf config set` with no arguments
- AND presses Escape or Ctrl+C in the fuzzy-finder
- THEN the CLI MUST exit with code 0 without modifying any configuration

#### Scenario: Invalid key format

- GIVEN the user runs `hf config set` without a dotted key
- WHEN the key does not contain a `.` separator
- THEN the command MUST exit with code 1
- AND MUST print `[ERROR] key must be in section.key format (e.g. hyperfleet.api-url)`

#### Scenario: Unknown section

- GIVEN the user runs `hf config set` with an unknown section prefix
- WHEN the section is not one of: hyperfleet, kubernetes, maestro, port-forward, database, rabbitmq, registry
- THEN the command MUST exit with code 1
- AND MUST print `[ERROR] unknown config section '<section>'`

#### Scenario: No active environment

- GIVEN no active environment is configured
- WHEN the user runs `hf config set <section.key> <value>`
- THEN the command MUST exit with code 1
- AND MUST print the no-active-environment error with guidance

### Requirement: hf-env

Environment management commands MUST be accessible exclusively at the top level as `hf env <subcommand>`. The `hf config env` group is removed. `hf env` with no arguments MUST launch an interactive fuzzy picker with a split-screen layout: left panel lists environments; right panel MUST display the highlighted environment's full YAML. A header MUST be shown above the item list explaining the operation and keyboard shortcuts.

#### Scenario: hf env bare — interactive picker

- GIVEN one or more environment profiles exist
- WHEN the user runs `hf env` with no arguments
- THEN the CLI MUST display a fuzzy-searchable list of environment names in the left panel
- AND MUST display the full colorized YAML of the highlighted environment in the right preview panel
- AND the currently active environment MUST be marked with a checkmark in the list
- AND a header MUST appear above the item list describing the operation and keyboard shortcuts
- WHEN the user selects an environment
- THEN the CLI MUST activate it and print the full active config (same as `hf config show`)
- WHEN the user aborts (Esc / Ctrl+C)
- THEN the CLI MUST exit cleanly with no changes

#### Scenario: Create environment via hf env create

- GIVEN no environment named `<name>` exists
- WHEN the user runs `hf env create <name>`
- THEN the CLI MUST copy the bundled template to `~/.config/hf/environments/<name>.yaml`
- AND set `active-environment: <name>` in `state.yaml`
- AND print the absolute path of the new file with a message to edit it

#### Scenario: Name already exists

- GIVEN an environment named `<name>` already exists
- WHEN the user runs `hf env create <name>`
- THEN the CLI MUST exit with code 1
- AND MUST print `[ERROR] environment '<name>' already exists`

#### Scenario: Name not provided

- GIVEN the user runs `hf env create` with no argument
- THEN the CLI MUST show the command usage and exit with code 1

#### Scenario: List environments via hf env list

- GIVEN one or more environment profiles exist
- WHEN the user runs `hf env list` (or `hf env ls`)
- THEN the CLI MUST display a table of environment names with an active marker

#### Scenario: Activate environment via hf env activate

- GIVEN the user runs `hf env activate <name>` and the environment exists
- THEN the CLI MUST set `<name>` as the active environment and print a confirmation

#### Scenario: Delete environment via hf env delete

- GIVEN the user runs `hf env delete <name>` (or `hf env rm <name>`)
- AND the environment file exists
- THEN the CLI MUST display: `Delete environment '<name>'? Type 'yes' to confirm:`
- AND MUST read the user's response from stdin
- AND if the user types exactly `yes`, the CLI MUST delete the environment file
- AND if the deleted environment was active, MUST clear the `active-environment` key from `state.yaml`
- AND MUST print `[INFO] Environment '<name>' deleted`
- AND if the user types anything other than `yes`, the CLI MUST print `Aborted` and exit with code 0 without deleting

#### Scenario: Delete non-existent environment

- GIVEN no environment named `<name>` exists
- WHEN the user runs `hf env delete <name>`
- THEN the CLI MUST print `[ERROR] environment '<name>' not found`
- AND exit with code 1

#### Scenario: hf env bare — no environments

- GIVEN no environment profiles exist
- WHEN the user runs `hf env` with no arguments
- THEN the CLI MUST print the command help block
- AND MUST print a message directing the user to run `hf env create <name>`

### Requirement: hf-env-show

`hf env show <name>` MUST display the file location and values of a named environment profile without activating it.

#### Scenario: Show named environment

- GIVEN an environment named `<name>` exists
- WHEN the user runs `hf env show <name>`
- THEN the CLI MUST print the absolute file path of the environment file
- AND display all key-value pairs defined in that file, with secrets redacted
- AND the currently active environment MUST NOT be changed
- AND if `<name>` is the active environment the path MUST be prefixed with `[active]`

#### Scenario: Non-existent environment

- GIVEN no environment named `<name>` exists
- WHEN the user runs `hf env show <name>`
- THEN the CLI MUST print `[ERROR] environment '<name>' not found`
- AND exit with code 1

### Requirement: Show Current Cluster and NodePool IDs

`hf cluster id` and `hf nodepool id` SHALL display the currently active resource IDs from state.

NOTE on severity: `hf cluster id` and `hf nodepool id` use `[WARN]` + exit 0 when no ID is set because they are purely informational — the absence of a stored ID is not an error condition for these commands. All other commands that *require* a cluster-id or nodepool-id to function use `[ERROR]` + exit 1.

#### Scenario: Display cluster ID

- **GIVEN** a cluster-id is set in state.yaml
- **WHEN** the user runs `hf cluster id`
- **THEN** the CLI MUST print the value of `cluster-id` from `~/.config/hf/state.yaml`

#### Scenario: No cluster ID set

- **GIVEN** no cluster-id is set in state.yaml
- **WHEN** the user runs `hf cluster id`
- **THEN** the CLI MUST print `[WARN] No cluster-id set in state`
- **AND** exit with code 0

#### Scenario: Display nodepool ID

- **GIVEN** a nodepool-id is set in state.yaml
- **WHEN** the user runs `hf nodepool id`
- **THEN** the CLI MUST print the value of `nodepool-id` from `~/.config/hf/state.yaml`

#### Scenario: No nodepool ID set

- **GIVEN** no nodepool-id is set in state.yaml
- **WHEN** the user runs `hf nodepool id`
- **THEN** the CLI MUST print `[WARN] No nodepool-id set in state`
- **AND** exit with code 0

---

### Requirement: active-env-guard

Commands that require a configured target MUST fail when no active environment is set.

#### Scenario: Command requires active env, none set

- GIVEN no active environment is configured
- WHEN the user runs any of: `hf config show`, `hf config set`
- THEN the command MUST exit with code 1
- AND MUST print:
  ```
  [ERROR] no active environment
    → run 'hf env create <name>' to create one
    → run 'hf env activate <name>' to activate an existing one
  ```

#### Scenario: Always-available commands

- GIVEN no active environment is configured
- WHEN the user runs any of: `hf env`, `hf env list`, `hf env create`, `hf env activate`, `hf env show`
- THEN the command MUST succeed normally
