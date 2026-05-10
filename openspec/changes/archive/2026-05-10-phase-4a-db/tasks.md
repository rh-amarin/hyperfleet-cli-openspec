# Tasks: phase-4a-db

- [x] 1. Add `github.com/jackc/pgx/v5` dependency (`go get github.com/jackc/pgx/v5`)
- [x] 2. Create `internal/db/db.go` — Config struct, NewFromConfig, Pool, Query, Exec, Querier interface, defaultQuerier adapter
- [x] 3. Create `internal/db/db_test.go` — DSN construction, NULL rendering, truncation tests
- [x] 4. Create `cmd/db.go` — `hf db query`, `hf db delete`, `hf db config` Cobra commands; register via init()
- [x] 5. Create `cmd/db_test.go` — mock Querier, test query table/JSON/0-rows/file-error, delete unknown target/denied, config masking
- [x] 6. Run `go build ./...`, `go vet ./...`, `go test ./...`; save output to `verification_proof/`
- [x] 7. Archive change
