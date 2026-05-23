# Tasks: Auto Port-Forward

- [x] 1.1 Add `auto-port-forward: "false"` to `internal/config/assets/config-template.yaml`
- [x] 2.1 Add `HF_MAESTRO_HTTP` → `{"maestro", "http-endpoint"}` to envVarMap
- [x] 2.2 Add `HF_MAESTRO_GRPC` → `{"maestro", "grpc-endpoint"}` to envVarMap
- [x] 3.1 Add `findFreePort()` to `internal/kube/kube.go`
- [x] 3.2 Add `EphemeralPortForward()` to `internal/kube/kube.go`
- [x] 4.1 Add `autoPortForwardStop` var to `cmd/root.go`
- [x] 4.2 Extend `PersistentPreRunE` to call `startAutoPortForwards(s)` when enabled
- [x] 4.3 Implement `startAutoPortForwards(s *config.Store)` in `cmd/root.go`
- [x] 4.4 Add `PersistentPostRunE` to `rootCmd`
- [x] 5.1 Add `TestEphemeralPortForward_PodNotFound` to `internal/kube/kube_test.go`
- [x] 5.2 Add `TestFindFreePort` to `internal/kube/kube_test.go`
- [x] 5.3 Add `TestAutoPortForward_DisabledByDefault` to `cmd/root_test.go`
- [x] 6.1 `go build ./...` — zero errors
- [x] 6.2 `go vet ./...` — zero warnings
- [x] 6.3 `go test ./...` — all pass
- [x] 6.4 Live: maestro consumers returned live data via auto-forwarded port
