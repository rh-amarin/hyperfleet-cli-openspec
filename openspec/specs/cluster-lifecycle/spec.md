# Cluster Lifecycle Specification

## Purpose

Provide CLI commands for full CRUD lifecycle management of HyperFleet clusters, including creation, retrieval, listing, searching, patching, and deletion. All cluster operations interact with the HyperFleet API at `/api/hyperfleet/v1/clusters`.
## Requirements
### Requirement: Create Cluster

`hf cluster create` SHALL additionally accept positional arguments `[name] [region] [version]`.

#### Scenario: Create cluster with positional arguments (MODIFIED)

- WHEN the user runs `hf cluster create <name> [region] [version]`
- THEN the CLI MUST use positional args for name, region, and version
- AND positional args take precedence over the `--name` flag and built-in defaults (`my-cluster`, `us-east-1`, `4.15.0`)

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

`hf cluster delete` SHALL accept an optional cluster ID, falling back to the configured cluster-id.

#### Scenario: Delete cluster (MODIFIED — optional ID, output)

- WHEN the user runs `hf cluster delete [cluster_id]`
- THEN the CLI MUST use the provided ID, or the configured cluster-id if none is provided
- AND the CLI MUST output the deleted cluster JSON subject to the `--output` flag

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

The statuses table SHALL include a FINALIZED column in addition to AVAILABLE.

#### Scenario: Get statuses table (MODIFIED — add FINALIZED column)

- WHEN the user runs `hf cluster statuses --output table`
- THEN the CLI MUST output columns: ADAPTER, GEN, AVAILABLE, FINALIZED
- AND AVAILABLE and FINALIZED columns MUST be color-coded dots: green=True, red=False, `-`=not present

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

