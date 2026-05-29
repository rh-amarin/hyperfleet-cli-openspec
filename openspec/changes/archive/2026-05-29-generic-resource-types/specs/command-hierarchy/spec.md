## ADDED Requirements

### Requirement: Generic Resource Command Group

The CLI SHALL provide a top-level command group `hf resource` with alias `hf rs` for config-defined HyperFleet API resource types.

#### Scenario: Command group registration

- GIVEN the root command is initialized
- WHEN the user runs `hf resource --help` or `hf rs --help`
- THEN the CLI MUST show the `resource` command group
- AND `hf rs` MUST be a registered alias for `hf resource`
- AND this group MUST NOT be named `resources` (that name is reserved for the cluster+nodepool overview)

#### Scenario: Distinction from hf resources

- GIVEN `hf resources` displays combined cluster and nodepool overview
- WHEN a user runs `hf resource <type> list`
- THEN the CLI MUST operate on config-defined API resource types
- AND MUST NOT invoke the cluster+nodepool overview command

#### Scenario: Dynamic type subcommands

- GIVEN the active environment defines resource types under `resource-types`
- WHEN the user runs `hf resource --help` after config load
- THEN the CLI MUST register one subcommand group per configured type name (e.g. `channels`, `versions`)
- AND each type group MUST expose subcommands: `list`, `get`, `create`, `search`, `patch`, `delete`, `id`

#### Scenario: No configured types

- GIVEN the active environment has no `resource-types` section or an empty map
- WHEN the user runs `hf resource --help`
- THEN the CLI MUST show at minimum the `types` subcommand
- AND MUST NOT fail with a parse error

### Requirement: Resource Types Subcommand

The CLI SHALL provide `hf resource types` to display configured resource types and their relationships.

#### Scenario: List configured types

- GIVEN resource types are defined in the active environment
- WHEN the user runs `hf resource types`
- THEN the CLI MUST print each type name, its API path template, its parent (if any), and its state-key
- AND MUST indicate whether each state-key is currently set in `state.yaml`

#### Scenario: Parent chain display

- GIVEN a child type with `parent: channels`
- WHEN the user runs `hf resource types`
- THEN the output MUST show the parent relationship
- AND MUST indicate which parent state-key is required before child commands can run

## MODIFIED Requirements

### Requirement: Cobra Command Tree

The CLI SHALL use [spf13/cobra](https://github.com/spf13/cobra) for command routing, flag parsing, and help generation.

#### Scenario: Command hierarchy

- GIVEN Cobra is the CLI framework
- WHEN commands are registered
- THEN the command tree MUST include:
  ```
  hf
  ├── resource | rs
  │   ├── types
  │   └── <type-name>          # dynamically registered from resource-types config
  │       ├── list
  │       ├── get
  │       ├── create
  │       ├── search
  │       ├── patch
  │       ├── delete
  │       └── id
  ```
- AND dynamic `<type-name>` groups MUST be registered after the active environment is loaded
- AND existing `hf cluster`, `hf nodepool`, and `hf resources` commands MUST remain unchanged
