# Design: version-build-time

## `internal/version/version.go`

```go
var Version   = "dev"
var BuildTime = "unknown"

func String() string {
    return Version + " (built " + BuildTime + ")"
}
```

`BuildTime` defaults to `"unknown"` so plain `go build` / `go run` still works without
producing an empty or misleading output.

## `Makefile`

```makefile
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -ldflags "-X $(MODULE)/internal/version.Version=$(VERSION) \
                         -X $(MODULE)/internal/version.BuildTime=$(BUILD_TIME)"
```

## `.goreleaser.yaml`

```yaml
ldflags:
  - >-
    -s -w
    -X github.com/rh-amarin/hyperfleet-cli/internal/version.Version={{.Version}}
    -X github.com/rh-amarin/hyperfleet-cli/internal/version.BuildTime={{.Date}}
```

goreleaser's `{{.Date}}` resolves to the RFC 3339 timestamp of the release build.

## `internal/version/version_test.go`

Replace `TestStringReturnsVersion` (which checked `String() == Version`, now false) with:

- `TestStringContainsVersion` — `strings.Contains(String(), Version)`
- `TestStringContainsBuildTime` — `strings.Contains(String(), BuildTime)`
- `TestBuildTimeNonEmpty` — `BuildTime != ""`

## Impact

| File | Change |
|---|---|
| `internal/version/version.go` | Add `BuildTime`; update `String()` |
| `internal/version/version_test.go` | Replace `TestStringReturnsVersion`; add 3 tests |
| `Makefile` | Add `BUILD_TIME`; extend `LDFLAGS` |
| `.goreleaser.yaml` | Extend ldflags with `BuildTime` |
