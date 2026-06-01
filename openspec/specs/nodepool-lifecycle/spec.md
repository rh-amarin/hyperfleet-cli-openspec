# NodePool Lifecycle Specification

## Purpose

Provide CLI commands for full CRUD lifecycle management of HyperFleet nodepools. Nodepools are always scoped to a parent cluster, requiring a `cluster-id` to be set in config. All nodepool operations interact with the HyperFleet API at `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools`.

## Prerequisites

**cluster-id required**: All nodepool commands require `cluster-id` to be set in state. If it is not set, the CLI MUST display:
```
[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.
```
AND exit with code 1 before making any API call.

**nodepool-id required for single-resource commands**: `hf nodepool patch`, `hf nodepool delete`, `hf nodepool conditions`, and `hf nodepool statuses` additionally require `nodepool-id` to be set in state (unless an explicit ID argument is provided). If cluster-id is set but nodepool-id is not, the CLI MUST display:
```
[ERROR] No nodepool-id set in state. Run 'hf nodepool create' or 'hf nodepool search <name>' first.
```
AND exit with code 1.
## Requirements
### Requirement: NodePool Lifecycle via hf rs nodepools

The CLI SHALL expose all nodepool lifecycle operations under `hf rs nodepools` (alias `hf resource nodepools`). API paths, request bodies, and response handling SHALL match the behaviors previously normative for `hf nodepool` commands, as specified in `rs-entity-commands`.

#### Scenario: hf nodepool command removed

- GIVEN the rs-entity-commands change is complete
- WHEN the user runs `hf nodepool list`
- THEN the CLI MUST NOT register `hf nodepool`
- AND the user MUST use `hf rs nodepools list` instead

