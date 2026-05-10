# Delta: NodePool Lifecycle — Template-Driven Create

Modifies the **Create NodePool** requirement in `openspec/specs/nodepool-lifecycle/spec.md`.

---

## Requirement: Create NodePool (MODIFIED)

`hf nodepool create` SHALL load the request body from a JSON template file rather than from hardcoded defaults.

### Template resolution (same rules as cluster create)

- WHEN `-f <path>` is supplied, the CLI MUST read the request body from that file.
- OTHERWISE the CLI MUST look for `<config-dir>/nodepool-template.json`.
  - IF the file does not exist, the CLI MUST write the built-in default template to that path, print `[INFO] Created default nodepool template at <path>`, and proceed.
  - IF the file exists, the CLI MUST read it.
- The CLI MUST parse the file as JSON. If parsing fails, it MUST exit with `[ERROR] loading template: <reason>`.

### Name override

- IF a positional `<name>` argument is provided, the CLI MUST set `body["name"]` to that value, overriding the template's `name` field.
- IF `--name` is provided (and no positional arg), the CLI MUST set `body["name"]` to the flag value.
- Positional arg takes precedence over `--name`.

### Flag overrides (unchanged behaviour, applied after template load)

- `--type <type>` (non-empty) → sets `body["spec"]["platform"]["type"]`
- `--replicas <n>` (n > 0) → sets `body["spec"]["replicas"]`

### New flag

| Flag | Short | Description |
|---|---|---|
| `--file <path>` | `-f` | Use this JSON file as the request body instead of the config-dir template |

### New positional argument

`hf nodepool create` Use is updated from `"create"` to `"create [name]"`. The command accepts 0 or 1 positional args (`cobra.MaximumNArgs(1)`).

### Built-in default template

The binary embeds `cmd/assets/nodepool-template.json`:

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

### Scenarios

#### Scenario: Create nodepool with default template (no template file exists)

- GIVEN no `nodepool-template.json` exists in the config dir
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST write the built-in default to `<config-dir>/nodepool-template.json`
- AND print `[INFO] Created default nodepool template at <path>`
- AND create the nodepool using the default payload

#### Scenario: Create nodepool using existing config-dir template

- GIVEN `<config-dir>/nodepool-template.json` exists with a custom payload
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST use the custom payload as the request body

#### Scenario: Create nodepool with `-f` override

- GIVEN a file at `/tmp/my-nodepool.json` containing a valid nodepool payload
- WHEN the user runs `hf nodepool create -f /tmp/my-nodepool.json`
- THEN the CLI MUST use that file's content as the request body
- AND MUST NOT read or write the config-dir template

#### Scenario: Create nodepool with name argument

- GIVEN a template with `"name": "my-nodepool"`
- WHEN the user runs `hf nodepool create workers`
- THEN the CLI MUST set `name` to `"workers"` in the request body

#### Scenario: Create nodepool with no arguments (defaults)

- GIVEN the config-dir template exists (or is auto-created)
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST NOT show a usage message
- AND MUST proceed with creation using the template payload

#### Scenario: Malformed template file

- GIVEN a template file containing invalid JSON
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST exit with `[ERROR] loading template: <reason>` and code 1
