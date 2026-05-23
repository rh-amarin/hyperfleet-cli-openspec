# Spec Delta: NodePool Lifecycle — `--search` flag on `nodepool list`

**Applies to:** `openspec/specs/nodepool-lifecycle/spec.md`
**Section:** Requirement: List NodePools

## Changes

### Requirement: List NodePools — add `--search` flag (MODIFIED)

Add the following to the existing List NodePools requirement.

#### Scenario: List nodepools with TSL search filter

- GIVEN an active environment is configured
- WHEN the user runs `hf nodepool list --search "<tsl-expression>"`
- THEN the CLI MUST append `?search=<url-encoded-tsl-expression>` to the `GET /api/hyperfleet/v1/nodepools` request
- AND the API filters results server-side using TSL
- AND the CLI MUST render the filtered result using the existing `--output` pipeline (json|table|yaml)
- AND the `--search` flag MUST be composable with `--output`, `--watch`, `--no-color`

#### Scenario: List nodepools — no `--search` flag (unchanged behaviour)

- GIVEN no `--search` flag is provided
- WHEN the user runs `hf nodepool list`
- THEN the CLI MUST send `GET /api/hyperfleet/v1/nodepools` with no `search` parameter
- AND behaviour is identical to the current implementation

#### Scenario: Filter nodepools by parent cluster

- GIVEN a cluster ID `<cluster_id>` is known
- WHEN the user runs `hf nodepool list --search "owner_id='<cluster_id>'"`
- THEN the CLI MUST return only nodepools belonging to that cluster
- NOTE: This is the cross-resource join use case. The `owner_id` field is unique to nodepools.

#### Scenario: Invalid TSL expression

- GIVEN the user provides a malformed TSL expression
- WHEN the user runs `hf nodepool list --search "<bad-expr>"`
- THEN the API returns HTTP 400 with an RFC 7807 error body
- AND the CLI MUST render the error JSON via `handleAPIError` (exit 0)

## TSL Field Reference (NodePools)

All cluster fields plus:

| Field | Operator support | Example |
|---|---|---|
| `owner_id` | `=`, `in` | `owner_id='019e30db-ce3f-7d63-b8ba-c81c728222f1'` |
| `name` | `=`, `!=`, `in` | `name='worker-pool'` |
| `labels.<key>` | `=`, `!=` | `labels.role='worker'` |
| `status.conditions.<Type>` | `=` only | `status.conditions.Ready='True'` |

## Flag Specification

```
--search string    TSL filter expression passed to the API search parameter.
                   Examples:
                     owner_id='<cluster-id>'
                     labels.role='worker'
                     status.conditions.Ready='True'
```
