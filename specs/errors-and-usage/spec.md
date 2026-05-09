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
- THEN the API MUST return the appropriate error response
- AND the CLI MUST output it as-is and exit with code 0

### Requirement: Default Argument Behavior

Commands that create resources SHALL use defaults rather than showing usage when no arguments are provided.

#### Scenario: Cluster create with no arguments

- GIVEN no arguments are provided
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST create a cluster using defaults (name=`my-cluster`, region=`us-east-1`, version=`4.15.0`)
- AND MUST NOT show a usage message

#### Scenario: NodePool create with no arguments

- GIVEN no arguments are provided
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST create 1 nodepool named `my-nodepool-1` with type `m4`
- AND MUST NOT show a usage message

### Requirement: Usage Messages for Required Arguments

Commands that require arguments SHALL show usage when arguments are missing.

#### Scenario: Cluster patch with no arguments

- GIVEN no arguments are provided
- WHEN the user runs `hf cluster patch`
- THEN the CLI MUST display: `Usage: hf cluster patch {spec|labels} [cluster_id]`
- AND list argument descriptions
- AND exit with code 1

#### Scenario: NodePool patch with no arguments

- GIVEN no arguments are provided
- WHEN the user runs `hf nodepool patch`
- THEN the CLI MUST display: `Usage: hf nodepool patch {spec|labels} [nodepool_id]`
- AND list argument descriptions
- AND exit with code 1
- NOTE: `spec` patches the `spec.counter` field (a counter integer stored as a string); the patch payload is valid JSON containing at least a `counter` property. No spec file is required.

#### Scenario: Adapter post-status with no arguments

- GIVEN no arguments are provided
- WHEN the user runs `hf cluster adapter post-status`
- THEN the CLI MUST display usage with argument descriptions and an example
- AND exit with code 1

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

