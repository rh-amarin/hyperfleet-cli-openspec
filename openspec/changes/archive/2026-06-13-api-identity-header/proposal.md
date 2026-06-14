## Why

Some HyperFleet API deployments require a caller-identity header (e.g., `X-HyperFleet-Identity`) to enforce access control or audit logging. There is currently no way to inject a custom header into every API call without modifying source code.

## What Changes

- Add two new config keys under the `hyperfleet` section: `identity-header` (header name) and `identity-value` (header value).
- When `identity-header` is set, the API client MUST attach the configured header and value to every outgoing HTTP request.
- When `identity-header` is not set (default), behaviour is unchanged — no header is injected.
- The feature is opt-in: leaving both keys unset produces no observable change.

## Capabilities

### New Capabilities

- `api-identity-header`: Configurable per-request identity header injected into all HyperFleet API HTTP calls, controlled by `hyperfleet.identity-header` and `hyperfleet.identity-value`.

### Modified Capabilities

- `api-client`: The API client requirement for typed HTTP methods must be extended — when an identity header is configured, it MUST be included in every request (GET, POST, PATCH, DELETE).
- `config-model`: The environment profile schema must include the two new `hyperfleet.identity-header` and `hyperfleet.identity-value` keys.

## Impact

- `internal/api/` — `NewClient` and/or the shared request builder gains the identity header logic.
- `internal/config/` — `Store.Get("hyperfleet", "identity-header")` and `Store.Get("hyperfleet", "identity-value")` must resolve through the standard precedence chain.
- `cmd/` — root command wiring passes the two resolved values into `api.NewClient`.
- `cmd/assets/config-template.yaml` — template gains commented-out `identity-header` and `identity-value` keys so new environments are aware of the option.
- No breaking changes; the keys are optional and default to empty.

## Testing Scope

| Package | Test cases |
|---|---|
| `internal/api` | Header present when configured; header absent when not configured; header name and value match config; all HTTP methods (GET, POST, PATCH, DELETE) carry the header |
| `internal/config` | `identity-header` and `identity-value` resolve via standard precedence (env file → default empty) |

Live cluster verification step requires access to `localhost:8000` (or the configured `api-url`) with `X-HyperFleet-Identity: openspec@test.com`.
