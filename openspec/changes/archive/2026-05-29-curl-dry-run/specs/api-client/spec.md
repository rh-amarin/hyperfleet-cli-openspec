## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Dry-Run Sentinel Error

The API client package SHALL expose a sentinel error for dry-run mode.

#### Scenario: ErrDryRun identity

- GIVEN `api.ErrDryRun` is defined
- WHEN a caller receives an error from a curl-mode request
- THEN `errors.Is(err, api.ErrDryRun)` MUST be true
- AND the error MUST NOT be an `APIError`
