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
- **WHEN** the user runs `hf env list`
- **THEN** the CLI MUST list each environment by filename (without `.yaml`)
- **AND** mark the active environment with `✓`; inactive environments MUST be prefixed with two spaces
- **AND** if no environments exist, the CLI MUST print `No environments configured. Run 'hf env create <name>' to create one.` and exit 0

#### Scenario: Activate environment

- **GIVEN** a named environment exists
- **WHEN** the user runs `hf env activate <name>`
- **THEN** the CLI MUST set `active-environment: <name>` in `state.yaml`
- **AND** subsequent `Get()` calls MUST read values from that environment file

#### Scenario: Activate non-existent environment

- **GIVEN** no environment named `<name>` exists
- **WHEN** the user runs `hf env activate <name>`
- **THEN** the CLI MUST print `[ERROR] environment '<name>' not found`
- **AND** exit with code 1

### Requirement: Config Directory Initialization

The CLI SHALL initialize the config directory structure on first run.

#### Scenario: First run

- **GIVEN** the config directory does not exist
- **WHEN** the CLI is first invoked
- **THEN** the CLI MUST create `~/.config/hf/` if it does not exist
- **AND** create `~/.config/hf/environments/` if it does not exist
- **AND** create `state.yaml` as an empty file if it does not exist
- **AND** NOT create `config.yaml`

### Requirement: Active State File

The CLI SHALL maintain a `state.yaml` file for runtime state separate from environment configuration.

#### Scenario: Generic resource state keys

- GIVEN a resource type named `channels`
- WHEN the user runs `hf rs channels search <name>` and a match is found
- THEN the CLI MUST write the matched resource ID to `state.yaml` under the key `channels`
- AND generic resource state keys MUST coexist with `clusters`, `nodepools`, and `active-environment`

### Requirement: Environment Variable Overrides

The CLI SHALL support environment variable overrides for key configuration properties.

#### Scenario: Supported environment variables

- **GIVEN** the following mappings exist:
  | Environment Variable | Config Path           |
  |---------------------|-----------------------|
  | `HF_API_URL`        | `hyperfleet.api-url`  |
  | `HF_API_VERSION`    | `hyperfleet.api-version` |
  | `HF_TOKEN`          | `hyperfleet.token`    |
  | `HF_CONTEXT`        | `kubernetes.context`  |
  | `HF_NAMESPACE`      | `hyperfleet.namespace` |
- **WHEN** any of these environment variables are set
- **THEN** the corresponding config value MUST use the environment variable
- **AND** the environment variable MUST take precedence over file-based config and environment profiles

### Requirement: Secret Handling

The CLI SHALL protect sensitive configuration values.

#### Scenario: Display secrets

- **GIVEN** a property is a secret (token, database.password, rabbitmq.password)
- **WHEN** `hf env show` displays the property
- **THEN** the value MUST be shown as `<set>` if non-empty or `<not set>` if empty
- **AND** the actual value MUST NOT be displayed in config show output

#### Scenario: Display empty string vs unset values

- **GIVEN** a non-secret config property may be set to an empty string or be absent entirely
- **WHEN** `hf env show` displays the property
- **THEN** a property set to an empty string MUST display as `''` (quoted empty string)
- **AND** a property whose key is absent from config MUST display as `<not set>`
- **AND** in JSON output, an empty string MUST appear as `""` and an absent key MUST be omitted

### Requirement: Config File Path Override

The CLI SHALL support overriding the config directory location.

#### Scenario: Custom config path

- **GIVEN** the `--config` flag or `HF_CONFIG_DIR` environment variable is set
- **WHEN** the CLI loads configuration
- **THEN** it MUST look for `state.yaml` and `environments/` in the specified directory instead of `~/.config/hf/`

### Requirement: HyperFleet application namespace config key

The CLI SHALL read the HyperFleet application namespace from `hyperfleet.namespace` (not `kubernetes.namespace`).

#### Scenario: Namespace resolved from hyperfleet section

- **WHEN** `s.Get("hyperfleet", "namespace")` is called
- **THEN** it MUST resolve using the standard precedence chain: `HF_NAMESPACE` env var > active env file `hyperfleet.namespace` > built-in default

#### Scenario: Legacy key not used

- **GIVEN** a user's environment YAML contains `kubernetes.namespace` but not `hyperfleet.namespace`
- **WHEN** the CLI reads the namespace
- **THEN** it MUST NOT read from `kubernetes.namespace` — the user MUST migrate the key to `hyperfleet.namespace`

### Requirement: Resource Types Configuration

The CLI SHALL support a structured `resource-types` section in environment YAML files defining configurable HyperFleet API resource types.

#### Scenario: Resource type definition fields

- GIVEN an environment file at `~/.config/hf/environments/<name>.yaml`
- WHEN a resource type is defined under `resource-types.<type-name>`
- THEN it MUST support fields:
  - `path` (required): API path relative to `{api-url}/api/hyperfleet/{api-version}/`
  - `parent` (optional): name of the immediate parent resource type
  - `path-param` (optional): placeholder name this type's ID fills in child paths; when omitted, derived from the entity name (e.g. `clusters` → `cluster_id`)
  - `create-template` (optional): filename under `<config-dir>/templates/` for create payloads
- AND the map key `<type-name>` MUST be used as both the CLI subcommand name and the flat key in `state.yaml` when the type is selected
- AND `state-key` MUST NOT be a supported configuration field

#### Scenario: Config template default

- GIVEN a newly created environment from the embedded config template
- WHEN the environment file is written
- THEN it MUST include `resource-types` with at least `clusters` and `nodepools` entries:
  ```yaml
  resource-types:
    clusters:
      path: clusters
      create-template: clusters.json
    nodepools:
      parent: clusters
      path: "clusters/{cluster_id}/nodepools"
      create-template: nodepools.json
  ```

### Requirement: State File Path

The config Store MUST expose the absolute path to `state.yaml`.

#### Scenario: StateFilePath returns state.yaml location

- **WHEN** `Store.StateFilePath()` is called
- **THEN** it MUST return `<config-dir>/state.yaml`

### Requirement: Identity Header Config Keys

The CLI SHALL read optional identity header configuration from the `hyperfleet` section of the active environment file.

#### Scenario: identity-header and identity-value resolve from environment file

- **GIVEN** the active environment file contains:
  ```yaml
  hyperfleet:
    identity-header: X-HyperFleet-Identity
    identity-value: openspec@test.com
  ```
- **WHEN** `s.Get("hyperfleet", "identity-header")` and `s.Get("hyperfleet", "identity-value")` are called
- **THEN** they MUST return `X-HyperFleet-Identity` and `openspec@test.com` respectively
- **AND** these values MUST be passed to `api.NewClient` as `identityHeader` and `identityValue`

#### Scenario: identity-header defaults to empty

- **GIVEN** the active environment file does not contain `hyperfleet.identity-header`
- **WHEN** `s.Get("hyperfleet", "identity-header")` is called
- **THEN** it MUST return an empty string
- **AND** no identity header MUST be injected into API requests

#### Scenario: Config template includes identity header keys as comments

- **GIVEN** a newly created environment from the embedded config template
- **WHEN** the environment file is written
- **THEN** it MUST include commented-out `identity-header` and `identity-value` keys under the `hyperfleet` section so operators are aware of the option

