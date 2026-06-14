## 1. Port-based Cleanup in stopPortForwardsBeforeStart

- [x] 1.1 In `cmd/kube.go`, add a second cleanup pass to `stopPortForwardsBeforeStart`: after the PID-file stop loop, iterate `services` and for each call `kube.PIDForPort(svc.localPort)`; if a PID is found send SIGTERM and log `[INFO] Killed stray process <pid> on port <port>`
- [x] 1.2 Add a test in `cmd/kube_test.go` covering the stray-process scenario: no PID file tracked, but `PIDForPort` returns a PID → verify the stray-process log line is emitted

## 2. Verification

- [x] 2.1 Run `go build ./...` and `go vet ./...` — zero errors
- [x] 2.2 Run `go test ./...` — capture output to `verification_proof/unit_tests.txt`
- [x] 2.3 Live verify: start a port-forward manually with `kubectl`, then run `hf kube port-forward start` and confirm it cleanly kills the stray process and starts — save output to `verification_proof/live_pf_restart.txt`
