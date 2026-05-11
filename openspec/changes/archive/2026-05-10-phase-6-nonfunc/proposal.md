## Why

The CLI is feature-complete but lacks the tooling to build, release, and distribute it reliably. Shell completions, CI pipelines, and a GoReleaser configuration are needed before the project can be shipped to end users.

## What Changes

- Add `cmd/completion.go` — Cobra-driven shell completion for bash, zsh, fish, and powershell
- Register `--output` flag completion across all data-producing commands
- Add `.goreleaser.yaml` — cross-platform builds (linux/darwin/windows × amd64/arm64) with version injection
- Add `.github/workflows/ci.yml` — build + vet + test on every push/PR to `main`
- Add `.github/workflows/release.yml` — GoReleaser triggered on version tags
- Extend `Makefile` with `completions` and `lint` targets
- Extend `.gitignore` to cover `bin/`, `dist/`, build artifacts

## Capabilities

### New Capabilities

- `shell-completions`: `hf completion [bash|zsh|fish|powershell]` generates shell completion scripts via Cobra built-ins
- `goreleaser-config`: GoReleaser config for cross-platform release builds with ldflags version injection
- `github-actions-ci`: CI workflow that builds, vets, and tests on every push and PR to `main`
- `github-actions-release`: Release workflow that runs GoReleaser on version tags

### Modified Capabilities

<!-- No existing spec-level requirements are changing -->

## Impact

- **cmd/**: new `completion.go`, new `completion_test.go`
- **Root command**: `--output` flag gets `ValidArgsFunction` for tab completion
- **Makefile**: new `completions` and `lint` targets
- **.gitignore**: added `bin/`, `dist/`, `*.exe`, `*.test`
- **New files**: `.goreleaser.yaml`, `.github/workflows/ci.yml`, `.github/workflows/release.yml`
- **No API or runtime behavior changes** — all additions are build/release tooling or CLI UX

### Testing Scope

- `cmd/completion_test.go`: test each of the 4 shells — verify output contains expected markers (`bash`, `#compdef`, etc.)
- No HTTP or cluster access needed; completion generation is purely local

### Verification

- `go build ./...`, `go vet ./...`, `go test ./...` — no live cluster needed
- Optionally: `goreleaser check` to validate `.goreleaser.yaml` syntax (if goreleaser is installed)
