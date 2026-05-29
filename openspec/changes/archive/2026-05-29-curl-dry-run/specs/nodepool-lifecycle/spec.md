## ADDED Requirements

### Requirement: NodePool Create Dry-Run

`hf nodepool create` SHALL support dry-run via the global `--curl` flag.

#### Scenario: Create with curl prints POST only

- GIVEN the active environment is configured with a valid `cluster-id`
- WHEN the user runs `hf nodepool create [--name <name>] --curl`
- THEN the CLI MUST print a POST curl for the nodepool create endpoint with the resolved template body to stderr
- AND MUST NOT perform any HTTP request (including duplicate-check GET)
- AND MUST NOT print created nodepool JSON to stdout
- AND MUST NOT update `nodepool-id` in state
- AND exit with code 0
