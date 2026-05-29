## Why

The HyperFleet API is growing beyond clusters and nodepools (e.g. channels, versions). Hardcoding a new top-level Cobra group per entity does not scale and duplicates the same CRUD/search/state pattern already implemented for cluster and nodepool. Operators need a config-driven way to register arbitrary API resource types with parent/child relationships without rebuilding the CLI.

## What Changes

- Add top-level command group `hf resource` with alias `hf rs` (distinct from existing `hf resources` cluster+nodepool overview)
- Add `resource-types` structured section to environment YAML defining per-type API paths, state keys, optional parent links, and create templates
- Dynamically register per-type subcommands (`list`, `get`, `create`, `search`, `patch`, `delete`, `id`) from active environment config
- Add `hf resource types` to show configured type graph, parent requirements, and active state keys
- Resolve child API paths using parent state keys (same model as cluster-id → nodepool paths)
- Reuse existing API client, output formatting, `-i` fuzzy selection, and `--curl` dry-run behavior
- Keep `hf cluster`, `hf nodepool`, and `hf resources` unchanged

## Capabilities

### New Capabilities

- `generic-resource-lifecycle`: Config-driven CRUD, search/id state management, parent context requirements, and command registration for arbitrary HyperFleet API resource types

### Modified Capabilities

- `command-hierarchy`: Add `hf resource` / `hf rs` tree; clarify distinction from `hf resources`
- `config-model`: Add `resource-types` section schema, validation, and state-key conventions

## Impact

- `internal/config/resource_types.go` — parse, validate, path resolution
- `internal/config/config.go` — load structured `resource-types` from env file alongside flat sections
- `internal/config/assets/config-template.yaml` — empty `resource-types:` default
- `internal/resource/generic.go` — generic JSON map type for API responses
- `cmd/resource.go` — new command group, dynamic registration, CRUD handlers
- `cmd/templates.go` — extend `loadTemplate` for per-type template files
- `cmd/config.go` — surface configured type names in config show
- `cmd/root.go` — register `resourceCmd` and `rsCmd` alias
- Tests: `internal/config/resource_types_test.go`, `cmd/resource_test.go`

## Testing Scope

- `internal/config/resource_types`: parse valid YAML; reject unknown parent, cycles, duplicate state-keys; path-param derivation; `ResolveResourcePath` with single and multi-level parent chains
- `cmd/resource`: httptest for root type list/get/create/search; child type fails without parent state; child list succeeds with parent state; `--curl` dry-run; state-key writes on search/create/id
- `cmd/config`: `hf config show` lists configured resource type names when present

## Live Verification

- Configure `channels` and `versions` resource types in an active environment profile
- Run `hf resource types`, `hf resource channels list`, `hf resource channels search <name>`, `hf resource versions list`
- Save output to `verification_proof/live.txt`
