## Context

`pfStartCmd` calls `kube.StartPortForward()` once per service. That function internally runs `FindRunningPod()` to locate the target pod, then spawns the daemon subprocess. The namespace and pod name are consumed internally but never returned to the caller, so `pfStartCmd` cannot include them in its output.

`pfStatusCmd` renders the post-start status table inline; the rendering logic is duplicated between the two commands if we want `pfStartCmd` to show the same view.

## Goals / Non-Goals

**Goals:**
- Surface the resolved namespace and pod name in the `[INFO] Started …` line.
- Show the same process-alive status table that `port-forward status` produces, automatically at the end of `port-forward start`.
- Share rendering logic between `pfStartCmd` and `pfStatusCmd` so there is one path to maintain.

**Non-Goals:**
- Waiting for the tunnel to be fully established before printing status (that would require polling the port).
- Storing namespace/pod in the PID file (overkill; status only needs PID and port).
- Changing any CLI flags or config keys.

## Decisions

**D1 — Introduce `StartResult` struct.**
`StartPortForward` currently returns `(pid, localPort, remotePort int, error)`. Change the return to `(StartResult, error)` where `StartResult` carries `Name`, `PID`, `LocalPort`, `RemotePort`, `Namespace`, and `PodName`. This is the minimal, clean way to pass the pod info back to `pfStartCmd` without a second Kubernetes API call.

Alternative considered: pass a pre-resolved pod name into `StartPortForward`. Rejected — it would split the "find pod" concern across two callers.

**D2 — Extract `printPortForwardStatus(w io.Writer, noColor bool)` in `cmd/kube.go`.**
Both `pfStartCmd` and `pfStatusCmd` call this helper. It reads PID files via `ListPortForwards()` and prints the bullet table exactly as today. Deduplication here prevents the two commands from diverging silently.

**D3 — Per-service line format.**
Updated from:
```
[INFO] Started hyperfleet-api: localhost:8000 → 8000 (pid 12345)
```
to:
```
[INFO] Started hyperfleet-api (amarin-ns1/hyperfleet-api-7d9f8b-xkj2p): localhost:8000 → 8000 (pid 12345)
```
The `(namespace/podName)` token is parenthesised after the service name so the port mapping remains readable.

**D4 — No delay before inline status.**
`StartPortForward` writes the PID file before returning, so by the time all services are started the PID files are present and the processes are alive. Showing the status table immediately is correct; there is no race to defend against at this level (the status check only tests process liveness, not tunnel health).

## Risks / Trade-offs

- **`StartResult` is an internal type** — no public API surface changes; impact is limited to `internal/kube` and `cmd/kube.go`.
- **Pod not found** — when `FindRunningPod` returns an error, `StartPortForward` currently logs a warning and still attempts the daemon. In that case `StartResult.PodName` will be `""` and the per-service line will omit the parenthesised token rather than crashing.
- **Status table timing** — processes are spawned non-blocking; there is a small window where a process might die immediately after start. The status table reflects reality at print time, which is the same behavior as running `hf kube port-forward status` right after start.

## Open Questions

None — scope is fully bounded to two files.
