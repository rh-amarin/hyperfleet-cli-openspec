## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: HyperFleet application namespace config key

The CLI SHALL read the HyperFleet application namespace from `hyperfleet.namespace` (not `kubernetes.namespace`).

#### Scenario: Namespace resolved from hyperfleet section

- **WHEN** `s.Get("hyperfleet", "namespace")` is called
- **THEN** it MUST resolve using the standard precedence chain: `HF_NAMESPACE` env var > active env file `hyperfleet.namespace` > built-in default

#### Scenario: Legacy key not used

- **GIVEN** a user's environment YAML contains `kubernetes.namespace` but not `hyperfleet.namespace`
- **WHEN** the CLI reads the namespace
- **THEN** it MUST NOT read from `kubernetes.namespace` — the user MUST migrate the key to `hyperfleet.namespace`
