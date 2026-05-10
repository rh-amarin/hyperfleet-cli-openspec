# resources-table Specification

## Purpose
TBD - created by archiving change phase-5-tables-adapter. Update Purpose after archive.
## Requirements
### Requirement: Combined Resources Table

The CLI SHALL display a combined table of clusters and their nodepools via `hf resources` and `hf table`.

#### Scenario: Display combined table with clusters and nodepools

- GIVEN clusters and nodepools exist
- WHEN the user runs `hf resources --output table` or `hf table --output table`
- THEN the CLI MUST fetch all clusters, their adapter statuses, their nodepools, and each nodepool's adapter statuses
- AND output a table with fixed columns ID, NAME, GEN plus dynamic condition columns and dynamic adapter columns
- AND cluster rows MUST show the cluster's own conditions and adapter statuses
- AND nodepool rows MUST appear indented (two spaces prefix on ID and NAME) immediately after their parent cluster

#### Scenario: Dynamic condition columns

- GIVEN clusters have status conditions of types Available and Reconciled
- WHEN the user runs `hf resources --output table`
- THEN the table MUST include one column per unique condition type across all resources
- AND condition types ending in Successful MUST be excluded

#### Scenario: Dynamic adapter columns

- GIVEN cluster adapters cl-deployment and nodepool adapter np-configmap exist
- WHEN the user runs `hf resources --output table`
- THEN the table MUST include columns for both cl-deployment and np-configmap
- AND cluster rows MUST show `-` in np-configmap column and vice versa

#### Scenario: hf table alias

- GIVEN the user runs `hf table`
- WHEN the command executes
- THEN it MUST produce the same output as `hf resources`

#### Scenario: JSON output skips per-resource fetching

- GIVEN the user runs `hf resources --output json`
- THEN the CLI MUST output the raw clusters list JSON response
- AND MUST NOT fetch nodepools or adapter statuses

#### Scenario: Deletion marker

- GIVEN a cluster has deleted_time set
- WHEN the user runs `hf resources --output table`
- THEN the GEN column for that cluster MUST append a red X mark
- AND adapter columns for that cluster MUST use the Finalized condition instead of Available

