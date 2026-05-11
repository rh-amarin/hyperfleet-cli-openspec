## MODIFIED Requirements

### Requirement: Usage Messages for Required Arguments

All user-facing commands with required positional arguments MUST display the full Cobra help text (not a bare arg-count error) when invoked with zero arguments.

#### Scenario: Any command with required positional args called with zero args

- **GIVEN** a command requires one or more positional arguments
- **AND** the user runs that command with zero arguments
- **THEN** the CLI MUST display the full Cobra help text for that command (Usage, Flags, description)
- **AND** exit with code 1
- **AND** MUST NOT print the bare Cobra message "accepts N arg(s), received 0"

This applies to all user-facing commands including but not limited to:
`hf config get`, `hf config set`, `hf config env create`, `hf config env activate`,
`hf config env delete`, `hf config env show`, `hf cluster update`, `hf cluster adapter post-status`,
`hf nodepool update`, `hf nodepool delete`, `hf pubsub publish cluster`, `hf pubsub publish nodepool`,
`hf kube debug`, `hf db exec`, `hf db delete`.

Exception: commands that use defaults when no args are provided (e.g. `hf cluster create`, `hf nodepool create`) retain that behaviour and MUST NOT show usage.

Exception: internal/hidden commands (e.g. `hf kube _daemon`) retain `cobra.ExactArgs` and are not user-facing.
