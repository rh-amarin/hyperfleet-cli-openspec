# Adapter Status Specification

## Purpose

Provide CLI commands to simulate adapter status reporting for clusters and nodepools. These commands allow manual posting of adapter conditions to the HyperFleet API, enabling testing of the convergence logic without running real adapters.

## Requirements

### Requirement: Post Cluster Adapter Status

The CLI SHALL post adapter status conditions for the current cluster.

#### Scenario: Post status with True

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster adapter post-status <adapter_name> True <generation>`
- THEN the CLI MUST send POST to `/api/hyperfleet/v1/clusters/{cluster_id}/statuses`
- AND the request payload MUST include:
  - `adapter`: the adapter name (e.g., `cl-deployment`, `cl-namespace`)
  - `conditions`: an array of 4 conditions with types `Available`, `Applied`, `Health`, `Finalized`, all with status `True`
  - `observed_generation`: the provided generation
  - `observed_time`: current ISO8601 timestamp
- AND each condition MUST have `reason: "ManualStatusPost"` and `message: "Status posted via hf adapter post-status"`

**Example** — `hf cluster adapter post-status cl-deployment True 3`:

Request payload:
```json
{
  "adapter": "cl-deployment",
  "observed_generation": 3,
  "observed_time": "2026-04-24T16:19:06Z",
  "conditions": [
    {"type": "Available", "status": "True", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"},
    {"type": "Applied",   "status": "True", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"},
    {"type": "Health",    "status": "True", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"},
    {"type": "Finalized", "status": "True", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"}
  ]
}
```

Response (HTTP 200):
```json
{
  "adapter": "cl-deployment",
  "observed_generation": 3,
  "observed_time": "2026-04-24T16:19:06Z",
  "last_report_time": "2026-04-24T16:19:06Z",
  "conditions": [
    {"type": "Available", "status": "True", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"},
    {"type": "Applied",   "status": "True", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"},
    {"type": "Health",    "status": "True", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"},
    {"type": "Finalized", "status": "True", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"}
  ]
}
```

#### Scenario: Post status with False

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster adapter post-status <adapter_name> False <generation>`
- THEN all 4 condition statuses MUST be set to `False`

**Example** — `hf cluster adapter post-status cl-job False 3`: same payload shape with all `"status": "False"`.

#### Scenario: Post status with Unknown

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster adapter post-status <adapter_name> Unknown <generation>`
- THEN all 4 condition statuses MUST be set to `Unknown`
- AND the API returns HTTP 204 No Content; the CLI MUST handle this gracefully (exit 0, print empty object)

#### Scenario: Missing required arguments

- GIVEN no arguments are provided
- WHEN the user runs `hf cluster adapter post-status`
- THEN the CLI MUST display usage information
- AND exit with code 1

**Example** usage:
```
hf cluster adapter post-status <adapter_name> <True|False|Unknown> <generation>
```

#### Scenario: Invalid status value

- GIVEN an invalid status value is provided (not `True`, `False`, or `Unknown`)
- WHEN the user runs `hf cluster adapter post-status <adapter> <invalid>`
- THEN the CLI MUST output `[ERROR] Invalid status value '<value>'. Must be one of: True, False, Unknown.` to stderr
- AND exit with code 1 without making any HTTP request

#### Scenario: Output format

- GIVEN `hf cluster adapter post-status` or `hf nodepool adapter post-status` completes
- WHEN the API responds with HTTP 200 or HTTP 204
- THEN the CLI MUST output the API response subject to the `--output` flag (default: JSON)
- AND on HTTP 200, the CLI MUST output the full `AdapterStatus` JSON returned by the API
- AND on HTTP 204 (returned for `Unknown` status), the CLI MUST output an empty JSON object `{}`
- AND exit with code 0 in both cases

### Requirement: Post NodePool Adapter Status

The CLI SHALL post adapter status conditions for a nodepool.

#### Scenario: Post nodepool adapter status

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool adapter post-status <adapter_name> <True|False|Unknown> <generation> [nodepool_id]`
- THEN the CLI MUST send POST to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}/statuses`
- AND the payload MUST follow the same structure as cluster adapter status posting
- AND the adapter name for nodepools is typically `np-configmap`
- AND the optional `nodepool_id` 4th arg overrides the nodepool-id from state

#### Scenario: Nodepool convergence after all adapters report

- GIVEN a nodepool's only required adapter is `np-configmap`
- WHEN `np-configmap` reports `Available=True` at the nodepool's current generation
- THEN the nodepool's `Reconciled` condition MUST flip to `True`
- AND the `Available` condition MUST flip to `True`

### Requirement: Adapter Status Model

The system SHALL follow a defined convergence model for adapter statuses.

#### Scenario: Cluster convergence

- GIVEN a cluster with required adapters: `cl-deployment`, `cl-invalid-resource`, `cl-job`, `cl-maestro`, `cl-namespace`, `cl-precondition-error`
- WHEN ALL required adapters report `Available=True` at the cluster's current generation
- THEN the cluster's `Reconciled` condition MUST become `True`
- AND each adapter MUST generate a per-adapter condition named `<AdapterName>Successful` (e.g., `ClDeploymentSuccessful`)

#### Scenario: Nodepool convergence

- GIVEN a nodepool with required adapter: `np-configmap`
- WHEN ALL required adapters report `Available=True` at the nodepool's current generation
- THEN the nodepool's `Reconciled` condition MUST become `True`

#### Scenario: Partial adapter reporting

- GIVEN some but not all required adapters have reported
- WHEN the user queries conditions
- THEN `Reconciled` MUST remain `False` with reason `MissingRequiredAdapters`
- AND the message MUST list which adapters are missing

