# HyperFleet CLI — Claude Code Instructions

## Environment Setup

Ensure the OpenSpec CLI is available at the start of every session:

```bash
openspec --version 2>/dev/null || npm install -g @fission-ai/openspec@latest
```

## Project Overview

`hf` is a self-contained Go CLI for managing HyperFleet clusters. It replaces a suite of bash scripts with a single binary — no external tools required (no `kubectl`, `psql`, `gcloud`, `gh`, etc.).

- **Module:** `github.com/rh-amarin/hyperfleet-cli`
- **Language:** Go 1.22+
- **Binary name:** `hf`
- **Real API:** configured via `hf config` (run `hf config doctor` to check connectivity)

## Common Commands

```bash
make build          # build ./bin/hf
make test           # go test ./...
go vet ./...        # static analysis
go build ./...      # verify compilation
```

## Architecture

| Layer | Package | Purpose |
|---|---|---|
| Commands | `cmd/` | Cobra subcommands, one file per domain |
| API client | `internal/api/` | Generic `Get[T]`, `Post[T]`, `Patch[T]`, `Delete`; RFC 9457 errors |
| Config | `internal/config/` | File-per-property at `~/.config/hf/<key>`; `config.yaml` + `state.yaml` |
| Output | `internal/output/` | `Printer` dispatches `--output json|table|yaml`; colored dot renderer |
| Resources | `internal/resource/` | `Cluster`, `NodePool`, `AdapterStatus`, `Condition`, `ListResponse[T]` |
| DB | `internal/db/` | `pgxpool` wrapper; `Query` → headers+rows, `Exec` for DML |
| Kubernetes | `internal/kube/` | `client-go` port-forward, log streaming, pod exec |
| Maestro | `internal/maestro/` | HTTP client for Maestro API |
| Pub/Sub | `internal/pubsub/` | GCP and RabbitMQ CloudEvent publishing |
| Repos | `internal/repos/` | `go-github` client for registry overview |

**Key patterns:**
- Config precedence: CLI flags > `HF_*` env vars > env profile > `config.yaml` > defaults
- Commands reuse `internal/` packages — never shell out to other `hf` subcommands
- `create` with no args uses defaults; does not show usage
- All data-producing commands support `--output json|table|yaml`
- External `hf-<name>` executables on `PATH` are auto-delegated as plugins
- API errors follow RFC 9457 (`APIError{Code, Detail, Status, Title, TraceID}`)

## Testing Conventions

- Test framework: Go standard library only (`testing` + `net/http/httptest`). No third-party frameworks.
- HTTP tests use `httptest.NewServer` — never mock the HTTP client directly.
- Integration tests are tagged `//go:build integration` and skipped without cluster access.

## OpenSpec Workflow

All changes follow the **spec-driven** workflow. Never implement code outside a change folder.

```
openspec new change "<name>"     # scaffold openspec/changes/<name>/
openspec status --change "<name>" --json
openspec instructions <artifact> --change "<name>" --json
# implement tasks, check off in tasks.md as you go
openspec archive --change "<name>"   # when all tasks done and verified
```

**Slash commands available:**
- `/opsx:propose` — create change + generate all artifacts (proposal, design, tasks)
- `/opsx:apply` — implement tasks from a change, one at a time
- `/opsx:explore` — read specs and understand scope before proposing
- `/opsx:archive` — merge delta specs into `openspec/specs/`, move to `archive/`

**Change folder structure:**
```
openspec/changes/<name>/
├── .openspec.yaml      # schema, createdAt, dependsOn
├── proposal.md         # what & why
├── design.md           # how (packages, key decisions)
├── tasks.md            # numbered checklist — check off immediately as done
└── specs/<domain>/spec.md   # delta against openspec/specs/
```

**Rules:**
- Check tasks off in `tasks.md` immediately — do not batch.
- If design needs to change during implementation, update `design.md` first.
- Archive one change before starting the next in the same dependency chain.
- Only archive when all tasks are checked and verification passes.

## Definition of Done

Every change MUST satisfy all of the following before archiving:

1. All new packages and functions have unit tests (using `httptest.NewServer` where HTTP is involved).
2. `go test ./...` passes with zero failures.
3. `go build ./...` and `go vet ./...` pass with no errors.
4. Test output is captured and saved to `verification_proof/` as `.txt` files (one per `go test` command).
5. `verification_proof/` files are committed and included in the change.
6. Live verification against the real cluster is performed and its output saved to `verification_proof/`.

## Spec Source of Truth

`openspec/specs/` is the authoritative spec. Changes that modify behavior must update the relevant spec via a delta in `openspec/changes/<name>/specs/`. The index is at `openspec/specs/index.md`.

## Config Keys Reference

| Section | Key | Default |
|---|---|---|
| `database` | `host` | `localhost` |
| `database` | `port` | `5432` |
| `database` | `name` | `hyperfleet` |
| `database` | `user` | `hyperfleet` |
| `database` | `password` | _(set via config — never hardcode)_ |
| `hyperfleet` | `api-url` | `http://localhost:8000` |
| `hyperfleet` | `api-version` | `v1` |
| `maestro` | `http-endpoint` | `http://localhost:8100` |
| `maestro` | `grpc-endpoint` | `localhost:8090` |
