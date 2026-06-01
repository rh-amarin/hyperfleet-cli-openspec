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

#### Scenario: Get with explicit ID

- GIVEN a resource ID argument is provided
- WHEN the user runs `hf resource channels get <id>`
- THEN the CLI MUST GET `{collection-path}/{id}`

#### Scenario: Get with state ID

- GIVEN no ID argument and the type's state-key is set
- WHEN the user runs `hf resource channels get`
- THEN the CLI MUST use the state-key value as the ID

#### Scenario: Interactive get

- GIVEN `-i` flag and no ID argument
- WHEN the user runs `hf resource channels get -i`
- THEN the CLI MUST list resources and prompt for fuzzy selection
- AND `-i --curl` MUST return `[ERROR] --curl cannot be used with interactive mode`

### Requirement: Create Generic Resource

`hf resource <type> create` SHALL POST a JSON body to the resolved collection path.

#### Scenario: Create with template file

- GIVEN `create-template: channels.json` in config
- AND a file exists at `<config-dir>/templates/channels.json`
- WHEN the user runs `hf resource channels create`
- THEN the CLI MUST POST the template body to the collection path
- AND `--name` MAY override the `name` field in the body

#### Scenario: Create with --file

- WHEN the user runs `hf resource channels create --file <path>`
- THEN the CLI MUST use the file content as the request body

#### Scenario: Create sets state

- GIVEN a successful create response includes an `id` field
- WHEN create completes
- THEN the CLI MUST write the new ID to the type's state-key in `state.yaml`
- AND MUST print `[INFO] <type> context set to '<id>'` on stderr

#### Scenario: Create dry-run

- GIVEN `--curl` is set
- WHEN the user runs `hf resource channels create`
- THEN the CLI MUST print the POST curl command
- AND MUST NOT send an HTTP request
- AND MUST exit 0

### Requirement: Search Generic Resource

`hf resource <type> search [name]` SHALL find resources by name and set active state.

#### Scenario: Search by name

- GIVEN resources exist in the API
- WHEN the user runs `hf resource channels search <name>`
- THEN the CLI MUST query with `search=name='<name>'`
- AND on match MUST persist the first match ID to the type's state-key
- AND MUST print `[INFO] <type> context set to '<id>'` on stderr

#### Scenario: Search no match

- GIVEN no resource matches
- WHEN the user runs `hf resource channels search <name>`
- THEN the CLI MUST print `[WARN] No <type> found matching '<name>'`
- AND exit 0

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

`hf resource <type> id` SHALL print or interactively set the active ID for the type.

#### Scenario: Print current ID

- GIVEN the type's state-key is set
- WHEN the user runs `hf resource channels id`
- THEN the CLI MUST print the state-key value to stdout

#### Scenario: Interactive set ID

- WHEN the user runs `hf resource channels id -i`
- THEN the CLI MUST fuzzy-select from listed resources
- AND MUST write the selected ID to the type's state-key

### Requirement: Active Environment Required

All `hf resource` commands SHALL require an active environment except `hf env *` lifecycle commands.

#### Scenario: No active environment

- GIVEN no active environment is configured
- WHEN the user runs `hf resource channels list`
- THEN the CLI MUST fail with the standard `[ERROR] no active environment` message
- AND exit with code 1

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

