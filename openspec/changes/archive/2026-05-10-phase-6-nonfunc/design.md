## Context

`hf` is feature-complete but is missing the build/release infrastructure needed to distribute the binary. Shell completions exist in Cobra but are not yet wired up to a subcommand. There is no CI pipeline, no release automation, and binaries are built only locally with `make build`.

## Goals / Non-Goals

**Goals:**
- Wire up Cobra's built-in completion generator as `hf completion [shell]`
- Register `--output` flag completions on the root command so tab-completion works for all subcommands
- Provide a `.goreleaser.yaml` that cross-compiles for linux/darwin/windows × amd64/arm64 and injects the version at build time
- Provide a CI workflow (`ci.yml`) that runs build + vet + test on every push/PR to `main`
- Provide a release workflow (`release.yml`) that runs GoReleaser on version tags

**Non-Goals:**
- Homebrew tap / package manager distribution (future phase)
- Code signing or notarization
- Docker image builds
- SBOM or provenance attestation
- Changelogs beyond what GoReleaser generates automatically

## Decisions

### Completion command: Cobra built-ins

Cobra provides `GenBashCompletionV2`, `GenZshCompletion`, `GenFishCompletion`, and `GenPowerShellCompletionWithDesc` on the root command. Using these avoids any external dependency and stays consistent with every other major Cobra-based CLI (kubectl, helm, gh).

`ValidArgs` / `ValidArgsFunction` on each command lets Cobra generate meaningful completions at no extra cost.

**Alternative considered:** generate static completion scripts at build time and embed them — rejected because they would go stale as commands evolve.

### GoReleaser v2 config

GoReleaser v2 is the current stable schema. `CGO_ENABLED=0` ensures purely static binaries that work in any Linux container. Version is injected via `-X github.com/rh-amarin/hyperfleet-cli/internal/version.Version={{.Version}}` — the `internal/version` package already exists from prior phases.

**Alternative considered:** `ko` or `docker buildx` — overkill for a CLI with no container-native distribution target.

### GitHub Actions: ubuntu-latest only

The CLI is a single binary with no platform-specific runtime requirements. Building on `ubuntu-latest` is sufficient to validate correctness. Cross-compilation is GoReleaser's job, not the test runner's.

### `.gitignore` additions

`bin/` is where `make build` writes the binary. `dist/` is where GoReleaser writes release artifacts. Both are purely local and must never be committed.

## Risks / Trade-offs

- [Risk] `internal/version.Version` ldflag path must match exactly → Mitigation: verify with `go build -ldflags "-X github.com/rh-amarin/hyperfleet-cli/internal/version.Version=test" ./...` during implementation
- [Risk] GoReleaser `version: 2` schema may require a specific goreleaser binary version → Mitigation: pin the goreleaser-action to `v6` which ships a compatible version; `goreleaser check` validates locally

## Migration Plan

No migration needed — purely additive. The existing `make build` and `go test ./...` workflows are unaffected.

## Open Questions

None. All decisions are resolved.
