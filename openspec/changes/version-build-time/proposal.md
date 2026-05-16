## Why

`hf version` currently prints only the version string (e.g. `v0.3.1` or `dev`). When
diagnosing issues with a deployed binary it is also useful to know exactly when it was
compiled, because the same version tag can produce different binaries if the source tree
was dirty or the binary was rebuilt from a branch tip without a new tag.

Showing the compile time gives operators a second, unambiguous identifier for the build.

## What Changes

**`internal/version/version.go`**
- Add `var BuildTime = "unknown"` — injected at link time; falls back to `"unknown"` for
  `go run` / plain `go build` invocations.
- Update `String()` to return `"<version> (built <buildtime>)"`.

**`Makefile`**
- Add `BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)` variable.
- Extend `LDFLAGS` with `-X $(MODULE)/internal/version.BuildTime=$(BUILD_TIME)`.

**`.goreleaser.yaml`**
- Extend ldflags entry with `-X github.com/rh-amarin/hyperfleet-cli/internal/version.BuildTime={{.Date}}`.

## Example output

```
$ hf version
v0.3.1 (built 2026-05-16T14:30:00Z)
```

Or without ldflags injection:
```
$ go run . version
dev (built unknown)
```

## Testing Scope

| Package | Test cases |
|---|---|
| `internal/version` (`version_test.go`) | `TestBuildTimeNonEmpty` — `BuildTime` is not empty; `TestStringContainsVersion` — `String()` includes `Version`; `TestStringContainsBuildTime` — `String()` includes `BuildTime`; update `TestStringReturnsVersion` → `TestStringFormat` to match new format |

No HTTP or live cluster access required — all tests are purely in-process.
Live verification (step d): run `hf version` from a binary built with `make build`.
