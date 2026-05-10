## ADDED Requirements

### Requirement: List Clusters

The CLI SHALL list all clusters via GET /clusters.

#### Scenario: List clusters as JSON

- GIVEN an active environment is configured
- WHEN the user runs `hf cluster list`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters`
- AND output the full `ListResponse[Cluster]` as pretty-printed JSON

#### Scenario: List clusters as table

- GIVEN an active environment is configured
- WHEN the user runs `hf cluster list --output table`
- THEN the CLI MUST output a table with columns: ID, NAME, GEN, STATUS
- AND STATUS MUST be derived from conditions: green dot if Available=True AND Reconciled=True, otherwise red dot (or plain text in no-color mode)

### Requirement: Update Cluster

The CLI SHALL update a cluster's fields via PATCH.

#### Scenario: Update cluster name

- GIVEN a valid cluster ID and --name flag
- WHEN the user runs `hf cluster update <id> --name <new-name>`
- THEN the CLI MUST send PATCH to `/api/hyperfleet/v1/clusters/<id>` with `{"name": "<new-name>"}`
- AND output the updated Cluster as JSON

#### Scenario: Update cluster replicas

- GIVEN a valid cluster ID and --replicas flag
- WHEN the user runs `hf cluster update <id> --replicas <n>`
- THEN the CLI MUST send PATCH to `/api/hyperfleet/v1/clusters/<id>` with `{"spec": {"replicas": "<n>"}}`
- AND output the updated Cluster as JSON
