# Configuration Model Specification

## Purpose

Define the configuration model for the HyperFleet CLI. Configuration is managed through self-contained environment files, each fully defining all settings for one target environment. A state file tracks transient runtime state such as the currently active environment and cluster/nodepool selection.

## Requirements

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

### Requirement: Active State File

The CLI SHALL manage active state separately from configuration.

#### Scenario: Active state file

- **GIVEN** the CLI is managing runtime state
- **WHEN** state changes occur (e.g., cluster selection)
- **THEN** the state MUST be stored at `~/.config/hf/state.yaml`
- **AND** properties MUST be top-level (flat, not nested):
  ```yaml
  active-environment: "kind"
  cluster-id: "019dbf43-65c5-7562-9077-e0a2331a1070"
  cluster-name: "test-1e317d46"
  nodepool-id: "019dbf43-7199-7ea6-b786-d617fc793c28"
  ```
- **AND** the file MUST be updated atomically (write to temp, then rename)

#### Scenario: Set cluster context

- **GIVEN** a cluster is created or found via search
- **WHEN** the CLI updates the active cluster
- **THEN** `cluster-id` and `cluster-name` MUST be updated in `state.yaml`

#### Scenario: Set nodepool context

- **GIVEN** a nodepool is created or found via search
- **WHEN** the CLI updates the active nodepool
- **THEN** `nodepool-id` MUST be updated in `state.yaml`

### Requirement: Environment Variable Overrides

The CLI SHALL support environment variable overrides for key configuration properties.

#### Scenario: Supported environment variables

- **GIVEN** the following mappings exist:
  | Environment Variable | Config Path |
  |---------------------|-------------|
  | `HF_API_URL` | `hyperfleet.api-url` |
  | `HF_API_VERSION` | `hyperfleet.api-version` |
  | `HF_TOKEN` | `hyperfleet.token` |
  | `HF_CONTEXT` | `kubernetes.context` |
  | `HF_NAMESPACE` | `kubernetes.namespace` |
- **WHEN** any of these environment variables are set
- **THEN** the corresponding config value MUST use the environment variable
- **AND** the environment variable MUST take precedence over file-based config and environment profiles

### Requirement: Secret Handling

The CLI SHALL protect sensitive configuration values.

#### Scenario: Display secrets

- **GIVEN** a property is a secret (token, database.password, rabbitmq.password)
- **WHEN** `hf config show` displays the property
- **THEN** the value MUST be shown as `<set>` if non-empty or `<not set>` if empty
- **AND** the actual value MUST NOT be displayed in config show output

#### Scenario: Display empty string vs unset values

- **GIVEN** a non-secret config property may be set to an empty string or be absent entirely
- **WHEN** `hf config show` displays the property
- **THEN** a property set to an empty string MUST display as `''` (quoted empty string)
- **AND** a property whose key is absent from config MUST display as `<not set>`
- **AND** in JSON output, an empty string MUST appear as `""` and an absent key MUST be omitted

### Requirement: Config File Path Override

The CLI SHALL support overriding the config directory location.

#### Scenario: Custom config path

- **GIVEN** the `--config` flag or `HF_CONFIG_DIR` environment variable is set
- **WHEN** the CLI loads configuration
- **THEN** it MUST look for `state.yaml` and `environments/` in the specified directory instead of `~/.config/hf/`
