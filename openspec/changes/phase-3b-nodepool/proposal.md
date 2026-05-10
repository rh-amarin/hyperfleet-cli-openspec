# Phase 3b: NodePool Lifecycle

## What

Implement the full `hf nodepool` command tree — list, get, create, update, delete, conditions, statuses — mirroring the existing `hf cluster` implementation.

## Why

Phase 3a delivered cluster lifecycle management. NodePools are the next natural resource in the HyperFleet hierarchy. Operators need the same CRUD primitives for nodepools that they already have for clusters.

## Scope

- `cmd/nodepool.go`: all 7 subcommands using existing `internal/api`, `internal/config`, `internal/resource`, `internal/output` packages
- `cmd/nodepool_test.go`: httptest-based unit tests for all subcommands
- No new packages required; reuse helpers from `cmd/cluster.go` (same package)
