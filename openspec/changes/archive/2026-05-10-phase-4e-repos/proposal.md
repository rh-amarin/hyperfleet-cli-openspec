## Why

The `hf repos` command is a stub with no implementation. Developers need a quick status overview of all HyperFleet GitHub repositories — including latest commits, open PRs, and container image tags from Quay — without having to visit multiple services manually.

## What Changes

- `internal/repos/`: New package implementing a GitHub + Quay registry client using `go-github` and the Quay HTTP API.
- `cmd/repos.go`: Full implementation of `hf repos` replacing the empty stub.
- `go.mod`/`go.sum`: Add `github.com/google/go-github/v62` and `golang.org/x/oauth2` dependencies.

## Capabilities

### New Capabilities

- `repos`: CLI command displaying a status table for all HyperFleet repositories, with parallel data fetching from GitHub and Quay.

### Modified Capabilities

<!-- none — the existing openspec/specs/repos/spec.md is already authoritative -->

## Impact

- New package: `internal/repos/`
- Modified file: `cmd/repos.go`
- New dependencies: `go-github/v62`, `golang.org/x/oauth2`
- Config section: `registry` (uses `github-token`, `quay-namespace`)

## Testing Scope

- `internal/repos/`: Unit tests using `httptest.NewServer` to mock the GitHub REST API and Quay API responses.
- `cmd/repos.go`: Integration via `internal/repos`; no additional command tests required (command logic is thin).

Verification steps that require live cluster access: none — this feature queries GitHub and Quay, not the HyperFleet API. Live verification can be done with a valid `GITHUB_TOKEN`.
