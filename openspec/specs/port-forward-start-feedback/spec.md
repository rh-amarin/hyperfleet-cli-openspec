# port-forward-start-feedback Specification

## Purpose
TBD - created by archiving change port-forward-start-feedback. Update Purpose after archive.
## Requirements
### Requirement: Port forward start enriched output

The CLI SHALL display the resolved namespace and pod name for each service when starting port forwards, and SHALL automatically show a process-alive status table after all services have been started.

#### Scenario: Namespace and pod shown in start line

- **WHEN** the user runs `hf kube port-forward start`
- **THEN** each `[INFO] Started …` line MUST include the namespace and pod name in the format:
  `[INFO] Started <name> (<namespace>/<podName>): localhost:<localPort> → <remotePort> (pid <pid>)`

#### Scenario: Status table shown after start

- **WHEN** `hf kube port-forward start` completes starting all services
- **THEN** the CLI MUST print the port-forward status table (identical to `hf kube port-forward status` output) immediately after the last `[INFO] Started …` line

#### Scenario: Pod not found — start line omits pod token

- **GIVEN** a service's target pod does not exist or is not Running
- **WHEN** the CLI starts that service
- **THEN** the `[INFO] Started …` line MUST omit the `(<namespace>/<podName>)` token
- **AND** the existing `[WARN]` message MUST still be shown before the start line

