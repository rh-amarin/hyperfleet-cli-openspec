## Context

`hf repos` is already registered as a Cobra command with no subcommands or `RunE`. The spec at `openspec/specs/repos/spec.md` defines: a table of REPOSITORY, COMMIT, PR URL, PR BRANCH, QUAY TAG, QUAY ALIASES for seven `openshift-hyperfleet/*` repos, with parallel fetching and table/JSON/YAML output.

The original spec mentions `gh` CLI for fetching, but the project mandate is to be a self-contained binary — no shelling out. We use `go-github` instead.

## Goals / Non-Goals

**Goals:**
- Implement `internal/repos.Client` that fetches GitHub commit and PR data via the GitHub REST API.
- Implement Quay tag fetching via the Quay.io REST API (unauthenticated for public repos).
- Implement `hf repos` command with parallel fetch and table/JSON/YAML output.
- Full unit test coverage using `httptest.NewServer`.

**Non-Goals:**
- Filtering by repository, date range, or branch.
- Writing back to GitHub (creating PRs, etc.).
- Supporting private Quay repositories (no auth needed for openshift-hyperfleet public images).

## Decisions

### go-github for GitHub API
Using `github.com/google/go-github/v62/github` with `golang.org/x/oauth2` for token auth. Token sourced from: `GITHUB_TOKEN` env var > `registry.github-token` config key. Anonymous (unauthenticated) requests have a 60 req/h rate limit; authenticated requests allow 5000/h. With 7 repos × 2 API calls = 14 calls, anonymous is fine but authenticated is safer.

### Quay REST API for image tags
Quay exposes `GET https://quay.io/api/v1/repository/{namespace}/{repo}/tag/?limit=10&page=1` returning JSON with a `tags` array. No auth needed for public repos. Each tag has `name`, `start_ts` (Unix timestamp), `manifest_digest`. We pick the most recent non-alias tag for QUAY TAG and aliases (tags pointing to same digest) for QUAY ALIASES.

### Parallel fetch
One goroutine per repository, using `sync.WaitGroup` + a results channel. Results collected and sorted to match the hardcoded repo order.

### Config keys
- `registry.github-token`: GitHub personal access token (optional; falls back to `GITHUB_TOKEN` env var).
- `registry.quay-namespace`: Quay namespace (default: `openshift-hyperfleet`).

### No active-env guard for `hf repos`
The `repos` command queries GitHub and Quay, not the HyperFleet API. However, the root `PersistentPreRunE` still runs and requires an active environment. We follow the same pattern as other commands (env required) — the command will fail cleanly if no env is set.

## Risks / Trade-offs

- [Risk] GitHub rate-limiting (60 req/h unauthenticated) → Mitigation: use token auth when available; 7 repos × 2 calls = 14 well within limit.
- [Risk] Quay API schema changes → Mitigation: parse defensively, fall back to `-` on any error.
- [Risk] go-github v62 API surface may differ from expectation → Mitigation: use the direct REST approach for GitHub commit/PR data to minimize surface area.

## Migration Plan

1. `go get` new dependencies.
2. Create `internal/repos/` package.
3. Replace `cmd/repos.go` stub with full implementation.
4. Run `go test ./...` and `go build ./...`.
5. Verify output manually with a valid `GITHUB_TOKEN`.

## Open Questions

- Should `hf repos` bypass the active-env requirement? Currently: No (follows standard guard). Can be changed later.
