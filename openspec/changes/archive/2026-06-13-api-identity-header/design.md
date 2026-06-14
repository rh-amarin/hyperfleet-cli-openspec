## Context

The HyperFleet API client (`internal/api/Client`) already supports Bearer token auth and verbose logging injected at construction time via `NewClient`. The root command wires these values from the config store into `NewClient` before dispatch.

Adding a configurable identity header follows the same pattern: resolve two values from config (`hyperfleet.identity-header`, `hyperfleet.identity-value`), pass them into `NewClient`, and set the header in the shared `do()` method before the request is sent.

## Goals / Non-Goals

**Goals**
- Inject a single configurable header into every API HTTP call when `identity-header` is set.
- Zero behaviour change when `identity-header` is empty (opt-in).
- Follow existing config precedence (env var → env file → default empty).

**Non-Goals**
- Multiple custom headers (a single identity header covers the stated requirement).
- Exposing the identity value as a secret (it is not a credential; no masking needed).
- Changing the `--curl` dry-run output format (the header will appear there naturally).

## Decisions

### Decision: Extend `NewClient` signature vs. functional options

**Choice**: Add `identityHeader, identityValue string` as two extra parameters to `NewClient`.

**Rationale**: `NewClient` already has four flat parameters; two more remain readable and avoid adding a new options pattern for a single feature. If the parameter list grows beyond ~6, a `ClientOptions` struct should be introduced — but that refactor is out of scope here.

**Alternative considered**: `ClientOptions` struct. Rejected because it requires a breaking change to all existing callers now, for a single new feature.

### Decision: Inject the header in `do()` rather than per-method

**Choice**: Set the identity header once in the shared `do()` method, alongside `Content-Type` and `Authorization`.

**Rationale**: All HTTP methods (GET, POST, PATCH, DELETE) call `do()`. A single injection point is correct by construction; there is no risk of forgetting it in a future method.

### Decision: Config keys in the `hyperfleet` section

**Choice**: `hyperfleet.identity-header` and `hyperfleet.identity-value`.

**Rationale**: Consistent with other `hyperfleet.*` keys (`api-url`, `api-version`, `token`, `namespace`). No new section is needed.

## Risks / Trade-offs

- [Risk] The identity value may appear in logs if verbose mode dumps raw headers. → Mitigation: identity value is not a secret per the spec; verbose logging of headers is acceptable.
- [Risk] The new `NewClient` parameters silently break compilation for any external callers. → Mitigation: this is an internal CLI binary; `NewClient` is not part of a public library API.

## Migration Plan

1. Update `NewClient` signature in `internal/api/client.go`.
2. Update all existing callers in `cmd/` (typically one root wiring call).
3. Add `identity-header` and `identity-value` (commented out) to `cmd/assets/config-template.yaml`.
4. Add unit tests in `internal/api/` using `httptest.NewServer`.
5. Verify live against `localhost:8000` with `X-HyperFleet-Identity: openspec@test.com`.

No rollback steps required — removing the two config keys restores prior behaviour.

## Open Questions

None. The scope is narrow and fully defined by the proposal.
