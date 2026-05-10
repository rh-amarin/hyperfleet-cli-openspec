# Maestro — Phase 4b Delta

## ADDED Requirements

### Requirement: Internal Maestro HTTP Client

The CLI SHALL provide an `internal/maestro` package with a typed HTTP client for the Maestro REST API.

#### Scenario: Client construction from config

- GIVEN `maestro.http-endpoint` is set in the active config
- WHEN `maestro.NewFromConfig(s)` is called
- THEN a `Client` is returned pointing to that endpoint with a 30-second timeout

#### Scenario: List all resource-bundles

- GIVEN a valid Maestro HTTP endpoint is configured
- WHEN `client.ListResourceBundles(ctx)` is called
- THEN the client MUST send GET to `/api/maestro/v1/resource-bundles`
- AND decode the response into `ResourceBundleList`

#### Scenario: List resource-bundles by consumer

- GIVEN a consumer name is provided
- WHEN `client.ListResourceBundlesByConsumer(ctx, consumer)` is called
- THEN the client MUST send GET to `/api/maestro/v1/resource-bundles?search=consumer_name='<consumer>'`

#### Scenario: Get a resource-bundle by name

- GIVEN a resource-bundle name is provided
- WHEN `client.GetResourceBundle(ctx, name)` is called
- THEN the client MUST list all bundles and return the one matching the name, or nil if not found

#### Scenario: Delete a resource-bundle by ID

- GIVEN a valid resource-bundle ID
- WHEN `client.DeleteResourceBundle(ctx, id)` is called
- THEN the client MUST send DELETE to `/api/maestro/v1/resource-bundles/<id>`

#### Scenario: List consumers

- GIVEN a valid Maestro HTTP endpoint
- WHEN `client.ListConsumers(ctx)` is called
- THEN the client MUST send GET to `/api/maestro/v1/consumers`
- AND decode the response into `ConsumerList`

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
