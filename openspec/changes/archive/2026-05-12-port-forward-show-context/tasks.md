## 1. Extend `internal/kube` — context-aware config building

- [x] 1.1 Add `ResolvedContext(kubeconfigPath, contextName string) (string, error)` to `internal/kube/kube.go`: use `clientcmd.NewNonInteractiveDeferredLoadingClientConfig` with `ConfigOverrides{CurrentContext: contextName}` to return the context name that will be used (reads kubeconfig's `current-context` when `contextName == ""`)
- [x] 1.2 Change `BuildConfig(kubeconfigPath string)` to `BuildConfig(kubeconfigPath, contextName string)`: replace `clientcmd.BuildConfigFromFlags("", resolved)` with `clientcmd.NewNonInteractiveDeferredLoadingClientConfig` + `ConfigOverrides{CurrentContext: contextName}`; empty `contextName` preserves existing behavior
- [x] 1.3 Change `NewClientset(kubeconfigPath string)` to `NewClientset(kubeconfigPath, contextName string)` and thread `contextName` through to `BuildConfig`
- [x] 1.4 Update all internal `kube.go` callers of `BuildConfig`/`NewClientset` (`StartPortForward`, `RunPortForwardDaemon`, `RunCurlPod`) to accept and pass `contextName` through their own signatures
- [x] 1.5 Add `contextName` as a 6th positional arg to the daemon subprocess invocation in `StartPortForward` and parse it in `pfDaemonCmd` → `RunPortForwardDaemon`

## 2. Update `cmd/kube.go`

- [x] 2.1 In `pfStartCmd`: read `kubernetes.context` via `s.Get("kubernetes", "context")`; call `kube.ResolvedContext(kubeconfig, ctx)` and print `[INFO] Kubernetes context: <name>` (warn and continue on error) before any service start lines
- [x] 2.2 In `pfStatusCmd`: same context header before the bullet table
- [x] 2.3 Pass the resolved context value to all `kube.NewClientset` / `kube.BuildConfig` / `kube.StartPortForward` calls in `cmd/kube.go` (including `pfDaemonCmd`)

## 3. Update `cmd/logs.go`

- [x] 3.1 Update the three `kube.NewClientset(resolvedKubeconfig(s))` calls in `cmd/logs.go` to pass `""` as the second argument (no behavior change; keeps logs commands working)

## 4. Tests

- [x] 4.1 Add `TestResolvedContext` to `internal/kube/kube_test.go`: (a) returns current-context name from a test kubeconfig when no override given; (b) returns override name when `contextName` is non-empty and that context exists; (c) returns an error when the named context does not exist
- [x] 4.2 Update existing `TestBuildConfig_*` tests for the new two-argument signature (pass `""` as context — behavior unchanged)

## 5. Verify

- [x] 5.1 `go build ./...` succeeds
- [x] 5.2 `go vet ./...` passes
- [x] 5.3 `go test ./... 2>&1 | tee openspec/changes/port-forward-show-context/verification_proof/tests.txt`
- [x] 5.4 Run `hf kube port-forward start` and `hf kube port-forward status` against the live cluster; save output to `openspec/changes/port-forward-show-context/verification_proof/live.txt`; confirm `[INFO] Kubernetes context: …` appears as the first line
- [x] 5.5 Commit all changed files (implementation + tasks.md + verification_proof/) and push to main
