# Design: `--search` Flag for `cluster list` and `nodepool list`

## Packages Touched

| File | Change |
|---|---|
| `cmd/cluster.go` | Add `clusterListSearch string` flag var; pass to `fetchAndRenderClusterList`; build query path with URL-encoded search value |
| `cmd/nodepool.go` | Same pattern — `nodepoolListSearch string`; pass to `fetchAndRenderNodepoolList` |
| `cmd/cluster_test.go` | Unit tests for search path construction and URL encoding |
| `cmd/nodepool_test.go` | Same |

No changes to `internal/api/`, `internal/resource/`, `internal/output/`, or any other package.

## Flag Definition

```go
// cmd/cluster.go
var clusterListSearch string

clusterListCmd.Flags().StringVar(&clusterListSearch, "search", "",
    "TSL filter expression (e.g. \"labels.environment='prod'\")")
```

Same pattern in `cmd/nodepool.go` with `nodepoolListSearch`.

## Path Construction

The `fetchAndRenderClusterList` function currently calls:

```go
api.Get[resource.ListResponse[resource.Cluster]](ctx, client, "clusters")
```

With the flag, the path becomes:

```go
path := "clusters"
if clusterListSearch != "" {
    path = "clusters?search=" + url.QueryEscape(clusterListSearch)
}
api.Get[resource.ListResponse[resource.Cluster]](ctx, client, path)
```

`url.QueryEscape` from stdlib `net/url` handles single quotes, spaces, and all
special TSL characters. No additional dependency needed.

The API client already passes the path through as-is to the base URL:
`baseURL + path` → `http://host/api/hyperfleet/v1/clusters?search=...`

This is safe because `api.Client.baseURL` always ends with `/` (enforced in
`newAPIClient`) and the path never starts with `/`.

## Watch Mode Interaction

`clusterListSearch` is a package-level var. `fetchAndRenderClusterList` is called on
every tick during watch mode. The search expression is re-applied on every refresh, which
is the correct behaviour — the filter persists across ticks because the flag var is set
once at startup.

No watch-mode-specific changes needed.

## Signature Change for `fetchAndRenderClusterList`

The function currently has signature:
```go
func fetchAndRenderClusterList(cmd *cobra.Command, tick, frequencySecs int) error
```

It already has access to `clusterListSearch` as a package-level var (same pattern as
`outputFmt`, `noColor`, `verbose`, `curlMode`). No signature change required.

## Error Handling

Invalid TSL expressions (e.g. `not status.conditions.Ready='True'`) cause the API to
return HTTP 400 with an RFC 7807 body. `api.Get` returns `*api.APIError`. The existing
`handleAPIError(p, err)` call in `fetchAndRenderClusterList` already prints this and
exits 0. No additional error handling needed.

## Unit Test Strategy

Use `httptest.NewServer` to serve a mock `/clusters` endpoint. Assert:

1. **No flag**: server receives request with no `search` query param.
2. **Simple flag**: server receives `search=labels.environment%3D%27prod%27` (URL-encoded).
3. **Compound flag**: `search=labels.team%3D%27core%27+and+generation%3E1` (encoded).
4. **API error on bad TSL**: mock returns 400 RFC 7807; CLI exits 0 and prints error JSON.

The test captures the server's received URL via `r.URL.RawQuery` and asserts equality.

## Spec Delta Location

`openspec/changes/list-search-filter/specs/cluster-lifecycle/spec.md`
`openspec/changes/list-search-filter/specs/nodepool-lifecycle/spec.md`

Both add a new `### Requirement: Search Filter on List` section to the respective specs.
