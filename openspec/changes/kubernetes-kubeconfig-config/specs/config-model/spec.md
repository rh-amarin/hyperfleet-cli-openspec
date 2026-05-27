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
  | `HF_KUBECONFIG`     | `kubernetes.kubeconfig` |
  | `HF_NAMESPACE`      | `hyperfleet.namespace` |
- **WHEN** any of these environment variables are set
- **THEN** the corresponding config value MUST use the environment variable
- **AND** the environment variable MUST take precedence over file-based config and environment profiles
