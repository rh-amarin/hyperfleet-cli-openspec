# API Client Specification — curl-mode delta

## MODIFIED Requirements

### Requirement: Client Initialization

The API client SHALL be initialized from the resolved configuration store.

#### Scenario: Create client with curl mode

- GIVEN the `--curl` flag is set
- WHEN `api.NewClient(baseURL, token, verbose, curlMode)` is called with `curlMode=true`
- THEN the returned client MUST print a curl command to stderr before every HTTP request

#### Scenario: Create client without curl mode

- GIVEN the `--curl` flag is not set
- WHEN `api.NewClient(baseURL, token, verbose, false)` is called
- THEN the client MUST NOT print any curl output

### Requirement: Verbose Request Logging

The API client SHALL log request details when verbose mode is enabled.

#### Scenario: Curl mode enabled

- GIVEN the client is created with `curlMode=true`
- WHEN any HTTP request is sent
- THEN the client MUST write to stderr before executing the request:
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

#### Scenario: Curl mode disabled

- GIVEN the client is created with `curlMode=false`
- WHEN a request is sent
- THEN no curl output MUST be written to stderr
