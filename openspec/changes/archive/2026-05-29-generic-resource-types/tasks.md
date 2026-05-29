## 1. OpenSpec artifacts

- [x] 1.1 proposal.md, design.md, spec deltas, tasks.md complete

## 2. Config model

- [x] 2.1 Add empty `resource-types:` to config template
- [x] 2.2 Implement `internal/config/resource_types.go` — parse, validate, path resolution, ancestor chain
- [x] 2.3 Extend env file load to read structured `resource-types` without breaking flat sections
- [x] 2.4 Add unit tests for parse, validation, and path resolution

## 3. Generic resource type

- [x] 3.1 Add `internal/resource/generic.go` with `GenericResource` and list helpers
- [x] 3.2 Extend `loadTemplate` to resolve `{config-dir}/templates/{create-template}`

## 4. Command implementation

- [x] 4.1 Create `cmd/resource.go` — `resourceCmd`, `rsCmd` alias, `types` subcommand
- [x] 4.2 Dynamic registration via preload before Cobra parses args
- [x] 4.3 Implement shared CRUD handlers parameterized by `ResourceTypeDef`
- [x] 4.4 Parent state resolution and state-key writes on search/create/id
- [x] 4.5 Register commands in `cmd/root.go` / `cmd/resource.go` init

## 5. Config show

- [x] 5.1 Surface configured resource type names in `hf config show`

## 6. Tests and verification

- [x] 6.1 Add `cmd/resource_test.go` with httptest coverage
- [x] 6.2 Run `go test ./...`, `go vet ./...`, `go build ./...`
- [x] 6.3 Save test/vet/build output to `verification_proof/`
- [x] 6.4 Live verification (`hf resource types`, `hf resource --help`); saved to `verification_proof/live.txt`
