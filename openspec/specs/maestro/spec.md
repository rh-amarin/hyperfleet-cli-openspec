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

#### Scenario: List resources — table output

- WHEN the user runs `hf maestro list --output table`
- THEN the CLI MUST print one line per resource bundle:
  `<id>  <name>  v<version>`
- AND for each manifest in the bundle, print one indented child line:
  `  <kind>/<name>  <namespace>`
- AND if there are no resource bundles, print `No resource bundles.`

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

### Requirement: Internal Maestro HTTP Client

The CLI SHALL provide an `internal/maestro` package with a typed HTTP client for the Maestro REST API.

#### Scenario: Client construction with curl mode

- GIVEN the `--curl` flag is set
- WHEN `maestro.NewFromConfig(s, curlMode)` is called with `curlMode=true`
- THEN the returned client MUST operate in dry-run mode for all HTTP methods

#### Scenario: Curl output for GET requests (dry-run)

- GIVEN `curlMode=true`
- WHEN `client.get(ctx, path, v)` is called
- THEN the client MUST write to stderr before returning:
  ```
  [CURL] curl -s "<url>" \
    -H 'Accept: application/json'
  ```
- AND the URL MUST be double-quoted
- AND the client MUST NOT send the HTTP request
- AND the method MUST return `maestro.ErrDryRun`

#### Scenario: Curl output for DELETE requests (dry-run)

- GIVEN `curlMode=true`
- WHEN `client.delete(ctx, path)` is called
- THEN the client MUST write to stderr before returning:
  ```
  [CURL] curl -s -X DELETE "<url>"
  ```
- AND the URL MUST be double-quoted
- AND the client MUST NOT send the HTTP request
- AND the method MUST return `maestro.ErrDryRun`

#### Scenario: Curl mode disabled

- GIVEN `curlMode=false`
- WHEN any request is sent
- THEN no curl output MUST be written to stderr
- AND HTTP requests MUST execute normally

### Requirement: CLI Maestro List Command

The CLI SHALL implement `hf maestro list` to list resource-bundles filtered by the configured consumer.

#### Scenario: List resource-bundles for configured consumer

- GIVEN `maestro.consumer` is set in the active config
- WHEN the user runs `hf maestro list`
- THEN the CLI MUST call `ListResourceBundlesByConsumer` with the configured consumer name
- AND output the result as JSON (default)

#### Scenario: List all when no consumer configured

- GIVEN `maestro.consumer` is empty
- WHEN the user runs `hf maestro list`
- THEN the CLI MUST call `ListResourceBundles` and output all bundles as JSON

### Requirement: CLI Maestro Bundles Command

The CLI SHALL implement `hf maestro bundles` to list all resource-bundles with no consumer filter.

#### Scenario: List all bundles

- GIVEN a valid Maestro HTTP endpoint is configured
- WHEN the user runs `hf maestro bundles`
- THEN the CLI MUST send GET to `/api/maestro/v1/resource-bundles`
- AND output the full `ResourceBundleList` as JSON

### Requirement: CLI Maestro Consumers Command

The CLI SHALL implement `hf maestro consumers` to list all Maestro consumers.

#### Scenario: List consumers

- GIVEN a valid Maestro HTTP endpoint is configured
- WHEN the user runs `hf maestro consumers`
- THEN the CLI MUST send GET to `/api/maestro/v1/consumers`
- AND output the `ConsumerList` as JSON

### Requirement: CLI Maestro Get Command

The CLI SHALL implement `hf maestro get [name]` to retrieve a specific resource-bundle by name.

#### Scenario: Get by name argument

- GIVEN a name argument is provided
- WHEN the user runs `hf maestro get <name>`
- THEN the CLI MUST list all bundles, filter by name, and output the matching bundle as JSON

#### Scenario: Get — name not found

- GIVEN no bundle matches the provided name
- WHEN the user runs `hf maestro get <name>`
- THEN the CLI MUST display `[WARN] No resource bundle found matching '<name>'`
- AND exit with code 0

#### Scenario: Get with interactive selection

- GIVEN no name argument is provided
- WHEN the user runs `hf maestro get`
- THEN the CLI MUST list available resources and present a numbered interactive selection menu

### Requirement: CLI Maestro Delete Command

The CLI SHALL implement `hf maestro delete [name]` to delete a resource-bundle by name.

#### Scenario: Delete by name argument

- GIVEN a name argument is provided
- WHEN the user runs `hf maestro delete <name>`
- THEN the CLI MUST resolve the name to an ID and send DELETE to `/api/maestro/v1/resource-bundles/<id>`

#### Scenario: Delete — name not found

- GIVEN no bundle matches the provided name
- WHEN the user runs `hf maestro delete <name>`
- THEN the CLI MUST display `[WARN] No resource bundle found matching '<name>'`
- AND exit with code 0

#### Scenario: Delete with interactive selection

- GIVEN no name argument is provided
- WHEN the user runs `hf maestro delete`
- THEN the CLI MUST list available resources and present a numbered interactive selection menu

### Requirement: Maestro Dry-Run Sentinel Error

The maestro client package SHALL expose a sentinel error for dry-run mode.

#### Scenario: ErrDryRun identity

- GIVEN `maestro.ErrDryRun` is defined
- WHEN a caller receives an error from a curl-mode request
- THEN `errors.Is(err, maestro.ErrDryRun)` MUST be true

