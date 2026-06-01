## MODIFIED Requirements

### Requirement: Resource Types Configuration

The CLI SHALL support a structured `resource-types` section in environment YAML files defining configurable HyperFleet API resource types.

#### Scenario: Resource type definition fields

- GIVEN an environment file at `~/.config/hf/environments/<name>.yaml`
- WHEN a resource type is defined under `resource-types.<type-name>`
- THEN it MUST support fields:
  - `path` (required): API path relative to `{api-url}/api/hyperfleet/{api-version}/`
  - `parent` (optional): name of the immediate parent resource type
  - `path-param` (optional): placeholder name this type's ID fills in child paths; when omitted, derived from the entity name (e.g. `clusters` → `cluster_id`)
  - `create-template` (optional): filename under `<config-dir>/templates/` for create payloads
- AND the map key `<type-name>` MUST be used as both the CLI subcommand name and the flat key in `state.yaml` when the type is selected
- AND `state-key` MUST NOT be a supported configuration field

#### Scenario: Config template default

- GIVEN a newly created environment from the embedded config template
- WHEN the environment file is written
- THEN it MUST include `resource-types` with at least `clusters` and `nodepools` entries:
  ```yaml
  resource-types:
    clusters:
      path: clusters
      create-template: clusters.json
    nodepools:
      parent: clusters
      path: "clusters/{cluster_id}/nodepools"
      create-template: nodepools.json
  ```

### Requirement: Active State File

The CLI SHALL maintain a `state.yaml` file for runtime state separate from environment configuration.

#### Scenario: Generic resource state keys

- GIVEN a resource type named `channels`
- WHEN the user runs `hf rs channels search <name>` and a match is found
- THEN the CLI MUST write the matched resource ID to `state.yaml` under the key `channels`
- AND generic resource state keys MUST coexist with `clusters`, `nodepools`, and `active-environment`
