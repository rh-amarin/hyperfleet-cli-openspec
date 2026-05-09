## ADDED Requirements

### Requirement: Config Store

The CLI SHALL provide a `Store` type in `internal/config` that loads, persists, and resolves configuration values from the split YAML files (`config.yaml` and `state.yaml`).

#### Scenario: Store loads defaults

- **WHEN** `Store.Load()` is called with no existing config files
- **THEN** all sections MUST be populated with built-in defaults
- **AND** `config.yaml` and `state.yaml` MUST be created atomically with mode `0600`

#### Scenario: Store resolves precedence

- **WHEN** `Store.Get(section, key)` is called
- **THEN** the value MUST be resolved in order: HF_* env var > active env profile > config.yaml > built-in default

#### Scenario: RequireActiveEnvironment guards commands

- **WHEN** `Store.RequireActiveEnvironment()` is called with no active environment set
- **THEN** it MUST return an error containing `[ERROR] no active environment`

### Requirement: API Client

The CLI SHALL provide generic typed HTTP functions in `internal/api` for all CRUD operations against the HyperFleet API.

#### Scenario: Generic GET decodes response

- **WHEN** `api.Get[T](ctx, client, path)` is called and the server returns 2xx JSON
- **THEN** the response MUST be decoded into type T without error

#### Scenario: RFC 7807 errors parsed

- **WHEN** the server returns a non-2xx response with `application/problem+json` content type
- **THEN** the error MUST be an `*api.APIError` with `Status`, `Title`, and `Detail` populated
- **AND** `APIError.Error()` MUST return `[{status}] {title}: {detail}`

### Requirement: Resource Types

The CLI SHALL provide Go struct types in `internal/resource` for all HyperFleet API resources.

#### Scenario: Cluster JSON round-trip

- **WHEN** a Cluster JSON blob from the API is unmarshaled and re-marshaled
- **THEN** all fields MUST be preserved without data loss

#### Scenario: ListResponse is generic

- **WHEN** `ListResponse[Cluster]` or `ListResponse[NodePool]` is unmarshaled
- **THEN** the `Items` field MUST contain the correct typed elements

### Requirement: Output Printer

The CLI SHALL provide a `Printer` type in `internal/output` that dispatches output to json, table, or yaml format.

#### Scenario: Printer writes JSON with indent

- **WHEN** `Printer.Print(v)` is called with format "json"
- **THEN** the output MUST be pretty-printed JSON with 2-space indentation and a trailing newline

#### Scenario: StatusDot renders condition status

- **WHEN** `StatusDot("True", noColor)` is called
- **THEN** it MUST return a green-colored dot (or "True" in no-color mode)

#### Scenario: DynamicColumns orders correctly

- **WHEN** `DynamicColumns(fixed, conditionTypes)` is called
- **THEN** columns MUST be ordered: fixed → Available → alpha others → Reconciled
