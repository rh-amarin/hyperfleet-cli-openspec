# Adapter Status Specification

## Purpose

Provide CLI commands to simulate adapter status reporting for clusters and nodepools. These commands allow manual posting of adapter conditions to the HyperFleet API, enabling testing of the convergence logic without running real adapters.
## Requirements
### Requirement: Post Cluster Adapter Status

The CLI SHALL post adapter status conditions for a cluster via `hf rs clusters adapter-report` (replacing `hf cluster adapter post-status`).

#### Scenario: Post status with True

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf rs clusters adapter-report <adapter_name> True <generation>`
- THEN the CLI MUST send PUT to `/api/hyperfleet/v1/clusters/{cluster_id}/statuses`
- AND the request payload MUST include four conditions with `reason: "ManualStatusPost"` and message referencing `hf rs adapter-report`

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

### Requirement: RS Adapter Report Command

The CLI SHALL provide `hf rs <entity> adapter-report` as the config-driven equivalent of legacy `hf cluster adapter post-status` and `hf nodepool adapter post-status`.

#### Scenario: PUT statuses path

- GIVEN a resolved resource id for type `<entity>`
- WHEN the user runs `hf rs <entity> adapter-report <adapter> <True|False|Unknown> <generation> [id]`
- THEN the CLI MUST PUT to `{collection-path}/{id}/statuses`
- AND the request body MUST match the Post Cluster Adapter Status payload shape (four conditions, `observed_generation`, `observed_time`)

#### Scenario: Command naming on rs

- GIVEN the user manages resources via `hf rs`
- WHEN simulating adapter reporting
- THEN the subcommand MUST be named `adapter-report` (not `adapter post-status`)

