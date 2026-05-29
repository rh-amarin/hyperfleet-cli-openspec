# API Client Specification

## Purpose

Provide a shared, type-safe HTTP client for all HyperFleet API operations, handling
authentication, error parsing, and verbose logging so that command implementations
can focus on business logic rather than HTTP plumbing.
## Requirements
### Requirement: Resource href Construction

The API client package SHALL expose a helper for building canonical href URLs for HyperFleet resources.

#### Scenario: Build resource href

- GIVEN the client is initialized with `api-url` and `api-version`
- WHEN `api.ResourceHref(resourcePath string) string` is called
- THEN it MUST return `{api-url}/api/hyperfleet/{api-version}/{resourcePath}`
- AND all CloudEvent payloads, list responses, and any other place that requires a resource URL MUST use this helper to guarantee consistency

Examples:
- Cluster href: `api.ResourceHref("clusters/{cluster_id}")`
- NodePool href: `api.ResourceHref("clusters/{cluster_id}/nodepools/{nodepool_id}")`

No hardcoded hostnames or path prefixes are allowed outside this helper.

Note: Other specs use `/api/hyperfleet/v1/` in endpoint path examples. The `v1` is illustrative only — the actual version is always substituted from the `hyperfleet.api-version` configuration value (default: `v1`). All paths are constructed through this helper or the client's base URL.

---

### Requirement: Client Initialization

The API client SHALL be initialized from the resolved configuration store.

#### Scenario: Create client with curl mode

- GIVEN the `--curl` flag is set
- WHEN `api.NewClient(baseURL, token, verbose, curlMode)` is called with `curlMode=true`
- THEN the returned client MUST operate in dry-run mode for all HTTP methods

#### Scenario: Create client without curl mode

- GIVEN the `--curl` flag is not set
- WHEN `api.NewClient(baseURL, token, verbose, false)` is called
- THEN the client MUST NOT print any curl output
- AND MUST execute HTTP requests normally

### Requirement: Typed HTTP Methods

The API client SHALL provide generic typed methods for CRUD operations.

#### Scenario: GET request with typed response

- GIVEN a valid API path (e.g., `clusters/{id}`)
- WHEN `Get[T]` is called with the path
- THEN the client MUST send an HTTP GET request to `{baseURL}{path}`
- AND set `Content-Type: application/json` and `Accept: application/json` headers
- AND unmarshal the response body into type `T`
- AND return the typed result and nil error on 2xx status

#### Scenario: POST request with typed request and response

- GIVEN a valid API path and a request body
- WHEN `Post[T]` is called with the path and body
- THEN the client MUST marshal the body as JSON
- AND send an HTTP POST request
- AND unmarshal the 2xx response body into type `T`

#### Scenario: PATCH request with typed request and response

- GIVEN a valid API path and a partial update body
- WHEN `Patch[T]` is called with the path and body
- THEN the client MUST marshal the body as JSON
- AND send an HTTP PATCH request
- AND unmarshal the 2xx response body into type `T`

#### Scenario: DELETE request with typed response

- GIVEN a valid API path
- WHEN `Delete[T]` is called with the path
- THEN the client MUST send an HTTP DELETE request
- AND unmarshal the response body into type `T`
- AND return the deleted resource (the HyperFleet API returns the full object on delete)

### Requirement: RFC 7807 Error Parsing

The API client SHALL parse non-2xx responses as RFC 7807 Problem Details.

#### Scenario: Parse 404 API error

- GIVEN the API returns a 404 response with an RFC 9457 JSON body (content-type `application/problem+json`)
- WHEN the client parses the response
- THEN it MUST return an `APIError` with required fields: `type`, `title`, `status`
- AND optional fields: `detail`, `instance`, `code`, `timestamp`, `trace_id`, `errors` ([]ValidationError)
- AND the `APIError` MUST implement Go's `error` interface
- AND `Error()` MUST return a formatted string: `[{status}] {title}: {detail}`

#### Scenario: Parse validation error with field-level details

- GIVEN the API returns a 400 response with `errors` array containing field-level validation failures
- WHEN the client parses the response
- THEN the `APIError.Errors` field MUST contain `ValidationError` entries with `field`, `message`, and optional `value`, `constraint`

#### Scenario: Non-JSON error response

- GIVEN the API returns a non-2xx response with a non-JSON body (e.g., plain text, HTML)
- WHEN the client parses the response
- THEN it MUST return an `APIError` with the `status` field set to the HTTP status code
- AND `detail` set to the raw response body (truncated to 500 characters) followed by `... [truncated]` if truncated
- AND `title` set to the HTTP status text
- AND if the response body appears to be HTML (starts with `<!` or `<html`), the `detail` MUST be prefixed with: `Received HTML response (possibly not the HyperFleet API). Verify the API URL with 'hf config show'.`

#### Scenario: Network error

- GIVEN the API is unreachable (connection refused, DNS failure, timeout)
- WHEN the client attempts a request
- THEN it MUST return a Go error (not an `APIError`)
- AND the error message MUST include the original network error details

#### Scenario: Request timeout

- GIVEN the API does not respond within the 30-second timeout
- WHEN the client times out
- THEN it MUST return a Go error (not an `APIError`) with message: `[ERROR] Request to <URL> timed out after 30s. Check your network connection and API server.`
- AND exit with a non-zero code

### Requirement: Verbose Request Logging

The API client SHALL log request details when verbose mode is enabled.

#### Scenario: Curl mode enabled (dry-run)

- GIVEN the client is created with `curlMode=true`
- WHEN any typed HTTP method (`Get`, `Post`, `Patch`, `Delete`, `Put`) is called
- THEN the client MUST write to stderr before returning:
  ```
  [CURL] curl -s -X <METHOD> "<URL>" \
    -H 'Accept: application/json' \
    [-H 'Content-Type: application/json' \]
    [-H 'Authorization: Bearer <token>' \]
    [-d '<json-body>']
  ```
- AND `Content-Type` header line MUST only appear when the request has a body
- AND `Authorization` header line MUST only appear when a token is configured
- AND the URL MUST be double-quoted to safely handle single quotes in query parameters
- AND the output MUST go to stderr so stdout data output is not affected
- AND the client MUST NOT send the HTTP request (no call to the HTTP transport)
- AND the method MUST return the zero value for `T` and `api.ErrDryRun`

#### Scenario: Curl mode disabled

- GIVEN the client is created with `curlMode=false`
- WHEN a request is sent
- THEN no curl output MUST be written to stderr
- AND the HTTP request MUST be executed normally

### Requirement: Context Propagation

The API client SHALL support Go context for cancellation and timeouts.

#### Scenario: Request with context cancellation

- GIVEN a context is cancelled while a request is in-flight
- WHEN the client detects the cancellation
- THEN the HTTP request MUST be aborted
- AND the client MUST return `context.Canceled` or `context.DeadlineExceeded`

### Requirement: Dry-Run Sentinel Error

The API client package SHALL expose a sentinel error for dry-run mode.

#### Scenario: ErrDryRun identity

- GIVEN `api.ErrDryRun` is defined
- WHEN a caller receives an error from a curl-mode request
- THEN `errors.Is(err, api.ErrDryRun)` MUST be true
- AND the error MUST NOT be an `APIError`

