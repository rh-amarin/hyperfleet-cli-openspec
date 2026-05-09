# Repository Status Specification

## Purpose

Provide a CLI command to display a status overview of all HyperFleet GitHub repositories, including latest commits, open PRs, and container image tags from Quay.

## Requirements

### Requirement: Show Repository Status Table

The CLI SHALL display a status table for all HyperFleet GitHub repositories.

#### Scenario: Display repos table

- GIVEN the user has GitHub CLI (`gh`) authenticated
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
- AND COMMIT MUST show the short hash of the latest commit (fetched via `gh`)
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
