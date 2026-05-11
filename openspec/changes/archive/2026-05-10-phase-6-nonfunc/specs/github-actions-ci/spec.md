# GitHub Actions CI Spec

## ADDED Requirements

### Requirement: Continuous integration on push and PR

The project SHALL provide a GitHub Actions workflow that validates every push and pull request to `main`.

#### Scenario: Push to main triggers CI

- GIVEN a developer pushes commits to `main`
- WHEN the push event fires
- THEN the CI workflow MUST run `go build ./...`, `go vet ./...`, and `go test ./...`
- AND the workflow MUST fail if any step exits non-zero

#### Scenario: Pull request to main triggers CI

- GIVEN a developer opens or updates a pull request targeting `main`
- WHEN the pull_request event fires
- THEN the CI workflow MUST run the same build, vet, and test steps
- AND the workflow MUST fail if any step exits non-zero

#### Scenario: Go version is pinned via go.mod

- GIVEN the workflow runs
- WHEN `actions/setup-go` installs Go
- THEN it MUST read the Go version from `go.mod`
- AND MUST enable the Go module cache
