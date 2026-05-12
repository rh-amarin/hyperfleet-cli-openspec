# Kubernetes Utilities Specification

## Purpose

Provide CLI commands for Kubernetes-related operations including port-forwarding to HyperFleet services, in-cluster curl execution, debug pod creation, and log tailing for pods and adapters.

## Implementation

Implemented in Go using `k8s.io/client-go` — no `kubectl` binary dependency.
GKE auth plugin bypass: set `HF_KUBE_TOKEN` env var to override bearer token.

Package `internal/kube` provides:
- `BuildConfig(kubeconfigPath string) (*rest.Config, error)` — resolves: flag → `KUBECONFIG` env → `~/.kube/config`
- `NewClientset(kubeconfigPath string) (kubernetes.Interface, error)`
- `StartPortForward(kubeconfigPath, namespace, name, podPattern string, localPort, remotePort int) (*PortForward, error)`
- `StopPortForward(service string) error`
- `ListPortForwards() ([]PortForward, error)`
- `IsProcessAlive(pid int) bool`
- `FindRunningPod(ctx, cs, namespace, pattern string) (string, error)`
- `RunPortForwardDaemon(kubeconfigPath, namespace, service string, localPort, remotePort int) error`
- `StreamLogs(ctx, cs, namespace, podPattern string, w io.Writer) error`
- `StreamLogsFiltered(ctx, cs, namespace, podPattern, clusterID string, w io.Writer) error`
- `RunCurlPod(ctx, kubeconfigPath, namespace string, curlArgs []string, w io.Writer) error`
- `CreateDebugPod(ctx, cs, namespace, pattern string) (string, error)`
## Requirements
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

The CLI SHALL tail logs from pods matching a name pattern.

#### Scenario: Tail logs for matching pods

- WHEN the user runs `hf logs [pattern]`
- THEN if `stern` is available in PATH, the CLI MUST delegate to `stern <pattern> -n <namespace>`
- AND if stern is not available, fan out goroutine log streaming across all pods matching pattern
- AND prefix each line with `[pod-name]`

### Requirement: Adapter Log Tailing

The CLI SHALL tail adapter logs filtered by the current cluster ID.

#### Scenario: Tail adapter logs

- WHEN the user runs `hf logs adapter [pattern] [--cluster-id <id>]`
- THEN the CLI MUST search for pods matching `adapter` (or `adapter-<pattern>`)
- AND filter log lines to those containing `cluster_id=<id>` (logfmt format)
- AND skip JSON/OpenTelemetry span lines (lines starting with `{`)
- AND display matching lines as `[pod] <time>  <LEVEL>  <msg>`
- AND resolve cluster-id from `--cluster-id` flag, else from active config state (`cfgStore.State().ClusterID`)

### Requirement: CollectLogs

Package `internal/kube` SHALL provide a `CollectLogs` function that fetches pod logs
for a bounded time window and returns all lines as a flat slice.

```go
func CollectLogs(ctx context.Context, cs kubernetes.Interface, namespace, podPattern string, sinceSeconds int64) ([]string, error)
```

#### Scenario: Collect logs from matching pods

- GIVEN pods exist in the namespace whose names contain `podPattern`
- WHEN `CollectLogs` is called with a positive `sinceSeconds` value
- THEN it MUST return all log lines from all matching pods as `[]string`
- AND lines from different pods MUST be combined into a single slice

#### Scenario: No matching pods

- GIVEN no pods in the namespace match `podPattern`
- WHEN `CollectLogs` is called
- THEN it MUST return an empty slice with no error

#### Scenario: Pod list error

- GIVEN the Kubernetes API returns an error listing pods
- WHEN `CollectLogs` is called
- THEN it MUST return the error immediately

### Requirement: ParseLogfmt exported

Package `internal/kube` SHALL export `ParseLogfmt(line string) map[string]string`
so that other packages can reuse logfmt parsing without duplication.

#### Scenario: Parse logfmt line

- WHEN `ParseLogfmt` is called with a valid logfmt line
- THEN it MUST return a map of all key-value pairs including quoted values

### Requirement: insights package

Package `internal/insights` SHALL provide three pure log-parsing functions that
operate on `[]string` log lines and return structured summary types.

#### Scenario: ParseAPILogs extracts completed requests

- WHEN `ParseAPILogs` is called with API pod log lines
- THEN it MUST parse only lines where `message == "HTTP request completed"`
- AND group counts by `method + path` with UUIDs normalised to `:id`
- AND track `OK` count (status_code < 400) and `Err` count (status_code >= 400) per group

#### Scenario: ParseSentinelLogs extracts cycle summaries

- WHEN `ParseSentinelLogs` is called with sentinel pod log lines
- THEN it MUST parse only lines where `message` starts with `"Trigger cycle completed"`
- AND accumulate per-topic cycle count, published count, and skipped count

#### Scenario: ParseAdapterLogs extracts adapter activity

- WHEN `ParseAdapterLogs` is called with adapter pod log lines
- THEN it MUST count `"Processing event"` messages per `component` as executions
- AND count `"Phase <name>: RUNNING"` messages per `component` and phase name

### Requirement: Log Insights Command

The CLI SHALL provide `hf logs insights [-s <duration>]` that fetches logs from
running pods and displays a human-readable summary of recent system activity.

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

