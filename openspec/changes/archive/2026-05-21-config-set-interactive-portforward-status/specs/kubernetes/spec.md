## ADDED Requirements

### Requirement: Port Forward Bare Invocation

When `hf kube port-forward` is invoked with no subcommand, the CLI SHALL display the command help block and then show the current port-forward status. If any port-forward is not connected, the user SHALL be prompted to start all port-forwards.

#### Scenario: hf kube port-forward bare — help and status shown

- **WHEN** the user runs `hf kube port-forward` with no subcommand
- **THEN** the CLI MUST print the command help block (including "Usage:" and available subcommands) to stdout
- **AND** MUST resolve and display the active Kubernetes context (same as `hf kube port-forward status`)
- **AND** MUST display the current port-forward connectivity status for all tracked port-forwards

#### Scenario: hf kube port-forward bare — no port-forwards tracked

- **WHEN** the user runs `hf kube port-forward` with no subcommand
- **AND** no port-forward PID files exist
- **THEN** the CLI MUST print the help block
- **AND** MUST print `No port-forwards tracked.`
- **AND** MUST exit with code 0 without prompting

#### Scenario: hf kube port-forward bare — all connected, no prompt

- **WHEN** the user runs `hf kube port-forward` with no subcommand
- **AND** all tracked port-forwards pass their connectivity check
- **THEN** the CLI MUST display the status table with green checkmarks
- **AND** MUST exit with code 0 without prompting to start

#### Scenario: hf kube port-forward bare — some down, user confirms start

- **WHEN** the user runs `hf kube port-forward` with no subcommand
- **AND** at least one tracked port-forward fails its connectivity check
- **THEN** the CLI MUST display the status table with red ✗ for the failing service(s)
- **AND** MUST print `Some port-forwards are down. Run 'hf kube port-forward start'? [y/N]: `
- **AND** if the user enters `y`, the CLI MUST start all port-forwards (same behaviour as `hf kube port-forward start`)

#### Scenario: hf kube port-forward bare — some down, user declines start

- **WHEN** the user runs `hf kube port-forward` with no subcommand
- **AND** at least one tracked port-forward fails its connectivity check
- **AND** the user enters anything other than `y` at the prompt
- **THEN** the CLI MUST exit with code 0 without starting any port-forwards
