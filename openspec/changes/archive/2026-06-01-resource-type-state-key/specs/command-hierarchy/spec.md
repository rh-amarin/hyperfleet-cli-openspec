## MODIFIED Requirements

### Requirement: Resource Types Subcommand

`hf rs types` (alias `hf resource types`) SHALL list configured resource types from the active environment.

#### Scenario: Types output shows entity state keys

- GIVEN resource types are configured
- WHEN the user runs `hf rs types`
- THEN the CLI MUST print each type name, its API path template, its parent (if any), and whether the entity name key is set in `state.yaml`
- AND MUST indicate which parent entity state is required before child commands can run
