# Kubernetes Utilities Specification

## Purpose

Provide CLI commands for Kubernetes-related operations including port-forwarding to HyperFleet services, in-cluster curl execution, debug pod creation, and log tailing for pods and adapters. Implemented using `k8s.io/client-go` — no `kubectl` binary dependency.
## Requirements
### Requirement: Port Forward Management

The CLI SHALL manage port forwards to HyperFleet services running in Kubernetes. When a matching Kubernetes Service exists in the target namespace and exposes the configured remote port, the CLI MUST port-forward via that Service. When no suitable Service exists, the CLI MUST fall back to resolving a pod by name pattern (existing behavior).

Predefined services:
| name           | service name   | pod pattern (fallback) | namespace config key       | local port | remote port |
|----------------|----------------|------------------------|----------------------------|------------|-------------|
| hyperfleet-api | hyperfleet-api | hyperfleet-api         | `hyperfleet.namespace`     | 8000       | 8000        |
| postgresql     | postgresql     | postgresql             | `hyperfleet.namespace`     | 5432       | 5432        |
| maestro-http   | maestro        | maestro                | `maestro.namespace`        | 8100       | 8000        |
| maestro-grpc   | maestro        | maestro                | `maestro.namespace`        | 8090       | 8090        |
| rabbitmq       | rabbitmq       | rabbitmq               | `hyperfleet.namespace`     | 15672      | 15672       |

The HyperFleet application namespace is read from `hyperfleet.namespace`. Maestro namespace remains `maestro.namespace`. RabbitMQ local port defaults to `15672` and MAY be overridden via `port-forward.rabbitmq-mgmt-port`.

PID files stored at `~/.config/hf/pf-<name>.pid`.

`StartPortForward` returns a `StartResult` struct with `Name`, `PID`, `LocalPort`, `RemotePort`, `Namespace`, `TargetKind` (`service` or `pod`), and `TargetName` fields.

#### Scenario: Start port forwards — namespace shown in output

- **GIVEN** kubeconfig is accessible
- **WHEN** the user runs `hf kube port-forward start`
- **THEN** the first line of output MUST be `[INFO] Kubernetes context: <contextName>`
- **AND** the CLI MUST stop all tracked port-forwards before starting new ones
- **AND** print `[INFO] Stopped <name>` for each stopped forward
- **AND** the CLI MUST start background port-forward processes for all 5 predefined services
- **AND** print `[INFO] Started <name> (<namespace>/svc/<serviceName>): localhost:<localPort> → <remotePort> (pid <pid>)` for each service forwarded via a Kubernetes Service
- **AND** print `[INFO] Started <name> (<namespace>/pod/<podName>): localhost:<localPort> → <remotePort> (pid <pid>)` for each service forwarded via pod fallback
- **AND** wait 1 second after the last start line
- **AND** display the connectivity status table (using protocol-aware checks) after the wait

#### Scenario: Start port forward — single service

- **WHEN** the user runs `hf kube port-forward start <name>`
- **THEN** the CLI MUST stop the tracked port-forward for `<name>` when one exists
- **AND** the CLI MUST start the named predefined service only
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

### Requirement: Protocol-Aware Connectivity Probes

The CLI SHALL test connectivity to each predefined HyperFleet service using that service's native protocol. Each connectivity check MUST use a short timeout (≤ 2 seconds) so that `status` remains responsive when a service is down.

#### Scenario: API connectivity check passes

- **WHEN** the `hyperfleet-api` port-forward connectivity check runs and the port is bound to an HTTP server that responds
- **THEN** the check MUST succeed (display green)

#### Scenario: API connectivity check fails

- **WHEN** the `hyperfleet-api` connectivity check runs and no process is listening on the port
- **THEN** the check MUST fail (display red) within 2 seconds

#### Scenario: Maestro HTTP connectivity check fails

- **WHEN** the `maestro-http` connectivity check runs and no process is listening on the port
- **THEN** the check MUST fail (display red) within 2 seconds

#### Scenario: Postgres connectivity check fails when port is closed

- **WHEN** the `postgresql` connectivity check runs and no process is listening on the port
- **THEN** the check MUST fail (display red) within 2 seconds

#### Scenario: Maestro gRPC connectivity check fails when port is closed

- **WHEN** the `maestro-grpc` connectivity check runs and no process is listening on the port
- **THEN** the check MUST fail (display red) within 2 seconds

#### Scenario: Maestro gRPC treats UNIMPLEMENTED as pass

