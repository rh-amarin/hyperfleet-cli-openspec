# Maestro Operations Specification

## Purpose

Provide CLI commands for managing Maestro resources, which are used by HyperFleet adapters to deploy Kubernetes manifests to managed clusters. All maestro commands use the Maestro HTTP API directly — no external `maestro-cli` tool is required.

## Requirements

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

### Requirement: List Maestro Resource Bundles

The CLI SHALL list maestro resource bundles via the HTTP API.

#### Scenario: List bundles

- GIVEN maestro-http-endpoint is configured
- WHEN the user runs `hf maestro bundles`
- THEN the CLI MUST send GET to `/api/maestro/v1/resource-bundles`
- AND output the response subject to the `--output` flag (default: JSON)
- AND the response contains `kind: ResourceBundleList` with `items` of resource bundles including full Kubernetes manifests, manifest_configs, metadata (labels/annotations), and per-resource feedback status

### Requirement: List Maestro Consumers

The CLI SHALL list maestro consumers via the HTTP API.

#### Scenario: List consumers

- GIVEN maestro-http-endpoint is configured
- WHEN the user runs `hf maestro consumers`
- THEN the CLI MUST send GET to `/api/maestro/v1/consumers`
- AND output the response subject to the `--output` flag (default: JSON)
- AND the JSON has shape `{"items": [{id, kind: "Consumer", name}], "kind": "ConsumerList", "total": N}`

### Requirement: Get Maestro Resource

The CLI SHALL retrieve a specific maestro resource by name via the HTTP API.

#### Scenario: Get by name

- GIVEN maestro-http-endpoint is configured
- WHEN the user runs `hf maestro get <name>`
- THEN the CLI MUST send GET to `/api/maestro/v1/resource-bundles` and filter by name
- AND output the matching resource bundle subject to the `--output` flag (default: JSON)

#### Scenario: Get — name not found

- GIVEN no resource bundle matches the provided name
- WHEN the user runs `hf maestro get <name>`
- THEN the CLI MUST display `[WARN] No resource bundle found matching '<name>'`
- AND exit with code 0

#### Scenario: Get with interactive selection

- GIVEN no name argument is provided
- WHEN the user runs `hf maestro get`
- THEN the CLI MUST list available resources and present an interactive selection menu

### Requirement: Delete Maestro Resource

The CLI SHALL delete a maestro resource by name via the HTTP API.

#### Scenario: Delete by name

- GIVEN maestro-http-endpoint is configured
- WHEN the user runs `hf maestro delete <name>`
- THEN the CLI MUST send DELETE to `/api/maestro/v1/resource-bundles/<id>` after resolving the name to an ID
- AND output the result subject to the `--output` flag (default: JSON)

#### Scenario: Delete with interactive selection

- GIVEN no name argument is provided
- WHEN the user runs `hf maestro delete`
- THEN the CLI MUST list available resources and present an interactive selection menu
