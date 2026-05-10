# Proposal: Template-Driven Cluster and NodePool Create

## Why

`hf cluster create` and `hf nodepool create` use hardcoded payloads. Users who need to customise the request body — labels, spec fields, instance type, region, version — must pass individual CLI flags or live with the built-in defaults. Adding a flag per field is not scalable and produces a sprawling API surface.

## What Changes

Replace the hardcoded payload construction in both create commands with a **JSON template** loaded from the config directory. The template is the authoritative source for the full API request body; the command only overrides the `name` field if a name is provided as a positional argument.

### Behaviour

**Template resolution order:**

1. Path given via `-f <file>` flag (user-supplied, any location)
2. `<config-dir>/cluster-template.json` / `<config-dir>/nodepool-template.json` (per-environment default)
3. Built-in default (auto-written to the config dir on first use)

**Auto-creation:** If the per-environment template file does not exist, the CLI writes the built-in default JSON to that path before using it. The user is informed via an `[INFO]` message pointing to the created file.

**Name override:** If a positional argument is supplied (`hf cluster create <name>` or `hf nodepool create <name>`), that value replaces the `name` field in the loaded template. All other template fields are used as-is.

**Existing flags retained:** `--name` (cluster and nodepool), `--nodepool-id` (cluster), and `--replicas`, `--type` (nodepool) continue to work and override the corresponding template fields, consistent with the current precedence model.

### New flag

| Command | Flag | Description |
|---|---|---|
| `hf cluster create` | `-f, --file <path>` | Load request body from this JSON file instead of the default template |
| `hf nodepool create` | `-f, --file <path>` | Load request body from this JSON file instead of the default template |

### Static default templates (embedded in binary)

Stored in `assets/` and embedded via `go:embed`. These are the seed files written to the config dir on first use.

**`assets/cluster-template.json`** — mirrors current hardcoded defaults:
```json
{
  "kind": "Cluster",
  "name": "my-cluster",
  "labels": {
    "counter": "1",
    "environment": "development",
    "shard": "1",
    "team": "core"
  },
  "spec": {
    "counter": "1",
    "region": "us-east-1",
    "version": "4.15.0"
  }
}
```

**`assets/nodepool-template.json`** — mirrors current hardcoded defaults:
```json
{
  "kind": "NodePool",
  "name": "my-nodepool",
  "labels": {
    "counter": "1"
  },
  "spec": {
    "counter": "1",
    "platform": { "type": "m4" },
    "replicas": 1
  }
}
```

## Capabilities Added

- Users can edit `cluster-template.json` / `nodepool-template.json` in their config dir to permanently customise defaults without flags
- Per-invocation override via `-f` enables scripting and CI pipelines with custom payloads
- First-run auto-creation removes the need to manually locate or copy template files
- Binary embeds the defaults — no external file distribution needed

## Out of Scope

- Template variables / interpolation (values are used as-is from the JSON)
- YAML template support (JSON only)
- Templating for commands other than `cluster create` and `nodepool create`

## Impact

| Area | Change |
|---|---|
| `cmd/cluster.go` | Load template, apply name override, add `-f` flag |
| `cmd/nodepool.go` | Load template, apply name override, add `-f` flag |
| `assets/` (new) | Two static JSON template files, embedded via `go:embed` |
| `internal/config/` | `TemplateFor(name string)` helper: resolve path, auto-create, return `[]byte` |
| `openspec/specs/cluster-lifecycle/spec.md` | Delta: template-driven create |
| `openspec/specs/nodepool-lifecycle/spec.md` | Delta: template-driven create |

## Testing Scope

- Unit tests for `config.TemplateFor`: auto-creation, `-f` path, missing file error
- Updated `TestClusterCreate` / `TestNodePoolCreate` in `cmd/`: template loaded, name override applied, `-f` respected
- `go test ./...`, `go build ./...`, `go vet ./...` all pass
