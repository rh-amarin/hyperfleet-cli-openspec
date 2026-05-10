## 1. cmd/cluster.go — Subcommands

- [ ] 1.1 Implement `hf cluster list` — GET /clusters, JSON + table output
- [ ] 1.2 Implement `hf cluster get [id]` — GET /clusters/{id}, JSON + table, ID resolution from state
- [ ] 1.3 Implement `hf cluster create` — duplicate check, POST, persist cluster-id to state, INFO stderr
- [ ] 1.4 Implement `hf cluster update <id>` — PATCH /clusters/{id} with --name/--replicas flags
- [ ] 1.5 Implement `hf cluster delete <id>` — DELETE /clusters/{id}, silent success, RFC 7807 on 404
- [ ] 1.6 Implement `hf cluster conditions [id]` — extract conditions from cluster, JSON + table
- [ ] 1.7 Implement `hf cluster statuses [id]` — GET /clusters/{id}/statuses, JSON + table (ADAPTER GEN AVAILABLE)

## 2. cmd/cluster_test.go — Unit Tests

- [ ] 2.1 Test `cluster list` happy path (200 with items)
- [ ] 2.2 Test `cluster get` happy path (200)
- [ ] 2.3 Test `cluster get` 404 propagated as API error JSON (exit 0)
- [ ] 2.4 Test `cluster create` happy path (201)
- [ ] 2.5 Test `cluster create` duplicate guard (WARN, no POST)
- [ ] 2.6 Test `cluster update` happy path (200)
- [ ] 2.7 Test `cluster delete` happy path (200 or 204, silent)
- [ ] 2.8 Test `cluster delete` 404 error case
- [ ] 2.9 Test `cluster conditions` happy path (JSON output)
- [ ] 2.10 Test `cluster statuses` happy path (JSON output)

## 3. Verification

- [ ] 3.1 Run `go build ./...` — save output to `verification_proof/build.txt`
- [ ] 3.2 Run `go vet ./...` — save output to `verification_proof/vet.txt`
- [ ] 3.3 Run `go test ./...` — save output to `verification_proof/test.txt`
- [ ] 3.4 Commit verification_proof/ files
