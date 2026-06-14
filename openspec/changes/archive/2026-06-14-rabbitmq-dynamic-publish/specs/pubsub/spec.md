# Delta: pubsub — RabbitMQ Dynamic Publish

## Removed Requirements

- **Publish Cluster Change Event to RabbitMQ** (`hf rabbitmq publish cluster <exchange>`) — replaced by dynamic command
- **Publish NodePool Change Event to RabbitMQ** (`hf rabbitmq publish nodepool <exchange>`) — replaced by dynamic command

## New Requirement: Publish Any Resource Reconcile Event to RabbitMQ

The CLI SHALL publish a reconcile CloudEvent for any resource type defined in the active environment's `resource-types` configuration.

### Scenario: Unknown resource type

- GIVEN an active environment with `resource-types` configured
- WHEN the user runs `hf rabbitmq publish <resource-type> <exchange>` where `<resource-type>` is not defined in `resource-types`
- THEN the CLI MUST display `[ERROR] unknown resource type "<resource-type>"`
- AND exit with code 1

### Scenario: Missing state for resource

- GIVEN an active environment with `resource-types` configured
- WHEN the user runs `hf rabbitmq publish <resource-type> <exchange>`
- AND the resource's ID or any required ancestor ID is not set in state
- THEN the CLI MUST display the appropriate missing-state error
- AND exit with code 1

### Scenario: Publish root resource event

- GIVEN a root resource type (no parent) with its ID set in state
- WHEN the user runs `hf rabbitmq publish <resource-type> <exchange> [routing-key]`
- THEN the CLI MUST print the CloudEvent JSON to stdout
- AND publish the following CloudEvent 1.0 message via the RabbitMQ HTTP Management API:
  ```json
  {
    "specversion": "1.0",
    "type": "com.redhat.hyperfleet.<typeName>.reconcile.v1",
    "source": "/hyperfleet/service/sentinel",
    "id": "<resource-id>",
    "time": "<UTC ISO8601>",
    "datacontenttype": "application/json",
    "data": {
      "id": "<resource-id>",
      "kind": "<typeName>",
      "href": "{api-url}/api/hyperfleet/{api-version}/{path}/{resource-id}"
    }
  }
  ```
- AND `<typeName>` MUST be the `resource-types` key verbatim (plural, e.g. `clusters`)
- AND the `href` MUST be constructed from the configured `api-url` and `api-version`
- AND print `[INFO] Published <typeName> <id> to exchange <exchange>` on success

### Scenario: Publish child resource event

- GIVEN a child resource type with its ID and all ancestor IDs set in state
- WHEN the user runs `hf rabbitmq publish <resource-type> <exchange> [routing-key]`
- THEN the CloudEvent data MUST include `owner_references` for the immediate parent:
  ```json
  {
    "id": "<resource-id>",
    "kind": "<typeName>",
    "href": "...",
    "owner_references": {
      "id": "<parent-id>",
      "kind": "<parentTypeName>",
      "href": "{api-url}/api/hyperfleet/{api-version}/{parent-path}/{parent-id}"
    }
  }
  ```
- AND `owner_references` MUST reference only the immediate parent (not the full ancestor chain)
