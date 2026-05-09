## Why

The HyperFleet CLI needs four shared foundation packages — `internal/config`, `internal/api`, `internal/resource`, and `internal/output` — before any command implementations can be built. Without these packages, every command would duplicate HTTP plumbing, config loading, type definitions, and output formatting. This is Phase 1 of a multi-phase build.

## What Changes

- **New**: `internal/config` — split YAML config model (`config.yaml` + `state.yaml`), environment profiles, active-env guard, config precedence chain
- **New**: `internal/api` — generic typed HTTP client (`Get[T]`, `Post[T]`, `Patch[T]`, `Delete[T]`), RFC 7807 error parsing, verbose debug logging
- **New**: `internal/resource` — all API struct types (`Cluster`, `NodePool`, `ClusterStatus`, `NodePoolStatus`, `AdapterStatus`, `CloudEvent`, `ListResponse[T]`, `ObjectReference`, `ValidationError`)
- **New**: `internal/output` — multi-format `Printer` (json/table/yaml), colored dot renderer, dynamic column ordering, colored JSON output

## Capabilities

### New Capabilities
- `config-model`: Split YAML configuration model with env profiles, state management, and precedence chain
- `api-client`: Generic typed HTTP client with RFC 7807 error handling and verbose logging
- `resource-types`: Go struct types for all HyperFleet API resources with JSON tags
- `output-formatting`: Multi-format printer with colored dots, dynamic columns, and colored JSON

### Modified Capabilities
- `config-registry`: Config storage implementation changes from single-file to split YAML model (delta spec documents this transition)

## Impact

- Adds dependencies: `gopkg.in/yaml.v3` for YAML serialization
- All subsequent command packages (`cmd/`) will depend on these four packages
- No changes to existing `cmd/` files in this phase — only new `internal/` packages

## Testing Scope

| Package | Test Cases |
|---|---|
| `internal/config` | Load defaults, precedence chain (flags > env vars > profile > file > defaults), atomic writes, env profile deep merge, `RequireActiveEnvironment` error |
| `internal/api` | `Get[T]` happy path, `Post[T]` happy path, RFC 7807 error parsing (400, 404), non-JSON error body, HTML error body, verbose logging, Bearer token injection |
| `internal/resource` | JSON round-trip for `Cluster`, `NodePool`, `ListResponse[T]`, `AdapterStatus` |
| `internal/output` | JSON output format, table output format, YAML output format, dot renderer (True/False/Unknown/absent), no-color mode, dynamic column ordering |

Live cluster access is NOT required for this phase — all tests use `httptest.NewServer` or file I/O.
