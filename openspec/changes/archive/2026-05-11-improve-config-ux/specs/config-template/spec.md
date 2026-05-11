## ADDED Requirements

### Requirement: Bundled environment template

The CLI SHALL ship a default environment template embedded in the binary, used as the seed for new environment files.

#### Scenario: Template file location and embedding

- **WHEN** the binary is compiled
- **THEN** the file at `cmd/assets/config-template.yaml` MUST be embedded using `//go:embed`
- **AND** it MUST contain all configuration sections with their default values:
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

#### Scenario: Template consistency with built-in defaults

- **WHEN** the `internal/config` package resolves a value via built-in defaults
- **THEN** the value MUST match the corresponding value in `cmd/assets/config-template.yaml`
- **AND** a unit test MUST assert this consistency by parsing the template file and comparing each key against the `defaults` map
