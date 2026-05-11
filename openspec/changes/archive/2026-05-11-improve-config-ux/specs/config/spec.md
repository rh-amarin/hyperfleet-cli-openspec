## REMOVED Requirements

### Requirement: hf-config-env-new
**Reason:** Replaced by `hf config env create` (see MODIFIED: hf-config-env-create below). Interactive prompting and flag-based creation are removed. The command is purely name-driven with template seeding.
**Migration:** Use `hf config env create <name>` instead.

## MODIFIED Requirements

### Requirement: Show Configuration

`hf config` (with no subcommand) and `hf config show` display the resolved configuration of the active environment.

#### Scenario: Active environment set

- **GIVEN** an active environment is configured in state.yaml
- **WHEN** the user runs `hf config` or `hf config show`
- **THEN** the active environment name MUST be printed at the top
- **AND** config values MUST be shown grouped by section
- **AND** secrets (token, database.password, rabbitmq.password) MUST be shown as `<set>` or `<not set>`

#### Scenario: No active environment

- **GIVEN** no active environment is configured
- **WHEN** the user runs `hf config` or `hf config show`
- **THEN** the command MUST exit with code 1
- **AND** MUST print:
  ```
  [ERROR] no active environment
    → run 'hf config env create <name>' to create one
    → run 'hf config env activate <name>' to activate an existing one
  ```

#### Scenario: Show config for a named environment

- **GIVEN** environment profiles exist in `~/.config/hf/environments/`
- **WHEN** the user runs `hf config show <env-name>`
- **THEN** the CLI MUST display the absolute file path of that environment file
- **AND** display the values defined in that file grouped by section
- **AND** secrets MUST be shown as `<set>` or `<not set>`
- **AND** the active environment MUST NOT be changed

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

### Requirement: hf-config-env-create

`hf config env create <name>` creates a new named environment from the bundled template and immediately activates it.

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

### Requirement: active-env-guard

Commands that require a configured target must fail when no active environment is set.

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

## REMOVED Requirements

### Requirement: hf-config-doctor (doctor subcommand)
**Reason:** Removed to simplify the command surface. Connectivity issues are surfaced through natural API errors when commands run.
**Migration:** None required.
