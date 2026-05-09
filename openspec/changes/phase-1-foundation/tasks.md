# Tasks: phase-1-foundation

## Setup

- [ ] 1. Add `gopkg.in/yaml.v3` and `golang.org/x/term` to go.mod via `go get`

## internal/resource

- [ ] 2. Create `internal/resource/resource.go` with all type definitions: `Cluster`, `NodePool`, `ClusterStatus`, `NodePoolStatus`, `ResourceCondition`, `AdapterCondition`, `AdapterStatus`, `AdapterStatusMetadata`, `AdapterStatusCreateRequest`, `ConditionRequest`, `CloudEvent`, `ListResponse[T]`, `ObjectReference`, `ValidationError`
- [ ] 3. Create `internal/resource/resource_test.go` with JSON round-trip tests for `Cluster`, `NodePool`, `ListResponse[Cluster]`, `AdapterStatus`

## internal/config

- [ ] 4. Create `internal/config/config.go` with `Store`, `New()`, `Load()`, `Get()`, `Set()`, `GetState()`, `SetState()`, `ActiveEnvironment()`, `RequireActiveEnvironment()`, `ClusterID()`, `NodePoolID()`, deep-merge, atomic write helpers
- [ ] 5. Create `internal/config/config_test.go` with tests: defaults loaded, precedence (env var overrides file), atomic state write, env profile deep-merge, `RequireActiveEnvironment` error when no active env

## internal/api

- [ ] 6. Create `internal/api/errors.go` with `APIError`, `ValidationError`, `parseError()` function
- [ ] 7. Create `internal/api/client.go` with `Client`, `NewClient()`, `ResourceHref()`, generic `Get[T]`, `Post[T]`, `Patch[T]`, `Delete[T]`
- [ ] 8. Create `internal/api/client_test.go` with tests: GET happy path, POST happy path, 404 RFC 7807 parsing, 400 with validation errors, non-JSON error, HTML error, verbose logging, Bearer token header

## internal/output

- [ ] 9. Create `internal/output/dots.go` with `StatusDot()` (True=green●, False=red●, Unknown=yellow●, absent=dash)
- [ ] 10. Create `internal/output/columns.go` with `DynamicColumns()` (fixed + Available + alpha others + Reconciled)
- [ ] 11. Create `internal/output/printer.go` with `Printer`, `NewPrinter()`, `Print()`, `PrintTable()`, `Warn()`, `Info()`, `Error()`, colored JSON output
- [ ] 12. Create `internal/output/printer_test.go` with tests: JSON output, table output, YAML output, dot renderer (all statuses), no-color mode, dynamic column ordering

## Verification

- [ ] 13. Run `go build ./...` and save output to `verification_proof/build.txt`
- [ ] 14. Run `go vet ./...` and save output to `verification_proof/vet.txt`
- [ ] 15. Run `go test ./...` and save output to `verification_proof/test.txt`
