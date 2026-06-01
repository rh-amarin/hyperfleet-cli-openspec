# generic-resource-lifecycle Specification

## Purpose
TBD - created by archiving change generic-resource-types. Update Purpose after archive.
## Requirements
### Requirement: List Generic Resources

`hf resource <type> list` SHALL list resources of the configured type from the resolved API path.

#### Scenario: List root type

- GIVEN a root resource type `channels` with `path: channels`
- WHEN the user runs `hf resource channels list`
- THEN the CLI MUST GET the resolved collection path
- AND MUST support `--output json|table|yaml`
- AND MUST support `--search <query>` as a `search=` query parameter when provided

#### Scenario: List child type

- GIVEN a child type `versions` with parent `channels`
- AND parent state is set
- WHEN the user runs `hf resource versions list`
- THEN the CLI MUST GET the parent-resolved collection path

#### Scenario: List with watch

- GIVEN `--watch` is supported on list commands
- WHEN the user runs `hf resource <type> list --watch`
- THEN the CLI MUST refresh the list on an interval matching cluster/nodepool watch behavior
- AND `--watch --curl` MUST print one curl command and exit without polling

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

### Requirement: Patch Generic Resource

`hf resource <type> patch {spec|labels} [id]` SHALL update a resource.

When `--file` is omitted, the CLI SHALL perform counter increment patch (legacy cluster/nodepool behavior). When `--file` is set, the CLI SHALL PATCH `{collection-path}/{id}/{section}` with the file body.

#### Scenario: Patch spec with file

- WHEN the user runs `hf resource channels patch spec [id] --file <path>`
- THEN the CLI MUST PATCH `{collection-path}/{id}/spec` with the file body

#### Scenario: Patch labels with file

- WHEN the user runs `hf resource channels patch labels [id] --file <path>`
- THEN the CLI MUST PATCH `{collection-path}/{id}/labels` with the file body

#### Scenario: Patch counter without file

- GIVEN no `--file` flag and entity is `clusters` or `nodepools`
- WHEN the user runs `hf rs <entity> patch spec [id]`
- THEN the CLI MUST increment the counter and PATCH using legacy cluster/nodepool patch semantics

### Requirement: Delete Generic Resource

`hf resource <type> delete [id]` SHALL DELETE a resource by ID.

#### Scenario: Delete with confirmation

- GIVEN no `-y` / force flag pattern matches cluster delete
- WHEN the user runs `hf resource channels delete <id>`
- THEN the CLI MUST send DELETE to `{collection-path}/{id}`
- AND MUST support `-i` for interactive ID selection

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

### Requirement: API Error Handling

Generic resource commands SHALL use the same API error handling as cluster commands.

#### Scenario: RFC 7807 error response

- GIVEN the API returns an RFC 7807 error
- WHEN a generic resource command fails
- THEN the CLI MUST print the error as JSON to stdout and exit 0
- AND non-API errors MUST exit 1 with `[ERROR]` prefix

### Requirement: Hierarchical RS Overview

`hf rs` with no subcommand SHALL render a hierarchical table of all configured resource types, nesting children under parents.

#### Scenario: Default table output

- GIVEN resource types are configured
- WHEN the user runs `hf rs` without `--output`
- THEN the default output format MUST be `table`

#### Scenario: Partial load warnings

- GIVEN one or more collection GETs fail during overview build
- WHEN the user runs `hf rs`
- THEN the CLI MUST print `[WARN]` lines at the top describing each failure
- AND MUST still render rows that loaded successfully
- AND MUST exit 0

