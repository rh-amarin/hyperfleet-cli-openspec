# adapter-status-commands Specification

## Purpose
TBD - created by archiving change phase-5-tables-adapter. Update Purpose after archive.
## Requirements
### Requirement: Post Cluster Adapter Status via CLI

The CLI SHALL provide a command to post adapter status to a cluster.

#### Scenario: Post True status to cluster

- GIVEN a cluster-id is set in state
- WHEN the user runs `hf cluster adapter post-status <adapter_name> True <generation>`
- THEN the CLI MUST send POST to `/api/hyperfleet/{version}/clusters/{cluster_id}/statuses`
- AND the request body MUST include adapter, observed_generation, observed_time, and 4 conditions (Available, Applied, Health, Finalized) all with status True

#### Scenario: Post False status to cluster

- GIVEN a cluster-id is set in state
- WHEN the user runs `hf cluster adapter post-status <adapter_name> False <generation>`
- THEN all 4 condition statuses MUST be set to False

#### Scenario: Reject invalid status value

- GIVEN an invalid status value
- WHEN the user runs `hf cluster adapter post-status <adapter> INVALID 1`
- THEN the CLI MUST output `[ERROR] Invalid status value 'INVALID'. Must be one of: True, False, Unknown.`
- AND exit with code 1 without making any HTTP request

### Requirement: Post NodePool Adapter Status via CLI

The CLI SHALL provide a command to post adapter status to a nodepool.

#### Scenario: Post True status to nodepool

- GIVEN cluster-id and nodepool-id are set in state
- WHEN the user runs `hf nodepool adapter post-status <adapter_name> True <generation>`
- THEN the CLI MUST send POST to `/api/hyperfleet/{version}/clusters/{cluster_id}/nodepools/{nodepool_id}/statuses`
- AND the request body MUST follow the same structure as cluster adapter status posting

#### Scenario: Override nodepool ID via positional arg

- GIVEN cluster-id is set in state
- WHEN the user runs `hf nodepool adapter post-status <adapter> True 1 <explicit_nodepool_id>`
- THEN the CLI MUST use the explicit 4th argument as the nodepool ID

