# Design: curl-mode

## Context

The CLI makes HTTP calls through two independent clients:
- `internal/api.Client` — all HyperFleet REST API calls (clusters, nodepools, etc.)
- `internal/maestro.Client` — all Maestro HTTP API calls

Both clients are constructed per-command using `newAPIClient(s)` and `newMaestroClient()` helpers in `cmd/`. The existing `verbose bool` field on `api.Client` (passed from the `--verbose` global flag) is the direct precedent for this pattern.

## Goals / Non-Goals

**Goals:**
- Print a complete, copy-pasteable `curl` command to stderr for every HTTP call when `--curl` is set.
- Cover both `api.Client` (HyperFleet API) and `maestro.Client` (Maestro API).
- Not pollute stdout (data output stays clean for piping and scripting).

**Non-Goals:**
- Masking the bearer token (the point is copy-paste usability).
- Pretty-printing JSON bodies (compact JSON is fine for `-d`).
- Covering `internal/kube` or `internal/repos` HTTP calls (those are not user-facing API interactions).

## Decisions

### Decision 1: Thread `curlMode` through client constructors, not a global

**Chosen:** Add `curlMode bool` field to each client struct, pass via constructor (`NewClient(..., curlMode bool)`, `NewFromConfig(s, curlMode bool)`), read from the `curlMode` global in the `cmd/` wiring layer.

**Alternative considered:** Read `os.Getenv` or a package-level global directly inside `internal/`. Rejected — `internal/` packages must not depend on CLI flag state. The constructor parameter pattern is already established by `verbose bool` on `api.Client`.

### Decision 2: Print in `do()` / `get()` / `delete()` — not via `http.RoundTripper`

**Chosen:** Add the print call directly inside each client's request method, immediately before `c.http.Do(req)`.

**Alternative considered:** Wrap `http.DefaultTransport` with a custom `http.RoundTripper` that intercepts all requests. Rejected because:
1. Reading `req.Body` (an `io.Reader`) in the transport requires draining and restoring it — more fragile.
2. The `do()` method already has the marshalled `bodyBytes` available as a `[]byte` before the body reader is consumed — simpler and more correct to use it there.
3. The verbose logging already uses the same `do()` / method-level pattern; consistency matters.

### Decision 3: Double-quote URLs, single-quote header values

**Chosen:** `curl -s -X GET "<url>" \\\n  -H 'Header: value'`

URLs are double-quoted because Maestro consumer search URLs contain literal single quotes (e.g., `?search=consumer_name='cluster1'`), which would break a single-quoted shell string. Header values don't contain double quotes in practice, so single-quoting them is safe and conventional.

### Decision 4: `printCurlCommand` as a package-private helper in `internal/api`; inline in `internal/maestro`

Maestro's curl output is simpler (no body, no auth header) and only two methods. Sharing via a new package would be over-engineering. The helper in `internal/api/client.go` is private and self-contained.

## Output Format

```
[CURL] curl -s -X GET "http://localhost:8000/api/hyperfleet/v1/clusters" \
  -H 'Accept: application/json' \
  -H 'Authorization: Bearer eyJh...'
```

For requests with a body (POST/PATCH):
```
[CURL] curl -s -X POST "http://localhost:8000/api/hyperfleet/v1/clusters" \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer eyJh...' \
  -d '{"name":"my-cluster","region":"eu-west-1"}'
```

Maestro GET (no auth header, GET is default so no `-X GET`):
```
[CURL] curl -s "http://localhost:8100/api/maestro/v1/resource-bundles?search=consumer_name='cluster1'" \
  -H 'Accept: application/json'
```

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Token leaks in shell history if user pastes full command | This is intentional — the flag is a debugging tool. Document clearly that `--curl` prints live tokens. |
| Single quotes in body JSON would break `-d 'body'` | Go struct fields that marshal to JSON won't contain single quotes in practice; field names and UUIDs are alphanumeric. |
| Future clients (e.g., a new internal package) won't get curl logging automatically | Acceptable — only document that `--curl` covers HyperFleet and Maestro API calls. |

## Migration Plan

All existing `api.NewClient` and `maestro.NewFromConfig` call sites get a new final `false` argument. These are:
- `internal/api/client_test.go` — 6 call sites
- `internal/maestro/maestro_test.go` — 1 call site
- `cmd/cluster.go` `newAPIClient()` — 1 call site
- `cmd/maestro.go` `newMaestroClient()` and inline — 2 call sites

No external consumers; this is an internal binary.
