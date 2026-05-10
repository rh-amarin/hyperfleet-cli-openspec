# github-actions-release Specification

## Purpose
TBD - created by archiving change phase-6-nonfunc. Update Purpose after archive.
## Requirements
### Requirement: Automated release on version tags

The project SHALL provide a GitHub Actions workflow that publishes a release when a version tag is pushed.

#### Scenario: Tag push triggers release

- GIVEN a developer pushes a tag matching `v*` (e.g. `v1.0.0`)
- WHEN the push tag event fires
- THEN the release workflow MUST run GoReleaser with `release --clean`
- AND GoReleaser MUST produce and upload release assets to the GitHub release

#### Scenario: Full git history is available to GoReleaser

- GIVEN the release workflow runs
- WHEN the repository is checked out
- THEN `fetch-depth: 0` MUST be set so GoReleaser can generate a full changelog

#### Scenario: GITHUB_TOKEN is passed to GoReleaser

- GIVEN the release workflow runs
- WHEN GoReleaser attempts to create a GitHub release and upload assets
- THEN it MUST have access to `GITHUB_TOKEN` from secrets
- AND the token MUST be passed via the `env` block of the goreleaser-action step

