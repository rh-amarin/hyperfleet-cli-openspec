# Cluster and NodePool ID Commands Specification

## Purpose

Provide `hf cluster id` and `hf nodepool id` commands that print the currently active cluster or nodepool ID from local state. These commands are scriptable, fast, and require no API call.

## Requirements

### Requirement: `hf cluster id`

The CLI SHALL implement `hf cluster id` to print the active cluster ID.

#### Scenario: Cluster ID is set

- **GIVEN** `cluster-id` is set in the active state
- **WHEN** the user runs `hf cluster id`
- **THEN** the CLI MUST print the cluster ID to stdout and exit 0

#### Scenario: Cluster ID is not set

- **GIVEN** no `cluster-id` is set in the active state
- **WHEN** the user runs `hf cluster id`
- **THEN** the CLI MUST print `[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.` and exit 1

---

### Requirement: `hf nodepool id`

The CLI SHALL implement `hf nodepool id` to print the active nodepool ID.

#### Scenario: NodePool ID is set

- **GIVEN** `nodepool-id` is set in the active state
- **WHEN** the user runs `hf nodepool id`
- **THEN** the CLI MUST print the nodepool ID to stdout and exit 0

#### Scenario: NodePool ID is not set

- **GIVEN** no `nodepool-id` is set in the active state
- **WHEN** the user runs `hf nodepool id`
- **THEN** the CLI MUST print `[ERROR] No nodepool-id set in state. Run 'hf nodepool create' or 'hf nodepool search <name>' first.` and exit 1

---

### Requirement: Nodepool API paths are cluster-scoped

All `hf nodepool` subcommands SHALL send requests to the cluster-scoped endpoint `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/...`, not the flat `/api/hyperfleet/v1/nodepools/...` endpoint.

#### Scenario: Create uses cluster-scoped path

- **GIVEN** `cluster-id` is set in state
- **WHEN** the user runs `hf nodepool create`
- **THEN** the CLI MUST POST to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools`

#### Scenario: Missing cluster-id is rejected before any API call

- **GIVEN** `cluster-id` is NOT set in state
- **WHEN** the user runs any `hf nodepool` subcommand that requires an API call
- **THEN** the CLI MUST print `[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.` and exit 1 without making an API call
