## 1. Add --curl persistent flag (cmd/root.go)

- [x] 1.1 Add `curlMode bool` to the global flag vars block
- [x] 1.2 Register `--curl` in `init()` as a persistent flag: `"print equivalent curl command for each API request"`

## 2. Update internal/api/client.go

- [x] 2.1 Add `curlMode bool` field to `Client` struct
- [x] 2.2 Add `curlMode bool` as 4th parameter to `NewClient`; set the field in the returned struct
- [x] 2.3 In `do()`: hoist body JSON bytes into a `bodyBytes []byte` variable before creating `bodyReader`, so they are available for curl printing
- [x] 2.4 In `do()`: after headers are set and before `c.http.Do(req)`, call `printCurlCommand(os.Stderr, method, url, c.token, bodyBytes)` when `c.curlMode`
- [x] 2.5 Add package-private `printCurlCommand(w io.Writer, method, url, token string, body []byte)` function:
  - Print `[CURL] curl -s -X <method> "<url>"`
  - Print ` \\\n  -H 'Accept: application/json'`
  - If body non-empty: print ` \\\n  -H 'Content-Type: application/json'`
  - If token non-empty: print ` \\\n  -H 'Authorization: Bearer <token>'`
  - If body non-empty: print ` \\\n  -d '<body>'`
  - Print final newline

## 3. Update internal/maestro/maestro.go

- [x] 3.1 Add `curlMode bool` field to `Client` struct
- [x] 3.2 Add `curlMode bool` as 2nd parameter to `NewFromConfig`; set the field in the returned struct
- [x] 3.3 In `get()`: after building the request and before `c.http.Do(req)`, if `c.curlMode` print: `[CURL] curl -s "<url>" \\\n  -H 'Accept: application/json'\n`
- [x] 3.4 In `delete()`: after building the request and before `c.http.Do(req)`, if `c.curlMode` print: `[CURL] curl -s -X DELETE "<url>"\n`
- [x] 3.5 Add `"os"` import

## 4. Update call sites

- [x] 4.1 `cmd/cluster.go` `newAPIClient()`: pass `curlMode` as 4th arg to `api.NewClient`
- [x] 4.2 `cmd/maestro.go` `newMaestroClient()`: pass `curlMode` as 2nd arg to `maestro.NewFromConfig`
- [x] 4.3 `cmd/maestro.go` inline `maestro.NewFromConfig(s)` in `maestroListCmd.RunE`: pass `curlMode`
- [x] 4.4 `internal/api/client_test.go`: add `false` as 4th arg to all 6 `api.NewClient` calls
- [x] 4.5 `internal/maestro/maestro_test.go`: add `false` as 2nd arg to the `maestro.NewFromConfig` call in `newTestClient`

## 5. Verify

- [x] 5.1 `go build ./...` succeeds
- [x] 5.2 `go vet ./...` reports no issues
- [x] 5.3 `go test ./...` passes — capture output to `verification_proof/tests.txt`
- [x] 5.4 Live verification: run `hf cluster list --curl` and `hf maestro list --curl` against the real cluster; save output to `verification_proof/live.txt`
