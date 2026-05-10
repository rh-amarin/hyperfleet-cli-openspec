## Context

`hf cluster create` and `hf nodepool create` build their API request bodies from hardcoded Go literals. Any customisation requires a CLI flag — and not every field is exposed as a flag. Users who run these commands repeatedly in different environments (different labels, spec fields, instance types) have no way to change the defaults persistently without editing Go code.

The config directory (`~/.config/hf/` or `$HF_CONFIG_DIR`) already holds per-environment settings and state. Two JSON template files in that directory give users a single, persistent, inspectable place to tune what `create` sends to the API.

## Goals / Non-Goals

**Goals:**
- Both create commands load their request body from a JSON template file in the config dir.
- If the template file does not exist it is auto-created from the built-in default and the user is told where.
- `-f <path>` lets the user supply any JSON file instead.
- A positional `<name>` argument overrides the `name` field in the template.
- Existing flags (`--name`, `--replicas`, `--type`, `--nodepool-id`) continue to work and override the corresponding template fields.
- Built-in defaults are embedded in the binary (`go:embed`) — no external file required at install time.

**Non-Goals:**
- Template variable interpolation.
- YAML template format.
- Templating for any command other than `cluster create` and `nodepool create`.

## Decisions

### D1 — Assets live in `cmd/assets/` and are embedded from there

`go:embed` requires paths to be within (or below) the directory containing the file with the directive. Placing the JSON files at `cmd/assets/cluster-template.json` and `cmd/assets/nodepool-template.json` allows `cmd/templates.go` to embed them directly without an extra package.

**Alternative considered:** top-level `assets/` with an `internal/templates/` shim — rejected as unnecessary indirection.

### D2 — `loadTemplate(configDir, resource, flagFile string) (map[string]any, bool, error)` in `cmd/templates.go`

A single package-level helper handles all three cases:
1. `flagFile != ""` → `os.ReadFile(flagFile)`, parse, return (created=false)
2. `flagFile == ""` → check `<configDir>/<resource>-template.json`
   - exists → read, parse, return (created=false)
   - missing → write embedded default, parse, return (created=true)

Returning a `bool` lets the caller emit an `[INFO]` message only when the file was just created, without duplicating the path logic.

**Alternative considered:** auto-create on every command invocation even with `-f` — rejected; `-f` is an explicit override and should not touch the config dir.

### D3 — Name override: positional arg > `--name` flag > template field

The template's `name` field is the lowest-priority source. The caller applies overrides in the same order as before:
1. Start with the parsed template as the body.
2. If `--name` flag is non-empty → `body["name"] = flagValue`.
3. If a positional arg is present → `body["name"] = args[0]` (positional takes precedence, consistent with existing cluster create behaviour).

### D4 — Cluster `[region] [version]` positional args are preserved

`hf cluster create` already documents `[name] [region] [version]`. These args continue to override the corresponding template spec fields so existing scripts are not broken.

### D5 — Nodepool create gains `[name]` positional arg

Currently `nodepoolCreateCmd.Use` is `"create"` with no positional args. It is updated to `"create [name]"` and `cobra.MaximumNArgs(1)` is added. The template-driven body replaces the hardcoded body.

### D6 — Other flags patch the body after template load

Before POSTing, the command applies flag overrides to the parsed template map:
- Cluster: `--replicas` → `body["spec"].(map[string]any)["replicas"]`; `--nodepool-id` → `body["nodepool_id"]`
- NodePool: `--type` → `body["spec"].(map[string]any)["platform"].(map[string]any)["type"]`; `--replicas` → `body["spec"].(map[string]any)["replicas"]`

When a flag is at its zero value (empty string / 0) it is ignored, preserving the template value.

## File Map

| File | Action |
|---|---|
| `cmd/assets/cluster-template.json` | New — default cluster payload |
| `cmd/assets/nodepool-template.json` | New — default nodepool payload |
| `cmd/templates.go` | New — `go:embed` + `loadTemplate()` |
| `cmd/cluster.go` | Modified — add `-f` flag, load template, apply overrides |
| `cmd/nodepool.go` | Modified — add `[name]` positional arg, `-f` flag, load template, apply overrides |
| `cmd/templates_test.go` | New — unit tests for `loadTemplate` |
| `cmd/cluster_test.go` | New — tests for template-driven cluster create |
| `cmd/nodepool_test.go` | New or extended — tests for template-driven nodepool create |

## Risks / Trade-offs

- **Behaviour change:** Cluster and nodepool create now read from a template file, so editing `cluster-template.json` in the config dir changes the payload for all future invocations. This is the desired behaviour but users upgrading from an older binary who have no template file will have one written silently on first use. The `[INFO]` message mitigates surprise.
- **`spec` field structure:** The template's `spec` and `labels` fields must be valid JSON objects; the CLI does not validate beyond JSON parsing. A malformed template causes a clear `[ERROR] loading template` message.
- **Duplicate check unchanged:** The existing name-based duplicate check runs before the POST, using whatever `name` ends up in the body. No change needed.

## Open Questions

_(none)_
