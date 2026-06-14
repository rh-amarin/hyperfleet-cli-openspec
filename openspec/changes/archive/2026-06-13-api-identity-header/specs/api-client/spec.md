## MODIFIED Requirements

### Requirement: Client Initialization

The API client SHALL be initialized from the resolved configuration store.

#### Scenario: Create client with curl mode

- GIVEN the `--curl` flag is set
- WHEN `api.NewClient(baseURL, token, verbose, curlMode, identityHeader, identityValue)` is called with `curlMode=true`
- THEN the returned client MUST operate in dry-run mode for all HTTP methods

#### Scenario: Create client without curl mode

- GIVEN the `--curl` flag is not set
- WHEN `api.NewClient(baseURL, token, verbose, false, identityHeader, identityValue)` is called
- THEN the client MUST NOT print any curl output
- AND MUST execute HTTP requests normally

#### Scenario: Create client with identity header

- GIVEN `identityHeader` is a non-empty string
- WHEN `api.NewClient(baseURL, token, verbose, curlMode, identityHeader, identityValue)` is called
- THEN the returned client MUST store the identity header name and value
- AND attach them to every outgoing HTTP request via the shared `do()` method

#### Scenario: Create client without identity header

- GIVEN `identityHeader` is an empty string
- WHEN `api.NewClient` is called
- THEN the client MUST NOT add any identity header to requests

### Requirement: Verbose Request Logging

The API client SHALL log request details when verbose mode is enabled.

#### Scenario: Curl mode enabled (dry-run) with identity header

- GIVEN the client is created with `curlMode=true` and a non-empty `identityHeader`
- WHEN any typed HTTP method (`Get`, `Post`, `Patch`, `Delete`) is called
- THEN the client MUST write to stderr:
  ```
  [CURL] curl -s -X <METHOD> "<URL>" \
    -H 'Accept: application/json' \
    [-H 'Content-Type: application/json' \]
    [-H 'Authorization: Bearer <token>' \]
    [-H '<identity-header>: <identity-value>' \]
    [-d '<json-body>']
  ```
- AND the identity header line MUST only appear when `identityHeader` is non-empty
- AND the client MUST NOT send the HTTP request
- AND the method MUST return the zero value for `T` and `api.ErrDryRun`
