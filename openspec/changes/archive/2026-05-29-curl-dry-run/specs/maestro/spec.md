## MODIFIED Requirements

### Requirement: Internal Maestro HTTP Client

The CLI SHALL provide an `internal/maestro` package with a typed HTTP client for the Maestro REST API.

#### Scenario: Client construction with curl mode

- GIVEN the `--curl` flag is set
- WHEN `maestro.NewFromConfig(s, curlMode)` is called with `curlMode=true`
- THEN the returned client MUST operate in dry-run mode for all HTTP methods

#### Scenario: Curl output for GET requests (dry-run)

- GIVEN `curlMode=true`
- WHEN `client.get(ctx, path, v)` is called
- THEN the client MUST write to stderr before returning:
  ```
  [CURL] curl -s "<url>" \
    -H 'Accept: application/json'
  ```
- AND the URL MUST be double-quoted
- AND the client MUST NOT send the HTTP request
- AND the method MUST return `maestro.ErrDryRun`

#### Scenario: Curl output for DELETE requests (dry-run)

- GIVEN `curlMode=true`
- WHEN `client.delete(ctx, path)` is called
- THEN the client MUST write to stderr before returning:
  ```
  [CURL] curl -s -X DELETE "<url>"
  ```
- AND the URL MUST be double-quoted
- AND the client MUST NOT send the HTTP request
- AND the method MUST return `maestro.ErrDryRun`

#### Scenario: Curl mode disabled

- GIVEN `curlMode=false`
- WHEN any request is sent
- THEN no curl output MUST be written to stderr
- AND HTTP requests MUST execute normally

## ADDED Requirements

### Requirement: Maestro Dry-Run Sentinel Error

The maestro client package SHALL expose a sentinel error for dry-run mode.

#### Scenario: ErrDryRun identity

- GIVEN `maestro.ErrDryRun` is defined
- WHEN a caller receives an error from a curl-mode request
- THEN `errors.Is(err, maestro.ErrDryRun)` MUST be true
