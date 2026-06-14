## ADDED Requirements

### Requirement: Identity Header Config Keys

The CLI SHALL read optional identity header configuration from the `hyperfleet` section of the active environment file.

#### Scenario: identity-header and identity-value resolve from environment file

- **GIVEN** the active environment file contains:
  ```yaml
  hyperfleet:
    identity-header: X-HyperFleet-Identity
    identity-value: openspec@test.com
  ```
- **WHEN** `s.Get("hyperfleet", "identity-header")` and `s.Get("hyperfleet", "identity-value")` are called
- **THEN** they MUST return `X-HyperFleet-Identity` and `openspec@test.com` respectively
- **AND** these values MUST be passed to `api.NewClient` as `identityHeader` and `identityValue`

#### Scenario: identity-header defaults to empty

- **GIVEN** the active environment file does not contain `hyperfleet.identity-header`
- **WHEN** `s.Get("hyperfleet", "identity-header")` is called
- **THEN** it MUST return an empty string
- **AND** no identity header MUST be injected into API requests

#### Scenario: Config template includes identity header keys as comments

- **GIVEN** a newly created environment from the embedded config template
- **WHEN** the environment file is written
- **THEN** it MUST include commented-out `identity-header` and `identity-value` keys under the `hyperfleet` section so operators are aware of the option
