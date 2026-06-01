## MODIFIED Requirements

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

## ADDED Requirements

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
