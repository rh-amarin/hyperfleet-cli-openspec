## MODIFIED Requirements

### Requirement: Post Cluster Adapter Status

The CLI SHALL post adapter status conditions for a cluster via `hf rs clusters adapter-report` (replacing `hf cluster adapter post-status`).

#### Scenario: Post status with True

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf rs clusters adapter-report <adapter_name> True <generation>`
- THEN the CLI MUST send PUT to `/api/hyperfleet/v1/clusters/{cluster_id}/statuses`
- AND the request payload MUST include four conditions with `reason: "ManualStatusPost"` and message referencing `hf rs adapter-report`
