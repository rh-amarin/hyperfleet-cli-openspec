# Phase 1 Foundation — Implementation Spec

## Summary

This delta captures the implementation requirements for the four foundation packages introduced in Phase 1. These packages have no existing specs — they are new capabilities.

Refer to:
- `openspec/specs/config-model/spec.md` for config requirements
- `openspec/specs/api-client/spec.md` for API client requirements
- `openspec/specs/resource-types/spec.md` for resource type requirements
- `openspec/specs/output-formatting/spec.md` for output formatting requirements

## Implementation Notes

- `internal/config.Store` MUST NOT use global state — callers pass the store around explicitly
- `internal/api` generic functions are package-level (`Get[T]`, `Post[T]`, etc.), not methods, to work around Go's lack of method-level type parameters
- `internal/resource` types MUST be pure value types with no methods beyond what the Go runtime provides
- `internal/output.Printer` MUST write to `io.Writer` (not hardcoded `os.Stdout`) to enable testing
- All test files use `testing.T.TempDir()` or `bytes.Buffer` — no real file system or network access
