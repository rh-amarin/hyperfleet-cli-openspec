## MODIFIED Requirements

### Requirement: List Maestro Resources

The CLI SHALL list maestro resources via the Maestro HTTP API.

#### Scenario: List resources

- GIVEN maestro-consumer and maestro-http-endpoint are configured
- WHEN the user runs `hf maestro list`
- THEN the CLI MUST send GET to `/api/maestro/v1/resource-bundles` filtered by consumer
- AND output the response subject to the `--output` flag (default: JSON)
- AND the JSON items MUST contain:
  - `id`: UUID
  - `name`: resource name (e.g., `mw-<cluster-uuid>`)
  - `consumerName`: the consumer (e.g., `cluster1`)
  - `version`: integer version number
  - `manifestCount`: number of Kubernetes manifests
  - `manifests`: array of `{kind, name, namespace}` summaries
  - `conditions`: array of `{type, status, reason}` (Applied, Available)

#### Scenario: List resources — table output

- WHEN the user runs `hf maestro list --output table`
- THEN the CLI MUST print one line per resource bundle:
  `<id>  <name>  v<version>`
- AND for each manifest in the bundle, print one indented child line:
  `  <kind>/<name>  <namespace>`
- AND if there are no resource bundles, print `No resource bundles.`
