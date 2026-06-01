## Why

The `state-key` field in `resource-types` duplicates the map key (entity name) and forces users to maintain two identifiers per type (`channels` vs `channel-id`). Using the entity name as the `state.yaml` key simplifies configuration and aligns CLI subcommands with runtime state.

## What Changes

- **BREAKING** — Remove `state-key` from the `resource-types` schema; the map key (entity name) is always the state key.
- **BREAKING** — `state.yaml` keys change from `cluster-id` / `nodepool-id` to `clusters` / `nodepools` (and similarly for custom types: `channels` not `channel-id`).
- Derive API path placeholders from entity name when `path-param` is omitted (`clusters` → `{cluster_id}`).
- Wire optional `path-param` on a type when the default derivation is insufficient.
- Update `ClusterID()` / `NodePoolID()` and legacy messaging/pubsub paths to use entity-name state keys.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `config-model`: Remove `state-key` field from resource-types schema; document entity-name state keys.
- `generic-resource-lifecycle`: State reads/writes use entity name, not configurable state-key.
- `command-hierarchy`: `hf rs types` displays entity name as state key.
- `rs-entity-commands`: Scenarios referencing `cluster-id` state use `clusters` / `nodepools`.

## Impact

- `internal/config/resource_types.go`: parse, path resolution, path-param derivation
- `internal/config/config.go`: `ClusterID`, `NodePoolID`
- `internal/config/assets/config-template.yaml`: remove state-key entries
- `cmd/pubsub.go`, `cmd/rabbitmq.go`, `cmd/logs.go`, `cmd/cluster.go`, `cmd/nodepool.go`: state key strings
- Tests and README

## Testing Scope

- `internal/config/resource_types_test.go`: parsing without state-key; path resolution with entity-name state; path-param override
- `cmd/resource_test.go`, `cmd/helpers_test.go`: updated fixtures
- `internal/config/config_test.go`: ClusterID/NodePoolID with `clusters` key

Live verification: `hf rs clusters search`, confirm `clusters` written to state.yaml; `hf rs nodepools list` with parent state set.
