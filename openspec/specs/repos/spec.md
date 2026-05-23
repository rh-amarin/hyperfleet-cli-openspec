# Repository Status Specification

## Purpose

Provide a CLI command to display a status overview of all HyperFleet GitHub repositories, including latest commits, open PRs, and container image tags from Quay.
## Requirements
### Requirement: Show Repository Status Table

The CLI SHALL display a status table for all HyperFleet GitHub repositories.

#### Scenario: Display repos table

- GIVEN a GitHub token is configured (via `GITHUB_TOKEN` env or `registry.github-token` config)
- WHEN the user runs `hf repos`
- THEN the CLI MUST display a formatted table with columns: REPOSITORY, COMMIT, PR URL, PR BRANCH, QUAY TAG, QUAY ALIASES
- AND the following repositories MUST be tracked:
  - `openshift-hyperfleet/hyperfleet-api-spec`
  - `openshift-hyperfleet/hyperfleet-api`
  - `openshift-hyperfleet/hyperfleet-sentinel`
  - `openshift-hyperfleet/hyperfleet-adapter`
  - `openshift-hyperfleet/hyperfleet-infra`
  - `openshift-hyperfleet/hyperfleet-e2e`
  - `openshift-hyperfleet/architecture`
- AND COMMIT MUST show the short hash of the latest commit (fetched via GitHub API)
- AND PR URL MUST show the URL of the latest open PR (or `-` if none)
- AND PR BRANCH MUST show the source branch of the latest open PR (or `-`)
- AND QUAY TAG MUST show the latest container image tag with date (or `-`)
- AND QUAY ALIASES MUST show any alias tags (or `-`)

#### Scenario: Output format

- GIVEN the user runs `hf repos`
- THEN the default output format MUST be table
- AND `--output json` MUST output the same data as a JSON array of objects with the same field names
- AND `--output yaml` MUST output the same data as YAML

#### Scenario: Performance

- GIVEN multiple repositories need to be queried
- WHEN the user runs `hf repos`
- THEN the CLI MUST fetch data in parallel for performance
- AND the total execution time SHOULD be significantly less than serial execution

### Requirement: GitHub token configuration
The CLI SHALL read the GitHub API token from `registry.github-token` config key or the `GITHUB_TOKEN` environment variable. When neither is set, the client SHALL make unauthenticated requests (60 req/h rate limit applies).

#### Scenario: Token from environment variable
- **WHEN** `GITHUB_TOKEN` is set in the environment
- **THEN** the repos client SHALL use it for authenticated GitHub API requests

#### Scenario: Token from config
- **WHEN** `registry.github-token` is set in config.yaml
- **THEN** the repos client SHALL use it for authenticated GitHub API requests

### Requirement: Quay namespace configuration
The CLI SHALL use `registry.quay-namespace` config key to determine the Quay.io namespace (default: `openshift-hyperfleet`).

#### Scenario: Default namespace
- **WHEN** `registry.quay-namespace` is not configured
- **THEN** the client SHALL use `openshift-hyperfleet` as the Quay namespace

### Requirement: go-github client (no shell-out)
The implementation SHALL use the `go-github` library for GitHub API access instead of shelling out to the `gh` CLI, in compliance with the project's self-contained binary mandate.

#### Scenario: No external tool dependency
- **WHEN** `hf repos` is invoked
- **THEN** no child process for `gh` or any other external tool SHALL be spawned

