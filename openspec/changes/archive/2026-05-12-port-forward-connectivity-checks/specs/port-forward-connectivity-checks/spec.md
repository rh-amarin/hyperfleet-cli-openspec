## ADDED Requirements

### Requirement: Protocol-aware port-forward connectivity probes

Package `internal/kube` SHALL provide exported functions to test connectivity to each predefined HyperFleet service using that service's native protocol.

```go
func CheckAPIConnectivity(port int) error
func CheckMaestroHTTPConnectivity(port int) error
func CheckPostgresConnectivity(port int, host, dbname, user, password string) error
func CheckMaestroGRPCConnectivity(port int) error
```

Each function MUST use a short timeout (≤ 2 seconds) so that `status` remains responsive when a service is down.

#### Scenario: API connectivity check passes

- **WHEN** `CheckAPIConnectivity` is called and the port is bound to an HTTP server that responds
- **THEN** it MUST return `nil`

#### Scenario: API connectivity check fails

- **WHEN** `CheckAPIConnectivity` is called and no process is listening on the port
- **THEN** it MUST return a non-nil error

#### Scenario: Maestro HTTP connectivity check passes

- **WHEN** `CheckMaestroHTTPConnectivity` is called and the port is bound to an HTTP server that responds
- **THEN** it MUST return `nil`

#### Scenario: Maestro HTTP connectivity check fails

- **WHEN** `CheckMaestroHTTPConnectivity` is called and no process is listening on the port
- **THEN** it MUST return a non-nil error

#### Scenario: Postgres connectivity check fails when port is closed

- **WHEN** `CheckPostgresConnectivity` is called and no process is listening on the port
- **THEN** it MUST return a non-nil error within the timeout

#### Scenario: Maestro gRPC connectivity check fails when port is closed

- **WHEN** `CheckMaestroGRPCConnectivity` is called and no process is listening on the port
- **THEN** it MUST return a non-nil error within the timeout

#### Scenario: Maestro gRPC treats UNIMPLEMENTED as pass

- **WHEN** `CheckMaestroGRPCConnectivity` is called and the gRPC server responds but does not implement the Health service
- **THEN** it MUST return `nil` (server is reachable, Health check not mandatory)
