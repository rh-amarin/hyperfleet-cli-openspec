## 1. API client dry-run

- [x] 1.1 Add `var ErrDryRun = errors.New("dry-run")` to `internal/api`
- [x] 1.2 In `Client.do()`: when `curlMode`, print curl and return `ErrDryRun` without `http.Do`
- [x] 1.3 Remove `WithoutCurl()` and `PrintCurl()` if no longer needed after cmd updates
- [x] 1.4 Add unit tests: curl mode does not hit transport; returns `ErrDryRun`

## 2. Maestro client dry-run

- [x] 2.1 Add `maestro.ErrDryRun` sentinel
- [x] 2.2 In `get()` and `delete()`: when `curlMode`, print curl and return `ErrDryRun` without HTTP
- [x] 2.3 Add unit tests for maestro dry-run

## 3. Command wiring

- [x] 3.1 Update `--curl` flag help text in `cmd/root.go`
- [x] 3.2 Extend `handleAPIError` (or add helper) to treat `api.ErrDryRun` as success (exit 0, no output)
- [x] 3.3 Handle `maestro.ErrDryRun` in maestro commands
- [x] 3.4 `cluster create` / `nodepool create`: skip duplicate GET when `curlMode`; remove `WithoutCurl`/`PrintCurl` workaround
- [x] 3.5 Watch commands: when `curlMode`, print first fetch curl and return without entering watch loop
- [x] 3.6 Interactive commands: reject `--curl` + `-i` with clear error
- [x] 3.7 Audit other cmd paths (tui, ui, pubsub, resources) for `ErrDryRun` handling

## 4. Tests and verification

- [x] 4.1 `TestClusterList_CurlDryRun` — curl on stderr, no JSON stdout, exit 0
- [x] 4.2 Update `TestClusterCreate_CurlModeDuplicateShowsPost` for dry-run semantics (no GET curl, no HTTP)
- [x] 4.3 Add maestro list dry-run test
- [x] 4.4 Run `go test ./...`, `go vet ./...`; save output to `verification_proof/`
- [x] 4.5 Live verification: `hf cluster list --curl`, `hf cluster create --curl`, `hf maestro list --curl`
