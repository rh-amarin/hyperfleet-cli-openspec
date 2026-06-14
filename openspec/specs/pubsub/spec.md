# Pub/Sub and Messaging Specification

## Purpose

Provide CLI commands for publishing CloudEvent messages to GCP Pub/Sub topics and RabbitMQ exchanges, and for listing Pub/Sub topics and subscriptions. These events trigger adapter reconciliation in the HyperFleet system.

## CloudEvent href Convention

All `href` values in CloudEvent payloads MUST be constructed from the configured `api-url` and `api-version` using the same pattern as the API client:

```
{hyperfleet.api-url}/api/hyperfleet/{hyperfleet.api-version}/{resource-path}
```

Examples:
- Cluster: `{api-url}/api/hyperfleet/{api-version}/clusters/{cluster_id}`
- NodePool: `{api-url}/api/hyperfleet/{api-version}/clusters/{cluster_id}/nodepools/{nodepool_id}`

This ensures hrefs are always consistent with the configured environment and never contain hardcoded hostnames.

## Requirements

### Requirement: List Pub/Sub Topics and Subscriptions

The CLI SHALL list GCP Pub/Sub topics and their subscriptions.

#### Scenario: Missing GCP credentials

- GIVEN GCP credentials are not configured (no `GOOGLE_APPLICATION_CREDENTIALS`, no gcloud ADC, no GCE metadata)
- WHEN the user runs `hf pubsub list` or any `hf pubsub publish` command
- THEN the CLI MUST display: `[ERROR] GCP credentials not found. Run 'gcloud auth application-default login' or set GOOGLE_APPLICATION_CREDENTIALS`
- AND exit with code 1
- AND other `hf` commands MUST continue to work normally

#### Scenario: List all topics

- GIVEN gcp-project is configured
- WHEN the user runs `hf pubsub list`
- THEN the CLI MUST print an `[INFO]` line identifying the project
- AND list all Pub/Sub topics in the configured GCP project
- AND for each topic, list its subscriptions indented beneath it
- AND topics MUST appear at the left margin
- AND subscriptions MUST be indented with 4 spaces

#### Scenario: List with no topics

- GIVEN gcp-project is configured but no topics exist in the project
- WHEN the user runs `hf pubsub list`
- THEN the CLI MUST print the `[INFO]` project line
- AND print `No topics found.`
- AND exit with code 0

#### Scenario: List with filter

- GIVEN gcp-project is configured
- WHEN the user runs `hf pubsub list <filter_term>`
- THEN the CLI MUST print `[INFO] Filtering by: <filter_term>`
- AND filter both topics AND subscriptions by the provided substring
- AND only show topics/subscriptions whose name contains the term

### Requirement: Publish Cluster Change Event to Pub/Sub

The CLI SHALL publish a cluster reconcile event to a GCP Pub/Sub topic.

#### Scenario: Publish cluster event — no cluster-id in state

- GIVEN no cluster-id is set in state
- WHEN the user runs `hf pubsub publish cluster <topic>`
- THEN the CLI MUST display `[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.`
- AND exit with code 1

#### Scenario: Publish cluster event

- GIVEN gcp-project and cluster-id are configured
- WHEN the user runs `hf pubsub publish cluster <topic>`
- THEN the CLI MUST print the CloudEvent JSON to stdout
- AND publish the following CloudEvent 1.0 message to the specified topic via the GCP Pub/Sub SDK:
  ```json
  {
    "specversion": "1.0",
    "type": "com.redhat.hyperfleet.cluster.reconcile.v1",
    "source": "/hyperfleet/service/sentinel",
    "id": "<cluster_id>",
    "time": "<UTC ISO8601>",
    "datacontenttype": "application/json",
    "data": {
      "id": "<cluster_id>",
      "kind": "Cluster",
      "href": "{api-url}/api/hyperfleet/{api-version}/clusters/<cluster_id>"
    }
  }
  ```
- AND the cluster-id MUST be read from state (no HyperFleet API fetch)
- AND the href MUST be constructed using the configured `api-url` and `api-version`
- AND print `[INFO] Published cluster <id> to topic <topic> (msg-id: <id>)` on success
- AND on publish failure display `[ERROR] Failed to publish: <error>` on stderr and exit 1

### Requirement: Publish NodePool Change Event to Pub/Sub

The CLI SHALL publish a nodepool reconcile event to a GCP Pub/Sub topic.

#### Scenario: Publish nodepool event — missing state

- GIVEN no cluster-id or no nodepool-id is set in state
- WHEN the user runs `hf pubsub publish nodepool <topic>`
- THEN the CLI MUST display the appropriate missing-state error (`[ERROR] No cluster-id ...` or `[ERROR] No nodepool-id ...`)
- AND exit with code 1

#### Scenario: Publish nodepool event

- GIVEN gcp-project, cluster-id, and nodepool-id are configured
- WHEN the user runs `hf pubsub publish nodepool <topic>`
- THEN the CLI MUST print the CloudEvent JSON to stdout
- AND publish the following CloudEvent 1.0 message to the specified topic:
  ```json
  {
    "specversion": "1.0",
    "type": "com.redhat.hyperfleet.nodepool.reconcile.v1",
    "source": "/hyperfleet/service/sentinel",
    "id": "<nodepool_id>",
    "time": "<UTC ISO8601>",
    "datacontenttype": "application/json",
    "data": {
      "id": "<nodepool_id>",
      "kind": "NodePool",
      "href": "{api-url}/api/hyperfleet/{api-version}/clusters/<cluster_id>/nodepools/<nodepool_id>",
      "owner_references": {
        "id": "<cluster_id>",
        "kind": "Cluster",
        "href": "{api-url}/api/hyperfleet/{api-version}/clusters/<cluster_id>"
      }
    }
  }
  ```
- AND both cluster-id and nodepool-id MUST be read from state (no HyperFleet API fetch)
- AND all hrefs MUST be constructed using the configured `api-url` and `api-version`

### Requirement: Publish Any Resource Reconcile Event to RabbitMQ

The CLI SHALL publish a reconcile CloudEvent for any resource type defined in the active environment's `resource-types` configuration.

#### Scenario: Unknown resource type

- GIVEN an active environment with `resource-types` configured
- WHEN the user runs `hf rabbitmq publish <resource-type> <exchange>` where `<resource-type>` is not defined in `resource-types`
- THEN the CLI MUST display `[ERROR] unknown resource type "<resource-type>"`
- AND exit with code 1

#### Scenario: Missing state for resource

- GIVEN an active environment with `resource-types` configured
- WHEN the user runs `hf rabbitmq publish <resource-type> <exchange>`
- AND the resource's ID or any required ancestor ID is not set in state
- THEN the CLI MUST display the appropriate missing-state error
- AND exit with code 1

#### Scenario: Publish root resource event

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
- AND the vhost MUST be URL-encoded (`/` becomes `%2F`)
- AND routing-key defaults to empty string when not provided
- AND print `[INFO] Published <typeName> <id> to exchange <exchange>` on success

#### Scenario: Publish child resource event

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

