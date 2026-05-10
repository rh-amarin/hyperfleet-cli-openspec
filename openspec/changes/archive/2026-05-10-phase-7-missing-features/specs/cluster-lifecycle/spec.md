# Cluster Lifecycle — Phase 7 Delta

## MODIFIED Requirements

### Requirement: Create Cluster

`hf cluster create` SHALL additionally accept positional arguments `[name] [region] [version]`.

#### Scenario: Create cluster with positional arguments (MODIFIED)

- WHEN the user runs `hf cluster create <name> [region] [version]`
- THEN the CLI MUST use positional args for name, region, and version
- AND positional args take precedence over the `--name` flag and built-in defaults (`my-cluster`, `us-east-1`, `4.15.0`)

### Requirement: Delete Cluster

`hf cluster delete` SHALL accept an optional cluster ID, falling back to the configured cluster-id.

#### Scenario: Delete cluster (MODIFIED — optional ID, output)

- WHEN the user runs `hf cluster delete [cluster_id]`
- THEN the CLI MUST use the provided ID, or the configured cluster-id if none is provided
- AND the CLI MUST output the deleted cluster JSON subject to the `--output` flag

### Requirement: Get Cluster Adapter Statuses

The statuses table SHALL include a FINALIZED column in addition to AVAILABLE.

#### Scenario: Get statuses table (MODIFIED — add FINALIZED column)

- WHEN the user runs `hf cluster statuses --output table`
- THEN the CLI MUST output columns: ADAPTER, GEN, AVAILABLE, FINALIZED
- AND AVAILABLE and FINALIZED columns MUST be color-coded dots: green=True, red=False, `-`=not present
