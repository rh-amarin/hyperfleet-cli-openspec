## MODIFIED Requirements

### Requirement: Generic Resource Command Group

The CLI SHALL provide a top-level command group `hf resource` with alias `hf rs` for config-defined HyperFleet API resource types. This group is the **canonical** interface for cluster and nodepool operations when those types appear in `resource-types`.

#### Scenario: Command group registration

- GIVEN the root command is initialized
- WHEN the user runs `hf resource --help` or `hf rs --help`
- THEN the CLI MUST show the `resource` command group
- AND `hf rs` MUST be a registered alias for `hf resource`
- AND the root command MUST NOT register a separate `resources` command (former combined overview)

#### Scenario: Overview is hf rs with no subcommand

- GIVEN `clusters` and `nodepools` are configured under `resource-types`
- WHEN the user runs `hf rs` with no subcommand
- THEN the CLI MUST render the combined cluster+nodepool operational overview
- AND MUST NOT require `hf table` or `hf resources`

#### Scenario: Dynamic type subcommands

- GIVEN the active environment defines resource types under `resource-types`
- WHEN the user runs `hf resource --help` after config load
- THEN the CLI MUST register one subcommand group per configured type name (e.g. `clusters`, `nodepools`, `channels`)
- AND each type group MUST expose: `list`, `table`, `get`, `create`, `search`, `patch`, `delete`, `conditions`, `statuses`, `adapter-report`, `id`
- AND MUST expose `force-delete` for types that support it (`clusters` via `delete --force`, `nodepools` via `force-delete`)

#### Scenario: No configured types

- GIVEN the active environment has no `resource-types` section or an empty map
- WHEN the user runs `hf resource --help`
- THEN the CLI MUST show at minimum the `types` subcommand and support `hf rs` overview when types are empty
- AND MUST NOT fail with a parse error

### Requirement: Cobra Command Tree

The CLI SHALL use [spf13/cobra](https://github.com/spf13/cobra) for command routing, flag parsing, and help generation.

#### Scenario: Command hierarchy

- GIVEN Cobra is the CLI framework
- WHEN commands are registered after migration is complete
- THEN the command tree MUST include:
  ```
  hf
  ├── resource | rs              # overview when invoked with no subcommand
  │   ├── types
  │   └── <entity>               # from resource-types (includes clusters, nodepools)
  │       ├── list
  │       ├── table
  │       ├── get
  │       ├── create
  │       ├── search
  │       ├── patch
  │       ├── delete
  │       ├── force-delete       # nodepools; clusters use delete --force
  │       ├── conditions
  │       ├── statuses
  │       ├── adapter-report
  │       └── id
  ```
- AND dynamic `<entity>` groups MUST be registered after the active environment is loaded
- AND `hf cluster`, `hf nodepool`, `hf resources`, and `hf table` MUST NOT be registered

#### Scenario: Legacy commands removed from root help

- GIVEN this change is complete
- WHEN the user runs `hf --help`
- THEN `hf cluster` and `hf nodepool` MUST NOT appear
- AND `hf table` and `hf resources` MUST NOT appear
- AND cluster/nodepool operations MUST appear only under `hf rs`

## ADDED Requirements

### Requirement: Deprecation of Legacy Commands (transition release)

During the transition release before removal, legacy commands MAY remain registered. If present, they SHALL delegate to `hf rs` and print a deprecation warning on stderr.

#### Scenario: Deprecated cluster list

- WHEN the user runs `hf cluster list`
- THEN stderr MUST include a deprecation warning naming `hf rs clusters list`
- AND stdout MUST match `hf rs clusters list` for the same flags and environment

#### Scenario: Deprecated hf table

- WHEN the user runs `hf table --watch`
- THEN stderr MUST include a deprecation warning naming `hf rs --watch`
- AND table output MUST match `hf rs --watch`
