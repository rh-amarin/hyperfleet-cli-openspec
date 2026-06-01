## Context

Each `resource-types.<name>` entry currently requires `state-key`, typically `<singular>-id`, while the CLI subcommand is `<name>`. State is written on search/create/id to `state-key`, and ancestor path resolution keys `ancestorIDs` maps by state-key.

## Goals / Non-Goals

**Goals:**
- `StateKey` always equals the resource type name (YAML map key).
- Remove `state-key` from config schema and template.
- Derive `{path_param}` placeholders from entity name when `path-param` is unset.

**Non-Goals:**
- Migrating `cluster-name` or other non-ID state keys.
- Automatic migration of existing `state.yaml` files (document manual rename).

## Decisions

### 1. State key = entity name

**Decision:** At parse time, set `ResourceTypeDef.StateKey = name`. Ignore `state-key` if present in YAML (no error).

**Why:** Single identifier; map key is already required and unique.

### 2. Path placeholder derivation

**Decision:** `placeholderInChildPaths(def)` returns `def.PathParam` if set, else `derivePathParamFromTypeName(def.Name)`:
- Strip trailing `s` from plural names, replace `-` with `_`, append `_id` (`clusters` → `cluster_id`, `channels` → `channel_id`).

**Why:** Replaces old `-id` suffix heuristic on state-key; matches existing API path templates.

### 3. Built-in cluster/nodepool state

**Decision:** `ClusterID()` reads `clusters`; `NodePoolID()` reads `nodepools`. Error messages reference `hf rs clusters search`.

**Why:** Matches bundled template entity names.

## Risks / Trade-offs

- **Breaking state.yaml** — Users rename `cluster-id` → `clusters`, `nodepool-id` → `nodepools`. Accepted; documented in proposal.

## Open Questions

None.
