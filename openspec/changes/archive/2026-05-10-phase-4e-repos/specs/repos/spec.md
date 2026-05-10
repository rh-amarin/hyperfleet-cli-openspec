## ADDED Requirements

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
