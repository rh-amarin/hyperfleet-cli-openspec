# Cluster Lifecycle Specification

## Purpose

Provide CLI commands for full CRUD lifecycle management of HyperFleet clusters, including creation, retrieval, listing, searching, patching, and deletion. All cluster operations interact with the HyperFleet API at `/api/hyperfleet/v1/clusters`.
## Requirements
### Requirement: Create Cluster

The CLI SHALL create a new HyperFleet cluster with configurable name, region, and version.

#### Scenario: Create cluster with explicit arguments

- GIVEN the API is reachable and api-url, api-version are configured
- WHEN the user runs `hf cluster create <name> [region] [version]`
- THEN the CLI MUST send a POST request to `/api/hyperfleet/v1/clusters`
- AND the request payload MUST include:
  - `name`: the provided cluster name
  - `labels`: `{"counter": "1", "environment": "development", "shard": "1", "team": "core"}`
  - `spec`: `{"counter": "1", "region": "<region>", "version": "<version>"}`
- AND the CLI MUST output the JSON response containing the created cluster object
- AND the response MUST include `id`, `kind: "Cluster"`, `generation: 1`, `status.conditions`
- AND the CLI MUST persist the cluster ID from the API response to active state via `config.SetClusterID`
- AND the CLI MUST print `[INFO] Cluster context set to '<id>'` on stderr after persisting

**Example** — `hf cluster create test-cluster-alpha us-east1 1.28`:

Request payload:
```json
{
  "kind": "Cluster",
  "name": "test-cluster-alpha",
  "labels": {"counter": "1", "environment": "development", "shard": "1", "team": "core"},
  "spec": {"counter": "1", "region": "us-east1", "version": "1.28"}
}
```

Response:
```json
{
  "id": "019dc049-43a8-7a42-b44a-8d7f89e9e10f",
  "kind": "Cluster",
  "generation": 1,
  "name": "test-cluster-alpha",
  "labels": {"counter": "1", "environment": "development", "shard": "1", "team": "core"},
  "spec": {"counter": "1", "region": "us-east1", "version": "1.28"},
  "status": {
    "conditions": [
      {
        "type": "Available",
        "status": "False",
        "reason": "AdaptersNotAtSameGeneration",
        "message": "Required adapters do not report a consistent Available state",
        "observed_generation": 1
      },
      {
        "type": "Reconciled",
        "status": "False",
        "reason": "MissingRequiredAdapters",
        "message": "Required adapters not reporting Available=True: [cl-deployment, cl-invalid-resource, cl-job, cl-maestro, cl-namespace, cl-precondition-error]. Currently reporting: []",
        "observed_generation": 1
      }
    ]
  },
  "created_by": "user@example.com",
  "created_time": "2026-04-24T16:00:00Z",
  "href": "/api/hyperfleet/v1/clusters/019dc049-43a8-7a42-b44a-8d7f89e9e10f"
}
```

Stderr: `[INFO] Cluster context set to '019dc049-43a8-7a42-b44a-8d7f89e9e10f'`

#### Scenario: Create cluster with default arguments

- GIVEN no arguments are provided
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST use defaults: name=`my-cluster`, region=`us-east-1`, version=`4.15.0`
- AND the CLI MUST NOT show a usage message — it MUST proceed with creation using defaults

#### Scenario: Create duplicate cluster

- GIVEN a cluster with the same name already exists
- WHEN the user runs `hf cluster create <existing-name>`
- THEN the CLI MUST first query `GET /api/hyperfleet/v1/clusters?search=name='<name>'` to check for an existing cluster
- AND if a cluster with that name is found, the CLI MUST print `[WARN] Cluster '<name>' already exists, skipping creation` and exit with code 0 without sending a POST
- AND if no existing cluster is found, proceed with the POST request normally

**Example** — `hf cluster create test-cluster-alpha` when `test-cluster-alpha` already exists:
```
[WARN] Cluster 'test-cluster-alpha' already exists, skipping creation
```

#### Scenario: Initial cluster status conditions

