# Delta: Cluster Lifecycle — Template-Driven Create

Modifies the **Create Cluster** requirement in `openspec/specs/cluster-lifecycle/spec.md`.

---

## Requirement: Create Cluster (MODIFIED)

`hf cluster create` SHALL load the request body from a JSON template file rather than from hardcoded defaults.

### Template resolution

- WHEN `-f <path>` is supplied, the CLI MUST read the request body from that file.
- OTHERWISE the CLI MUST look for `<config-dir>/cluster-template.json`.
  - IF the file does not exist, the CLI MUST write the built-in default template to that path, print `[INFO] Created default cluster template at <path>`, and proceed.
  - IF the file exists, the CLI MUST read it.
- The CLI MUST parse the file as JSON. If parsing fails, it MUST exit with `[ERROR] loading template: <reason>`.

### Name override

- IF a positional `<name>` argument is provided, the CLI MUST set `body["name"]` to that value, overriding the template's `name` field.
- IF `--name` is provided (and no positional arg), the CLI MUST set `body["name"]` to the flag value.
- Positional arg takes precedence over `--name`.

### Positional `[region]` and `[version]` overrides (preserved)

- IF `args[1]` is provided, the CLI MUST set `body["spec"]["region"]` to that value.
- IF `args[2]` is provided, the CLI MUST set `body["spec"]["version"]` to that value.

### Flag overrides (unchanged behaviour, applied after template load)

- `--replicas <n>` (n > 0) → sets `body["spec"]["replicas"]`
- `--nodepool-id <id>` (non-empty) → sets `body["nodepool_id"]`

### New flag

| Flag | Short | Description |
|---|---|---|
| `--file <path>` | `-f` | Use this JSON file as the request body instead of the config-dir template |

### Built-in default template

The binary embeds `cmd/assets/cluster-template.json`:

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

### Scenarios

#### Scenario: Create cluster with default template (no template file exists)

- GIVEN no `cluster-template.json` exists in the config dir
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST write the built-in default to `<config-dir>/cluster-template.json`
- AND print `[INFO] Created default cluster template at <path>`
- AND create the cluster using the default payload

#### Scenario: Create cluster using existing config-dir template

- GIVEN `<config-dir>/cluster-template.json` exists with a custom payload
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST use the custom payload as the request body

#### Scenario: Create cluster with `-f` override

- GIVEN a file at `/tmp/my-cluster.json` containing a valid cluster payload
- WHEN the user runs `hf cluster create -f /tmp/my-cluster.json`
- THEN the CLI MUST use that file's content as the request body
- AND MUST NOT read or write the config-dir template

#### Scenario: Create cluster with name argument

- GIVEN a template with `"name": "my-cluster"`
- WHEN the user runs `hf cluster create prod-cluster`
- THEN the CLI MUST set `name` to `"prod-cluster"` in the request body

#### Scenario: Malformed template file

- GIVEN a template file containing invalid JSON
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST exit with `[ERROR] loading template: <reason>` and code 1
