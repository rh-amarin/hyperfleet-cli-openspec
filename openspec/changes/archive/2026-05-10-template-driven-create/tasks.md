# Tasks: template-driven-create

## 1. Static asset files

- [x] 1.1 Create `cmd/assets/cluster-template.json` with the default cluster payload (kind, name, labels, spec with region/version/counter)
- [x] 1.2 Create `cmd/assets/nodepool-template.json` with the default nodepool payload (kind, name, labels, spec with platform/type/replicas/counter)

## 2. `cmd/templates.go` — embed and loader

- [x] 2.1 Create `cmd/templates.go`: embed both asset files via `//go:embed assets/cluster-template.json` and `//go:embed assets/nodepool-template.json`
- [x] 2.2 Implement `loadTemplate(configDir, resource, flagFile string) (map[string]any, bool, error)`:
  - `flagFile != ""` → read from that path, parse JSON, return (map, false, err)
  - `flagFile == ""` → check `<configDir>/<resource>-template.json`; if missing, write embedded default and return (map, true, nil); if present, read and parse

## 3. `cmd/cluster.go` — template-driven create

- [x] 3.1 Add `clusterCreateFile string` flag var and register `-f, --file` flag on `clusterCreateCmd` in `init()`
- [x] 3.2 In `clusterCreateCmd.RunE`: call `loadTemplate(s.ConfigDir(), "cluster", clusterCreateFile)` to get the base body
- [x] 3.3 Print `[INFO] Created default cluster template at <path>` when `loadTemplate` returns `created=true`
- [x] 3.4 Apply name override: positional `args[0]` > `--name` flag > template's `name` field
- [x] 3.5 Apply region/version overrides from `args[1]` / `args[2]` into `body["spec"]`
- [x] 3.6 Apply `--replicas` and `--nodepool-id` flag overrides into the body map after template load
- [x] 3.7 Remove the old hardcoded `spec` and `body` literal construction

## 4. `cmd/nodepool.go` — template-driven create

- [x] 4.1 Add `nodepoolCreateFile string` flag var and register `-f, --file` flag on `nodepoolCreateCmd` in `init()`
- [x] 4.2 Update `nodepoolCreateCmd.Use` to `"create [name]"` and add `Args: cobra.MaximumNArgs(1)`
- [x] 4.3 In `nodepoolCreateCmd.RunE`: call `loadTemplate(s.ConfigDir(), "nodepool", nodepoolCreateFile)` to get the base body
- [x] 4.4 Print `[INFO] Created default nodepool template at <path>` when `loadTemplate` returns `created=true`
- [x] 4.5 Apply name override: positional `args[0]` > `--name` flag > template's `name` field
- [x] 4.6 Apply `--type` flag override into `body["spec"]["platform"]["type"]` when non-empty
- [x] 4.7 Apply `--replicas` flag override into `body["spec"]["replicas"]` when > 0
- [x] 4.8 Remove the old hardcoded `spec` and `body` literal construction

## 5. Tests

- [x] 5.1 Create `cmd/templates_test.go`: test `loadTemplate` with missing config-dir file (auto-creates), existing file, and `-f` explicit path
- [x] 5.2 Add `TestClusterCreate_Template` in `cmd/cluster_test.go`: mock HTTP server returns 201; assert request body comes from template; assert name override from positional arg
- [x] 5.3 Add `TestClusterCreate_FileFlag` in `cmd/cluster_test.go`: `-f` with a custom JSON file; assert body sent matches file contents
- [x] 5.4 Add `TestClusterCreate_MalformedTemplate` in `cmd/cluster_test.go`: invalid JSON template file; assert `[ERROR] loading template` returned
- [x] 5.5 Add `TestNodepoolCreate_Template` in `cmd/nodepool_test.go`: mock HTTP server; assert body comes from template; assert name override
- [x] 5.6 Add `TestNodepoolCreate_FileFlag` in `cmd/nodepool_test.go`: `-f` with custom JSON; assert body matches
- [x] 5.7 Add `TestNodepoolCreate_MalformedTemplate` in `cmd/nodepool_test.go`: invalid JSON; assert error

## 6. Verification

- [x] 6.1 Run `go build ./...` — must pass with no errors
- [x] 6.2 Run `go vet ./...` — must pass with no errors
- [x] 6.3 Run `go test ./cmd/... -v` — must pass; save output to `verification_proof/cmd_test.txt`
- [x] 6.4 Run `go test ./... -v` — must pass; save output to `verification_proof/all_test.txt`
