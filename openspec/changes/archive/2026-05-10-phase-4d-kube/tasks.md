## 1. Dependencies

- [x] 1.1 Add `k8s.io/client-go`, `k8s.io/api`, `k8s.io/apimachinery` to go.mod

## 2. internal/kube package

- [x] 2.1 Create `internal/kube/kube.go` with `BuildConfig`, `NewClientset`, `IsProcessAlive`, `FindRunningPod`
- [x] 2.2 Implement `StartPortForward`, `StopPortForward`, `ListPortForwards`, `RunPortForwardDaemon`
- [x] 2.3 Implement `StreamLogs`, `StreamLogsFiltered`
- [x] 2.4 Implement `RunCurlPod`, `CreateDebugPod`
- [x] 2.5 Write unit tests in `internal/kube/kube_test.go`

## 3. cmd/kube.go — port-forward, curl, debug

- [x] 3.1 Add `hf kube port-forward` subcommand group
- [x] 3.2 Add `hf kube port-forward start [name] [localPort:remotePort]`
- [x] 3.3 Add `hf kube port-forward stop [name]`
- [x] 3.4 Add `hf kube port-forward status`
- [x] 3.5 Add `hf kube curl [--] [curl-args]`
- [x] 3.6 Add `hf kube debug <partial-deployment-name>`

## 4. cmd/logs.go — stream and adapter

- [x] 4.1 Implement default `hf logs [pattern]` (stern delegate or fan-out)
- [x] 4.2 Add `hf logs adapter [pattern] [--cluster-id]`

## 5. Verification

- [x] 5.1 `go build ./...` passes — save to `verification_proof/build.txt`
- [x] 5.2 `go vet ./...` passes — save to `verification_proof/vet.txt`
- [x] 5.3 `go test ./...` passes — save to `verification_proof/test.txt`
- [x] 5.4 Commit verification_proof files
