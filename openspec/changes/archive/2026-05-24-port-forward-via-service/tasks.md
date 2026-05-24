## 1. Target resolution (`internal/kube`)

- [x] 1.1 Add `ResolvePortForwardTarget` with service lookup (match `spec.ports[].port` to remote port) and pod-pattern fallback
- [x] 1.2 Update `StartResult`: replace `PodName` with `TargetKind` and `TargetName`
- [x] 1.3 Update `StartPortForward` to call `ResolvePortForwardTarget` and pass target kind/name to daemon
- [x] 1.4 Update `RunPortForwardDaemon` to accept target kind + name and build service or pod portforward URL
- [x] 1.5 Update `EphemeralPortForward` to use `ResolvePortForwardTarget` (service-first)

## 2. Command layer (`cmd/`)

- [x] 2.1 Add `serviceName` to `serviceSpec` and predefined service table in `servicesForArgs`
- [x] 2.2 Update `_daemon` subcommand args: insert `targetKind` and `targetName` positional parameters
- [x] 2.3 Update `pfStartCmd` start output to use `svc/` and `pod/` prefixes per spec
- [x] 2.4 Update generic port-forward path to set both `serviceName` and `podPattern` to CLI arg

## 3. Tests

- [x] 3.1 Unit tests: `ResolvePortForwardTarget` — service found, service missing → pod, pod not ready warn, neither found
- [x] 3.2 Unit tests: update existing `StartPortForward` / `EphemeralPortForward` tests for new fields and fallback
- [x] 3.3 Run `go test ./internal/kube/... ./cmd/...` and save output to `verification_proof/kube_test.txt`

## 4. Verification

- [x] 4.1 Run `go build ./...` and `go vet ./...`; save outputs to `verification_proof/`
- [x] 4.2 Live: `hf kube port-forward start` — confirm `svc/` in start lines; save to `verification_proof/live.txt`
- [x] 4.3 Live: `hf kube port-forward status` — confirm connectivity checks pass; append to `verification_proof/live.txt`
