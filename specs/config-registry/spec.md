# Delta for config-registry

## MODIFIED Requirements

### Requirement: Configuration Storage
The CLI SHALL store configuration as two YAML files in `~/.config/hf/`:
- `config.yaml` for static settings (sections: hyperfleet, kubernetes, maestro, port-forward, database, rabbitmq, registry)
- `state.yaml` for active runtime state (flat top-level keys: active-environment, cluster-id, cluster-name, nodepool-id)

#### Scenario: YAML config storage
- GIVEN the CLI is initialized
- WHEN any configuration property is set
- THEN the value MUST be written into the appropriate section of `config.yaml`
- AND `config.yaml` MUST be written atomically with mode `0600`

#### Scenario: YAML state storage
- GIVEN runtime state changes (cluster selection, active environment)
- WHEN state is updated
- THEN the value MUST be written into `state.yaml` as a flat top-level key
- AND `state.yaml` MUST be written atomically (write temp, then rename)

### Requirement: Environment Profiles
Named environment profiles SHALL be stored as `~/.config/hf/environments/<name>.yaml`,
using the same nested YAML structure as `config.yaml` with only the overriding keys present.
Activation writes `active-environment: <name>` to `state.yaml` — it does NOT copy files.

(Previously: `~/.config/hf/<env>.<key>` flat files, activation by file copy. Superseded.)

#### Scenario: Environment file storage
- GIVEN an environment named `kind` overrides `kubernetes.context` and `kubernetes.namespace`
- WHEN the environment is stored
- THEN it MUST be at `~/.config/hf/environments/kind.yaml` with nested YAML structure

#### Scenario: Activate environment
- GIVEN a named environment exists in `environments/`
- WHEN the user runs `hf config env activate <name>`
- THEN `active-environment: <name>` MUST be written to `state.yaml`
- AND subsequent config reads MUST deep-merge the env YAML on top of config.yaml at runtime
- AND `config.yaml` MUST NOT be modified
