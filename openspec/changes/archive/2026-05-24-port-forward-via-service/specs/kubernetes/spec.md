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

PID files stored at `~/.config/hf/pf-<name>.pid`.

`StartPortForward` returns a `StartResult` struct with `Name`, `PID`, `LocalPort`, `RemotePort`, `Namespace`, `TargetKind` (`service` or `pod`), and `TargetName` fields.

#### Scenario: Start port forwards — namespace shown in output

- **GIVEN** kubeconfig is accessible
- **WHEN** the user runs `hf kube port-forward start`
- **THEN** the first line of output MUST be `[INFO] Kubernetes context: <contextName>`
- **AND** the CLI MUST start background port-forward processes for all 4 predefined services
- **AND** print `[INFO] Started <name> (<namespace>/svc/<serviceName>): localhost:<localPort> → <remotePort> (pid <pid>)` for each service forwarded via a Kubernetes Service
- **AND** print `[INFO] Started <name> (<namespace>/pod/<podName>): localhost:<localPort> → <remotePort> (pid <pid>)` for each service forwarded via pod fallback
- **AND** wait 1 second after the last start line
- **AND** display the connectivity status table (using protocol-aware checks) after the wait

#### Scenario: Start port forward — single service

- **WHEN** the user runs `hf kube port-forward start <name>`
- **THEN** the CLI MUST start the named predefined service only
- **AND** display the namespace-enriched start line and connectivity status table for that service

#### Scenario: Start generic port forward

- **WHEN** the user runs `hf kube port-forward start <service> <localPort:remotePort>`
- **THEN** the CLI MUST attempt to port-forward `<service>` as a Kubernetes Service in the HyperFleet namespace first
- **AND** if no suitable Service exists, MUST fall back to pod pattern matching using `<service>` as the pattern
- **AND** MUST start a port-forward for the resolved target

#### Scenario: Service preferred when available

- **GIVEN** a Kubernetes Service named `<serviceName>` exists in the target namespace
- **AND** the Service exposes a port matching the configured remote port
- **AND** the Service has ready Endpoints with a backing pod for that port
- **WHEN** the user runs `hf kube port-forward start` for the corresponding predefined service (or a generic forward using that service name)
- **THEN** the CLI MUST resolve the target pod from the Service Endpoints (same behavior as `kubectl port-forward svc/...`)
- **AND** MUST NOT require pod pattern matching when Endpoints resolution succeeds
- **AND** the start line MUST show the `svc/` prefix

#### Scenario: Service missing — pod fallback

- **GIVEN** no Kubernetes Service matching the expected name and remote port exists in the target namespace
- **AND** a pod whose name contains the fallback pod pattern exists
- **WHEN** the user runs `hf kube port-forward start [service]`
- **THEN** the CLI MUST resolve the target pod by pattern matching (existing behavior)
- **AND** port-forward via the pod subresource (`pods/<podName>/portforward`)
- **AND** the start line MUST show the `pod/` prefix

#### Scenario: Stop port forwards

- **GIVEN** port forwards are running
- **WHEN** the user runs `hf kube port-forward stop`
- **THEN** the CLI MUST terminate all running port-forward processes

#### Scenario: Stop named port forward

- **WHEN** the user runs `hf kube port-forward stop <name>`
- **THEN** the CLI MUST terminate the named port-forward only

#### Scenario: Check port forward status — context header

