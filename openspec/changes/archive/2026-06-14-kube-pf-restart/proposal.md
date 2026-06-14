## Why

`hf kube port-forward start` calls `stopPortForwardsBeforeStart` before launching daemons, but that function only terminates processes that are tracked in PID files. When PID files are missing — after a crash, a manual `rm`, or a forward started outside `hf` — the stop silently does nothing. The subsequent start then fails to bind the port, leaving the user with no running forward and no clear error.

## What Changes

- `stopPortForwardsBeforeStart` in `cmd/kube.go` adds a second cleanup pass: after stopping tracked PID-file entries, it kills any process still listening on each service's local port via `kube.PIDForPort` — regardless of whether a PID file existed
- Log line `[INFO] Killed stray process <pid> on port <port>` is emitted when a stray process is found and terminated
- No change to command signatures, flags, or user-visible subcommand structure

## Capabilities

### New Capabilities
_(none — this is a behavioural fix within an existing capability)_

### Modified Capabilities
- `kubernetes`: The "Start port forwards" scenario must be updated to require that the CLI also kills any process occupying a service's local port even when no PID file exists for that service

## Impact

- `cmd/kube.go` — `stopPortForwardsBeforeStart`: add port-based cleanup loop after PID-file-based cleanup
- `cmd/kube_test.go` — add test covering the stray-process scenario (no PID file, port occupied)

## Testing Scope

- `cmd/kube_test.go`: stray process on local port killed even when no PID file is tracked; existing start/stop tests continue to pass

Live cluster verification required to confirm port-forward starts cleanly when a stray process was occupying the port.
