## MODIFIED Requirements

### Requirement: Show Configuration

`hf config` and `hf config show` MUST display the active environment's configuration sections followed by a `state:` block at the bottom listing all non-empty runtime state variables.

#### Scenario: Active environment file path displayed

- **GIVEN** an active environment is configured
- **WHEN** the user runs `hf config` or `hf config show` (no argument)
- **THEN** the CLI MUST print the absolute path of the active environment file
- **AND** the path MUST appear before the YAML sections
- **AND** the path MUST be the resolved absolute path of `~/.config/hf/environments/<name>.yaml`
