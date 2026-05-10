## 1. Dependencies

- [ ] 1.1 Add `k8s.io/client-go`, `k8s.io/api`, `k8s.io/apimachinery` to go.mod

## 2. internal/kube package

- [ ] 2.1 Create `internal/kube/kube.go` with `BuildConfig`, `NewClientset`, `IsProcessAlive`, `FindRunningPod`
- [ ] 2.2 Implement `StartPortForward`, `StopPortForward`, `ListPortForwards`, `RunPortForwardDaemon`
- [ ] 2.3 Implement `StreamLogs`, `StreamLogsFiltered`
- [ ] 2.4 Implement `RunCurlPod`, `CreateDebugPod`
- [ ] 2.5 Write unit tests in `internal/kube/kube_test.go`

## 3. cmd/kube.go — port-forward, curl, debug

- [ ] 3.1 Add `hf kube port-forward` subcommand group
- [ ] 3.2 Add `hf kube port-forward start [name] [localPort:remotePort]`
- [ ] 3.3 Add `hf kube port-forward stop [name]`
- [ ] 3.4 Add `hf kube port-forward status`
- [ ] 3.5 Add `hf kube curl [--] [curl-args]`
- [ ] 3.6 Add `hf kube debug <partial-deployment-name>`

## 4. cmd/logs.go — stream and adapter

- [ ] 4.1 Implement default `hf logs [pattern]` (stern delegate or fan-out)
- [ ] 4.2 Add `hf logs adapter [pattern] [--cluster-id]`

## 5. Verification

- [ ] 5.1 `go build ./...` passes — save to `verification_proof/build.txt`
- [ ] 5.2 `go vet ./...` passes — save to `verification_proof/vet.txt`
- [ ] 5.3 `go test ./...` passes — save to `verification_proof/test.txt`
- [ ] 5.4 Commit verification_proof files
