## 1. Config

- [x] 1.1 Add `identity-header` and `identity-value` commented-out keys to `cmd/assets/config-template.yaml` under the `hyperfleet` section

## 2. API Client

- [x] 2.1 Extend `api.NewClient` signature to accept `identityHeader, identityValue string` as the last two parameters
- [x] 2.2 Store `identityHeader` and `identityValue` as fields on `api.Client`
- [x] 2.3 In `do()`, set the identity header on the request when `identityHeader` is non-empty (alongside `Content-Type` and `Authorization`)
- [x] 2.4 In `printCurlCommand`, include the identity header line when `identityHeader` is non-empty

## 3. Command Wiring

- [x] 3.1 In the root command wiring (where `api.NewClient` is called), resolve `s.Get("hyperfleet", "identity-header")` and `s.Get("hyperfleet", "identity-value")` and pass them to `NewClient`

## 4. Tests

- [x] 4.1 Add unit tests in `internal/api/` using `httptest.NewServer` verifying: header present on GET/POST/PATCH/DELETE when configured; header absent when `identityHeader` is empty
- [x] 4.2 Run `go test ./...` and save output to `verification_proof/go-test.txt`
- [x] 4.3 Run `go build ./...` and `go vet ./...` and save output to `verification_proof/go-build-vet.txt`

## 5. Live Verification

- [x] 5.1 Set `identity-header: X-HyperFleet-Identity` and `identity-value: openspec@test.com` in the active environment file
- [x] 5.2 Run a live API call (e.g., `hf cluster list`) against `localhost:8000` and confirm the header is sent (use `--curl` flag to inspect, then run live)
- [x] 5.3 Save the live output to `verification_proof/live-identity-header.txt`
