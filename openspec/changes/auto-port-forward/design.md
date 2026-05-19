# Design: Auto Port-Forward

## Architecture

### Hook point: `PersistentPreRunE` / `PersistentPostRunE` on rootCmd

`PersistentPreRunE` checks `s.Get("hyperfleet", "auto-port-forward") == "true"`.
If true, calls `startAutoPortForwards(s)` before any command runs.
`PersistentPostRunE` calls `autoPortForwardStop()` after the command completes.

### URL override via env vars

`config.Store.Get()` already checks env vars first. After port-forwards succeed:
- `os.Setenv("HF_API_URL", "http://127.0.0.1:<port>")` → all API commands auto-route
- `os.Setenv("HF_MAESTRO_HTTP", "http://127.0.0.1:<port>")` → all maestro HTTP calls route
- `os.Setenv("HF_MAESTRO_GRPC", "127.0.0.1:<port>")` → all gRPC calls route

No changes to any of the ~70 existing command files.

### `kube.EphemeralPortForward` (new function)

1. `NewClientset` + `BuildConfig` from kubeconfig
2. `FindRunningPod` → find target pod
3. `findFreePort()` → `net.Listen("tcp", "127.0.0.1:0")`, grab port, close
4. SPDY dialer setup (same pattern as `RunPortForwardDaemon`)
5. `portforward.New(...)` + goroutine
6. Wait for `readyCh` / `errCh` / 30s timeout
7. Return `localPort` + `sync.Once`-guarded stop func

### Remote port mapping

| Service | Namespace | Pod pattern | Remote port |
|---------|-----------|-------------|-------------|
| HyperFleet API | `hyperfleet.namespace` | `hyperfleet-api` | 8000 |
| Maestro HTTP | `maestro.namespace` | `maestro` | `port-forward.maestro-http-remote-port` (default 8000) |
| Maestro gRPC | `maestro.namespace` | `maestro` | `port-forward.maestro-grpc-remote-port` (default 8090) |

## Files Changed

| File | Change |
|------|--------|
| `internal/config/assets/config-template.yaml` | Add `hyperfleet.auto-port-forward: "false"` |
| `internal/config/config.go` | Add `HF_MAESTRO_HTTP` and `HF_MAESTRO_GRPC` to envVarMap |
| `internal/kube/kube.go` | Add `findFreePort()` and `EphemeralPortForward()` |
| `cmd/root.go` | Add `autoPortForwardStop` var, `startAutoPortForwards`, `PersistentPostRunE` |
| `internal/kube/kube_test.go` | Add `TestEphemeralPortForward_PodNotFound`, `TestFindFreePort` |
| `cmd/root_test.go` | Add `TestAutoPortForward_DisabledByDefault` |
