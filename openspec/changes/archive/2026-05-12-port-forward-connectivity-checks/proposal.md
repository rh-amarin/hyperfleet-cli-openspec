## Why

The current port-forward status check only verifies that a TCP connection can be opened on the local port — it cannot tell whether the tunnel is actually forwarding to a healthy service. Users need protocol-aware health checks so that a broken postgres connection or a non-responsive gRPC endpoint shows up as a failure rather than a green dot. Namespace visibility is also missing from the start output, making it hard to confirm which cluster namespace each forward targets. Additionally, `kubernetes.namespace` is a misnomer: the namespace is the HyperFleet application namespace, not a kubernetes config property.

## What Changes

- **Namespace in start output** — Each `[INFO] Started` line now includes the Kubernetes namespace (already present in `StartResult.Namespace`), even when no pod was resolved.
- **Config key rename** — `kubernetes.namespace` → `hyperfleet.namespace` in the config template and in every `s.Get("kubernetes", "namespace")` call site. The `HF_NAMESPACE` env-var mapping updates accordingly. **BREAKING** for anyone who has `kubernetes.namespace` in their environment YAML (migration: rename the key).
- **Post-start delay + status display** — After all port-forwards are started, wait 1 second, then display a connectivity status table that uses the protocol checks described below (not just TCP ping).
- **Protocol-aware status checks** — `hf kube port-forward status` performs a real protocol check per service instead of a TCP dial:
  - `hyperfleet-api` — HTTP GET `http://localhost:<port>/api/hyperfleet/v1` — pass if HTTP response received (any non-connection-error status).
  - `postgresql` — Open a `pgx` connection to `localhost:<port>` with configured credentials and call `Ping` — pass if `Ping` returns nil.
  - `maestro-http` — HTTP GET `http://localhost:<port>/api/maestro/v1` — pass if HTTP response received.
  - `maestro-grpc` — gRPC dial to `localhost:<port>` + `grpc.health.v1.Health/Check` — pass if response received with status SERVING.
- **Green ✓ / red ✗ indicators** — Status output uses ✓ (green) for pass and ✗ (red) for fail instead of the current colored bullet.

## Capabilities

### New Capabilities

- `port-forward-connectivity-checks`: Protocol-aware health check functions in `internal/kube` (one per service type: REST, Postgres, Maestro HTTP, Maestro gRPC), and updated status display logic in `cmd/kube.go`.

### Modified Capabilities

- `kubernetes`: Port-forward start output format (namespace visible), status check behavior (protocol-aware), post-start wait+display.
- `config-model`: `kubernetes.namespace` renamed to `hyperfleet.namespace`; `HF_NAMESPACE` env-var remapped.
- `config-template`: Remove `namespace` from `kubernetes:` section; add `namespace` to `hyperfleet:` section.

## Impact

- `cmd/kube.go` — `pfStartCmd`: add namespace to `[INFO] Started` line, add 1-second wait + connectivity status after all forwards start. `pfStatusCmd` / `printPortForwardStatus`: call protocol-specific checkers, render ✓/✗. `servicesForArgs`, `kubeCurlCmd`, `kubeDebugCmd`: change `s.Get("kubernetes", "namespace")` → `s.Get("hyperfleet", "namespace")`.
- `internal/config/config.go` — `envVarMap`: update `HF_NAMESPACE` target from `{"kubernetes", "namespace"}` to `{"hyperfleet", "namespace"}`.
- `internal/config/assets/config-template.yaml` — Move `namespace` key from `kubernetes:` section to `hyperfleet:` section.
- `internal/kube/kube.go` — Add exported health-check functions: `CheckAPIConnectivity`, `CheckPostgresConnectivity`, `CheckMaestroHTTPConnectivity`, `CheckMaestroGRPCConnectivity`.
- New Go dependency — `google.golang.org/grpc` (already vendored via client-go; need to add health proto); `jackc/pgx/v5` already in `go.mod` for DB commands.

## Testing Scope

| Package | Test cases needed |
|---|---|
| `internal/kube` | `TestCheckAPIConnectivity_OK`, `TestCheckAPIConnectivity_Down`, `TestCheckMaestroHTTPConnectivity_OK`, `TestCheckMaestroHTTPConnectivity_Down` (use `httptest.NewServer`); `TestCheckPostgresConnectivity_Down` (dial to closed port returns error); `TestCheckMaestroGRPCConnectivity_Down` (dial to closed port returns error) |
| `cmd` | `TestPFStatus_WithConnectivity` — mock port-forward PID files + httptest server for API, verify ✓/✗ rendered correctly |
| `internal/config` | `TestHFNamespace_EnvVar` — set `HF_NAMESPACE`, verify `s.Get("hyperfleet","namespace")` returns it; `TestHFNamespace_Profile` — set in env file, verify get |

Live cluster access is required for the final verification step: run `hf kube port-forward start` and `hf kube port-forward status` against the real GKE cluster and capture the output showing green ✓ for all four services.
