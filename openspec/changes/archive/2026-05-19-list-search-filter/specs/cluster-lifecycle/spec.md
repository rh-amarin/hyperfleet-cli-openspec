# Spec Delta: Cluster Lifecycle — `--search` flag on `cluster list`

**Applies to:** `openspec/specs/cluster-lifecycle/spec.md`
**Section:** Requirement: List Clusters

## Changes

### Requirement: List Clusters — add `--search` flag (MODIFIED)

Add the following to the existing List Clusters requirement.

#### Scenario: List clusters with TSL search filter

- GIVEN an active environment is configured
- WHEN the user runs `hf cluster list --search "<tsl-expression>"`
- THEN the CLI MUST append `?search=<url-encoded-tsl-expression>` to the `GET /api/hyperfleet/v1/clusters` request
- AND the API filters results server-side using TSL (Tree Search Language)
- AND the CLI MUST render the filtered result using the existing `--output` pipeline (json|table|yaml)
- AND the `--search` flag MUST be composable with `--output`, `--watch`, `--no-color`

#### Scenario: List clusters — no `--search` flag (unchanged behaviour)

- GIVEN no `--search` flag is provided
- WHEN the user runs `hf cluster list`
- THEN the CLI MUST send `GET /api/hyperfleet/v1/clusters` with no `search` parameter
- AND behaviour is identical to the current implementation

#### Scenario: Invalid TSL expression

- GIVEN the user provides a malformed TSL expression (e.g. `not status.conditions.Ready='True'`)
- WHEN the user runs `hf cluster list --search "<bad-expr>"`
- THEN the API returns HTTP 400 with an RFC 7807 error body
- AND the CLI MUST render the error JSON via `handleAPIError` (exit 0)
- EXAMPLE output:
  ```json
  {
    "code": "HYPERFLEET-VAL-001",
    "detail": "invalid search expression: ...",
    "status": 400,
    "title": "Validation Error"
  }
  ```

## TSL Field Reference (Clusters)

| Field | Operator support | Example |
|---|---|---|
| `name` | `=`, `!=`, `in` | `name='my-cluster'` |
| `id` | `=`, `in` | `id='019e30db-...'` |
| `generation` | `=`, `!=`, `<`, `<=`, `>`, `>=` | `generation>1` |
| `created_by` | `=`, `!=` | `created_by='user@example.com'` |
| `updated_by` | `=`, `!=` | `updated_by='user@example.com'` |
| `labels.<key>` | `=`, `!=` | `labels.environment='production'` |
| `status.conditions.<Type>` | `=` only | `status.conditions.Ready='True'` |

Compound expressions supported with `and` / `or`.
`not` is NOT supported with `status.conditions.*` fields (returns 400).

## Flag Specification

```
--search string    TSL filter expression passed to the API search parameter.
                   See https://github.com/yaacov/tree-search-language for syntax.
                   Examples:
                     labels.environment='production'
                     status.conditions.Ready='True'
                     labels.team='core' and generation>1
```

## Live API Evidence

Verified 2026-05-16 against `34.175.55.239:8000`:

```
GET /api/hyperfleet/v1/clusters?search=labels.environment='development'
→ 200 {"kind":"ClusterList","total":1,"items":[{"name":"my-cluster","labels":{"environment":"development",...}}]}

GET /api/hyperfleet/v1/clusters?search=labels.team='platform'
→ 200 {"kind":"ClusterList","total":2,"items":[...two platform clusters...]}

GET /api/hyperfleet/v1/clusters?search=labels.environment='production'
→ 200 {"kind":"ClusterList","total":0,"items":[]}
```
