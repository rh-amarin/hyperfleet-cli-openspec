# Spec: config

## Purpose

Manage HyperFleet CLI configuration through environment profiles and runtime state. Provides commands to show, get, and set configuration values, manage named environment profiles, and display active resource IDs.
## Requirements
### Requirement: Show Configuration

`hf config` and `hf config show` MUST display the active environment's configuration sections followed by a `state:` block at the bottom listing all non-empty runtime state variables.

#### Scenario: Active environment file path displayed

- **GIVEN** an active environment is configured
- **WHEN** the user runs `hf config` or `hf config show` (no argument)
- **THEN** the CLI MUST print the absolute path of the active environment file
- **AND** the path MUST appear before the YAML sections
- **AND** the path MUST be the resolved absolute path of `~/.config/hf/environments/<name>.yaml`

### Requirement: Set Configuration Value

The CLI SHALL allow setting individual configuration properties using dotted section.key notation.

#### Scenario: Set a config property

- **GIVEN** an active environment is configured
- **WHEN** the user runs `hf config set <section.key> <value>`
- **THEN** the value MUST be written into the active environment file
- **AND** subsequent reads MUST return the new value

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

---

### Requirement: hf-config-env-create

`hf config env create <name>` SHALL create a new named environment from the bundled template and immediately activate it.

#### Scenario: Create environment with new name

- **GIVEN** no environment named `<name>` exists
- **WHEN** the user runs `hf config env create <name>`
- **THEN** the CLI MUST copy `cmd/assets/config-template.yaml` to `~/.config/hf/environments/<name>.yaml`
- **AND** set `active-environment: <name>` in `state.yaml`
- **AND** print the absolute path of the new file with a message to edit it:
  ```
  Environment '<name>' created and activated.
  Edit your configuration: ~/.config/hf/environments/<name>.yaml
  ```

#### Scenario: Name already exists

- **GIVEN** an environment named `<name>` already exists
- **WHEN** the user runs `hf config env create <name>`
- **THEN** the CLI MUST exit with code 1
- **AND** MUST print `[ERROR] environment '<name>' already exists`

#### Scenario: Name not provided

- **GIVEN** the user runs `hf config env create` with no argument
- **WHEN** the command is evaluated
- **THEN** the CLI MUST show the command usage and exit with code 1

---

### Requirement: hf-config-env-show

`hf config env show <name>` SHALL display the file location and values of a named environment profile without activating it.

#### Scenario: Show named environment

- **GIVEN** an environment named `<name>` exists at `~/.config/hf/environments/<name>.yaml`
- **WHEN** the user runs `hf config env show <name>`
- **THEN** the CLI MUST print the absolute file path of the environment file
- **AND** display all key-value pairs defined in that file, grouped by section
- **AND** secrets MUST be shown as `<set>` or `<not set>`
- **AND** the currently active environment MUST NOT be changed

#### Scenario: Non-existent environment

- **GIVEN** no environment named `<name>` exists
- **WHEN** the user runs `hf config env show <name>`
- **THEN** the CLI MUST print `[ERROR] environment '<name>' not found`
- **AND** exit with code 1

---

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

- **GIVEN** no active environment is configured
- **WHEN** the user runs any of: `hf config show`, `hf config set`
- **THEN** the command MUST exit with code 1
- **AND** MUST print:
  ```
  [ERROR] no active environment
    → run 'hf config env create <name>' to create one
    → run 'hf config env activate <name>' to activate an existing one
  ```

#### Scenario: Always-available commands

- **GIVEN** no active environment is configured
- **WHEN** the user runs any of: `hf config env list`, `hf config env create`, `hf config env activate`, `hf config env show`
- **THEN** the command MUST succeed normally

### Requirement: Get Configuration Value

`hf config get <key>` SHALL retrieve a single configuration or state value. Use `section.key` dot notation for config values or a plain key for state values.

#### Scenario: Get a config value

- **GIVEN** an active environment is set
- **WHEN** the user runs `hf config get hyperfleet.api-url`
- **THEN** the CLI MUST print the resolved value of `api-url` in the `hyperfleet` section

#### Scenario: Get a state value

- **GIVEN** an active environment is set and cluster-id is present in state.yaml
- **WHEN** the user runs `hf config get cluster-id`
- **THEN** the CLI MUST print the value of `cluster-id` from state.yaml

#### Scenario: Key not found

- **GIVEN** an active environment is set
- **WHEN** the user runs `hf config get hyperfleet.nonexistent`
- **THEN** the CLI MUST print `[ERROR] Config key 'hyperfleet.nonexistent' not found`
- **AND** exit with code 1

#### Scenario: No arguments provided

- **GIVEN** the user runs `hf config get` with no arguments
- **THEN** the CLI MUST display the command's full help text
- **AND** exit with code 1

