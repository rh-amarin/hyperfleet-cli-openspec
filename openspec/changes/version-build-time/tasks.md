# Tasks: version-build-time

## `internal/version/version.go`

- [x] 1. Add `var BuildTime = "unknown"`
- [x] 2. Update `String()` to return `Version + " (built " + BuildTime + ")"`

## `internal/version/version_test.go`

- [x] 3. Replace `TestStringReturnsVersion` with `TestStringContainsVersion` — `String()` must contain `Version`
- [x] 4. Add `TestStringContainsBuildTime` — `String()` must contain `BuildTime`
- [x] 5. Add `TestBuildTimeNonEmpty` — `BuildTime` must not be empty

## `Makefile`

- [x] 6. Add `BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)`
- [x] 7. Extend `LDFLAGS` with `-X $(MODULE)/internal/version.BuildTime=$(BUILD_TIME)`

## `.goreleaser.yaml`

- [x] 8. Extend ldflags with `-X github.com/rh-amarin/hyperfleet-cli/internal/version.BuildTime={{.Date}}`

## Verify

- [x] 9. (a) `go build ./...` — must pass with zero errors
- [x] 10. (b) `go vet ./...` — must report no issues
- [x] 11. (c) `go test ./...` — must pass with zero failures; capture full output and save to `verification_proof/tests.txt`
- [x] 12. (d) Live verification — build with `make build` and run `bin/hf version`; capture output and save to `verification_proof/live.txt`
