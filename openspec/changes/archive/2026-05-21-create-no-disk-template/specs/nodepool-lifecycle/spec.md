# NodePool Lifecycle — delta spec for create-no-disk-template

Only the scenarios below change. All other scenarios from `openspec/specs/nodepool-lifecycle/spec.md` remain in force.

## MODIFIED Requirements

### Requirement: Create NodePool

`hf nodepool create` SHALL load the request body from a JSON template. The binary embeds a built-in default template (`cmd/assets/nodepool-template.json`). When no `--file` flag is given, the CLI MUST use the embedded default bytes directly in memory — it MUST NOT read from or write to `<config-dir>`.

#### Scenario: Create nodepool with default template (no template file exists)

- GIVEN no `--file` flag is provided
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST use the built-in embedded template in memory
- AND MUST NOT write any file to `<config-dir>`
- AND MUST NOT print `[INFO] Created default nodepool template at …`
- AND MUST create the nodepool using the default payload (`kind=NodePool`, `name=my-nodepool`, default labels and spec)

#### Scenario: Create nodepool using existing config-dir template

- GIVEN `<config-dir>/nodepool-template.json` exists on disk
- WHEN the user runs `hf nodepool create` without `--file`
- THEN the CLI MUST use the embedded default in memory
- AND MUST NOT read the on-disk file
- AND MUST NOT write any file to `<config-dir>`

#### Scenario: Create nodepool with no arguments

- GIVEN no `--file` flag and no positional arguments
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST NOT show a usage message
- AND MUST proceed with creation using the embedded template payload
- AND MUST NOT write any file to `<config-dir>`
