## Why

Port-forward today always resolves a pod by name pattern and forwards directly to that pod. When the backing pod is rescheduled or replaced, the tunnel breaks until the user restarts the forward. Kubernetes Services provide stable endpoints and match how `kubectl port-forward svc/...` works — the CLI should prefer Services when they exist and only fall back to pod lookup when no suitable Service is available.

## What Changes

- Port-forward target resolution prefers a Kubernetes **Service** resource (by name, with the configured remote port) before falling back to pod pattern matching.
- Predefined services (`hyperfleet-api`, `postgresql`, `maestro-http`, `maestro-grpc`) gain an explicit service name; pod pattern remains as fallback only.
- Generic port-forwards (`hf kube port-forward start <name> <local:remote>`) try `<name>` as a Service first, then as a pod pattern.
- Start output distinguishes service vs pod targets: `svc/<name>` vs `pod/<name>`.
- `StartResult` and the hidden `_daemon` subprocess carry target kind (service or pod) so the daemon uses the correct client-go subresource URL.
- Auto port-forward (`EphemeralPortForward`) follows the same service-first resolution.
- Pod-not-ready warnings apply only when the pod fallback path is used.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `kubernetes`: Port Forward Management — service-first target resolution, updated output format, pod fallback behavior, and generic port-forward resolution order.

## Impact

- **`internal/kube/kube.go`** — `FindService`, `ResolvePortForwardTarget`, service-based daemon URL, update `StartPortForward`, `RunPortForwardDaemon`, `EphemeralPortForward`.
- **`cmd/kube.go`** — `serviceSpec` adds `serviceName`; start output uses target kind; `_daemon` args include target kind.
- **`cmd/tui_portforward.go`** — inherits updated `StartPortForward` (no logic change expected).
- **`cmd/root.go`** — auto port-forward uses updated `EphemeralPortForward`.
- **`internal/kube/kube_test.go`** — unit tests for service-first resolution and fallback.
- **`openspec/specs/kubernetes/spec.md`** — updated via delta at archive.

## Testing Scope

| Package / area | Test cases |
|---|---|
| `internal/kube` | `ResolvePortForwardTarget` — service exists → service target; service missing → pod fallback; pod not ready → warn + continue; neither found → error |
| `internal/kube` | `RunPortForwardDaemon` — service URL vs pod URL (httptest or fake clientset where feasible) |
| `internal/kube` | `EphemeralPortForward` — service-first path |
| `cmd` | Start output shows `svc/` prefix when service used; `pod/` when pod fallback used |

## Live Verification

- Run `hf kube port-forward start` against real cluster; confirm start lines show `svc/<name>` for predefined services that have Services.
- Run `hf kube port-forward status` and confirm connectivity checks still pass.
- Stop/start cycle works; no regression for maestro dual-port forwards on the same service name.
