# Errors and Usage Specification

## Purpose

Define error handling behavior, usage messages, and edge cases across all CLI commands to ensure consistent and predictable behavior for the CLI reimplementation.

> API errors MUST follow RFC 7807 Problem Details format as defined in `api-client/spec.md` Requirement: RFC 7807 Error Parsing. This spec documents the CLI-level display behavior for those errors.
## Requirements
### Requirement: API Error Format

The CLI SHALL output HyperFleet API errors in RFC 7807 Problem Details format.

#### Scenario: Display API 404 error

- GIVEN a request is made for a non-existent resource
- WHEN the API returns a 404 response
- THEN the CLI MUST output the full error JSON containing:
  - `code`: error code (e.g., `HYPERFLEET-NTF-001`)
  - `detail`: human-readable description
  - `instance`: request path
  - `status`: HTTP status code as integer
  - `title`: short error title (e.g., `Resource Not Found`)
  - `trace_id`: UUID for tracing
  - `type`: URL to error documentation
  - `timestamp`: ISO8601 timestamp
- AND the CLI MUST exit with code 0 (the script does not check HTTP status)

### Requirement: Search Not Found Behavior

The CLI SHALL handle search-not-found cases gracefully.

#### Scenario: Cluster search returns no results

- GIVEN no cluster matches the search term
- WHEN the user runs `hf cluster search <nonexistent>`
- THEN the CLI MUST display `[WARN] No clusters found matching '<name>'` on stderr
- AND output an empty JSON array `[]` on stdout
- AND exit with code 0

### Requirement: Duplicate Creation Prevention

The CLI SHALL prevent duplicate resource creation.

#### Scenario: Create cluster with existing name

- GIVEN a cluster with the same name already exists
- WHEN the user runs `hf cluster create <existing-name>`
- THEN the CLI MUST search for an existing cluster with that name
- AND if found MUST print `[WARN] Cluster '<name>' already exists, skipping creation`
- AND exit with code 0
- AND MUST NOT send a POST create request

#### Scenario: Create cluster with existing name under dry-run

- GIVEN a cluster with the same name already exists
- WHEN the user runs `hf cluster create <existing-name> --curl`
- THEN the CLI MUST NOT perform the duplicate-check GET request
- AND MUST print the POST create curl to stderr
- AND MUST NOT print cluster data to stdout
- AND MUST NOT mutate state
- AND exit with code 0

#### Scenario: Create nodepool with existing name under dry-run

- GIVEN a nodepool with the same name already exists in the active cluster
- WHEN the user runs `hf nodepool create <existing-name> --curl`
- THEN the CLI MUST NOT perform the duplicate-check GET request
- AND MUST print the POST create curl to stderr
- AND MUST NOT mutate state
- AND exit with code 0

### Requirement: Default Argument Behavior

Commands that create resources SHALL use defaults rather than showing usage when no arguments are provided.

#### Scenario: Cluster create with no arguments

- GIVEN no arguments are provided
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST create a cluster using the embedded default template
- AND MUST NOT show a usage message

#### Scenario: NodePool create with no arguments

- GIVEN no arguments are provided
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST create a nodepool using the embedded default template
- AND MUST NOT show a usage message

### Requirement: Usage Messages for Required Arguments

All user-facing commands with required positional arguments MUST display the full Cobra help text (not a bare arg-count error) when invoked with zero arguments.

#### Scenario: Any command with required positional args called with zero args

- **GIVEN** a command requires one or more positional arguments
- **AND** the user runs that command with zero arguments
- **THEN** the CLI MUST display the full Cobra help text for that command (Usage, Flags, description)
- **AND** exit with code 1
- **AND** MUST NOT print the bare Cobra message "accepts N arg(s), received 0"

This applies to all user-facing commands including but not limited to:
`hf config set`, `hf env create`, `hf env activate`, `hf env delete`, `hf env show`,
`hf cluster adapter post-status`, `hf nodepool delete`,
`hf pubsub publish cluster`, `hf pubsub publish nodepool`,
`hf kube debug`, `hf db delete`.

Exception: commands that use defaults when no args are provided (e.g. `hf cluster create`, `hf nodepool create`) retain that behaviour and MUST NOT show usage.

Exception: internal/hidden commands (e.g. `hf kube _daemon`) retain `cobra.ExactArgs` and are not user-facing.

### Requirement: Exit Code Conventions

The CLI SHALL follow consistent exit code conventions.

#### Scenario: Successful operations

- GIVEN an operation completes successfully
- WHEN the CLI exits
- THEN the exit code MUST be 0

#### Scenario: Missing required arguments

- GIVEN required arguments are not provided
- WHEN the CLI shows a usage message
- THEN the exit code MUST be 1

#### Scenario: API errors

- GIVEN the API returns an error (4xx, 5xx)
- WHEN the CLI receives the error response
- THEN the exit code MUST be 0 (current behavior does not check HTTP status)
- AND the error response MUST be output as-is

