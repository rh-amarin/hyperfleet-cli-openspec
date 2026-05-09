# Configuration Model Specification

## Purpose

Define the split YAML configuration model for the HyperFleet CLI. Configuration is divided into two files: a static configuration file for connection properties and settings that rarely change, and an active state file for transient runtime state like the current cluster or nodepool selection. This replaces the file-per-property model from the shell scripts.

## Requirements

### Requirement: Split Configuration Files

The CLI SHALL use two YAML files stored in `~/.config/hf/`.

#### Scenario: Static configuration file

- GIVEN the CLI is initialized
- WHEN configuration is loaded
- THEN the static config MUST be stored at `~/.config/hf/config.yaml`
- AND it MUST contain the following sections with their properties:
  ```yaml
  hyperfleet:
    api-url: "http://localhost:8000"
    api-version: "v1"
    token: ""
    gcp-project: "hcm-hyperfleet"

  kubernetes:
    context: ""
    namespace: ""

  maestro:
    consumer: "cluster1"
    http-endpoint: "http://localhost:8100"
    grpc-endpoint: "localhost:8090"
    namespace: "maestro"  # Kubernetes namespace used for maestro pod port-forwarding

  port-forward:
    api-port: 8000
    pg-port: 5432
    maestro-http-port: 8100
    maestro-http-remote-port: 8000
    maestro-grpc-port: 8090
    maestro-grpc-remote-port: 8090

  database:
    host: "localhost"
    port: 5432
    name: "hyperfleet"
    user: "hyperfleet"
    password: "foobar-bizz-buzz"

  rabbitmq:
    host: "localhost"
    mgmt-port: 15672
    user: "guest"
    password: "guest"
    vhost: "/"

  registry:
    name: ""  # defaults to $USER at runtime
  ```

#### Scenario: Active state file

- GIVEN the CLI is managing runtime state
- WHEN state changes occur (e.g., cluster selection)
- THEN the state MUST be stored at `~/.config/hf/state.yaml`
- AND properties MUST be top-level (flat, not nested):
  ```yaml
  active-environment: "kind"
  cluster-id: "019dbf43-65c5-7562-9077-e0a2331a1070"
  cluster-name: "test-1e317d46"
  nodepool-id: "019dbf43-7199-7ea6-b786-d617fc793c28"
  ```
- AND the file MUST be updated atomically (write to temp, then rename)

#### Scenario: Canonical configuration defaults

- GIVEN the static configuration file is missing or a key is absent
- WHEN the CLI resolves a configuration value
- THEN the built-in defaults defined in the `static configuration file` scenario above MUST be used
- AND all other specs that reference configuration defaults MUST treat the values in this spec as authoritative
- AND no other spec may define a conflicting default for the same key

#### Scenario: Config directory creation

- GIVEN the config directory does not exist
- WHEN the CLI is first invoked
- THEN the CLI MUST create `~/.config/hf/` if it does not exist
- AND create `config.yaml` with default values if it does not exist
- AND create `state.yaml` as an empty file if it does not exist

### Requirement: Configuration Precedence

The CLI SHALL resolve configuration values with a defined precedence order.

#### Scenario: Precedence chain

- GIVEN multiple sources may define the same property
- WHEN the CLI resolves a configuration value
- THEN the precedence order MUST be (highest to lowest):
  1. CLI flags (`--api-url`, `--api-token`)
  2. Environment variables (`HF_API_URL`, `HF_API_VERSION`, `HF_TOKEN`, `HF_CONTEXT`, `HF_NAMESPACE`)
  3. Active environment overrides (from `environments/<name>.yaml`)
  4. Static config file (`config.yaml`)
  5. Built-in defaults

### Requirement: Environment Profiles

The CLI SHALL support named environment profiles that override static configuration.

#### Scenario: Environment file storage

- GIVEN environment profiles exist
- WHEN environments are stored
- THEN each environment MUST be stored as `~/.config/hf/environments/<name>.yaml`
- AND the file MUST contain only the properties that differ from the base config
- AND the file MUST use the same YAML structure as `config.yaml` (nested sections)
  ```yaml
  # ~/.config/hf/environments/gke-production.yaml
  hyperfleet:
    api-url: "https://hyperfleet.prod.internal:8000"
    token: "prod-token-xxx"
  kubernetes:
    context: "gke_project_region_cluster"
    namespace: "hyperfleet-prod"
  database:
    host: "db.prod.internal"
    password: "prod-password"
  ```

