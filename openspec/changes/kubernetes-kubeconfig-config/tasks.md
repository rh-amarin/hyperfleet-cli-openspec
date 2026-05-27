## 1. Config model

- [x] 1.1 Add `kubeconfig: ""` to `kubernetes` section in config template
- [x] 1.2 Add `HF_KUBECONFIG` → `kubernetes.kubeconfig` in `envVarMap`
- [x] 1.3 Add `kubeconfig` to `knownKeysForSection("kubernetes")`

## 2. Resolution

- [x] 2.1 Export `ResolveKubeconfig` from `internal/kube` (rename from unexported helper)
- [x] 2.2 Update `resolvedKubeconfig(s)` to use flag → config → `ResolveKubeconfig("")`
- [x] 2.3 Update kubeconfig-not-found error message

## 3. Tests and verification

- [x] 3.1 Add config tests for `kubernetes.kubeconfig` and `HF_KUBECONFIG`
- [x] 3.2 Add `resolvedKubeconfig` precedence test in `cmd/kube_test.go`
- [x] 3.3 Run `go test ./...`, `go vet ./...`, save output to `verification_proof/`