- **WHEN** the `maestro-grpc` connectivity check runs and the gRPC server responds but does not implement the Health service
- **THEN** the check MUST succeed (server is reachable, Health check not mandatory)

### Requirement: In-Cluster Curl

The CLI SHALL execute curl commands from inside the Kubernetes cluster.

#### Scenario: Run curl from in-cluster pod

- WHEN the user runs `hf kube curl [--] [curl-flags...] <url>`
- THEN the CLI MUST create or reuse a persistent pod named `hf-curl` running `curlimages/curl:latest`
- AND execute the curl command inside the pod via SPDY exec
- AND display the curl output
- NOTE: curl flags starting with `-` must be preceded by `--` to avoid Cobra flag parsing

### Requirement: Debug Pod Creation

The CLI SHALL create debug pods from existing deployment templates.

#### Scenario: Create debug pod

- GIVEN a deployment exists in the cluster
- WHEN the user runs `hf kube debug <partial-deployment-name>`
- THEN the CLI MUST find a deployment whose name contains the partial name
- AND create a pod using the same spec with liveness/readiness probes removed and `restartPolicy: Never`
- AND wait up to 3 minutes for the pod to reach Running phase
- AND print `[INFO] Debug pod ready: <pod-name>` and the kubectl exec command
- AND exec into the pod with an interactive shell session that persists until the user exits

### Requirement: Pod Log Tailing

The CLI SHALL tail logs from pods matching a name pattern using client-go log streaming.

#### Scenario: Tail logs for matching pods

- WHEN the user runs `hf logs [pattern]`
- THEN the CLI MUST fan out goroutine log streaming across all pods matching pattern
- AND prefix each line with `[pod-name]`

### Requirement: Adapter Log Tailing

The CLI SHALL tail adapter logs filtered by the current cluster ID.

#### Scenario: Tail adapter logs

- WHEN the user runs `hf logs adapter [pattern] [--cluster-id <id>]`
- THEN the CLI MUST search for pods matching `adapter` (or `adapter-<pattern>`)
- AND filter log lines to those containing `cluster_id=<id>` (logfmt format)
- AND skip JSON/OpenTelemetry span lines (lines starting with `{`)
- AND display matching lines as `[pod] <time>  <LEVEL>  <msg>`
- AND resolve cluster-id from `--cluster-id` flag, else from active config state

### Requirement: Bounded Log Collection

The CLI SHALL support collecting pod logs for a bounded time window for use in analysis commands.

#### Scenario: Collect logs from matching pods

- GIVEN pods exist in the namespace whose names match the pattern
- WHEN log collection is requested for a specific time window
- THEN the CLI MUST return all log lines from all matching pods combined into a single result
- AND lines from different pods MUST be combined in the order they are received

#### Scenario: No matching pods

- GIVEN no pods in the namespace match the pattern
- WHEN log collection is requested
- THEN the CLI MUST return an empty result with no error

#### Scenario: Pod list error

- GIVEN the Kubernetes API returns an error listing pods
- WHEN log collection is requested
- THEN the CLI MUST propagate the error immediately

### Requirement: Log Insights Command

The CLI SHALL provide `hf logs insights [-s <duration>]` that fetches logs from running pods and displays a human-readable summary of recent system activity.

#### Scenario: Run log insights with default window

- WHEN the user runs `hf logs insights`
- THEN the CLI MUST fetch logs from the last 1 minute from pods matching `api`, `sentinel`, and `adapter`
- AND display an API section with request counts grouped by `METHOD /normalised/path` and OK/error split
- AND display a Sentinel section with cycle and published-message counts per topic
- AND display an Adapter section with execution counts and phase activity per adapter component

#### Scenario: Run log insights with custom window

- WHEN the user runs `hf logs insights -s 5m`
- THEN the CLI MUST fetch logs from the last 5 minutes
- AND all output sections reflect the extended window

#### Scenario: Invalid duration

- WHEN the user runs `hf logs insights -s notaduration`
- THEN the CLI MUST display `[ERROR] invalid --since value "notaduration": ...`
- AND exit with code 1

#### Scenario: No active environment

- GIVEN no environment is activated
- WHEN the user runs `hf logs insights`
- THEN the CLI MUST fail with `[ERROR]` and exit 1

#### Scenario: No activity in window

- GIVEN pods exist but emitted no relevant log lines in the time window
- WHEN the user runs `hf logs insights`
- THEN the CLI MUST display `(no activity)` for that section

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

