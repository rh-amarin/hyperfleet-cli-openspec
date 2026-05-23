## MODIFIED Requirements

### Requirement: Delete NodePool

`hf nodepool delete` SHALL accept an optional nodepool ID, falling back to the configured nodepool-id in state when no explicit ID is provided (consistent with `hf nodepool get`, `patch`, `conditions`, and `statuses`). When `--force` is passed, it SHALL POST to `clusters/<clusterID>/nodepools/<id>/force-delete` instead of sending DELETE. An optional `--reason` flag provides an audit reason.

#### Scenario: Delete nodepool by explicit ID

- **WHEN** the user runs `hf nodepool delete <nodepool_id>`
- **THEN** the CLI MUST send DELETE to `clusters/<clusterID>/nodepools/<nodepool_id>`
- **AND** the CLI MUST output the deleted nodepool object subject to the `--output` flag

#### Scenario: Delete current nodepool (state fallback — FIXED)

- **GIVEN** nodepool-id is set in state and no explicit ID is provided
- **WHEN** the user runs `hf nodepool delete`
- **THEN** the CLI MUST resolve the nodepool-id from state and send DELETE to `clusters/<clusterID>/nodepools/<nodepool_id>`
- **AND** MUST NOT return an error about missing nodepool ID

#### Scenario: Delete nodepool with no ID and no state

- **GIVEN** no nodepool-id is set in state and no argument is provided and `-i` is not used
- **WHEN** the user runs `hf nodepool delete`
- **THEN** the CLI MUST display `[ERROR] no active nodepool — run 'hf nodepool use <id>' or pass --nodepool-id`
- **AND** exit with code 1

#### Scenario: Force-delete nodepool

- **WHEN** the user runs `hf nodepool delete --force [nodepool_id]`
- **THEN** the CLI MUST POST to `clusters/<clusterID>/nodepools/<id>/force-delete` with body `{"reason": "<reason>"}`
- **AND** print `[INFO] NodePool '<id>' force-deleted` on success
- **AND** MUST NOT send a DELETE request

#### Scenario: Force-delete nodepool with reason

- **WHEN** the user runs `hf nodepool delete --force --reason "stuck" [nodepool_id]`
- **THEN** the CLI MUST POST `{"reason": "stuck"}` to `clusters/<clusterID>/nodepools/<id>/force-delete`
