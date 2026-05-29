## Context

Cluster and nodepool commands hardcode API paths (`clusters`, `clusters/{id}/nodepools`) and state keys (`cluster-id`, `nodepool-id`). New API entities like channels/versions follow the same REST pattern but are not yet exposed in the CLI.

Environment files today unmarshal as `map[string]map[string]string` for flat sections. Resource type definitions require structured YAML under a top-level `resource-types` key.

Cobra commands are registered in `init()` before config is loaded. Dynamic per-type subcommands must register lazily on first `hf resource` invocation after `loadConfig()`, guarded by `sync.Once`.

## Decisions

### D1 — Command name: `hf resource` / `hf rs`

`hf resources` is reserved for the combined cluster+nodepool overview (`cmd/resources.go`). The singular `resource` group holds config-defined types. Alias `rs` matches common kubectl-style abbreviations.

### D2 — Config schema: flat `resource-types` map with `parent` pointer

```yaml
resource-types:
  channels:
    path: channels
    state-key: channel-id
    create-template: channels.json
  versions:
    parent: channels
    path: "channels/{channel_id}/versions"
    state-key: version-id
    path-param: channel_id
    create-template: versions.json
```

Each map key is the CLI subcommand name. `parent` names the immediate parent type. `path` is relative to `{api-url}/api/hyperfleet/{api-version}/`. Placeholders `{channel_id}` are filled from ancestor state keys at runtime.

Default `path-param`: derive from `state-key` by replacing `-id` suffix with `_id` (`channel-id` → `channel_id`).

### D3 — Path resolution walks ancestor chain

For type `T` with parent chain `[root, …, parent(T)]`:

1. Collect each ancestor's `state-key` value from `state.yaml`
2. If any required ancestor ID is empty, error with guidance to run `hf resource <parent> search`
3. Substitute all `{path_param}` placeholders in `T.path`

Validation at load: parent exists, no cycles, unique state-keys, root types have no placeholders in path.

### D4 — State keys in state.yaml

Generic types write flat keys in `state.yaml` (e.g. `channel-id`, `version-id`) via `SetState`. Same stale-child behavior as cluster/nodepool: changing parent context does not auto-clear child keys.

### D5 — Generic API type

Use `map[string]any` wrapped as `resource.GenericResource` for get/list/create responses. Table output shows `id`, `name`, `kind` when present; falls back to JSON for `--output json|yaml`.

### D6 — Create templates

Extend `loadTemplate(resource, flagFile)` to accept a template filename from config. Resolution order: `--file` flag → read `{config-dir}/templates/{create-template}` → error if missing (no embedded defaults for generic types in v1).

### D7 — Dynamic Cobra registration

```go
var registerResourceCommands sync.Once

resourceCmd.PersistentPreRunE = func(cmd, args) error {
    s, err := loadConfig()
    // ...
    registerResourceCommands.Do(func() {
        types, _ := s.ResourceTypes()
        for _, t := range types {
            resourceCmd.AddCommand(newTypeCmd(t))
        }
    })
    return nil
}
```

Subcommands per type: `list`, `get`, `create`, `search`, `patch`, `delete`, `id` — mirroring cluster/nodepool minus `conditions`/`statuses`/`adapter` (out of scope v1).

### D8 — Built-in types stay first-class

`hf cluster` and `hf nodepool` remain unchanged. Generic machinery is only for config-defined types.

## Risks

- **Env file round-trip**: `Set()` currently rewrites env files as flat string maps, which would strip `resource-types`. Generic types are read-only via config file edit in v1; document that `hf config set` does not manage `resource-types`.
- **Empty config**: `hf resource` with no types configured shows only `types` subcommand and helpful message.
- **Completion**: Dynamic completion for type names deferred; static `types` subcommand suffices for v1.
