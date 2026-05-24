## Context

`internal/kube.StartPortForward` calls `FindRunningPod`, then spawns a detached `hf kube port-forward _daemon` subprocess that opens an SPDY tunnel to `pods/<podName>/portforward`. Predefined services in `cmd/kube.go` map logical names to pod patterns (e.g. `maestro-http` → pod pattern `maestro`). The same pod-resolution path is used by `EphemeralPortForward` for auto port-forward.

Kubernetes supports port-forwarding to Services via `services/<name>/portforward`. The apiserver selects a backing pod from the Service endpoints — the same behavior as `kubectl port-forward svc/...`. client-go's `portforward.New` works identically; only the REST subresource URL changes.

## Goals / Non-Goals

**Goals:**

- Prefer Service port-forward when a Service with the expected name exists in the target namespace and exposes the configured remote port.
- Preserve pod-pattern fallback for clusters without Services or with differently named Services.
- Surface target kind in start output and `StartResult` for debugging.
- Apply service-first resolution consistently to persistent forwards, ephemeral auto-forwards, and generic CLI invocations.

**Non-Goals:**

- Endpoints-based pod selection when no Service exists (keep existing pod pattern logic).
- Headless Service edge cases beyond standard ClusterIP/NodePort port matching.
- Changing connectivity probe logic, PID file format, or stop/status behavior.
- Adding new config keys for service names (use predefined mapping + CLI arg as service name).

## Decisions

### D1 — Service-first resolution helper

Add `ResolvePortForwardTarget(ctx, cs, namespace, serviceName, podPattern string, remotePort int) (targetKind, targetName string, warn error)` in `internal/kube`:

1. If `serviceName` is non-empty, `GET` the Service in `namespace`. If found and any `spec.ports[].port == remotePort` (or `targetPort` matches numerically), return `("service", serviceName, nil)`.
2. Else call `FindRunningPod`. On success return `("pod", podName, nil)`. On `PodNotReadyError` return `("pod", podName, err)` so caller can warn and continue.
3. If neither resolves, return error.

**Why:** Single entry point shared by `StartPortForward`, `EphemeralPortForward`, and tests.

### D2 — Predefined service name mapping

Extend `serviceSpec` with `serviceName`:

| name | serviceName | podPattern (fallback) |
|------|-------------|----------------------|
| hyperfleet-api | hyperfleet-api | hyperfleet-api |
| postgresql | postgresql | postgresql |
| maestro-http | maestro | maestro |
| maestro-grpc | maestro | maestro |

Maestro HTTP and gRPC share the `maestro` Service but use different remote ports (8000 vs 8090); port matching in D1 selects the correct Service port entry.

### D3 — Service resolution via Endpoints (kubectl-compatible)

When a Service matches, resolve a backing pod from the Service's **Endpoints** object (same port number), then port-forward via `pods/<podName>/portforward`. This matches `kubectl port-forward svc/...` behavior — the apiserver `services/portforward` subresource is not used (not supported on all clusters).

The daemon subprocess still receives only the resolved pod name (6 positional args, unchanged from before).

**Alternative considered:** Direct `services/<name>/portforward` REST subresource. Rejected: fails on GKE with "the server could not find the requested resource"; kubectl resolves endpoints client-side instead.

### D4 — Start output format

- Service: `[INFO] Started <name> (<namespace>/svc/<serviceName>): localhost:<localPort> → <remotePort> (pid <pid>)`
- Pod: `[INFO] Started <name> (<namespace>/pod/<podName>): localhost:<localPort> → <remotePort> (pid <pid>)`

Replace the current ambiguous `(<namespace>/<podName>)` vs `(<namespace>)` split with explicit kind prefix.

### D5 — StartResult fields

Replace `PodName string` with:

```go
TargetKind string // "service" or "pod"
TargetName string
```

Keep backward compatibility in tests by updating assertions; no external API consumers.

### D6 — Generic port-forward resolution

For `hf kube port-forward start <name> <local:remote>`, set `serviceName = name` and `podPattern = name`. Resolution follows D1.

### D7 — Service port matching

Match `remotePort` against Service `spec.ports[].port` (the Service port number, not container targetPort name). This aligns with kubectl's `local:servicePort` semantics and existing config where maestro-http remote port is 8000.

If Service exists but no port matches, fall through to pod pattern (do not fail early — Service may exist for a different purpose).

## Risks / Trade-offs

- **[Risk] Service forwards to a different pod than the one the user expects]** → Same as kubectl; document that service forwards use apiserver endpoint selection. Mitigation: start output shows `svc/` so users know.
- **[Risk] Daemon arg change breaks in-flight PIDs from old hf version]** → Old daemons keep running until stopped; new starts use new format. No migration needed.
- **[Risk] Maestro service has both ports; wrong port selected]** → Explicit remote port check in D1 prevents cross-wiring.

## Migration Plan

No config migration. Users restart port-forwards after upgrade (`hf kube port-forward stop && hf kube port-forward start`). Behavior change is transparent when Services exist.

## Open Questions

_(none)_
