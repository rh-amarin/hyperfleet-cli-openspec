# Cluster Lifecycle Specification

## Purpose

Provide CLI commands for full CRUD lifecycle management of HyperFleet clusters, including creation, retrieval, listing, searching, patching, and deletion. All cluster operations interact with the HyperFleet API at `/api/hyperfleet/v1/clusters`.
## Requirements
### Requirement: Cluster Lifecycle via hf rs clusters

The CLI SHALL expose all cluster lifecycle operations under `hf rs clusters` (alias `hf resource clusters`). API paths, request bodies, and response handling SHALL match the behaviors previously normative for `hf cluster` commands, as specified in `rs-entity-commands`.

#### Scenario: hf cluster command removed

- GIVEN the rs-entity-commands change is complete
- WHEN the user runs `hf cluster list`
- THEN the CLI MUST NOT register `hf cluster`
- AND the user MUST use `hf rs clusters list` instead

