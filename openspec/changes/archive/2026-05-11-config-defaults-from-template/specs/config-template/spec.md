# Delta for config-template

## MODIFIED Requirements

### Requirement: Bundled environment template

The CLI SHALL ship a default environment template embedded in the binary. The template serves two purposes: it seeds new environment files created by `hf config env create`, and it provides the built-in default values for the `internal/config` package.

#### Scenario: Template file location and embedding

- **WHEN** the binary is compiled
- **THEN** the file at `internal/config/assets/config-template.yaml` MUST be embedded using `//go:embed` in the `internal/config` package
- **AND** the embedded bytes MUST be exported as `config.ConfigTemplateYAML []byte`
- **AND** `cmd/config.go` MUST use `config.ConfigTemplateYAML` (not its own embed) when seeding new environment files

#### Scenario: Template drives built-in defaults

- **WHEN** the `internal/config` package is initialized
- **THEN** it MUST parse `config-template.yaml` into the `defaults` map at `init()` time
- **AND** `Store.Get()` MUST return values from this parsed map when no active environment profile supplies a value
- **AND** there MUST be no separate hardcoded default map in Go source code

#### Scenario: Template content

- **WHEN** the binary is compiled
- **THEN** `internal/config/assets/config-template.yaml` MUST contain all configuration sections with their default values:
  ```yaml
  hyperfleet:
    api-url: "http://localhost:8000"
    api-version: "v1"
    token: ""
    gcp-project: "hcm-hyperfleet"

  kubernetes:
    context: ""
    namespace: ""

  maestro:
    consumer: "cluster1"
    http-endpoint: "http://localhost:8100"
    grpc-endpoint: "localhost:8090"
    namespace: "maestro"

  port-forward:
    api-port: "8000"
    pg-port: "5432"
    maestro-http-port: "8100"
    maestro-http-remote-port: "8000"
    maestro-grpc-port: "8090"
    maestro-grpc-remote-port: "8090"

  database:
    host: "localhost"
    port: "5432"
    name: "hyperfleet"
    user: "hyperfleet"
    password: "foobar-bizz-buzz"

  rabbitmq:
    host: "localhost"
    mgmt-port: "15672"
    user: "guest"
    password: "guest"
    vhost: "/"

  registry:
    name: ""
  ```

