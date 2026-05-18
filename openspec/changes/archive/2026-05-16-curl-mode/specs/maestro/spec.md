# Maestro Operations Specification — curl-mode delta

## MODIFIED Requirements

### Requirement: Internal Maestro HTTP Client

The CLI SHALL provide an `internal/maestro` package with a typed HTTP client for the Maestro REST API.

#### Scenario: Client construction with curl mode

- GIVEN the `--curl` flag is set
- WHEN `maestro.NewFromConfig(s, curlMode)` is called with `curlMode=true`
- THEN the returned client MUST print a curl command to stderr before every HTTP request

#### Scenario: Curl output for GET requests

- GIVEN `curlMode=true`
- WHEN `client.get(ctx, path, v)` is called
- THEN the client MUST write to stderr before executing the request:
  ```
  [CURL] curl -s "<url>" \
    -H 'Accept: application/json'
  ```
- AND the URL MUST be double-quoted

#### Scenario: Curl output for DELETE requests

- GIVEN `curlMode=true`
- WHEN `client.delete(ctx, path)` is called
- THEN the client MUST write to stderr before executing the request:
  ```
  [CURL] curl -s -X DELETE "<url>"
  ```
- AND the URL MUST be double-quoted

#### Scenario: Curl mode disabled

- GIVEN `curlMode=false`
- WHEN any request is sent
- THEN no curl output MUST be written to stderr
