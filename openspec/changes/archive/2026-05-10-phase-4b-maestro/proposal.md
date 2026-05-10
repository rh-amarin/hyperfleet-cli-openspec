# Phase 4b: Maestro — Internal Client and hf maestro Commands

## Why

HyperFleet adapters use Maestro to deploy Kubernetes manifests to managed clusters. Operators currently must use raw `curl` calls or the Maestro web UI to inspect resource-bundles and consumers. The `hf maestro` command surfaces these operations as first-class CLI commands consistent with the rest of the `hf` tool.

## What Changes

- `internal/maestro/maestro.go` — new typed HTTP client wrapping the Maestro REST API
- `internal/maestro/maestro_test.go` — unit tests using `httptest.NewServer`
- `cmd/maestro.go` — full `hf maestro` subcommand tree (list, get, delete, bundles, consumers)
- `cmd/maestro_test.go` — command-level tests
- Delta spec at `openspec/changes/phase-4b-maestro/specs/maestro/spec.md`
