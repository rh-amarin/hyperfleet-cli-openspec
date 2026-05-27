## ADDED Requirements

### Requirement: Kubeconfig Path Resolution

The CLI SHALL resolve the kubeconfig file path used for all Kubernetes operations (port-forward, logs, debug, auto-port-forward) with a defined precedence order.

#### Scenario: Kubeconfig precedence chain

- **WHEN** the CLI needs a kubeconfig path
- **THEN** the precedence order MUST be (highest to lowest):
  1. `--kubeconfig` CLI flag
  2. `HF_KUBECONFIG` environment variable or `kubernetes.kubeconfig` from the active environment file
  3. `KUBECONFIG` environment variable
  4. `~/.kube/config`
- **AND** an empty `kubernetes.kubeconfig` value MUST NOT block lower-precedence sources

#### Scenario: Per-environment kubeconfig

- **GIVEN** the active environment file sets `kubernetes.kubeconfig: /path/to/kubeconfig`
- **AND** `KUBECONFIG` is unset
- **WHEN** the user runs any command that uses Kubernetes (e.g. `hf kube port-forward start`)
- **THEN** the CLI MUST use `/path/to/kubeconfig`

## MODIFIED Requirements

### Requirement: Port Forward Management

The CLI SHALL manage port forwards to HyperFleet services running in Kubernetes. When a matching Kubernetes Service exists in the target namespace and exposes the configured remote port, the CLI MUST port-forward via that Service. When no suitable Service exists, the CLI MUST fall back to resolving a pod by name pattern (existing behavior).

Predefined services:
| name           | service name   | pod pattern (fallback) | namespace config key       | local port | remote port |
|----------------|----------------|------------------------|----------------------------|------------|-------------|
| hyperfleet-api | hyperfleet-api | hyperfleet-api         | `hyperfleet.namespace`     | 8000       | 8000        |
| postgresql     | postgresql     | postgresql             | `hyperfleet.namespace`     | 5432       | 5432        |
| maestro-http   | maestro        | maestro                | `maestro.namespace`        | 8100       | 8000        |
| maestro-grpc   | maestro        | maestro                | `maestro.namespace`        | 8090       | 8090        |

The HyperFleet application namespace is read from `hyperfleet.namespace`. Maestro namespace remains `maestro.namespace`.

Kubeconfig path is read from `kubernetes.kubeconfig` (see Kubeconfig Path Resolution requirement).

PID files stored at `~/.config/hf/pf-<name>.pid`.

`StartPortForward` returns a `StartResult` struct with `Name`, `PID`, `LocalPort`, `RemotePort`, `Namespace`, `TargetKind` (`service` or `pod`), and `TargetName` fields.

#### Scenario: Kubeconfig not found

- **GIVEN** the kubeconfig file is not found at the resolved path
- **WHEN** the user runs `hf kube port-forward start`
- **THEN** the CLI MUST display `[ERROR] kubeconfig not found at <path>. Set kubernetes.kubeconfig, KUBECONFIG, or use --kubeconfig.`
