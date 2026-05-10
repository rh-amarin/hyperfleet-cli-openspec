## Why

The `hf cluster` command tree (stub only as of Phase 1) needs a full implementation so operators can manage the complete cluster lifecycle — create, inspect, update, delete, and monitor conditions/statuses — via a single binary without any external tools.

## What Changes

- `hf cluster list` — GET /clusters, JSON array or table with STATUS column derived from conditions
- `hf cluster get [id]` — GET /clusters/{id}, JSON or single-row table; uses state cluster-id when id omitted
- `hf cluster create` — POST /clusters with optional --name/--replicas/--nodepool-id flags; checks for duplicates first; persists cluster-id to state
- `hf cluster update <id>` — PATCH /clusters/{id} with --name and/or --replicas flags
- `hf cluster delete <id>` — DELETE /clusters/{id}; silent success; RFC 7807 error on 404
- `hf cluster conditions <id>` — GET /clusters/{id}/conditions returning []ResourceCondition; table: TYPE STATUS LAST TRANSITION REASON MESSAGE
- `hf cluster statuses <id>` — GET /clusters/{id}/statuses returning []AdapterStatus; table: ADAPTER GEN AVAILABLE

## Capabilities

### New Capabilities

- `cluster-lifecycle`: Full CRUD lifecycle for clusters — list, get, create (with duplicate check), update, delete, conditions, statuses

### Modified Capabilities

_(none — this extends the stub cluster.go with real implementations)_

## Impact

- `cmd/cluster.go`: replaces stub with full Cobra subcommand tree
- `cmd/cluster_test.go`: new file with httptest-based unit tests for all endpoints
- No new internal packages required; uses `internal/api`, `internal/config`, `internal/resource`, `internal/output`
- API: `/api/hyperfleet/v1/clusters` and `/api/hyperfleet/v1/clusters/{id}/conditions|statuses`

## Testing Scope

| Package / File       | Test cases needed                                                        |
|----------------------|--------------------------------------------------------------------------|
| `cmd/cluster_test.go` | list (200), get (200), get 404 propagated as API error, create (201), create duplicate guard, update (200), delete (200), delete 404, conditions (200), statuses (200) |

Live cluster access required for: none (all verified via httptest; live verification is best-effort).
