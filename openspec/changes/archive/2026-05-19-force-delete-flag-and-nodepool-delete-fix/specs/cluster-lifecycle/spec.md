## MODIFIED Requirements

### Requirement: Delete Cluster

`hf cluster delete` SHALL accept an optional cluster ID, falling back to the configured cluster-id. When `--force` is passed, it SHALL POST to `clusters/<id>/force-delete` instead of sending DELETE. An optional `--reason` flag provides an audit reason for force deletions.

#### Scenario: Delete cluster (unchanged)

- **WHEN** the user runs `hf cluster delete [cluster_id]`
- **THEN** the CLI MUST use the provided ID, or the configured cluster-id if none is provided
- **AND** the CLI MUST send DELETE to `clusters/<id>`
- **AND** the CLI MUST output the deleted cluster JSON subject to the `--output` flag

#### Scenario: Force-delete cluster

- **WHEN** the user runs `hf cluster delete --force [cluster_id]`
- **THEN** the CLI MUST POST to `clusters/<id>/force-delete` with body `{"reason": "<reason>"}`
- **AND** print `[INFO] Cluster '<id>' force-deleted` on success
- **AND** MUST NOT send a DELETE request

#### Scenario: Force-delete cluster with reason

- **WHEN** the user runs `hf cluster delete --force --reason "stuck" [cluster_id]`
- **THEN** the CLI MUST POST `{"reason": "stuck"}` to `clusters/<id>/force-delete`
