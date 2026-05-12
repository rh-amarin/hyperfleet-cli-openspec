## Why

`hf kube port-forward start` and `status` give no indication of which Kubernetes cluster they are targeting. The `kubernetes.context` config key exists and is mapped to `HF_CONTEXT`, but `BuildConfig` ignores it — it always uses the kubeconfig's `current-context` implicitly. A user working across multiple clusters (dev, staging, prod) has no way to confirm which context is active without running `kubectl config current-context` separately, and cannot override it via `hf` config.

## What Changes

- `BuildConfig` is updated to accept an optional context name; when non-empty it selects that context from the kubeconfig instead of the current-context.
- `resolvedKubeconfig` (or its caller) additionally resolves and returns the `kubernetes.context` config value so it can be passed through.
- A new helper `ResolvedContext(kubeconfigPath, contextOverride string) (string, error)` in `internal/kube` returns the context name that will actually be used (either the override or the file's current-context), without establishing a connection.
- `port-forward start` prints one header line before starting services: `[INFO] Kubernetes context: <context>`.
- `port-forward status` prints the same header line before the bullet table.
- The `kubernetes.context` config key is now honoured: if set, it overrides the kubeconfig's current-context for all `hf kube` operations.

## Capabilities

### New Capabilities

- `port-forward-show-context`: Display the resolved Kubernetes context at the top of `port-forward start` and `port-forward status` output.

### Modified Capabilities

- `kubernetes`: `BuildConfig` gains a context parameter; `kubernetes.context` config key is now wired up and effective. The `port-forward start` and `port-forward status` output format changes.

## Impact

- **`internal/kube/kube.go`**: `BuildConfig` signature change (adds `contextName string`); new `ResolvedContext` helper.
- **`cmd/kube.go`**: reads `kubernetes.context` from config, passes it to `BuildConfig`; prints context header in `pfStartCmd` and `pfStatusCmd`.
- **`internal/kube/kube_test.go`**: unit tests for `ResolvedContext`.
- No changes to config keys, CLI flags, or external APIs. `BuildConfig` callers outside `cmd/kube.go` (if any) must be updated for the new signature.

## Testing Scope

- `internal/kube`: unit test for `ResolvedContext` — returns current-context when no override given; returns override when provided; errors when named context does not exist in the kubeconfig.
- Live verification required: run `hf kube port-forward start` and `hf kube port-forward status` against the real cluster and confirm the context header appears.
