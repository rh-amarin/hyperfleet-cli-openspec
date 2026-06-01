## MODIFIED Requirements

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

#### Scenario: Config template default

- GIVEN a newly created environment from the embedded config template
- WHEN the environment file is written
- THEN it MUST include `resource-types` with at least `clusters` and `nodepools` entries:
  ```yaml
  resource-types:
    clusters:
      path: clusters
      state-key: cluster-id
      create-template: clusters.json
    nodepools:
      parent: clusters
      path: "clusters/{cluster_id}/nodepools"
      state-key: nodepool-id
      path-param: cluster_id
  ```
