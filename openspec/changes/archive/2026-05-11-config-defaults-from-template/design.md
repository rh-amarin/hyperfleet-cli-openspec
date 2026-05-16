## Context

`internal/config/config.go` has a `var defaults map[string]map[string]string` (lines 19–61) whose values are identical to `cmd/assets/config-template.yaml`. The `cmd` package embeds the template file (`cmd/assets.go`) and a test (`cmd/assets_test.go:TestConfigTemplateMatchesDefaults`) asserts they never drift. The embed in `cmd` is also used at line 249 of `cmd/config.go` to seed new environment files.

Go's `//go:embed` only allows embedding files within the package's own directory tree, so `internal/config` cannot embed a file from `cmd/assets/`.

## Goals / Non-Goals

**Goals:**
- One copy of the default values: `internal/config/assets/config-template.yaml`
- No hardcoded default values in Go source code
- `cmd` still gets the template bytes (for env-file seeding) without owning the embed
- Drift between defaults and template is structurally impossible

**Non-Goals:**
- Changing any default values
- Changing the precedence chain (env vars > profile > defaults)
- Modifying any CLI user-visible behaviour

## Decisions

### Move the YAML file into `internal/config/assets/`

The embed directive must point to a file within the package directory. Moving the file into `internal/config/assets/config-template.yaml` lets the config package own both the template and the defaults it produces.

**Alternative considered:** keep the file in `cmd/assets/` and inject the bytes via a `Store` constructor parameter. Rejected because it pushes the responsibility for supplying defaults into every call site (including `NewFromEnv()`), making the API awkward.

### Export `ConfigTemplateYAML []byte` from `internal/config`

`cmd/config.go` uses the template bytes directly to write new environment files (`os.WriteFile(profPath, configTemplateYAML, 0600)`). Exporting the variable from `internal/config` lets `cmd` continue doing this without its own embed.

### Parse the template at `init()` time

Loading the YAML once at package init keeps `Get()` free of error handling. If the embedded YAML is malformed the binary panics at startup, which is caught by `go test ./...` before release.

**Alternative considered:** parse lazily on first `Get()` call and return an error. Rejected — `Get()` currently returns a plain string; adding an error return would break the entire call graph.

### Remove `TestConfigTemplateMatchesDefaults`

The test existed only to detect drift between two copies of the same data. After this change there is one copy, so the test has no value and should be deleted.

## Risks / Trade-offs

- [Risk] If the embedded YAML file is ever malformed (e.g., bad merge conflict markers), the binary panics at startup. → Mitigation: `go test ./...` runs `init()` and will catch the panic before the binary ships.

## Migration Plan

No user-facing migration required. The change is internal. Environment files already on disk are unaffected — they continue to be read as-is.

## Open Questions

_(none)_
