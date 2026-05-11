## REMOVED Requirements

### Requirement: Split Configuration Files
**Reason:** Replaced by the self-contained environment file model. `config.yaml` is eliminated; configuration no longer cascades from a shared base file. Each environment file is complete and standalone.
**Migration:** Users must create an environment file with `hf config env create <name>` and edit it directly. There is no automatic migration of existing `config.yaml` data.

## MODIFIED Requirements

### Requirement: Configuration Precedence

The CLI SHALL resolve configuration values with a defined precedence order.

#### Scenario: Precedence chain

- **WHEN** the CLI resolves a configuration value
- **THEN** the precedence order MUST be (highest to lowest):
  1. CLI flags (`--api-url`, `--api-token`)
  2. Environment variables (`HF_API_URL`, `HF_API_VERSION`, `HF_TOKEN`, `HF_CONTEXT`, `HF_NAMESPACE`)
  3. Active environment file (`~/.config/hf/environments/<active-name>.yaml`)
  4. Built-in defaults
- **AND** there MUST be no `config.yaml` layer in the precedence chain

### Requirement: Environment Profiles

The CLI SHALL support named environment profiles, each fully defining the configuration for one target environment.

#### Scenario: Environment file storage

- **WHEN** an environment is created
- **THEN** it MUST be stored at `~/.config/hf/environments/<name>.yaml`
- **AND** the file MUST contain ALL configuration sections with their values (complete, not sparse)
- **AND** the file MUST be seeded from `cmd/assets/config-template.yaml`

#### Scenario: List environments

- **GIVEN** environment profiles exist in `~/.config/hf/environments/`
- **WHEN** the user runs `hf config env list`
- **THEN** the CLI MUST list each environment by filename (without `.yaml`)
- **AND** mark the active environment with `✓`; inactive environments MUST be prefixed with two spaces
- **AND** if no environments exist, the CLI MUST print `No environments configured. Run 'hf config env create <name>' to create one.` and exit 0

#### Scenario: Activate environment

- **GIVEN** a named environment exists
- **WHEN** the user runs `hf config env activate <name>`
- **THEN** the CLI MUST set `active-environment: <name>` in `state.yaml`
- **AND** subsequent `Get()` calls MUST read values from that environment file

#### Scenario: Activate non-existent environment

- **GIVEN** no environment named `<name>` exists
- **WHEN** the user runs `hf config env activate <name>`
- **THEN** the CLI MUST print `[ERROR] environment '<name>' not found`
- **AND** exit with code 1

### Requirement: Set Configuration Value

The CLI SHALL allow setting individual configuration properties using dotted section.key notation, writing to the active environment file.

#### Scenario: Set a config property

- **GIVEN** an active environment is set
- **WHEN** `Set(section, key, value)` is called
- **THEN** the value MUST be written into the active environment file at `~/.config/hf/environments/<active-name>.yaml`
- **AND** subsequent `Get()` calls MUST return the new value

#### Scenario: Set with no active environment

- **GIVEN** no active environment is configured
- **WHEN** `Set(section, key, value)` is called
- **THEN** it MUST return an error: `[ERROR] no active environment`

### Requirement: Config Directory Initialization

#### Scenario: First run

- **GIVEN** the config directory does not exist
- **WHEN** the CLI is first invoked
- **THEN** the CLI MUST create `~/.config/hf/` if it does not exist
- **AND** create `~/.config/hf/environments/` if it does not exist
- **AND** create `state.yaml` as an empty file if it does not exist
- **AND** NOT create `config.yaml`
