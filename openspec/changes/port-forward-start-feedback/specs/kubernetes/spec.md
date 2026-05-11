## MODIFIED Requirements

### Requirement: Port Forward Management

The CLI SHALL manage port forwards to HyperFleet services running in Kubernetes.

Predefined services:
| name           | pod pattern    | namespace  | local port | remote port |
|----------------|----------------|------------|------------|-------------|
| hyperfleet-api | hyperfleet-api | amarin-ns1 | 8000       | 8000        |
| postgresql     | postgresql     | amarin-ns1 | 5432       | 5432        |
| maestro-http   | maestro        | maestro    | 8100       | 8000        |
| maestro-grpc   | maestro        | maestro    | 8090       | 8090        |

Port values configurable via `cfg.PortForward.*` in `~/.config/hf/config.yaml`.
Maestro namespace configurable via `cfg.Maestro.Namespace`.

PID files stored at `~/.config/hf/pf-<name>.pid` — format: `<pid>\n<localPort>\n<remotePort>`.

`StartPortForward` returns a `StartResult` struct with `Name`, `PID`, `LocalPort`, `RemotePort`, `Namespace`, and `PodName` fields.

#### Scenario: Start port forwards

- GIVEN kubeconfig is accessible
- WHEN the user runs `hf kube port-forward start`
- THEN the CLI MUST start background port-forward processes for all 4 predefined services
- AND print `[INFO] Started <name> (<namespace>/<podName>): localhost:<localPort> → <remotePort> (pid <pid>)` for each service where the pod was found
- AND print `[INFO] Started <name>: localhost:<localPort> → <remotePort> (pid <pid>)` for services where no pod was found
- AND display the port-forward status table after the last start line

- WHEN the user runs `hf kube port-forward start <name>`
- THEN the CLI MUST start the named predefined service only
- AND display the enriched start line and status table for that service

- WHEN the user runs `hf kube port-forward start <service> <localPort:remotePort>`
- THEN the CLI MUST start a generic port-forward for any service/port combination

#### Scenario: Stop port forwards

- GIVEN port forwards are running
- WHEN the user runs `hf kube port-forward stop`
- THEN the CLI MUST terminate all running port-forward processes

- WHEN the user runs `hf kube port-forward stop <name>`
- THEN the CLI MUST terminate the named port-forward only

#### Scenario: Check port forward status

- WHEN the user runs `hf kube port-forward status`
- THEN the CLI MUST display one line per PID file:
  - For alive port forwards: `  ● <name> - localhost:<port> (PID: <pid>)` with a green bullet (●)
  - For dead or stale port forwards: `  ● <name> - localhost:<port> (PID: <pid>) [DEAD]` with a red bullet (●)

#### Scenario: Pod not running

- GIVEN the target service pod exists but is not in Running phase (e.g., Pending or CrashLoopBackOff)
- WHEN the user runs `hf kube port-forward start [service]`
- THEN the CLI MUST display `[WARN] <service>: pod not ready (phase: <phase>). Port-forward may not succeed.`
- AND attempt the port-forward anyway

#### Scenario: Port number validation

- GIVEN a custom `<localPort:remotePort>` argument is provided
- WHEN the user runs `hf kube port-forward start <service> <localPort:remotePort>`
- THEN both port values MUST be valid integers in the range 1–65535
- AND if either port is invalid, the CLI MUST display `[ERROR] Invalid port '<value>'. Must be an integer between 1 and 65535.` and exit 1

#### Scenario: Kubeconfig not found

- GIVEN the kubeconfig file is not found at the resolved path (flag → `KUBECONFIG` env → `~/.kube/config`)
- WHEN any `hf kube` command is invoked
- THEN the CLI MUST display `[ERROR] kubeconfig not found at <path>. Set KUBECONFIG or use --kubeconfig.`
- AND exit with code 1
