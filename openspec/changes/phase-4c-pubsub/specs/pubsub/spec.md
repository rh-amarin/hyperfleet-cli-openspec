# Pub/Sub and Messaging — Delta Spec

## ADDED Requirements

### Requirement: List Pub/Sub Topics and Subscriptions

The CLI SHALL list GCP Pub/Sub topics and their subscriptions.

#### Scenario: Missing GCP credentials

- GIVEN GCP credentials are not configured (no `GOOGLE_APPLICATION_CREDENTIALS`, no gcloud ADC, no GCE metadata)
- WHEN the user runs `hf pubsub list` or any `hf pubsub publish` command
- THEN the CLI MUST display: `[ERROR] GCP credentials not found. Run 'gcloud auth application-default login' or set GOOGLE_APPLICATION_CREDENTIALS`
- AND exit with code 1

#### Scenario: List all topics

- GIVEN gcp-project is configured
- WHEN the user runs `hf pubsub list`
- THEN the CLI MUST print an `[INFO]` line identifying the project
- AND list all Pub/Sub topics with subscriptions indented 4 spaces beneath each topic

#### Scenario: List with no topics

- GIVEN gcp-project is configured but no topics exist
- WHEN the user runs `hf pubsub list`
- THEN the CLI MUST print `No topics found.` and exit with code 0

#### Scenario: List with filter

- GIVEN gcp-project is configured
- WHEN the user runs `hf pubsub list <filter_term>`
- THEN the CLI MUST print `[INFO] Filtering by: <filter_term>`
- AND only show topics/subscriptions whose name contains the term

### Requirement: Publish Cluster Change Event to Pub/Sub

The CLI SHALL publish a cluster reconcile CloudEvent 1.0 to a GCP Pub/Sub topic.

#### Scenario: Publish cluster event — no cluster-id in state

- GIVEN no cluster-id is set in state
- WHEN the user runs `hf pubsub publish cluster <topic>`
- THEN the CLI MUST display `[ERROR] No cluster-id set in state.` and exit with code 1

#### Scenario: Publish cluster event

- GIVEN gcp-project and cluster-id are configured
- WHEN the user runs `hf pubsub publish cluster <topic>`
- THEN the CLI MUST print the CloudEvent JSON to stdout and publish it to the specified topic
- AND print `[INFO] Published cluster <id> to topic <topic> (msg-id: <id>)` on success

### Requirement: Publish NodePool Change Event to Pub/Sub

The CLI SHALL publish a nodepool reconcile CloudEvent 1.0 to a GCP Pub/Sub topic.

#### Scenario: Publish nodepool event — missing state

- GIVEN no cluster-id or no nodepool-id is set in state
- WHEN the user runs `hf pubsub publish nodepool <topic>`
- THEN the CLI MUST display the appropriate missing-state error and exit with code 1

#### Scenario: Publish nodepool event

- GIVEN gcp-project, cluster-id, and nodepool-id are configured
- WHEN the user runs `hf pubsub publish nodepool <topic>`
- THEN the CLI MUST print the CloudEvent JSON to stdout and publish it to the specified topic

### Requirement: Publish Cluster Change Event to RabbitMQ

The CLI SHALL publish a cluster reconcile event to a RabbitMQ exchange via the HTTP Management API.

#### Scenario: Publish cluster event to RabbitMQ

- GIVEN rabbitmq config and cluster-id are set
- WHEN the user runs `hf rabbitmq publish cluster <exchange> [routing-key]`
- THEN the CLI MUST print the CloudEvent JSON to stdout
- AND POST to `http://{host}:{mgmt-port}/api/exchanges/{vhost-encoded}/{exchange}/publish`
- AND vhost MUST be URL-encoded (`/` becomes `%2F`)
- AND routing-key defaults to empty string when not provided

### Requirement: Publish NodePool Change Event to RabbitMQ

The CLI SHALL publish a nodepool reconcile event to a RabbitMQ exchange via the HTTP Management API.

#### Scenario: Publish nodepool event to RabbitMQ

- GIVEN rabbitmq config, cluster-id, and nodepool-id are set
- WHEN the user runs `hf rabbitmq publish nodepool <exchange> [routing-key]`
- THEN the CLI MUST print the CloudEvent JSON to stdout and POST to the exchange endpoint
