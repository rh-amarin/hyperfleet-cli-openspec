# Tasks: create-no-disk-template

## Implementation

- [x] 1. `cmd/templates.go`: Change `loadTemplate` signature to `(configDir, resource, flagFile string) (map[string]any, error)` — remove `created` bool return value
- [x] 2. `cmd/templates.go`: Remove the `os.Stat` / `os.MkdirAll` / `os.WriteFile` branch; when `flagFile` is empty, use `embeddedDefault(resource)` bytes directly in memory
- [x] 3. `cmd/cluster.go`: Update `loadTemplate` call site — drop `created` variable, remove `if created { p.Info(...) }` block
- [x] 4. `cmd/nodepool.go`: Same update as cluster.go
- [x] 5. `cmd/templates_test.go`: Replace `TestLoadTemplate_AutoCreatesDefault` and `TestLoadTemplate_NodepoolDefault` — assert file NOT written, assert correct fields returned; drop `created` from all assertions; update `TestLoadTemplate_ExistingFile` and `TestLoadTemplate_FlagFileOverride` to drop `created` from signature; update `TestLoadTemplate_MalformedJSON` and `TestLoadTemplate_FlagFileMissing`

## Verification

- [x] 6. `go build ./...` — no errors
- [x] 7. `go vet ./...` — no warnings
- [x] 8. `go test ./cmd/...` — all tests pass; save output to `verification_proof/`
