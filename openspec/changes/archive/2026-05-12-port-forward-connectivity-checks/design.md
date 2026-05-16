## Context

`hf kube port-forward status` currently calls `kube.IsPortListening` which does a raw TCP dial to `127.0.0.1:<port>`. This tells us the local side of the tunnel is bound, but not whether the service behind it is responding correctly. Users are left unsure whether the forward is actually working. In addition, the namespace used for each service's pod lookup is not printed anywhere, and the `kubernetes.namespace` config key is semantically wrong — it is the HyperFleet application namespace, not a Kubernetes config setting.

Four predefined services are forwarded: `hyperfleet-api` (HTTP/REST), `postgresql` (pgx), `maestro-http` (HTTP/REST), `maestro-grpc` (gRPC). Each needs a different probe.

## Goals / Non-Goals

**Goals:**
- Add per-service protocol-level connectivity probes used by both `status` and the post-start display.
- Rename `kubernetes.namespace` → `hyperfleet.namespace` everywhere (config template, envVarMap, all call sites).
- Include namespace in the `[INFO] Started` line even when no pod was resolved.
- After starting all forwards, wait 1 second then call the same connectivity-display path used by `status`.
- Render ✓ (green) / ✗ (red) instead of the current bullet for both status and post-start views.

**Non-Goals:**
- Retry logic or automatic restart on failed probe.
- Protocol checks for custom/generic port-forwards (only the four predefined names get probes; unknown names fall back to TCP dial).
- Making postgres probe require a full authenticated query (a `Ping` over the existing pgx pool config is sufficient).

## Decisions

### Decision 1: Health-check functions in `internal/kube`, dispatched by service name

**Choice:** Add four exported functions to `internal/kube`:
```
CheckAPIConnectivity(port int) error
CheckMaestroHTTPConnectivity(port int) error
CheckPostgresConnectivity(port int, host, dbname, user, password string) error
CheckMaestroGRPCConnectivity(port int) error
```
A dispatch helper `CheckPortForwardConnectivity(name string, localPort int, s configGetter) error` maps service names to their prober.

**Why not put this in `cmd/`?** The check logic (HTTP client, pgx Ping, gRPC dial) belongs in `internal/kube` — it's reusable and keeps `cmd/kube.go` as thin as possible.

**Why not a single `Check(name, port)` with a switch in `cmd/`?** The service-specific helpers are independently testable via `httptest.NewServer` without mocking the whole command.

### Decision 2: gRPC health check via standard `grpc.health.v1.Health/Check`

**Choice:** Use `google.golang.org/grpc` (already an indirect dependency via `k8s.io/client-go`) and the standard gRPC health proto. Dial with `grpc.Dial("localhost:<port>", grpc.WithTransportCredentials(insecure.NewCredentials()))` with a 2-second timeout; call `Health/Check` with service name `""` (server-level health).

**Alternative considered:** TCP dial only for gRPC. Rejected — a TCP dial passing does not verify the gRPC server is actually serving; this would give the same false positives the current code has.

**Risk:** `grpc-health-probe` proto needs to be vendored. `google.golang.org/grpc/health/grpc_health_v1` is included in the grpc-go module already pulled by client-go, so no new module import is needed — just an explicit import in the health-check code.

### Decision 3: Postgres probe uses `pgxpool.New` + `Ping`, not a full query

**Choice:** Open a temporary `pgxpool` with a 2-second connection timeout, call `Ping(ctx)`, then close it.

**Why not reuse `internal/db`?** The DB package holds a long-lived pool. The probe needs an on-demand one-shot connection that can time out quickly and be discarded. Reusing the pool would entangle the status command with the DB session state.

**Why not just TCP dial to postgres port?** Same reason as gRPC — TCP connectivity does not verify postgres is accepting and authenticating connections.

### Decision 4: `hyperfleet.namespace` rename — update config template + all call sites atomically

**Choice:** In the same change, update `config-template.yaml`, `envVarMap` in `internal/config/config.go`, and all `s.Get("kubernetes", "namespace")` call sites in `cmd/kube.go`. No migration shim — this is a developer-facing config key and environments can be re-created.

**Alternative:** Keep `kubernetes.namespace` as an alias and add `hyperfleet.namespace`. Rejected — adds dead config surface for what is a straightforward rename.

### Decision 5: Post-start status uses the same `printPortForwardStatus` path

**Choice:** After the start loop, call `time.Sleep(time.Second)` then `printPortForwardStatus(cmd.OutOrStdout(), s)`. Pass the config store to `printPortForwardStatus` so it can supply postgres credentials to the connectivity probe.

**Why 1 second?** Enough time for the daemon subprocess to open the port-forward tunnel; not long enough to be annoying. Documented as approximate — not a health guarantee.

## Risks / Trade-offs

- [Postgres probe requires credentials] → `CheckPostgresConnectivity` reads `database.*` from the config store. If credentials are wrong, the probe fails with ✗ — but this is actually the correct behavior (the user can't connect, so the forward isn't useful).
- [gRPC server may not implement Health service] → If Maestro gRPC doesn't implement `grpc.health.v1`, the probe returns an error and renders ✗ even though the port is open. Mitigation: treat "unimplemented" gRPC status as pass, since the server is at least responding.
- [1-second post-start wait may be insufficient on slow clusters] → The wait is cosmetic; if the tunnel isn't ready yet, the ✗ is transient and `hf kube port-forward status` will show ✓ once the tunnel is established.

## Migration Plan

1. Rename `kubernetes.namespace` → `hyperfleet.namespace` in `config-template.yaml`.
2. Update `envVarMap` in `internal/config/config.go`.
3. Update all `s.Get("kubernetes", "namespace")` calls in `cmd/kube.go` and any other call sites.
4. Users with existing environment YAML files must rename the key: under `kubernetes:`, change `namespace: <value>` → under `hyperfleet:`, add `namespace: <value>`. The error message if the key is missing should guide them to `hf config set hyperfleet.namespace <value>`.

No rollback strategy needed — the CLI is a local binary and users can downgrade by reinstalling a prior build.