- **WHEN** the user runs `hf kube port-forward status`
- **THEN** the first line of output MUST be `[INFO] Kubernetes context: <contextName>`
- **AND** `<contextName>` MUST be the context that will actually be used (the `kubernetes.context` config override if set, otherwise the kubeconfig's current-context)
- **AND** the port-forward status table MUST follow

#### Scenario: Check port forward status — protocol-aware

- **WHEN** the user runs `hf kube port-forward status`
- **THEN** the CLI MUST display one line per PID file using protocol-aware connectivity checks:
  - `hyperfleet-api`: HTTP GET to `http://localhost:<port>/api/hyperfleet/v1`
  - `postgresql`: pgx Ping to `localhost:<port>` with configured credentials
  - `maestro-http`: HTTP GET to `http://localhost:<port>/api/maestro/v1`
  - `maestro-grpc`: gRPC Health/Check to `localhost:<port>`
- **AND** display `  ✓ <name> - localhost:<port> (PID: <pid>)` in green when the protocol check passes
- **AND** display `  ✗ <name> - localhost:<port> (PID: <pid>) [NOT CONNECTED]` in red when the protocol check fails
- **AND** use the same connectivity display for the post-start wait output

#### Scenario: Port forward status — unknown service falls back to TCP

- **GIVEN** a generic (non-predefined) port-forward PID file exists
- **WHEN** the user runs `hf kube port-forward status`
- **THEN** the CLI MUST use TCP dial as the connectivity check for that entry

#### Scenario: Context resolved from config override

- **GIVEN** `kubernetes.context` is set to a non-empty value in the active config (or via `HF_CONTEXT`)
- **WHEN** the user runs `hf kube port-forward start` or `hf kube port-forward status`
- **THEN** the context header MUST show the configured override value
- **AND** all Kubernetes API calls MUST use that context

#### Scenario: Context resolved from kubeconfig current-context

- **GIVEN** `kubernetes.context` is empty (default)
- **WHEN** the user runs `hf kube port-forward start` or `hf kube port-forward status`
- **THEN** the context header MUST show the name of the kubeconfig's current-context

#### Scenario: Context resolution failure

- **GIVEN** the kubeconfig file is missing or the named context does not exist
- **WHEN** the CLI attempts to resolve the context name
- **THEN** the CLI MUST print `[WARN] Could not resolve kubernetes context: <reason>` and continue

#### Scenario: Pod not running (pod fallback only)

- **GIVEN** no suitable Kubernetes Service exists for the target
- **AND** the fallback target pod exists but is not in Running phase
- **WHEN** the user runs `hf kube port-forward start [service]`
- **THEN** the CLI MUST display `[WARN] <service>: pod not ready (phase: <phase>). Port-forward may not succeed.`
- **AND** attempt the port-forward anyway

#### Scenario: Port number validation

- **GIVEN** a custom `<localPort:remotePort>` argument is provided
- **WHEN** the user runs `hf kube port-forward start <service> <localPort:remotePort>`
- **THEN** both port values MUST be valid integers in the range 1–65535
- **AND** if either port is invalid, the CLI MUST display `[ERROR] Invalid port '<value>'. Must be an integer between 1 and 65535.` and exit 1

#### Scenario: Kubeconfig not found

- **GIVEN** the kubeconfig file is not found at the resolved path
- **WHEN** any `hf kube` command is invoked
- **THEN** the CLI MUST display `[ERROR] kubeconfig not found at <path>. Set KUBECONFIG or use --kubeconfig.`
- **AND** exit with code 1

## ADDED Requirements

### Requirement: Ephemeral Port-Forward Service Resolution

When `hyperfleet.auto-port-forward` is enabled, ephemeral in-process port-forwards MUST use the same service-first resolution as persistent port-forwards before falling back to pod pattern matching.

#### Scenario: Auto port-forward uses service when available

- **GIVEN** `hyperfleet.auto-port-forward` is `"true"`
- **AND** a Kubernetes Service for `hyperfleet-api` exists with port 8000 and ready Endpoints
- **WHEN** any non-bypass command runs
- **THEN** the CLI MUST establish the ephemeral forward via the pod resolved from Service Endpoints
- **AND** MUST NOT require pod pattern matching when Endpoints resolution succeeds

#### Scenario: Auto port-forward falls back to pod

- **GIVEN** `hyperfleet.auto-port-forward` is `"true"`
- **AND** no suitable Kubernetes Service exists for the target
- **WHEN** any non-bypass command runs
- **THEN** the CLI MUST fall back to pod pattern resolution (existing behavior)
