## ADDED Requirements

### Requirement: Force-Delete Cluster

`hf cluster delete --force` SHALL POST to `clusters/<id>/force-delete` to permanently remove a cluster that is stuck in a Finalizing or unrecoverable state. An optional `--reason` flag passes an audit reason in the request body.

#### Scenario: Force-delete active cluster

- **WHEN** the user runs `hf cluster delete --force`
- **THEN** the CLI MUST resolve the cluster ID from state (or explicit arg) and POST to `clusters/<id>/force-delete` with body `{"reason": ""}`
- **AND** print `[INFO] Cluster '<id>' force-deleted` on success

#### Scenario: Force-delete with reason

- **WHEN** the user runs `hf cluster delete --force --reason "stuck in finalizing"`
- **THEN** the CLI MUST POST to `clusters/<id>/force-delete` with body `{"reason": "stuck in finalizing"}`

#### Scenario: Force-delete non-existent cluster

- **WHEN** the user runs `hf cluster delete --force <bad-id>`
- **THEN** the CLI MUST handle a 404 API error with `[ERROR] Cluster '<id>' not found`

#### Scenario: Regular delete is unchanged

- **WHEN** the user runs `hf cluster delete` without `--force`
- **THEN** the CLI MUST send DELETE to `clusters/<id>` as before (no change to existing behavior)
