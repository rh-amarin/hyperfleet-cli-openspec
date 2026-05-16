## MODIFIED Requirements

### Requirement: Port Forward Management

The CLI SHALL manage port forwards to HyperFleet services running in Kubernetes.

Predefined services:
| name           | pod pattern    | namespace config key       | local port | remote port |
|----------------|----------------|----------------------------|------------|-------------|
| hyperfleet-api | hyperfleet-api | `hyperfleet.namespace`     | 8000       | 8000        |
| postgresql     | postgresql     | `hyperfleet.namespace`     | 5432       | 5432        |
| maestro-http   | maestro        | `maestro.namespace`        | 8100       | 8000        |
| maestro-grpc   | maestro        | `maestro.namespace`        | 8090       | 8090        |

The HyperFleet application namespace is read from `hyperfleet.namespace` (previously `kubernetes.namespace`). Maestro namespace remains `maestro.namespace`.

PID files stored at `~/.config/hf/pf-<name>.pid` — format: `<pid>\n<localPort>\n<remotePort>`.

`StartPortForward` returns a `StartResult` struct with `Name`, `PID`, `LocalPort`, `RemotePort`, `Namespace`, and `PodName` fields.

#### Scenario: Start port forwards — namespace shown in output

- **GIVEN** kubeconfig is accessible
- **WHEN** the user runs `hf kube port-forward start`
- **THEN** the CLI MUST start background port-forward processes for all 4 predefined services
- **AND** print `[INFO] Started <name> (<namespace>/<podName>): localhost:<localPort> → <remotePort> (pid <pid>)` for each service where the pod was found
- **AND** print `[INFO] Started <name> (<namespace>): localhost:<localPort> → <remotePort> (pid <pid>)` for services where no pod name is available
- **AND** wait 1 second after the last start line
- **AND** display the connectivity status table (using protocol-aware checks) after the wait

#### Scenario: Start port forward — single service

- **WHEN** the user runs `hf kube port-forward start <name>`
- **THEN** the CLI MUST start the named predefined service only
- **AND** display the namespace-enriched start line and connectivity status table for that service

#### Scenario: Start generic port forward

- **WHEN** the user runs `hf kube port-forward start <service> <localPort:remotePort>`
- **THEN** the CLI MUST start a generic port-forward for any service/port combination

#### Scenario: Stop port forwards

- **GIVEN** port forwards are running
- **WHEN** the user runs `hf kube port-forward stop`
- **THEN** the CLI MUST terminate all running port-forward processes

#### Scenario: Stop named port forward

- **WHEN** the user runs `hf kube port-forward stop <name>`
- **THEN** the CLI MUST terminate the named port-forward only

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

#### Scenario: Pod not running

- **GIVEN** the target service pod exists but is not in Running phase
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
