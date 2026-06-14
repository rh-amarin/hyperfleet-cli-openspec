# API Identity Header Specification

## Purpose

Allow operators to inject a configurable identity header into every HyperFleet API HTTP request, enabling server-side access control and audit logging without modifying source code.

## ADDED Requirements

### Requirement: Configurable Identity Header Injection

The API client SHALL support an optional identity header that, when configured, is attached to every outgoing HTTP request.

#### Scenario: Identity header injected when configured

- **GIVEN** `hyperfleet.identity-header` is set to a non-empty string (e.g., `X-HyperFleet-Identity`)
- **AND** `hyperfleet.identity-value` is set (e.g., `openspec@test.com`)
- **WHEN** any API method (GET, POST, PATCH, DELETE) sends a request
- **THEN** the HTTP request MUST include the header `X-HyperFleet-Identity: openspec@test.com`

#### Scenario: Identity header absent when not configured

- **GIVEN** `hyperfleet.identity-header` is empty or unset
- **WHEN** any API method sends a request
- **THEN** no identity header MUST be added to the request
- **AND** all other request headers MUST be unchanged

#### Scenario: Identity header appears in curl dry-run output

- **GIVEN** `hyperfleet.identity-header` is configured and `--curl` flag is set
- **WHEN** any typed HTTP method is called
- **THEN** the curl output written to stderr MUST include `-H '<identity-header>: <identity-value>'`

#### Scenario: Empty identity-value with non-empty identity-header

- **GIVEN** `hyperfleet.identity-header` is set to a non-empty string
- **AND** `hyperfleet.identity-value` is empty
- **WHEN** any API method sends a request
- **THEN** the request MUST include the configured header with an empty value