- GIVEN a cluster was just created
- WHEN the API responds with the created cluster
- THEN the cluster MUST have initial conditions:
  - `Reconciled: False` with reason `MissingRequiredAdapters`
  - `Available: False` with reason `AdaptersNotAtSameGeneration`

### Requirement: Search Cluster

The CLI SHALL search for clusters by name and set the found cluster as the current context.

#### Scenario: Search with no arguments

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster search` with no arguments
- THEN the CLI MUST behave identically to `hf cluster get` — fetching and returning the current cluster from state

#### Scenario: Search with no arguments and no cluster in state

- GIVEN no cluster-id is set in state
- WHEN the user runs `hf cluster search` with no arguments
- THEN the CLI MUST display `[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.`
- AND exit with code 1

#### Scenario: Search for existing cluster

- GIVEN clusters exist in the API
- WHEN the user runs `hf cluster search <name>`
- THEN the CLI MUST query the API filtering by name
- AND output the matching clusters as a JSON array of full Cluster objects
- AND persist the found cluster's ID to active state via `config.SetClusterID`
- AND print `[INFO] Cluster context set to '<id>'` on stderr after persisting

#### Scenario: Search for non-existent cluster

- GIVEN no cluster matches the search name
- WHEN the user runs `hf cluster search <name>`
- THEN the CLI MUST display `[WARN] No clusters found matching '<name>'`
- AND output an empty JSON array `[]`
- AND exit with code 0

#### Scenario: Multiple matches

- GIVEN multiple clusters match the search name
- WHEN the user runs `hf cluster search <name>`
- THEN the CLI MUST display `[WARN] Multiple clusters found matching '<name>', using first result`
- AND set cluster-id to the first element in the returned `items` array
- AND persist that cluster-id to active state

### Requirement: Get Cluster

The CLI SHALL retrieve and display full details of a specific cluster.

#### Scenario: Get current cluster

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster get`
- THEN the CLI MUST send a GET request to `/api/hyperfleet/v1/clusters/{cluster_id}`
- AND output the full cluster JSON including: id, kind, name, generation, labels, spec, status.conditions, created_by, created_time, updated_by, updated_time, href

#### Scenario: Get cluster by explicit ID

- GIVEN a valid cluster ID is provided
- WHEN the user runs `hf cluster get <cluster_id>`
- THEN the CLI MUST use the provided ID instead of the configured cluster-id

#### Scenario: Get non-existent cluster

- GIVEN an invalid or non-existent cluster ID is used
- WHEN the user runs `hf cluster get <invalid_id>`
- THEN the CLI MUST output the API error response (RFC 7807 format)
- AND the error MUST contain code `HYPERFLEET-NTF-001`, status 404, title `Resource Not Found`
- AND the CLI MUST exit with code 0
- NOTE: Exit code 0 for API errors is intentional to maintain backwards compatibility with the original shell scripts. All API errors exit 0 and output the error JSON. See `errors-and-usage/spec.md` and `technical-architecture/spec.md` Error Handling Strategy.

**Example** — `hf cluster get 00000000-0000-0000-0000-000000000000`:
```json
{
  "code": "HYPERFLEET-NTF-001",
  "detail": "Cluster with id='00000000-0000-0000-0000-000000000000' not found",
  "status": 404,
  "title": "Resource Not Found",
  "type": "https://api.hyperfleet.io/errors/not-found"
}
```

### Requirement: Patch Cluster

The CLI SHALL increment a counter field in the cluster's spec or labels section, triggering a generation bump.

#### Scenario: Patch spec counter

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster patch spec`
- THEN the CLI MUST fetch the current cluster
- AND read the current `spec.counter` value as an integer (if absent, treat as `0`)
- AND increment it by 1
- AND send a PATCH to `/api/hyperfleet/v1/clusters/{cluster_id}` with the incremented counter as a string (e.g., `"2"`)
- AND display `[INFO] Incrementing spec.counter: <old> -> <new>` where `<old>` and `<new>` are integer strings (e.g., `1 -> 2`; first increment displays `0 -> 1`)
- AND the cluster's generation MUST increment

**Example** — `hf cluster patch spec` when current `spec.counter` is `"1"`:
```
[INFO] Incrementing spec.counter: 1 -> 2
```

#### Scenario: Patch labels counter

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster patch labels`
- THEN the CLI MUST fetch the current cluster
- AND read the current `labels.counter` value as an integer (if absent, treat as `0`)
- AND increment it by 1
- AND send a PATCH to `/api/hyperfleet/v1/clusters/{cluster_id}` with the incremented counter as a string
- AND display `[INFO] Incrementing labels.counter: <old> -> <new>`
- AND the cluster's generation MUST increment

**Example** — `hf cluster patch labels` when current `labels.counter` is `"1"`:
```
[INFO] Incrementing labels.counter: 1 -> 2
```

#### Scenario: Patch with no arguments

- GIVEN the user provides no arguments
- WHEN the user runs `hf cluster patch`
- THEN the CLI MUST display usage: `Usage: hf cluster patch {spec|labels} [cluster_id]`
- AND exit with code 1

**Example** — `hf cluster patch`:
```
Usage: hf.cluster.patch.sh spec|labels [cluster_id]

Arguments:
  spec|labels   Which section to increment the counter field in (required)
  cluster_id    Cluster ID (default: current cluster)
```

### Requirement: Delete Cluster

The CLI SHALL delete a cluster by ID.

#### Scenario: Delete cluster

- GIVEN a cluster exists
- WHEN the user runs `hf cluster delete [cluster_id]`
- THEN the CLI MUST send a DELETE request to `/api/hyperfleet/v1/clusters/{cluster_id}`
- AND the response MUST include the full cluster object with `deleted_by`, `deleted_time`, and incremented `generation`
- AND the CLI MUST output the deleted cluster object subject to the `--output` flag (default: JSON)

#### Scenario: Delete current cluster

- GIVEN a cluster-id is set in config and no explicit ID is provided
- WHEN the user runs `hf cluster delete`
- THEN the CLI MUST use the configured cluster-id

### Requirement: Get Cluster Conditions

The CLI SHALL display the generation and status conditions of a cluster.

#### Scenario: Get conditions

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster conditions`
- THEN the CLI MUST fetch the cluster from `/api/hyperfleet/v1/clusters/{cluster_id}`
- AND extract only `generation` and `status.conditions`
- AND output them as JSON

**Example** — `hf cluster conditions` immediately after creation (no adapters yet):
```json
{
  "generation": 1,
  "status": {
    "conditions": [
      {
        "type": "Available",
        "status": "False",
        "reason": "AdaptersNotAtSameGeneration",
        "message": "Required adapters do not report a consistent Available state",
        "observed_generation": 1
      },
      {
        "type": "Reconciled",
        "status": "False",
        "reason": "MissingRequiredAdapters",
        "message": "Required adapters not reporting Available=True: [cl-deployment, cl-invalid-resource, cl-job, cl-maestro, cl-namespace, cl-precondition-error]. Currently reporting: []",
        "observed_generation": 1
      }
    ]
  }
}
```

### Requirement: Get Cluster Conditions Table

The CLI SHALL display cluster conditions in a formatted table via the `--table` flag.

#### Scenario: Display conditions table

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster conditions --table`
- THEN the CLI MUST output a table with columns: TYPE, STATUS, LAST TRANSITION, REASON, MESSAGE
- AND status values MUST be color-coded: True=green, False=red, Unknown=yellow

**Example** — `hf cluster conditions --table` before any adapters report:
```
TYPE        STATUS  LAST TRANSITION      REASON                       MESSAGE
---         ---     ---                  ---                          ---
Available   False   2026-04-24T16:00:00Z AdaptersNotAtSameGeneration  Required adapters do not report a consistent Available state
Reconciled  False   2026-04-24T16:00:00Z MissingRequiredAdapters      Required adapters not reporting Available=True: [cl-deployment, ...]. Currently reporting: []
```

**Example** — `hf cluster conditions --table` after some adapters report (partial convergence):
```
TYPE                    STATUS  LAST TRANSITION      REASON               MESSAGE
---                     ---     ---                  ---                  ---
Available               False   2026-04-24T16:01:00Z AdaptersNotAtSame... ...
Reconciled              False   2026-04-24T16:01:00Z MissingRequired...   ...
ClDeploymentSuccessful  True    2026-04-24T16:01:00Z ManualStatusPost      Status posted via hf.adapter.status.sh
ClJobSuccessful         False   2026-04-24T16:01:00Z ManualStatusPost      Status posted via hf.adapter.status.sh
ClNamespaceSuccessful   True    2026-04-24T16:01:00Z ManualStatusPost      Status posted via hf.adapter.status.sh
```

### Requirement: Get Cluster Adapter Statuses

The CLI SHALL display adapter statuses for a cluster.

#### Scenario: Get statuses with no adapter reports

- GIVEN a newly created cluster with no adapter status reports
- WHEN the user runs `hf cluster statuses`
- THEN the CLI MUST output `{"items": [], "kind": "AdapterStatusList", "page": 1, "size": 0, "total": 0}`

#### Scenario: Get statuses with adapter reports

- GIVEN adapters have reported statuses for the cluster
- WHEN the user runs `hf cluster statuses`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters/{cluster_id}/statuses`
- AND output the `AdapterStatusList` with items containing: adapter name, conditions (Available, Applied, Health, Finalized), observed_generation, last_report_time

**Example** — `hf cluster statuses` after three adapters have reported at generation 3:
```json
{
  "items": [
    {
      "adapter": "cl-deployment",
      "observed_generation": 3,
      "last_report_time": "2026-04-24T16:19:06Z",
      "conditions": [
        {"type": "Available", "status": "True",  "reason": "ManualStatusPost"},
        {"type": "Applied",   "status": "True",  "reason": "ManualStatusPost"},
        {"type": "Health",    "status": "True",  "reason": "ManualStatusPost"},
        {"type": "Finalized", "status": "True",  "reason": "ManualStatusPost"}
      ]
    },
    {
      "adapter": "cl-job",
      "observed_generation": 3,
      "last_report_time": "2026-04-24T16:19:08Z",
      "conditions": [
        {"type": "Available", "status": "False", "reason": "ManualStatusPost"},
        {"type": "Applied",   "status": "False", "reason": "ManualStatusPost"},
        {"type": "Health",    "status": "False", "reason": "ManualStatusPost"},
        {"type": "Finalized", "status": "False", "reason": "ManualStatusPost"}
      ]
    },
    {
      "adapter": "cl-namespace",
      "observed_generation": 3,
      "last_report_time": "2026-04-24T16:19:10Z",
      "conditions": [
        {"type": "Available", "status": "True",  "reason": "ManualStatusPost"},
        {"type": "Applied",   "status": "True",  "reason": "ManualStatusPost"},
        {"type": "Health",    "status": "True",  "reason": "ManualStatusPost"},
        {"type": "Finalized", "status": "True",  "reason": "ManualStatusPost"}
      ]
    }
  ],
  "kind": "AdapterStatusList",
  "page": 1,
  "size": 3,
  "total": 3
}
```

#### Scenario: Get statuses table

- GIVEN adapters have reported statuses for the cluster
- WHEN the user runs `hf cluster statuses --table`
- THEN the CLI MUST output a formatted table with columns: ADAPTER, GEN, Available, Finalized
- AND each row MUST represent one adapter entry from the statuses list
- AND GEN MUST show the `observed_generation` value for that adapter
- AND Available and Finalized columns MUST be color-coded dots: green=True, red=False, yellow=Unknown, `-`=not present

**Example** — `hf cluster statuses --table` for the same three adapters above (colors shown in parentheses):
```
ADAPTER       GEN  Available  Finalized
---           ---  ---        ---
cl-deployment  3   ●(green)   ●(green)
cl-job         3   ●(red)     ●(red)
cl-namespace   3   ●(green)   ●(green)
```

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

