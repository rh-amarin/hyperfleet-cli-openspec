## 1. Config package

- [x] 1.1 Add `EnvFilePath(name string) string` method to `internal/config/config.go` returning `filepath.Join(s.dir, "environments", name+".yaml")`

## 2. Command

- [x] 2.1 In `configShowCmd.RunE` (no-argument path), call `s.EnvFilePath(active)` and print the path before the YAML sections

## 3. Tests

- [x] 3.1 Add `TestConfigShow_EnvFilePath` to `cmd/config_test.go` — assert the env file path line appears in `hf config show` output
- [x] 3.2 Run `go test ./...` and confirm zero failures

## 4. Verification

- [x] 4.1 Run `go build ./...` and `go vet ./...` — must pass
- [x] 4.2 Run `./bin/hf config` against the live cluster and save output to `verification_proof/live.txt`
- [x] 4.3 Save `go test ./...` output to `verification_proof/unit_tests.txt`
