# Tasks: phase-4a-db

- [ ] 1. Add `github.com/jackc/pgx/v5` dependency (`go get github.com/jackc/pgx/v5`)
- [ ] 2. Create `internal/db/db.go` — Config struct, NewFromConfig, Pool, Query, Exec, Querier interface, defaultQuerier adapter
- [ ] 3. Create `internal/db/db_test.go` — DSN construction, NULL rendering, truncation tests
- [ ] 4. Create `cmd/db.go` — `hf db query`, `hf db delete`, `hf db config` Cobra commands; register via init()
- [ ] 5. Create `cmd/db_test.go` — mock Querier, test query table/JSON/0-rows/file-error, delete unknown target/denied, config masking
- [ ] 6. Run `go build ./...`, `go vet ./...`, `go test ./...`; save output to `verification_proof/`
- [ ] 7. Archive change
