# Delta Spec: version-build-time

## `internal/version` package

### Variables

| Variable | Default | Injected via ldflag |
|---|---|---|
| `Version` | `"dev"` | `-X …/version.Version=<tag>` |
| `BuildTime` | `"unknown"` | `-X …/version.BuildTime=<RFC3339>` |

### `String() string`

Returns `"<Version> (built <BuildTime>)"`.

Examples:
- `make build`: `v0.3.1 (built 2026-05-16T14:30:00Z)`
- `go run . version` (no ldflags): `dev (built unknown)`

## Build system

### Makefile

```
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -ldflags "-X $(MODULE)/internal/version.Version=$(VERSION) \
                         -X $(MODULE)/internal/version.BuildTime=$(BUILD_TIME)"
```

### `.goreleaser.yaml`

```yaml
ldflags:
  - >-
    -s -w
    -X github.com/rh-amarin/hyperfleet-cli/internal/version.Version={{.Version}}
    -X github.com/rh-amarin/hyperfleet-cli/internal/version.BuildTime={{.Date}}
```