#### Scenario: List environments

- GIVEN environment profiles exist in `~/.config/hf/environments/`
- WHEN the user runs `hf config env list`
- THEN the CLI MUST list each environment by filename (without `.yaml`)
- AND show the count of overridden properties as `(N overrides)`
- AND mark the active environment (from `state.yaml` `active-environment`) with `✓` at the left margin; inactive environments MUST be prefixed with two spaces
- AND the output MUST follow this format:
  ```
    dev       (3 overrides)
  ✓ prod      (5 overrides)  ← active
    local     (2 overrides)
  ```
- AND if no environments exist, the CLI MUST print `[INFO] No environments configured. Run 'hf config env new' to create one.` and exit 0

#### Scenario: Activate environment

- GIVEN a named environment exists
- WHEN the user runs `hf config env activate <name>`
- THEN the CLI MUST set `active-environment: <name>` in `state.yaml`
- AND subsequent config reads MUST merge the environment overrides on top of the base config
- AND the CLI MUST NOT modify `config.yaml` — overrides are applied at runtime

#### Scenario: Activate non-existent environment

- GIVEN no environment named `<name>` exists at `~/.config/hf/environments/<name>.yaml`
- WHEN the user runs `hf config env activate <name>`
- THEN the CLI MUST print `[ERROR] environment '<name>' not found`
- AND exit with code 1

### Requirement: Environment Variable Overrides

The CLI SHALL support environment variable overrides for key configuration properties.

#### Scenario: Supported environment variables

- GIVEN the following mappings exist:
  | Environment Variable | Config Path |
  |---------------------|-------------|
  | `HF_API_URL` | `hyperfleet.api-url` |
  | `HF_API_VERSION` | `hyperfleet.api-version` |
  | `HF_TOKEN` | `hyperfleet.token` |
  | `HF_CONTEXT` | `kubernetes.context` |
  | `HF_NAMESPACE` | `kubernetes.namespace` |
- WHEN any of these environment variables are set
- THEN the corresponding config value MUST use the environment variable
- AND the environment variable MUST take precedence over file-based config and environment profiles

### Requirement: State Management

The CLI SHALL manage active state separately from configuration.

#### Scenario: Set cluster context

- GIVEN a cluster is created or found via search
- WHEN the CLI updates the active cluster
- THEN `cluster-id` and `cluster-name` MUST be updated in `state.yaml`
- AND the static `config.yaml` MUST NOT be modified

#### Scenario: Set nodepool context

- GIVEN a nodepool is created or found via search
- WHEN the CLI updates the active nodepool
- THEN `nodepool-id` MUST be updated in `state.yaml`

### Requirement: Secret Handling

The CLI SHALL protect sensitive configuration values.

#### Scenario: Display secrets

- GIVEN a property is a secret (token, database.password, rabbitmq.password)
- WHEN `hf config show` displays the property
- THEN the value MUST be shown as `<set>` if non-empty or `<not set>` if empty
- AND the actual value MUST NOT be displayed in config show output

#### Scenario: Display empty string vs unset values

- GIVEN a non-secret config property may be set to an empty string or be absent entirely
- WHEN `hf config show` displays the property
- THEN a property set to an empty string MUST display as `''` (quoted empty string)
- AND a property whose key is absent from config MUST display as `<not set>`
- AND in JSON output, an empty string MUST appear as `""` and an absent key MUST be omitted

#### Scenario: Secrets in config file

- GIVEN secrets are stored in `config.yaml`
- WHEN the file is written
- THEN the CLI SHOULD warn the user about file permissions
- AND the config file SHOULD be created with mode `0600` (owner read/write only)

### Requirement: Config File Path Override

The CLI SHALL support overriding the config directory location.

#### Scenario: Custom config path

- GIVEN the `--config` flag or `HF_CONFIG_DIR` environment variable is set
- WHEN the CLI loads configuration
- THEN it MUST look for `config.yaml`, `state.yaml`, and `environments/` in the specified directory instead of `~/.config/hf/`
