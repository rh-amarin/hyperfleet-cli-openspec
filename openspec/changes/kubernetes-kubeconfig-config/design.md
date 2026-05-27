## Context

`kubernetes.context` is resolved via `Store.Get("kubernetes", "context")` and passed to all kube operations. Kubeconfig path resolution lives in `cmd/kube.go` as `resolvedKubeconfig()`, which currently returns only the `--kubeconfig` CLI flag (empty string otherwise), delegating to `internal/kube.resolveKubeconfig()` for `KUBECONFIG` / `~/.kube/config`.

Auto-port-forward, port-forward start/stop, logs, and debug all call `resolvedKubeconfig(s)` but the store parameter is ignored.

## Decisions

### D1 — Config key: `kubernetes.kubeconfig`

Mirrors `kubernetes.context`. Empty string means "not configured" — fall through to `KUBECONFIG` then `~/.kube/config`. Stored as a plain path string (supports `~` expansion is out of scope; users provide absolute or relative paths as today with `--kubeconfig`).

### D2 — Precedence

```
--kubeconfig flag  >  HF_KUBECONFIG / env file  >  KUBECONFIG env  >  ~/.kube/config
```

Config profile overrides ambient `KUBECONFIG` so switching `hf config env activate` picks up the right kubeconfig without shell exports. CLI flag remains highest for one-off overrides.

`HF_KUBECONFIG` is registered in `envVarMap` so `Store.Get("kubernetes", "kubeconfig")` handles it automatically.

### D3 — Centralize in `cmd/kube.go`

Keep resolution in `resolvedKubeconfig(s)` rather than changing `internal/kube.resolveKubeconfig()`. The internal helper continues to handle the path-argument → KUBECONFIG → default chain when given an empty path; `resolvedKubeconfig` fills in the config layer before calling through.

Implementation:

```go
func resolvedKubeconfig(s interface{ Get(string, string) string }) string {
    if kubeConfigFlag != "" {
        return kubeConfigFlag
    }
    if v := s.Get("kubernetes", "kubeconfig"); v != "" {
        return v
    }
    return kube.ResolveKubeconfig("")
}
```

Export `ResolveKubeconfig` (rename from unexported `resolveKubeconfig`) or duplicate the KUBECONFIG/default tail inline. Prefer exporting to avoid duplication.

### D4 — Error message

Update `[ERROR] kubeconfig not found at <path>. Set KUBECONFIG or use --kubeconfig.` to also mention `kubernetes.kubeconfig` / `hf config set kubernetes.kubeconfig`.

## Risks

- **Relative paths**: resolved relative to cwd, same as `--kubeconfig` today. No change.
- **Existing users**: empty default preserves current behavior.
