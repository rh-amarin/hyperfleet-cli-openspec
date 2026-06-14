## Packages

| Package | Change |
|---|---|
| `cmd` | `stopPortForwardsBeforeStart` — add port-based cleanup pass |

## Key Decision: Two-Pass Stop

The existing first pass (PID-file stop) is kept unchanged. A second pass is appended:

```
for each service in services:
    if kube.PIDForPort(svc.localPort) succeeds:
        send SIGTERM to that PID
        log "[INFO] Killed stray process <pid> on port <port>"
```

This runs AFTER the first pass so the PID-file-managed processes are already terminated and their ports likely freed. The second pass catches anything still bound to the port.

The second pass iterates `services` (the slice already built by `servicesForArgs`), which is exactly the set of ports the CLI is about to occupy. No additional configuration is needed.

## Why Not Change `kube.StopPortForward`

`StopPortForward` is intentionally name-scoped — it owns a named tracking entry. Adding port-based fallback there would change its contract and could silently kill unrelated processes if the port happens to be in use by something not `hf`-managed. The cmd layer is the right place because it knows EXACTLY which ports it intends to reclaim.

## File Map

```
cmd/kube.go          — stopPortForwardsBeforeStart: add second pass
cmd/kube_test.go     — add stray-process test
openspec/changes/kube-pf-restart/specs/kubernetes/spec.md  — delta spec
```
