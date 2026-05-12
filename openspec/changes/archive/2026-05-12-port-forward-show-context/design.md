## Context

`BuildConfig` currently calls `clientcmd.BuildConfigFromFlags("", kubeconfigPath)`. The empty first argument means "use the kubeconfig's current-context". `kubernetes.context` is a first-class config key (mapped to `HF_CONTEXT`) but is never read by any kube function. The daemon subprocess (`_daemon`) receives its kubeconfig path as a positional arg but has no way to receive a context override.

Callers of `BuildConfig` / `NewClientset` today:
- `internal/kube/kube.go` — `NewClientset`, `StartPortForward`, `RunPortForwardDaemon`, `RunCurlPod`
- `cmd/kube.go` — `pfDaemonCmd`
- `cmd/logs.go` — three `NewClientset` calls in log-tailing commands

## Goals / Non-Goals

**Goals:**
- Print `[INFO] Kubernetes context: <context>` at the top of `port-forward start` and `port-forward status` output.
- Honour `kubernetes.context` (and `HF_CONTEXT`): when set, use that context instead of the kubeconfig's current-context for all `hf kube` operations.
- Keep all callers that don't need a context override working unchanged by passing `""`.

**Non-Goals:**
- Wiring context into `hf logs` commands (out of scope; they pass `""` and keep today's behavior).
- Adding a `--context` CLI flag (the config key is the override mechanism).
- Changing any other command's output.

## Decisions

**D1 — Add `contextName string` to `BuildConfig` and `NewClientset`.**
Replace `clientcmd.BuildConfigFromFlags("", resolved)` with:
```go
loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: resolved}
overrides := &clientcmd.ConfigOverrides{}
if contextName != "" {
    overrides.CurrentContext = contextName
}
cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
```
When `contextName == ""` this is functionally identical to the current call. All existing callers that don't need a context override pass `""`.

Alternative considered: a separate `BuildConfigWithContext` function. Rejected — duplication without benefit; the zero-value empty string is a clean default.

**D2 — Add `ResolvedContext(kubeconfigPath, contextName string) (string, error)` to `internal/kube`.**
Uses the same `clientcmd.NewNonInteractiveDeferredLoadingClientConfig` path to determine which context name will actually be used (reads the kubeconfig file's `current-context` when no override is given). Called by `pfStartCmd` and `pfStatusCmd` to get the display string before any other output.

**D3 — Thread context through the daemon subprocess args.**
`StartPortForward` spawns `hf kube port-forward _daemon <kubeconfig> <namespace> <podName> <localPort> <remotePort>`. Add `<context>` as a 6th positional arg. `pfDaemonCmd` passes it to `RunPortForwardDaemon`, which passes it to `BuildConfig`. When `context == ""` behavior is unchanged.

**D4 — Context is read from config in `pfStartCmd` / `pfStatusCmd`, not from a CLI flag.**
`s.Get("kubernetes", "context")` is the resolved value (config file → `HF_CONTEXT` env, following existing precedence). No new CLI flag is introduced.

**D5 — Output format.**
A single line is prepended to the existing output for both commands:
```
[INFO] Kubernetes context: gke_project_region_cluster
```
Printed to stdout via `fmt.Fprintln(cmd.OutOrStdout(), ...)`. No ANSI coloring (consistent with other `[INFO]` lines).

## Risks / Trade-offs

- **Daemon arg count change**: old daemon processes started before this change (with 5 args) do not receive the context arg. Those processes are already running and won't be restarted until `stop` + `start` — acceptable since the context arg defaults to `""` (current-context) which matches the old behavior.
- **`cmd/logs.go` callers**: pass `""` to `NewClientset` — no behavior change, but the signature update is mechanical work across 3 call sites.
- **`ResolvedContext` reads the kubeconfig file**: if the file is missing or the named context doesn't exist, it returns an error. `pfStartCmd`/`pfStatusCmd` should print a warning and continue (the subsequent `StartPortForward`/`ListPortForwards` calls will fail with a clearer error anyway).

## Open Questions

None.
