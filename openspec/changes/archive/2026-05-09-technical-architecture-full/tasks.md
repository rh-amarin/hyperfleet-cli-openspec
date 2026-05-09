## 3. Command Stub Files

- [x] 3.1 Create `cmd/cluster.go` — stub `clusterCmd` registered with `rootCmd` via `init()`
- [x] 3.2 Create `cmd/nodepool.go` — stub `nodepoolCmd` registered with `rootCmd` via `init()`
- [x] 3.3 Create `cmd/config.go` — stub `configCmd` registered with `rootCmd` via `init()`
- [x] 3.4 Create `cmd/db.go` — stub `dbCmd` registered with `rootCmd` via `init()`
- [x] 3.5 Create `cmd/maestro.go` — stub `maestroCmd` registered with `rootCmd` via `init()`
- [x] 3.6 Create `cmd/pubsub.go` — stub `pubsubCmd` registered with `rootCmd` via `init()`
- [x] 3.7 Create `cmd/rabbitmq.go` — stub `rabbitmqCmd` registered with `rootCmd` via `init()`
- [x] 3.8 Create `cmd/kube.go` — stub `kubeCmd` registered with `rootCmd` via `init()`
- [x] 3.9 Create `cmd/logs.go` — stub `logsCmd` registered with `rootCmd` via `init()`
- [x] 3.10 Create `cmd/repos.go` — stub `reposCmd` registered with `rootCmd` via `init()`
- [x] 3.11 Create `cmd/resources.go` — stub `resourcesCmd` registered with `rootCmd` via `init()`
- [x] 3.12 Create `cmd/version.go` — prints `internal/version.Version` to stdout; registered with `rootCmd` via `init()`
- [x] 3.13 Register completion command using Cobra's built-in completion support in `cmd/completion.go`

## 4. Version Package

- [x] 4.1 Create `internal/version/version.go` with `var Version = "dev"` and a `String() string` helper

## 5. Makefile

- [x] 5.1 Create `Makefile` with `build` target (`go build -ldflags ... -o bin/hf .`), `test` target (`go test ./...`), and `vet` target (`go vet ./...`)

## 6. Unit Tests

- [x] 6.1 Create `internal/version/version_test.go` — verify `Version` is non-empty and `String()` returns it
- [x] 6.2 Create `cmd/root_test.go` — verify `Execute()` returns nil for `--help`; verify all global flags are registered on the root command

## 7. Verify

- [x] 7.1 `go build ./...` succeeds with no errors — save output to `verification_proof/7.1-go-build.txt`
- [x] 7.2 `go vet ./...` reports no issues — save output to `verification_proof/7.2-go-vet.txt`
- [x] 7.3 `go test ./...` passes — save full output to `verification_proof/7.3-go-test.txt`
- [x] 7.4 Run `./bin/hf --help` against the built binary and save output to `verification_proof/7.4-hf-help.txt`
- [x] 7.4d (SKIPPED — live cluster not available in this environment) Live cluster verification — noted in `verification_proof/7.4d-live-cluster.txt`
- [x] 7.5 Run `./bin/hf version` and save output to `verification_proof/7.5-hf-version.txt`
