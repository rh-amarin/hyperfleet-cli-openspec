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

### Requirement: Get Cluster by ID

The CLI SHALL retrieve a single cluster by explicit ID or from active state.

#### Scenario: Get cluster by explicit ID

- GIVEN a valid cluster ID is provided
- WHEN the user runs `hf cluster get <id>`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters/<id>`
- AND output the full Cluster object as JSON by default

#### Scenario: Get cluster from state

- GIVEN no ID arg is provided and cluster-id is in state.yaml
- WHEN the user runs `hf cluster get`
- THEN the CLI MUST use the state cluster-id

#### Scenario: Get non-existent cluster

- GIVEN an invalid ID is provided
- WHEN the API returns 404
- THEN the CLI MUST output the RFC 7807 error JSON and exit with code 0

### Requirement: Create Cluster

The CLI SHALL create a new cluster with optional flags, check for duplicates, and persist the new cluster-id.

#### Scenario: Create cluster with flags

- GIVEN the user provides optional --name, --replicas, --nodepool-id flags
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST first GET `/api/hyperfleet/v1/clusters?search=name='<name>'` to check for duplicates
- AND if no duplicate, POST to `/api/hyperfleet/v1/clusters` with name, labels, spec
- AND persist the new cluster ID via `cfgStore.SetState("cluster-id", id)`
- AND print `[INFO] Cluster context set to '<id>'` on stderr

#### Scenario: Create cluster with defaults

- GIVEN no flags are provided
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST use defaults: name=`my-cluster`, region=`us-east-1`, version=`4.15.0`
- AND MUST NOT show usage

#### Scenario: Duplicate cluster creation

- GIVEN a cluster with the same name already exists
- WHEN the duplicate check returns a non-empty items list
- THEN the CLI MUST print `[WARN] Cluster '<name>' already exists, skipping creation` on stderr
- AND return nil (exit 0) without POSTing

### Requirement: Update Cluster

The CLI SHALL update a cluster's name and/or replicas via PATCH.

#### Scenario: Update cluster

- GIVEN a valid cluster ID and at least one flag (--name or --replicas)
- WHEN the user runs `hf cluster update <id>`
- THEN the CLI MUST send PATCH to `/api/hyperfleet/v1/clusters/<id>` with a JSON body containing only the provided fields
- AND output the updated Cluster as JSON

### Requirement: Delete Cluster

The CLI SHALL delete a cluster by ID with silent success.

#### Scenario: Delete existing cluster

- GIVEN a valid cluster ID
- WHEN the user runs `hf cluster delete <id>`
- THEN the CLI MUST send DELETE to `/api/hyperfleet/v1/clusters/<id>`
- AND produce no output on success (silent)

#### Scenario: Delete non-existent cluster

- GIVEN a non-existent cluster ID
- WHEN the API returns 404
- THEN the CLI MUST output the RFC 7807 error JSON and exit 0

### Requirement: Get Cluster Conditions

The CLI SHALL display cluster status conditions in JSON or table format.

#### Scenario: Get conditions as JSON

- GIVEN a valid cluster ID
- WHEN the user runs `hf cluster conditions <id>`
- THEN the CLI MUST GET `/api/hyperfleet/v1/clusters/<id>`
- AND output `{"generation": N, "status": {"conditions": [...]}}` as JSON

#### Scenario: Get conditions as table

- GIVEN a valid cluster ID
- WHEN the user runs `hf cluster conditions <id> --output table`
- THEN the CLI MUST output a table with columns: TYPE, STATUS, LAST TRANSITION, REASON, MESSAGE

### Requirement: Get Cluster Adapter Statuses

The CLI SHALL display adapter status reports for a cluster.

#### Scenario: Get statuses as JSON

- GIVEN a valid cluster ID
- WHEN the user runs `hf cluster statuses <id>`
- THEN the CLI MUST GET `/api/hyperfleet/v1/clusters/<id>/statuses`
- AND output the `ListResponse[AdapterStatus]` as JSON

#### Scenario: Get statuses as table

- GIVEN a valid cluster ID
- WHEN the user runs `hf cluster statuses <id> --output table`
- THEN the CLI MUST output a table with columns: ADAPTER, GEN, AVAILABLE
- AND AVAILABLE MUST show the status of the Available condition for each adapter
