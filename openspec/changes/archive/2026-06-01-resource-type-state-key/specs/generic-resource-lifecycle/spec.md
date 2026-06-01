## MODIFIED Requirements

### Requirement: Get Generic Resource

`hf resource <type> get [id]` SHALL fetch a single resource by ID.

#### Scenario: Get uses state when ID omitted

- GIVEN no ID argument and state for the entity name is set (e.g. `channels` in `state.yaml`)
- WHEN the user runs `hf resource channels get`
- THEN the CLI MUST use the entity-name state value as the ID

### Requirement: Create Generic Resource

`hf resource <type> create` SHALL create a resource and optionally persist its ID to state.

#### Scenario: Create persists entity state key

- GIVEN a successful create for type `channels`
- WHEN create completes
- THEN the CLI MUST write the new ID to `state.yaml` under the key `channels`

### Requirement: Search Generic Resource

`hf resource <type> search [name]` SHALL find a resource by name and set active context.

#### Scenario: Search persists entity state key

- GIVEN a search match for type `channels`
- WHEN search succeeds
- THEN the CLI MUST persist the first match ID to `state.yaml` under the key `channels`

### Requirement: Generic Resource ID Command

`hf resource <type> id` SHALL display or interactively set the active ID for the type.

#### Scenario: ID prints entity state key

- GIVEN state for the entity name is set
- WHEN the user runs `hf resource channels id`
- THEN the CLI MUST print the value stored under `channels` in `state.yaml`

#### Scenario: Interactive ID sets entity state key

- WHEN the user runs `hf resource channels id -i` and selects a resource
- THEN the CLI MUST write the selected ID to `state.yaml` under `channels`

### Requirement: Active Environment Required

All `hf resource` commands SHALL require an active environment except `hf env *` lifecycle commands.

#### Scenario: Child type requires parent entity state

- GIVEN type `nodepools` with parent `clusters`
- AND `clusters` is not set in `state.yaml`
- WHEN the user runs `hf rs nodepools list`
- THEN the CLI MUST fail with an error referencing the parent entity name `clusters`
