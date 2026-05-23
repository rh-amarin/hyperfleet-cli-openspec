# Proposal: `--search` Flag for `cluster list` and `nodepool list`

## Summary

Add a `--search` flag to `hf cluster list` and `hf nodepool list` that forwards a TSL
(Tree Search Language) expression directly to the API's native `search` query parameter,
enabling server-side filtering without any post-processing tool.

## Problem

`hf cluster list` and `hf nodepool list` currently return every resource the API has.
Users who want a subset — e.g., only production clusters, only reconciled node pools, or
clusters belonging to a specific team — must pipe the JSON output through `jq` or similar.
This contradicts the CLI's design goal of being fully self-contained with no external
tool dependencies.

The workaround (`hf cluster search <name>`) is narrowly scoped: it only matches by name,
and it has a side effect of updating the active cluster context — unsuitable for simple
listing workflows.

## Solution

Add `--search <expr>` to both list commands. When present, the CLI appends
`?search=<url-encoded-expr>` to the `GET /clusters` (or `/nodepools`) request. The API
processes the filter server-side and returns only matching items; the CLI renders the
result through the existing `--output json|table|yaml` pipeline unchanged.

## API Verification

The HyperFleet API already supports this parameter natively — no backend work required.
Confirmed against the live API at `34.175.55.239:8000` on 2026-05-16:

```
# Label filter — 1 result
GET /api/hyperfleet/v1/clusters?search=labels.environment='development'
→ {"kind":"ClusterList","page":1,"size":1,"total":1,"items":[...]}

# Label filter — 2 results
GET /api/hyperfleet/v1/clusters?search=labels.team='platform'
→ {"kind":"ClusterList","page":1,"size":2,"total":2,"items":[...]}

# No match
GET /api/hyperfleet/v1/clusters?search=labels.environment='production'
→ {"kind":"ClusterList","page":1,"size":0,"total":0,"items":[]}
```

The API uses TSL (Tree Search Language — github.com/yaacov/tree-search-language).
Searchable fields for clusters and nodepools include:

| Field | Example |
|---|---|
| `name` | `name='my-cluster'` |
| `id` | `id='019e...'` |
| `generation` | `generation>1` |
| `created_by` | `created_by='user@example.com'` |
| `labels.<key>` | `labels.environment='production'` |
| `status.conditions.<Type>` | `status.conditions.Ready='True'` |

Compound expressions use `and` / `or`: `labels.team='core' and status.conditions.Ready='True'`

**Condition queries are `=` only** — `!=`, `not`, `<`, `>` on conditions return 400.

## User-Facing Interface

```
# Filter by label
hf cluster list --search "labels.environment='production'"

# Filter by condition
hf cluster list --search "status.conditions.Ready='True'"

# Compound expression
hf cluster list --search "labels.team='platform' and generation>1"

# NodePools by parent cluster
hf nodepool list --search "owner_id='019e30db-...'"

# Combine with output format
hf cluster list --search "labels.environment='development'" --output table
```

## Scope

- `hf cluster list` — add `--search` flag
- `hf nodepool list` — add `--search` flag
- No other commands are affected
- No changes to `internal/` packages — pure `cmd/` layer change

## Out of Scope

- `--page` / `--page-size` / `--order` flags (separate proposal)
- `hf cluster search` command behaviour (unchanged)
- Shell completion for TSL expressions

## Risks

- **Invalid TSL expressions**: The API returns 400 with an RFC 7807 error body. The CLI
  already renders API errors via `handleAPIError`, so this is handled transparently.
- **URL encoding**: TSL expressions contain single quotes and spaces. The CLI must
  URL-encode the value before appending it to the request path. Failure to do so causes
  a malformed request. This must be covered by a unit test.
