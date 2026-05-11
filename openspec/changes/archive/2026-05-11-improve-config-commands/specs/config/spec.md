## MODIFIED Requirements

### Requirement: Show Configuration

`hf config` and `hf config show` MUST display the active environment's configuration sections followed by a `state:` block at the bottom listing all non-empty runtime state variables.

#### Scenario: Active environment set — state variables included

- **GIVEN** an active environment is configured and state.yaml contains any of: active-environment, cluster-id, cluster-name, nodepool-id
- **WHEN** the user runs `hf config` or `hf config show`
- **THEN** a `state:` section MUST appear last in the output, after all config sections
- **AND** it MUST list all non-empty state keys (active-environment, cluster-id, cluster-name, nodepool-id) as YAML key-value pairs
- **AND** the existing config sections (hyperfleet, kubernetes, …) MUST precede the state block

## ADDED Requirements

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
