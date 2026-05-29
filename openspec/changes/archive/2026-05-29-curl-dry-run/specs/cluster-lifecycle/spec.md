## ADDED Requirements

### Requirement: Cluster Create Dry-Run

`hf cluster create` SHALL support dry-run via the global `--curl` flag.

#### Scenario: Create with curl prints POST only

- GIVEN the active environment is configured
- WHEN the user runs `hf cluster create [--name <name>] --curl`
- THEN the CLI MUST print a POST curl for `clusters` with the resolved template body to stderr
- AND MUST NOT perform any HTTP request (including duplicate-check GET)
- AND MUST NOT print created cluster JSON to stdout
- AND MUST NOT update `cluster-id` in state
- AND exit with code 0
