# Spec: config

## Purpose

Manage HyperFleet CLI configuration through environment profiles and runtime state. Configuration is edited directly in YAML files; the CLI provides environment lifecycle commands and a read-only display command.

## Requirements

### Requirement: Show Environment Configuration

`hf env show [name]` MUST display the named environment's configuration sections. When `name` is omitted, the active environment MUST be shown. When the displayed environment is the active one, a `state:` block MUST follow the configuration sections listing all non-empty runtime state variables from `state.yaml`.

After all configuration and state output, the CLI MUST print:
- the absolute path of the environment file
- the absolute path of `state.yaml`
- a message directing the user to edit those files

When writing to an interactive terminal, section header keys MUST be rendered in bold cyan and separator lines MUST appear between the last configuration section and the `state:` block, and before the file paths. Color output is suppressed when `--no-color` is set, the `NO_COLOR` environment variable is set, or stdout is not a TTY.

Secret scalar values (`token`, `password`) MUST be redacted as `<set>` or `<not set>` in the display output.

#### Scenario: Show active environment with no argument

- **WHEN** the user runs `hf env show` with no argument
- **AND** an active environment is configured
- **THEN** the CLI MUST display the active environment file contents (secrets redacted)
- **AND** MUST append a `state:` block when non-empty state exists
- **AND** MUST print the environment file path with `[active]` suffix
- **AND** MUST print the state file path
- **AND** MUST print a message that the user can edit those files

#### Scenario: Show named environment

- **GIVEN** an environment named `<name>` exists
- **WHEN** the user runs `hf env show <name>`
- **THEN** the CLI MUST display all sections from that environment file with secrets redacted
- **AND** the currently active environment MUST NOT be changed
- **AND** if `<name>` is the active environment the `state:` block MUST be included
- **AND** file paths and the edit message MUST appear after the configuration output

#### Scenario: No active environment

- **GIVEN** no active environment is configured
- **WHEN** the user runs `hf env show` with no argument
- **THEN** the command MUST exit with code 1
- **AND** MUST print the no-active-environment error with guidance

#### Scenario: Non-existent environment

- **GIVEN** no environment named `<name>` exists
- **WHEN** the user runs `hf env show <name>`
- **THEN** the CLI MUST print `[ERROR] environment '<name>' not found`
- **AND** exit with code 1

#### Scenario: Section headers colorized on TTY

- **WHEN** the user runs `hf env show` and stdout is an interactive terminal
- **AND** `--no-color` is not set and `NO_COLOR` is not in the environment
- **THEN** top-level YAML section keys (e.g. `hyperfleet:`, `database:`, `state:`) MUST be rendered in bold cyan
- **AND** value lines MUST remain uncolored

#### Scenario: Color suppressed — non-TTY stdout

- **WHEN** stdout is not a TTY (e.g. output is piped)
- **WHEN** the user runs `hf env show`
- **THEN** the output MUST contain no ANSI escape codes
- **AND** separator lines MUST still appear

### Requirement: hf-env

Environment management commands MUST be accessible at the top level as `hf env <subcommand>`. There is no `hf config` command group. `hf env` with no arguments MUST launch an interactive fuzzy picker with a split-screen layout: left panel lists environments; right panel MUST display the highlighted environment's full YAML. A header MUST be shown above the item list explaining the operation and keyboard shortcuts.

#### Scenario: hf env bare — interactive picker

- GIVEN one or more environment profiles exist
- WHEN the user runs `hf env` with no arguments
- THEN the CLI MUST display a fuzzy-searchable list of environment names in the left panel
- AND MUST display the full colorized YAML of the highlighted environment in the right preview panel
- AND the currently active environment MUST be marked with a checkmark in the list
- AND a header MUST appear above the item list describing the operation and keyboard shortcuts
- WHEN the user selects an environment
- THEN the CLI MUST activate it and print the full active config (same as `hf env show`)
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
- THEN the CLI MUST delete the environment file
- AND if the deleted environment was active, MUST clear the `active-environment` key from `state.yaml`

#### Scenario: hf env bare — no environments

- GIVEN no environment profiles exist
- WHEN the user runs `hf env` with no arguments
- THEN the CLI MUST print the command help block
- AND MUST print a message directing the user to run `hf env create <name>`

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

#### Scenario: Always-available commands

- GIVEN no active environment is configured
- WHEN the user runs any of: `hf env`, `hf env list`, `hf env create`, `hf env activate`, `hf env show <name>`
- THEN the command MUST succeed normally (or fail only for other reasons such as a missing named environment)

#### Scenario: hf env show requires active env when name omitted

- GIVEN no active environment is configured
- WHEN the user runs `hf env show` with no argument
- THEN the command MUST exit with code 1
- AND MUST print:
  ```
  [ERROR] no active environment
    → run 'hf env create <name>' to create one
    → run 'hf env activate <name>' to activate an existing one
  ```
