## ADDED Requirements

### Requirement: Resource Types Configuration

The CLI SHALL support a structured `resource-types` section in environment YAML files defining configurable HyperFleet API resource types.

#### Scenario: Resource type definition fields

- GIVEN an environment file at `~/.config/hf/environments/<name>.yaml`
- WHEN a resource type is defined under `resource-types.<type-name>`
- THEN it MUST support fields:
  - `path` (required): API path relative to `{api-url}/api/hyperfleet/{api-version}/`
  - `state-key` (required): flat key written to `state.yaml` when the type is selected
  - `parent` (optional): name of the immediate parent resource type
  - `path-param` (optional): placeholder name this type's ID fills in child paths; defaults to `state-key` with `-id` replaced by `_id`
  - `create-template` (optional): filename under `<config-dir>/templates/` for create payloads
- AND the map key `<type-name>` MUST be used as the CLI subcommand name

#### Scenario: Root type example

- GIVEN this configuration:
  ```yaml
  resource-types:
    channels:
      path: channels
      state-key: channel-id
      create-template: channels.json
  ```
- WHEN the user runs `hf resource channels list`
- THEN the CLI MUST call GET `{api-url}/api/hyperfleet/{api-version}/channels`

#### Scenario: Child type example

- GIVEN this configuration:
  ```yaml
  resource-types:
    channels:
      path: channels
      state-key: channel-id
    versions:
      parent: channels
      path: "channels/{channel_id}/versions"
      state-key: version-id
      path-param: channel_id
  ```
- AND `channel-id` is set in `state.yaml` to `abc-123`
- WHEN the user runs `hf resource versions list`
- THEN the CLI MUST call GET `{api-url}/api/hyperfleet/{api-version}/channels/abc-123/versions`

#### Scenario: Missing parent state

- GIVEN a child type requires `parent: channels` with `state-key: channel-id`
- AND `channel-id` is not set in `state.yaml`
- WHEN the user runs any child type command except `types`
- THEN the CLI MUST print `[ERROR] No channel-id set in state. Run 'hf resource channels search <name>' first.`
- AND exit with code 1

#### Scenario: Config validation on load

- WHEN resource types are loaded from the active environment
- THEN the CLI MUST reject configurations where:
  - `parent` references a non-existent type name
  - a parent cycle exists
  - two types share the same `state-key`
  - a root type's `path` contains unresolved `{placeholders}`

#### Scenario: Default path-param derivation

- GIVEN a type with `state-key: channel-id` and no `path-param`
- WHEN path resolution runs for a child referencing `{channel_id}`
- THEN the CLI MUST derive `path-param` as `channel_id`

#### Scenario: Multi-level parent chain

- GIVEN types `channels` → `versions` → `releases` each with explicit paths and state keys
- WHEN the user runs a command on the deepest child type
- THEN the CLI MUST resolve all ancestor state keys from `state.yaml` and substitute all placeholders in the child path

#### Scenario: Config template default

- GIVEN a newly created environment from the embedded config template
- WHEN the environment file is written
- THEN it MUST include an empty `resource-types:` section (or omit the section entirely with empty parse result)

#### Scenario: hf config show includes type names

- GIVEN resource types are configured in the active environment
- WHEN the user runs `hf config show`
- THEN the output MUST list configured resource type names
- AND MUST NOT require `hf config set` to manage nested resource-type fields in v1

## MODIFIED Requirements

### Requirement: Active State File

The CLI SHALL maintain a `state.yaml` file for runtime state separate from environment configuration.

#### Scenario: Generic resource state keys

- GIVEN a resource type with `state-key: channel-id`
- WHEN the user runs `hf resource channels search <name>` and a match is found
- THEN the CLI MUST write the matched resource ID to `state.yaml` under `channel-id`
- AND generic resource state keys MUST coexist with existing keys (`cluster-id`, `nodepool-id`, `active-environment`)
