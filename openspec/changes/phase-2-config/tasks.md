## 1. Planning Artifacts

- [x] 1.1 Write proposal.md
- [x] 1.2 Write design.md
- [x] 1.3 Write specs/config/spec.md (delta)
- [x] 1.4 Write tasks.md

## 2. Active-Env Guard (cmd/root.go)

- [ ] 2.1 Add `isBypassCommand(cmd)` helper that returns true for `config env *`, `version`, `completion`, `help`
- [ ] 2.2 Update `PersistentPreRunE` to call `config.NewFromEnv().Load()` then `RequireActiveEnvironment()`, bypassing for bypass commands
- [ ] 2.3 Add test in `cmd/root_test.go` verifying guard fires for `config show` and is bypassed for `config env list` and `version`

## 3. cmd/config.go — Core Commands

- [ ] 3.1 Implement `hf config show` — load store, render config as YAML with secret masking
- [ ] 3.2 Implement `hf config get <section> <key>` — print resolved value or error
- [ ] 3.3 Implement `hf config set <section> <key> <value>` — validate section, write to config.yaml

## 4. cmd/config.go — Env Subcommands

- [ ] 4.1 Implement `hf config env list` (alias: `ls`) — table: NAME, API URL, ACTIVE
- [ ] 4.2 Implement `hf config env create <name>` — flags: --api-url, --api-token, --cluster-id, --nodepool-id
- [ ] 4.3 Implement `hf config env activate <name>` — set active-environment in state.yaml
- [ ] 4.4 Implement `hf config env delete <name>` (alias: `rm`) — remove file, clear state if active
- [ ] 4.5 Implement `hf config env show <name>` — print env file path and YAML contents

## 5. cmd/config.go — Doctor

- [ ] 5.1 Implement `hf config doctor` — 5s timeout HTTP GET to api-url, print OK/ERROR

## 6. Tests (cmd/config_test.go)

- [ ] 6.1 Test `hf config show` with active env set (check YAML output)
- [ ] 6.2 Test `hf config get` — found and not-found cases
- [ ] 6.3 Test `hf config set` — valid section writes; invalid section errors
- [ ] 6.4 Test `hf config env list` — lists environments, marks active
- [ ] 6.5 Test `hf config env create` — creates file; errors on duplicate
- [ ] 6.6 Test `hf config env activate` — sets state; errors on not-found
- [ ] 6.7 Test `hf config env delete` — removes file; clears state if active; errors on not-found
- [ ] 6.8 Test `hf config env show` — prints YAML; errors on not-found
- [ ] 6.9 Test `hf config doctor` — reachable (httptest.NewServer) and unreachable cases
- [ ] 6.10 Test active-env guard — bypassed for `config env activate`, fires for `config show`

## 7. Verification

- [ ] 7.1 Run `go build ./...` → capture to `verification_proof/build.txt`
- [ ] 7.2 Run `go vet ./...` → capture to `verification_proof/vet.txt`
- [ ] 7.3 Run `go test ./...` → capture to `verification_proof/test.txt`
- [ ] 7.4 Commit verification_proof/ files
