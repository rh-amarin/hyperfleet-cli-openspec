## 1. Go Module Initialization

- [x] 1.1 Create `go.mod` with module `github.com/rh-amarin/hyperfleet-cli`, Go 1.22, and `github.com/spf13/cobra` as the initial direct dependency
- [x] 1.2 Create `main.go` at the repo root that calls `cmd.Execute()` and exits non-zero on error

## 2. Root Command and Global Flags

- [x] 2.1 Create `cmd/root.go` defining `rootCmd` with `Use: "hf"`, short/long descriptions, and `PersistentPreRunE` that populates package-level flag variables
- [x] 2.2 Register persistent flags on `rootCmd`: `--config`, `--output` (default `"json"`), `--no-color`, `--verbose`/`-v`, `--api-url`, `--api-token`
- [x] 2.3 Create `cmd/execute.go` (or add `Execute()` to `root.go`) that calls `rootCmd.Execute()` and returns any error

## 3. Command Stub Files

- [ ] 3.1 Create `cmd/cluster.go` — stub `clusterCmd` registered with `rootCmd`, calls `cmd.Help()` when invoked directly
- [ ] 3.2 Create `cmd/nodepool.go` — stub `nodepoolCmd`
- [ ] 3.3 Create `cmd/config.go` — stub `configCmd`
- [ ] 3.4 Create `cmd/db.go` — stub `dbCmd`
- [ ] 3.5 Create `cmd/maestro.go` — stub `maestroCmd`
- [ ] 3.6 Create `cmd/pubsub.go` — stub `pubsubCmd`
- [ ] 3.7 Create `cmd/rabbitmq.go` — stub `rabbitmqCmd`
- [ ] 3.8 Create `cmd/kube.go` — stub `kubeCmd`
- [ ] 3.9 Create `cmd/logs.go` — stub `logsCmd`
- [ ] 3.10 Create `cmd/repos.go` — stub `reposCmd`
- [ ] 3.11 Create `cmd/resources.go` — stub `resourcesCmd`
- [ ] 3.12 Create `cmd/version.go` — prints `internal/version.Version` to stdout
- [ ] 3.13 Register `completion` command using Cobra's built-in `GenBashCompletion` etc. (can use `rootCmd.GenCompletionScript` approach)

## 4. Version Package

- [ ] 4.1 Create `internal/version/version.go` with `var Version = "dev"` and a `String() string` helper

## 5. Makefile

- [ ] 5.1 Create `Makefile` with `build` target (`go build -ldflags ... -o bin/hf .`), `test` target (`go test ./...`), and `vet` target (`go vet ./...`)

## 6. Unit Tests

- [ ] 6.1 Create `internal/version/version_test.go` — verify `Version` is non-empty and `String()` returns it
- [ ] 6.2 Create `cmd/root_test.go` — verify `Execute()` returns nil for `--help`; verify all global flags are registered on the root command

## 7. Verify

- [ ] 7.1 `go build ./...` succeeds with no errors
- [ ] 7.2 `go vet ./...` reports no issues
- [ ] 7.3 `go test ./...` passes — capture full output and save to `verification_proof/tests.txt`
- [ ] 7.4 Run `./bin/hf --help` against the built binary and save output to `verification_proof/7.4-hf-help.txt`
- [ ] 7.5 Run `./bin/hf version` and save output to `verification_proof/7.5-hf-version.txt`
