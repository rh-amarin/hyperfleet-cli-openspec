## 1. cmd/cluster.go — Subcommands

- [x] 1.1 Implement `hf cluster list` — GET /clusters, JSON + table output
- [x] 1.2 Implement `hf cluster get [id]` — GET /clusters/{id}, JSON + table, ID resolution from state
- [x] 1.3 Implement `hf cluster create` — duplicate check, POST, persist cluster-id to state, INFO stderr
- [x] 1.4 Implement `hf cluster update <id>` — PATCH /clusters/{id} with --name/--replicas flags
- [x] 1.5 Implement `hf cluster delete <id>` — DELETE /clusters/{id}, silent success, RFC 7807 on 404
- [x] 1.6 Implement `hf cluster conditions [id]` — extract conditions from cluster, JSON + table
- [x] 1.7 Implement `hf cluster statuses [id]` — GET /clusters/{id}/statuses, JSON + table (ADAPTER GEN AVAILABLE)

## 2. cmd/cluster_test.go — Unit Tests

- [x] 2.1 Test `cluster list` happy path (200 with items)
- [x] 2.2 Test `cluster get` happy path (200)
- [x] 2.3 Test `cluster get` 404 propagated as API error JSON (exit 0)
- [x] 2.4 Test `cluster create` happy path (201)
- [x] 2.5 Test `cluster create` duplicate guard (WARN, no POST)
- [x] 2.6 Test `cluster update` happy path (200)
- [x] 2.7 Test `cluster delete` happy path (200 or 204, silent)
- [x] 2.8 Test `cluster delete` 404 error case
- [x] 2.9 Test `cluster conditions` happy path (JSON output)
- [x] 2.10 Test `cluster statuses` happy path (JSON output)

## 3. Verification

- [x] 3.1 Run `go build ./...` — save output to `verification_proof/build.txt`
- [x] 3.2 Run `go vet ./...` — save output to `verification_proof/vet.txt`
- [x] 3.3 Run `go test ./...` — save output to `verification_proof/test.txt`
- [x] 3.4 Commit verification_proof/ files
