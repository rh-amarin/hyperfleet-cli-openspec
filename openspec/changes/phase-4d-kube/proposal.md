## Why

The `hf` CLI currently has stub `kube` and `logs` commands with no implementation. Engineers managing HyperFleet clusters need to port-forward to in-cluster services, tail pod logs, exec curl from inside the cluster, and create debug pods — all without requiring `kubectl` to be installed.

## What Changes

- **New**: `internal/kube` package using `k8s.io/client-go` for all Kubernetes operations
- **New**: `hf kube port-forward start|stop|status` — manage background port-forward processes to predefined services
- **New**: `hf kube port-forward start <name>` / `hf kube port-forward start <service> <localPort:remotePort>` — named or custom port forwards
- **New**: `hf kube curl [--] <curl-args>` — run curl from a persistent in-cluster pod
- **New**: `hf kube debug <partial-name>` — create and exec into a debug pod from deployment spec
- **New**: `hf logs [pattern]` — tail pod logs (delegates to stern if available, otherwise goroutine fan-out)
- **New**: `hf logs adapter [pattern] [--cluster-id]` — filtered adapter logs by cluster ID

## Capabilities

### New Capabilities

- `kubernetes`: Port-forward management, in-cluster curl, debug pod creation, and pod log tailing via `client-go`

### Modified Capabilities

_(none — this adds new commands; no existing requirement changes)_

## Impact

- **Dependencies**: Adds `k8s.io/client-go`, `k8s.io/api`, `k8s.io/apimachinery`
- **New package**: `internal/kube/` with all client-go logic
- **Modified files**: `cmd/kube.go`, `cmd/logs.go` (stubs become full implementations)
- **Config**: reads `kubernetes.context`, `kubernetes.namespace`, `port-forward.*`, `maestro.namespace` from config store
- **PID files**: `~/.config/hf/pf-<name>.pid`

## Testing Scope

| Package | Test Cases |
|---|---|
| `internal/kube` | `BuildConfig` (valid/missing kubeconfig), `IsProcessAlive`, `FindRunningPod` with fake clientset, `ListPortForwards` with PID file stubs |

Live cluster access required for: port-forward smoke tests, curl pod exec, debug pod creation, log streaming.
